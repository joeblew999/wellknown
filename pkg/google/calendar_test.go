package google

import (
	"strings"
	"testing"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

func TestCalendar(t *testing.T) {
	tests := []struct {
		name    string
		event   types.CalendarEvent
		want    string
		wantErr bool
	}{
		{
			name: "complete event with all fields",
			event: types.CalendarEvent{
				Title:       "Team Meeting",
				StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
				Location:    "Conference Room A",
				Description: "Quarterly planning meeting",
			},
			want:    "googlecalendar://render?action=CREATE&text=Team+Meeting&dates=20251026T140000Z/20251026T150000Z&location=Conference+Room+A&details=Quarterly+planning+meeting",
			wantErr: false,
		},
		{
			name: "minimal event with only required fields",
			event: types.CalendarEvent{
				Title:     "Quick Sync",
				StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 27, 10, 30, 0, 0, time.UTC),
			},
			want:    "googlecalendar://render?action=CREATE&text=Quick+Sync&dates=20251027T100000Z/20251027T103000Z",
			wantErr: false,
		},
		{
			name: "event with location but no description",
			event: types.CalendarEvent{
				Title:     "Client Visit",
				StartTime: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
				Location:  "123 Main St",
			},
			want:    "googlecalendar://render?action=CREATE&text=Client+Visit&dates=20251101T090000Z/20251101T100000Z&location=123+Main+St",
			wantErr: false,
		},
		{
			name: "event with special characters in title",
			event: types.CalendarEvent{
				Title:     "Project Review & Planning",
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
			want:    "googlecalendar://render?action=CREATE&text=Project+Review+%26+Planning&dates=20251026T140000Z/20251026T150000Z",
			wantErr: false,
		},
		{
			name: "missing title returns error",
			event: types.CalendarEvent{
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "missing start time returns error",
			event: types.CalendarEvent{
				Title:   "Meeting",
				EndTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "missing end time returns error",
			event: types.CalendarEvent{
				Title:     "Meeting",
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "end time before start time returns error",
			event: types.CalendarEvent{
				Title:     "Meeting",
				StartTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calendar(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calendar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Calendar() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalendarDeterministic ensures same input produces same output
func TestCalendarDeterministic(t *testing.T) {
	event := types.CalendarEvent{
		Title:       "Test Event",
		StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Location:    "Test Location",
		Description: "Test Description",
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
	first := results[0]
	for i, result := range results {
		if result != first {
			t.Errorf("Deterministic test failed: result[%d] = %v, want %v", i, result, first)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "midnight UTC",
			time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			want: "20250101T000000Z",
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
			got := formatTime(tt.time)
			if got != tt.want {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	if !strings.HasPrefix(url, "googlecalendar://render?") {
		t.Errorf("URL should start with 'googlecalendar://render?', got %v", url)
	}

	// Verify required parameters are present
	requiredParams := []string{"action=CREATE", "text=", "dates="}
	for _, param := range requiredParams {
		if !strings.Contains(url, param) {
			t.Errorf("URL should contain '%s', got %v", param, url)
		}
	}
}
