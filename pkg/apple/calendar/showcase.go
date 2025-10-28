package calendar

// ShowcaseExample represents a calendar example for the showcase page
type ShowcaseExample struct {
	Name        string                 `json:"name"`        // Display name for the card
	Description string                 `json:"description"` // Card description
	Data        map[string]interface{} `json:"data"`        // Actual form data
}

// GetName returns the example name for the showcase
func (e ShowcaseExample) GetName() string { return e.Name }

// GetDescription returns the example description
func (e ShowcaseExample) GetDescription() string { return e.Description }

// ShowcaseExamples provides examples for the Apple Calendar showcase page
var ShowcaseExamples = []ShowcaseExample{
	{
		Name:        "Team Meeting",
		Description: "Weekly team standup with calendar features",
		Data: map[string]interface{}{
			"title":       "Team Standup",
			"start":       "2025-11-01T10:00",
			"end":         "2025-11-01T10:30",
			"location":    "Conference Room A",
			"description": "Weekly sync to discuss progress and blockers",
		},
	},
	{
		Name:        "All-Day Conference",
		Description: "Full-day event demonstration",
		Data: map[string]interface{}{
			"title":       "Tech Conference 2025",
			"start":       "2025-12-10T00:00",
			"end":         "2025-12-11T00:00",
			"allDay":      true,
			"location":    "Convention Center",
			"description": "Annual technology conference with keynotes and workshops",
		},
	},
	{
		Name:        "Client Presentation",
		Description: "Important client meeting",
		Data: map[string]interface{}{
			"title":       "Q4 Business Review",
			"start":       "2025-11-15T14:00",
			"end":         "2025-11-15T16:00",
			"location":    "Executive Boardroom",
			"description": "Present Q4 results and discuss 2026 strategic plans",
		},
	},
	{
		Name:        "Lunch Break",
		Description: "Team lunch outing",
		Data: map[string]interface{}{
			"title":       "Team Lunch",
			"start":       "2025-11-20T12:00",
			"end":         "2025-11-20T13:30",
			"location":    "Downtown Restaurant",
			"description": "Celebrating the successful product launch!",
		},
	},
	{
		Name:        "Workshop",
		Description: "Full-day professional development",
		Data: map[string]interface{}{
			"title":       "Leadership Training",
			"start":       "2025-12-05T09:00",
			"end":         "2025-12-05T17:00",
			"location":    "Training Center",
			"description": "All-day workshop on effective leadership and team management",
		},
	},
	{
		Name:        "Birthday",
		Description: "All-day birthday celebration",
		Data: map[string]interface{}{
			"title":       "John's Birthday",
			"start":       "2025-11-30T00:00",
			"end":         "2025-12-01T00:00",
			"allDay":      true,
			"description": "Remember to wish John a happy birthday!",
		},
	},
}
