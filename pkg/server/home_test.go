package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHomepage ensures the homepage renders correctly with all services
func TestHomepage(t *testing.T) {
	// Setup test server
	mux, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	// Request homepage
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Verify 200 OK
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Debug: Print first 500 chars
	t.Logf("Response body (first 500 chars): %s", body[:min(500, len(body))])

	// Verify homepage title
	if !strings.Contains(body, "wellknown") {
		t.Error("Homepage should contain 'wellknown' title")
	}

	// Verify hero text
	if !strings.Contains(body, "Universal Go library") {
		t.Error("Homepage should contain hero description")
	}

	// Verify all services are shown
	expectedServices := []string{
		"Google Calendar",
		"Apple Calendar",
		"Google Maps",
		"Apple Maps",
	}

	for _, service := range expectedServices {
		if !strings.Contains(body, service) {
			t.Errorf("Homepage should contain service: %s", service)
		}
	}

	// Verify service cards exist (count actual div elements, not CSS class definitions)
	cardCount := strings.Count(body, `<div class="service-card">`)
	if cardCount < 4 {
		t.Errorf("Expected at least 4 service cards, found %d", cardCount)
	}

	// Verify Custom and Showcase links exist
	if !strings.Contains(body, "/google/calendar") {
		t.Error("Homepage should link to /google/calendar")
	}
	if !strings.Contains(body, "/google/calendar/showcase") {
		t.Error("Homepage should link to /google/calendar/showcase")
	}

	t.Log("✅ Homepage rendered successfully with all services")
}

// TestHomepageNotFound ensures non-root paths return 404
func TestHomepageNotFound(t *testing.T) {
	mux, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	// Request non-existent path
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Verify 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for /nonexistent, got %d", rec.Code)
	}

	t.Log("✅ 404 handling works correctly")
}
