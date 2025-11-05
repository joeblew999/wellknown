package ai

import (
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// GoogleTokenGetter creates a function that retrieves Google OAuth tokens from PocketBase
// This is for accessing Google APIs (Calendar, etc.) NOT for Anthropic AI.
// Returns a function that can be called to get the current Google access token.
func GoogleTokenGetter(app core.App, userID string) func() (string, error) {
	return func() (string, error) {
		// Query google_tokens collection for the user's token
		record, err := app.FindFirstRecordByFilter(
			"google_tokens",
			"user_id = {:user_id}",
			dbx.Params{
				"user_id": userID,
			},
		)
		if err != nil {
			return "", fmt.Errorf("Google token not found for user %s: %w", userID, err)
		}

		// Get token expiry
		expiry := record.GetDateTime("expiry")
		if expiry.IsZero() {
			return "", fmt.Errorf("token expiry not set")
		}

		// Check if token is expired (with 1-minute buffer)
		if time.Now().After(expiry.Time().Add(-1 * time.Minute)) {
			// TODO: Implement token refresh
			// For now, return error - refresh should be handled by OAuth flow
			return "", fmt.Errorf("token expired at %v", expiry.Time())
		}

		// Return access token
		accessToken := record.GetString("access_token")
		if accessToken == "" {
			return "", fmt.Errorf("access token is empty")
		}

		return accessToken, nil
	}
}

// NOTE: Anthropic OAuth token storage is NOT implemented yet.
// For Anthropic AI, use client.NewClientWithAPIKey(cfg.AI.Anthropic.APIKey) instead.
// When Anthropic OAuth is needed in the future, create a separate 'anthropic_tokens'
// collection and implement AnthropicTokenGetter similar to GoogleTokenGetter above.
