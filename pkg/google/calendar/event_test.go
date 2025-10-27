package calendar

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

// compareURLs compares two URLs by parsing them and comparing components
// This is more robust than string comparison as it handles parameter ordering
func compareURLs(t *testing.T, got, want string) {
	t.Helper()

	// Parse both URLs
	gotURL, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Failed to parse generated URL: %v", err)
	}

	wantURL, err := url.Parse(want)
	if err != nil {
		t.Fatalf("Failed to parse expected URL: %v", err)
	}

	// Compare scheme, host, and path
	if gotURL.Scheme != wantURL.Scheme {
		t.Errorf("URL scheme mismatch: got %q, want %q", gotURL.Scheme, wantURL.Scheme)
	}
	if gotURL.Host != wantURL.Host {
		t.Errorf("URL host mismatch: got %q, want %q", gotURL.Host, wantURL.Host)
	}
	if gotURL.Path != wantURL.Path {
		t.Errorf("URL path mismatch: got %q, want %q", gotURL.Path, wantURL.Path)
	}

	// Compare query parameters (order-independent)
	gotParams := gotURL.Query()
	wantParams := wantURL.Query()

	for key, wantVals := range wantParams {
		gotVals, exists := gotParams[key]
		if !exists {
			t.Errorf("Missing query parameter %q", key)
			continue
		}
		if len(gotVals) != len(wantVals) {
			t.Errorf("Query parameter %q: got %d values, want %d values", key, len(gotVals), len(wantVals))
			continue
		}
		for i, wantVal := range wantVals {
			if gotVals[i] != wantVal {
				t.Errorf("Query parameter %q[%d]: got %q, want %q", key, i, gotVals[i], wantVal)
			}
		}
	}

	// Check for unexpected parameters
	for key := range gotParams {
		if _, exists := wantParams[key]; !exists {
			t.Errorf("Unexpected query parameter %q", key)
		}
	}
}

// TestEvent_ValidCases tests all valid calendar events from testdata
func TestEvent_ValidCases(t *testing.T) {
	for _, tc := range ValidTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := tc.Event.GenerateURL()
			if err != nil {
				t.Fatalf("GenerateURL() error = %v, want nil", err)
			}
			// Use smart URL comparison that handles parameter ordering
			compareURLs(t, got, tc.ExpectedURL)
		})
	}
}

// TestEvent_ErrorCases tests all error cases from testdata
// NOTE: Now that GenerateURL() doesn't validate, we test Validate() explicitly.
// In production, validation is done via JSON Schema before calling GenerateURL().
func TestEvent_ErrorCases(t *testing.T) {
	for _, tc := range ErrorTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Event.Validate()
			if err == nil {
				t.Error("Validate() error = nil, want error")
			}
		})
	}
}

// TestEventDeterministic ensures the same input always produces the same output
func TestEventDeterministic(t *testing.T) {
	event := Event{
		Title:     "Test Event",
		StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
	}

	// Generate URL multiple times
	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		url, err := event.GenerateURL()
		if err != nil {
			t.Fatalf("GenerateURL() error = %v", err)
		}
		results[i] = url
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("GenerateURL() not deterministic: got different results\n  first: %v\n  got:   %v", results[0], results[i])
		}
	}
}

// TestEventURLStructure verifies the URL has the correct structure
func TestEventURLStructure(t *testing.T) {
	event := Event{
		Title:       "Test",
		StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Location:    "Location",
		Description: "Description",
	}

	url, err := event.GenerateURL()
	if err != nil {
		t.Fatalf("GenerateURL() error = %v", err)
	}

	// Verify URL structure
	if !strings.HasPrefix(url, "https://calendar.google.com/calendar/render?") {
		t.Errorf("URL should start with 'https://calendar.google.com/calendar/render?', got %v", url)
	}

	// Verify required parameters are present
	requiredParams := []string{"action=TEMPLATE", "text=", "dates="}
	for _, param := range requiredParams {
		if !strings.Contains(url, param) {
			t.Errorf("URL should contain '%s', got %v", param, url)
		}
	}
}
