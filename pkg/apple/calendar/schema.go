package calendar

// Schema defines the JSON Schema for Apple Calendar events.
// Apple Calendar uses ICS format which supports full RFC 5545 spec with advanced features.
const Schema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Apple Calendar Event",
  "description": "Create an Apple Calendar event with full ICS support (attendees, recurrence, reminders)",
  "properties": {
    "title": {
      "type": "string",
      "title": "Event Title",
      "description": "Name of the event",
      "minLength": 1,
      "examples": ["Team Meeting", "Weekly Standup"]
    },
    "start": {
      "type": "string",
      "format": "datetime-local",
      "title": "Start Time",
      "description": "When the event begins"
    },
    "end": {
      "type": "string",
      "format": "datetime-local",
      "title": "End Time",
      "description": "When the event ends"
    },
    "location": {
      "type": "string",
      "title": "Location",
      "description": "Where the event takes place",
      "examples": ["Conference Room A", "Zoom"]
    },
    "description": {
      "type": "string",
      "title": "Description",
      "description": "Details about the event"
    },
    "allDay": {
      "type": "boolean",
      "title": "All Day Event",
      "description": "Whether this is an all-day event",
      "default": false
    },
    "attendees": {
      "type": "array",
      "title": "Attendees",
      "description": "People invited to the event",
      "items": {
        "type": "object",
        "properties": {
          "email": {
            "type": "string",
            "format": "email",
            "title": "Email"
          },
          "name": {
            "type": "string",
            "title": "Name"
          },
          "required": {
            "type": "boolean",
            "title": "Required",
            "default": false
          }
        },
        "required": ["email"]
      }
    },
    "recurrence": {
      "type": "object",
      "title": "Recurrence",
      "description": "Make this a recurring event",
      "properties": {
        "frequency": {
          "type": "string",
          "title": "Frequency",
          "enum": ["DAILY", "WEEKLY", "MONTHLY", "YEARLY"]
        },
        "interval": {
          "type": "integer",
          "title": "Interval",
          "minimum": 1,
          "default": 1
        },
        "count": {
          "type": "integer",
          "title": "Number of Occurrences",
          "minimum": 1
        },
        "until": {
          "type": "string",
          "format": "date",
          "title": "Repeat Until"
        }
      },
      "required": ["frequency"]
    },
    "reminders": {
      "type": "array",
      "title": "Reminders",
      "description": "Alerts before the event",
      "items": {
        "type": "object",
        "properties": {
          "minutesBefore": {
            "type": "integer",
            "title": "Minutes Before",
            "minimum": 0,
            "examples": [15, 30, 60, 1440]
          }
        },
        "required": ["minutesBefore"]
      }
    }
  },
  "required": ["title", "start", "end"]
}`
