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
// RegisterCalendarRoutes handles Calendar API calls (list/create events using stored tokens)
func RegisterCalendarRoutes(wk *Wellknown, e *core.ServeEvent, registry *RouteRegistry) {
	// Pre-flight check: Validate required collections exist
	if _, err := wk.FindCollectionByNameOrId("google_tokens"); err != nil {
		log.Printf("⚠️  Calendar routes NOT registered: collection 'google_tokens' not found (migrations may not have run)")
		log.Printf("   Run 'go run . migrate up' to create required collections")
		return // Skip calendar routes registration
	}

	log.Println("✅ Calendar routes: Pre-flight checks passed")

	// Create route handler for Calendar domain
	handler := NewRouteHandler(registry, "Calendar", e)

	// Protected routes - require authentication
	handler.GET("/api/calendar/events", handleListEvents(wk),
		WithAuth(), WithDescription("List calendar events"))
	handler.POST("/api/calendar/events", handleCreateEvent(wk),
		WithAuth(), WithDescription("Create calendar event"))

	log.Println("✅ Calendar API routes registered (OAuth + Calendar API only)")
}

// handleListEvents lists user's calendar events
func handleListEvents(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Auth middleware ensures e.Auth is populated
		userID := e.Auth.Id

		// Get Google token for user
		token, err := getGoogleToken(wk, userID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to get Google token",
			})
		}

		// Create Calendar API client
		client := wk.oauthService.GoogleConfig.Client(context.Background(), token)
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
		// Auth middleware ensures e.Auth is populated
		userID := e.Auth.Id

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
		client := wk.oauthService.GoogleConfig.Client(context.Background(), token)
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
		newToken, err := wk.oauthService.GoogleConfig.TokenSource(context.Background(), token).Token()
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
