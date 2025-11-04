package wellknown

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joeblew999/wellknown/pb/codegen/models"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
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

	// Server-based OAuth routes (no JS SDK)
	e.Router.GET("/auth/google", handleGoogleLogin)
	e.Router.GET("/auth/google/callback", handleGoogleCallback(wk))
	e.Router.GET("/auth/logout", handleLogout(wk))
	e.Router.GET("/auth/status", handleAuthStatus(wk))

	log.Println("✅ Google OAuth routes registered (server-based)")
	return e.Next()
}

// handleGoogleLogin initiates the OAuth flow
func handleGoogleLogin(e *core.RequestEvent) error {
	// Generate state token for CSRF protection
	state := generateStateToken()

	// Store state in session cookie
	http.SetCookie(e.Response, &http.Cookie{
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
	return e.Redirect(http.StatusTemporaryRedirect, url)
}

// handleGoogleCallback processes the OAuth callback
func handleGoogleCallback(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Verify state token
		stateCookie, err := e.Request.Cookie("oauth_state")
		if err != nil {
			return e.String(http.StatusBadRequest, "Missing state cookie")
		}

		state := e.Request.URL.Query().Get("state")
		if state != stateCookie.Value {
			return e.String(http.StatusBadRequest, "Invalid state token")
		}

		// Exchange code for token
		code := e.Request.URL.Query().Get("code")
		token, err := googleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Failed to exchange code: %v", err)
			return e.String(http.StatusInternalServerError, "Failed to exchange code")
		}

		// Get user info
		client := googleOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Printf("Failed to get user info: %v", err)
			return e.String(http.StatusInternalServerError, "Failed to get user info")
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email string `json:"email"`
			ID    string `json:"id"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			log.Printf("Failed to decode user info: %v", err)
			return e.String(http.StatusInternalServerError, "Failed to decode user info")
		}

		// Find or create user in Pocketbase
		userCollection, err := wk.FindCollectionByNameOrId("users")
		if err != nil {
			return e.String(http.StatusInternalServerError, "Failed to find users collection")
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
				return e.String(http.StatusInternalServerError, "Failed to create user")
			}
			log.Printf("Created new user: %s", userInfo.Email)
		}

		// Store Google OAuth token
		if err := storeGoogleToken(wk, user.Id, token); err != nil {
			log.Printf("Failed to store token: %v", err)
			return e.String(http.StatusInternalServerError, "Failed to store token")
		}

		// Generate JWT token for the user
		authToken, err := user.NewAuthToken()
		if err != nil {
			log.Printf("Failed to generate auth token: %v", err)
			return e.String(http.StatusInternalServerError, "Failed to generate auth token")
		}

		// Set auth cookie
		http.SetCookie(e.Response, &http.Cookie{
			Name:     "pb_auth",
			Value:    authToken,
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to home or dashboard
		return e.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

// handleLogout logs out the user
func handleLogout(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Clear auth cookie
		http.SetCookie(e.Response, &http.Cookie{
			Name:     "pb_auth",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		return e.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

// handleAuthStatus returns current auth status
func handleAuthStatus(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authCookie, err := e.Request.Cookie("pb_auth")
		if err != nil || authCookie.Value == "" {
			return e.JSON(http.StatusOK, map[string]interface{}{
				"authenticated": false,
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"authenticated": true,
		})
	}
}

// storeGoogleToken stores the OAuth token in the database using type-safe proxy
func storeGoogleToken(wk *Wellknown, userID string, token *oauth2.Token) error {
	collection, err := wk.FindCollectionByNameOrId("google_tokens")
	if err != nil {
		return fmt.Errorf("failed to find google_tokens collection: %w", err)
	}

	// Check if token already exists for this user
	existingToken, err := wk.FindFirstRecordByFilter(collection.Name, "user_id = {:user_id}", map[string]any{
		"user_id": userID,
	})

	var tokenProxy *models.GoogleTokens
	if err != nil {
		// Token doesn't exist, create new proxy
		tokenProxy, err = models.NewProxy[models.GoogleTokens](wk)
		if err != nil {
			return fmt.Errorf("failed to create GoogleTokens proxy: %w", err)
		}
	} else {
		// Token exists, wrap it in proxy
		tokenProxy, err = models.WrapRecord[models.GoogleTokens](existingToken)
		if err != nil {
			return fmt.Errorf("failed to wrap record as GoogleTokens: %w", err)
		}
	}

	// Set token fields using type-safe setters
	tokenProxy.SetUserId(userID)
	tokenProxy.SetAccessToken(token.AccessToken)
	tokenProxy.SetRefreshToken(token.RefreshToken)
	tokenProxy.SetTokenType(token.TokenType)
	tokenProxy.SetExpiry(types.NowDateTime().Add(time.Until(token.Expiry)))

	return wk.Save(tokenProxy.ProxyRecord())
}

// generateStateToken generates a random state token for CSRF protection
func generateStateToken() string {
	// Simple implementation - in production use crypto/rand
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
