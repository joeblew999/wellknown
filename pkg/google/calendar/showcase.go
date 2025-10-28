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

// ShowcaseExamples provides examples for the Google Calendar showcase page
var ShowcaseExamples = []ShowcaseExample{
	{
		Name:        "Team Meeting",
		Description: "Weekly team standup meeting",
		Data: map[string]interface{}{
			"title":       "Team Standup",
			"start":       "2025-11-01T10:00",
			"end":         "2025-11-01T10:30",
			"location":    "Conference Room A",
			"description": "Weekly sync to discuss progress and blockers",
		},
	},
	{
		Name:        "Client Presentation",
		Description: "Quarterly business review with client",
		Data: map[string]interface{}{
			"title":       "Q4 Business Review",
			"start":       "2025-11-15T14:00",
			"end":         "2025-11-15T16:00",
			"location":    "Executive Boardroom",
			"description": "Present Q4 results and discuss 2026 plans",
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
		Description: "Professional development workshop",
		Data: map[string]interface{}{
			"title":       "Leadership Training Workshop",
			"start":       "2025-12-05T09:00",
			"end":         "2025-12-05T17:00",
			"location":    "Training Center",
			"description": "All-day workshop on effective leadership and team management",
		},
	},
	{
		Name:        "Sprint Planning",
		Description: "Plan next development sprint",
		Data: map[string]interface{}{
			"title":       "Sprint 12 Planning",
			"start":       "2025-11-25T10:00",
			"end":         "2025-11-25T12:00",
			"location":    "Zoom Meeting",
			"description": "Plan tasks and set goals for the next 2-week sprint",
		},
	},
	{
		Name:        "Code Review",
		Description: "Review pull requests and discuss architecture",
		Data: map[string]interface{}{
			"title":       "Weekly Code Review",
			"start":       "2025-11-08T15:00",
			"end":         "2025-11-08T16:00",
			"location":    "Engineering Room",
			"description": "Review PRs from the week and discuss any architectural concerns",
		},
	},
}
