package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// setupTestServer initializes templates and registers routes for testing
func setupTestServer() (*http.ServeMux, error) {
	// Initialize templates if not already done
	if Templates == nil {
		_, err := initTemplates()
		if err != nil {
			return nil, err
		}
	}

	// Set URLs for handlers
	LocalURL = "http://localhost:8080"
	MobileURL = "http://localhost:8080"

	// Create a fresh mux for testing
	mux := http.NewServeMux()

	// Clear registered services to avoid duplication
	ClearRegisteredServices()

	// Register routes on our test mux
	RegisterRoutes(mux)

	return mux, nil
}

// TestNavigationLinksAreValid ensures all links in navigation point to registered routes.
// This prevents dead links after route deletions (see PREVENTION.md Issue 1).
func TestNavigationLinksAreValid(t *testing.T) {
	// Create test server with templates initialized
	mux, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	// Get home page to extract navigation links
	req := httptest.NewRequest("GET", "/google/calendar", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Extract all href links from navigation
	// Pattern: href="/path/to/route"
	linkRegex := regexp.MustCompile(`href="(/[^"#]*)"`)
	matches := linkRegex.FindAllStringSubmatch(body, -1)

	if len(matches) == 0 {
		t.Fatal("No navigation links found in page")
	}

	// Track tested links to avoid duplicates
	tested := make(map[string]bool)
	deadLinks := []string{}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		link := match[1]

		// Skip external links, anchors, and already tested
		if strings.HasPrefix(link, "http") || strings.HasPrefix(link, "#") || tested[link] {
			continue
		}

		tested[link] = true

		// Test if link resolves to a valid route
		testReq := httptest.NewRequest("GET", link, nil)
		testRec := httptest.NewRecorder()
		mux.ServeHTTP(testRec, testReq)

		// Accept 200 OK or 3xx redirects
		// Reject 404 Not Found
		if testRec.Code == http.StatusNotFound {
			deadLinks = append(deadLinks, link)
			t.Errorf("Dead link found: %s returned 404", link)
		}
	}

	if len(deadLinks) > 0 {
		t.Errorf("Found %d dead navigation links: %v", len(deadLinks), deadLinks)
		t.Error("Fix: Update navigation in templates/base.html or register missing routes in server.go")
	}

	t.Logf("✅ Validated %d unique navigation links - all working", len(tested))
}

// TestAllRegisteredRoutesWork ensures all registered routes return valid responses.
// This is the inverse test - ensures registered routes actually work.
func TestAllRegisteredRoutesWork(t *testing.T) {
	// Create test server with templates initialized
	mux, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	// List of all routes that should be registered
	routes := []string{
		"/google/calendar",
		"/google/calendar/showcase",
		"/apple/calendar",
		"/apple/calendar/showcase",
		"/apple/calendar/download",
		"/google/maps",
		"/google/maps/showcase",
		"/apple/maps",
		"/apple/maps/showcase",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			// Download route requires query parameter, so 400 is acceptable
			if route == "/apple/calendar/download" && rec.Code == http.StatusBadRequest {
				t.Logf("✅ %s returns 400 (expected - requires event parameter)", route)
				return
			}

			// All other routes should return 200 OK
			if rec.Code != http.StatusOK {
				t.Errorf("Route %s returned status %d, expected 200", route, rec.Code)
			} else {
				t.Logf("✅ %s returns 200 OK", route)
			}
		})
	}
}
