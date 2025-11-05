package wellknown

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// RequireAuth returns authentication middleware that validates JWT and populates c.Auth
// PocketBase automatically populates c.Auth from Authorization header or pb_auth cookie
func RequireAuth() func(*core.RequestEvent) error {
	return func(c *core.RequestEvent) error {
		if c.Auth == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authentication required",
			})
		}
		return c.Next()
	}
}

// GetAuthRecord extracts the authenticated record from the request context
// Returns the record and true if authenticated, nil and false otherwise
// Use this helper to simplify auth checks in handlers
func GetAuthRecord(c *core.RequestEvent) (*core.Record, bool) {
	authRecord := c.Auth
	if authRecord == nil {
		return nil, false
	}
	return authRecord, true
}

// MustGetAuthRecord extracts the auth record or returns an error response
// Returns the record and nil error if authenticated, or nil record and JSON error
func MustGetAuthRecord(c *core.RequestEvent) (*core.Record, error) {
	if c.Auth == nil {
		return nil, c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authentication required",
		})
	}
	return c.Auth, nil
}
