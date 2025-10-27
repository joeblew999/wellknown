package calendar

// UISchema defines the UI layout for Google Calendar event forms
// This controls how the form is presented visually (layout, grouping, etc.)
// while the JSON Schema (schema.go) defines validation and data types.
const UISchema = `{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "Label",
      "text": "📅 Event Details"
    },
    {
      "type": "Control",
      "scope": "#/properties/title",
      "options": {
        "placeholder": "e.g., Team Meeting"
      }
    },
    {
      "type": "HorizontalLayout",
      "elements": [
        {
          "type": "Control",
          "scope": "#/properties/start",
          "label": "Start Time"
        },
        {
          "type": "Control",
          "scope": "#/properties/end",
          "label": "End Time"
        }
      ]
    },
    {
      "type": "Label",
      "text": "📍 Location & Description"
    },
    {
      "type": "Control",
      "scope": "#/properties/location",
      "options": {
        "placeholder": "Conference Room A, Zoom link, etc."
      }
    },
    {
      "type": "Control",
      "scope": "#/properties/description",
      "options": {
        "multi": true,
        "placeholder": "Add event details, agenda, or notes..."
      }
    }
  ]
}`
