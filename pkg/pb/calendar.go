package wellknown

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joeblew999/wellknown/pkg/pb/codegen/models"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// registerCalendarRoutes sets up Google Calendar API routes
// NOTE: Calendar URL/ICS generation is handled by the main server (pkg/server), not here!
// PocketBase only handles OAuth + Calendar API calls (list/create events using stored tokens)
func registerCalendarRoutes(wk *Wellknown, e *core.ServeEvent) error {
	// Protected routes - require authentication
	e.Router.GET("/api/calendar/events", handleListEvents(wk)).BindFunc(requireAuthMiddleware(wk))
	e.Router.POST("/api/calendar/events", handleCreateEvent(wk)).BindFunc(requireAuthMiddleware(wk))

	log.Println("✅ Calendar API routes registered (OAuth + Calendar API only)")
	return e.Next()
}

// requireAuthMiddleware checks for valid auth token
func requireAuthMiddleware(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authCookie, err := e.Request.Cookie("pb_auth")
		if err != nil || authCookie.Value == "" {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authentication required",
			})
		}

		// Store user ID in context (simplified - should verify token properly)
		// For now, pass through
		return e.Next()
	}
}

// handleListEvents lists user's calendar events
func handleListEvents(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get user from auth cookie (simplified - should extract from JWT)
		authCookie, _ := e.Request.Cookie("pb_auth")
		userID := authCookie.Value // TODO: Extract real user ID from JWT

		// Get Google token for user
		token, err := getGoogleToken(wk, userID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to get Google token",
			})
		}

		// Create Calendar API client
		client := googleOAuthConfig.Client(context.Background(), token)
		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create Calendar service",
			})
		}

		// List upcoming events
		t := time.Now().Format(time.RFC3339)
		events, err := srv.Events.List("primary").
			TimeMin(t).
			MaxResults(10).
			SingleEvents(true).
			OrderBy("startTime").
			Do()

		if err != nil {
			log.Printf("Failed to list events: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to list events",
			})
		}

		return e.JSON(http.StatusOK, events.Items)
	}
}

// handleCreateEvent creates a new calendar event
func handleCreateEvent(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get user from auth cookie (simplified)
		authCookie, _ := e.Request.Cookie("pb_auth")
		userID := authCookie.Value // TODO: Extract real user ID from JWT

		// Parse request body
		var eventData struct {
			Summary     string    `json:"summary"`
			Description string    `json:"description"`
			Location    string    `json:"location"`
			StartTime   time.Time `json:"start_time"`
			EndTime     time.Time `json:"end_time"`
		}

		if err := json.NewDecoder(e.Request.Body).Decode(&eventData); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
		}

		// Get Google token for user
		token, err := getGoogleToken(wk, userID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to get Google token",
			})
		}

		// Create Calendar API client
		client := googleOAuthConfig.Client(context.Background(), token)
		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create Calendar service",
			})
		}

		// Create event
		event := &calendar.Event{
			Summary:     eventData.Summary,
			Description: eventData.Description,
			Location:    eventData.Location,
			Start: &calendar.EventDateTime{
				DateTime: eventData.StartTime.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: eventData.EndTime.Format(time.RFC3339),
			},
		}

		createdEvent, err := srv.Events.Insert("primary", event).Do()
		if err != nil {
			log.Printf("Failed to create event: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create event",
			})
		}

		return e.JSON(http.StatusCreated, createdEvent)
	}
}

// getGoogleToken retrieves stored OAuth token for user using type-safe proxy
func getGoogleToken(wk *Wellknown, userID string) (*oauth2.Token, error) {
	collection, err := wk.FindCollectionByNameOrId("google_tokens")
	if err != nil {
		return nil, fmt.Errorf("failed to find google_tokens collection: %w", err)
	}

	record, err := wk.FindFirstRecordByFilter(collection.Name, "user_id = {:user_id}", map[string]any{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find token for user: %w", err)
	}

	// Wrap record in type-safe proxy
	tokenProxy, err := models.WrapRecord[models.GoogleTokens](record)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap record as GoogleTokens: %w", err)
	}

	// Parse expiry using type-safe getter
	expiry := tokenProxy.Expiry().Time()

	token := &oauth2.Token{
		AccessToken:  tokenProxy.AccessToken(),
		RefreshToken: tokenProxy.RefreshToken(),
		TokenType:    tokenProxy.TokenType(),
		Expiry:       expiry,
	}

	// Check if token needs refresh
	if time.Now().After(token.Expiry) {
		// Token expired, refresh it
		newToken, err := googleOAuthConfig.TokenSource(context.Background(), token).Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Update stored token
		if err := storeGoogleToken(wk, userID, newToken); err != nil {
			log.Printf("Warning: failed to update refreshed token: %v", err)
		}

		return newToken, nil
	}

	return token, nil
}
