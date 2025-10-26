package google

import (
	"strings"
	"testing"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// TestCalendar_ValidCases tests all valid calendar events from testdata
func TestCalendar_ValidCases(t *testing.T) {
	for _, tc := range CalendarEvents {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := Calendar(tc.Event)
			if err != nil {
				t.Fatalf("Calendar() error = %v, want nil", err)
			}
			if got != tc.ExpectedURL {
				t.Errorf("Calendar() = %v, want %v", got, tc.ExpectedURL)
			}
		})
	}
}

// TestCalendar_ErrorCases tests all error cases from testdata
func TestCalendar_ErrorCases(t *testing.T) {
	for _, tc := range ErrorTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := Calendar(tc.Event)
			if err == nil {
				t.Error("Calendar() error = nil, want error")
			}
		})
	}
}

// TestCalendarDeterministic ensures the same input always produces the same output
func TestCalendarDeterministic(t *testing.T) {
	event := types.CalendarEvent{
		Title:     "Test Event",
		StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
	}

	// Generate URL multiple times
	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		url, err := Calendar(event)
		if err != nil {
			t.Fatalf("Calendar() error = %v", err)
		}
		results[i] = url
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Calendar() not deterministic: got different results\n  first: %v\n  got:   %v", results[0], results[i])
		}
	}
}

// TestFormatTime tests the time formatting function
func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "midnight UTC",
			time: time.Date(2025, 10, 26, 0, 0, 0, 0, time.UTC),
			want: "20251026T000000Z",
		},
		{
			name: "afternoon time",
			time: time.Date(2025, 10, 26, 14, 30, 45, 0, time.UTC),
			want: "20251026T143045Z",
		},
		{
			name: "non-UTC timezone gets converted to UTC",
			time: time.Date(2025, 10, 26, 14, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			want: "20251026T220000Z", // 14:00 PST = 22:00 UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTime(tt.time); got != tt.want {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalendarURLStructure verifies the URL has the correct structure
func TestCalendarURLStructure(t *testing.T) {
	event := types.CalendarEvent{
		Title:       "Test",
		StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Location:    "Location",
		Description: "Description",
	}

	url, err := Calendar(event)
	if err != nil {
		t.Fatalf("Calendar() error = %v", err)
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
