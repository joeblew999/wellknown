package wellknown

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// registerCalendarRoutes sets up Google Calendar API routes
func registerCalendarRoutes(wk *Wellknown, e *core.ServeEvent) error {
	// Protected routes - require authentication
	e.Router.HandleFunc("GET /api/calendar/events", requireAuth(wk, handleListEvents(wk)))
	e.Router.HandleFunc("POST /api/calendar/events", requireAuth(wk, handleCreateEvent(wk)))

	log.Println("✅ Calendar API routes registered")
	return nil
}

// requireAuth middleware checks for valid auth token
func requireAuth(wk *Wellknown, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("pb_auth")
		if err != nil || authCookie.Value == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Store user ID in context (simplified - should verify token properly)
		// For now, pass through
		next(w, r)
	}
}

// handleListEvents lists user's calendar events
func handleListEvents(wk *Wellknown) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from auth cookie (simplified - should extract from JWT)
		authCookie, _ := r.Cookie("pb_auth")
		userID := authCookie.Value // TODO: Extract real user ID from JWT

		// Get Google token for user
		token, err := getGoogleToken(wk, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to get Google token",
			})
			return
		}

		// Create Calendar API client
		client := googleOAuthConfig.Client(context.Background(), token)
		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create Calendar service",
			})
			return
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to list events",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events.Items)
	}
}

// handleCreateEvent creates a new calendar event
func handleCreateEvent(wk *Wellknown) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from auth cookie (simplified)
		authCookie, _ := r.Cookie("pb_auth")
		userID := authCookie.Value // TODO: Extract real user ID from JWT

		// Parse request body
		var eventData struct {
			Summary     string    `json:"summary"`
			Description string    `json:"description"`
			Location    string    `json:"location"`
			StartTime   time.Time `json:"start_time"`
			EndTime     time.Time `json:"end_time"`
		}

		if err := json.NewDecoder(r.Body).Decode(&eventData); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Get Google token for user
		token, err := getGoogleToken(wk, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to get Google token",
			})
			return
		}

		// Create Calendar API client
		client := googleOAuthConfig.Client(context.Background(), token)
		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create Calendar service",
			})
			return
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create event",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdEvent)
	}
}

// getGoogleToken retrieves stored OAuth token for user
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

	// Parse expiry
	expiryStr := record.GetString("expiry")
	expiry, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		expiry = time.Now().Add(time.Hour) // Default to 1 hour if parse fails
	}

	token := &oauth2.Token{
		AccessToken:  record.GetString("access_token"),
		RefreshToken: record.GetString("refresh_token"),
		TokenType:    record.GetString("token_type"),
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
		if err := storeGoogleToken(wk, userID, newToken, record.GetString("email")); err != nil {
			log.Printf("Warning: failed to update refreshed token: %v", err)
		}

		return newToken, nil
	}

	return token, nil
}
