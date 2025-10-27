package wellknown

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

// registerOAuthRoutes sets up server-based Google OAuth routes
func registerOAuthRoutes(wk *Wellknown, e *core.ServeEvent) error {
	// Initialize Google OAuth config from environment
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}

	// Server-based OAuth routes (no JS SDK) - using standard http mux
	e.Router.HandleFunc("GET /auth/google", handleGoogleLogin)
	e.Router.HandleFunc("GET /auth/google/callback", handleGoogleCallback(wk))
	e.Router.HandleFunc("GET /auth/logout", handleLogout(wk))
	e.Router.HandleFunc("GET /auth/status", handleAuthStatus(wk))

	log.Println("✅ Google OAuth routes registered (server-based)")
	return nil
}

// handleGoogleLogin initiates the OAuth flow
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state := generateStateToken()

	// Store state in session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to Google OAuth consent page
	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handleGoogleCallback processes the OAuth callback
func handleGoogleCallback(wk *Wellknown) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify state token
		stateCookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, "Missing state cookie", http.StatusBadRequest)
			return
		}

		state := r.URL.Query().Get("state")
		if state != stateCookie.Value {
			http.Error(w, "Invalid state token", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		code := r.URL.Query().Get("code")
		token, err := googleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Failed to exchange code: %v", err)
			http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
			return
		}

		// Get user info
		client := googleOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Printf("Failed to get user info: %v", err)
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email string `json:"email"`
			ID    string `json:"id"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			log.Printf("Failed to decode user info: %v", err)
			http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
			return
		}

		// Find or create user in Pocketbase
		userCollection, err := wk.FindCollectionByNameOrId("users")
		if err != nil {
			http.Error(w, "Failed to find users collection", http.StatusInternalServerError)
			return
		}

		// Try to find existing user by email
		user, err := wk.FindFirstRecordByFilter(userCollection.Name, "email = {:email}", map[string]any{
			"email": userInfo.Email,
		})

		if err != nil {
			// User doesn't exist, create new one
			user = core.NewRecord(userCollection)
			user.Set("email", userInfo.Email)
			user.Set("name", userInfo.Name)
			user.Set("username", userInfo.Email) // Use email as username
			user.Set("verified", true)           // Auto-verify Google users
			// Set password (required for auth collection)
			user.SetPassword(generateStateToken()) // Random password since we use OAuth

			if err := wk.Save(user); err != nil {
				log.Printf("Failed to create user: %v", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
			log.Printf("Created new user: %s", userInfo.Email)
		}

		// Store Google OAuth token
		if err := storeGoogleToken(wk, user.Id, token, userInfo.Email); err != nil {
			log.Printf("Failed to store token: %v", err)
			http.Error(w, "Failed to store token", http.StatusInternalServerError)
			return
		}

		// Generate JWT token for the user
		authToken, err := wk.NewAuthToken(user.Collection().Name, user.Id)
		if err != nil {
			log.Printf("Failed to generate auth token: %v", err)
			http.Error(w, "Failed to generate auth token", http.StatusInternalServerError)
			return
		}

		// Set auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "pb_auth",
			Value:    authToken,
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to home or dashboard
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// handleLogout logs out the user
func handleLogout(wk *Wellknown) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "pb_auth",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// handleAuthStatus returns current auth status
func handleAuthStatus(wk *Wellknown) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("pb_auth")
		if err != nil || authCookie.Value == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"authenticated": false,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": true,
		})
	}
}

// storeGoogleToken stores the OAuth token in the database
func storeGoogleToken(wk *Wellknown, userID string, token *oauth2.Token, email string) error {
	collection, err := wk.FindCollectionByNameOrId("google_tokens")
	if err != nil {
		return fmt.Errorf("failed to find google_tokens collection: %w", err)
	}

	// Check if token already exists for this user
	existingToken, err := wk.FindFirstRecordByFilter(collection.Name, "user_id = {:user_id}", map[string]any{
		"user_id": userID,
	})

	var record *core.Record
	if err != nil {
		// Token doesn't exist, create new record
		record = core.NewRecord(collection)
	} else {
		// Token exists, update it
		record = existingToken
	}

	// Set token fields
	record.Set("user_id", userID)
	record.Set("access_token", token.AccessToken)
	record.Set("refresh_token", token.RefreshToken)
	record.Set("token_type", token.TokenType)
	record.Set("expiry", token.Expiry.Format(time.RFC3339))
	record.Set("email", email)

	return wk.Save(record)
}

// generateStateToken generates a random state token for CSRF protection
func generateStateToken() string {
	// Simple implementation - in production use crypto/rand
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
