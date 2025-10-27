package calendar

import "time"

// Frequency represents how often a recurring event repeats
type Frequency string

const (
	FrequencyDaily   Frequency = "DAILY"
	FrequencyWeekly  Frequency = "WEEKLY"
	FrequencyMonthly Frequency = "MONTHLY"
	FrequencyYearly  Frequency = "YEARLY"
)

// RecurrenceRule defines when and how an event repeats (RRULE in ICS)
type RecurrenceRule struct {
	Frequency  Frequency      // How often: DAILY, WEEKLY, MONTHLY, YEARLY
	Interval   int            // Every N days/weeks/months/years (default: 1)
	Count      *int           // Number of occurrences (nil = infinite)
	Until      *time.Time     // End date (nil = no end)
	ByDay      []time.Weekday // Which days of week (for WEEKLY)
	ByMonthDay []int          // Which days of month (for MONTHLY)
}

// AttendeeRole defines the role of an attendee
type AttendeeRole string

const (
	RoleReqParticipant AttendeeRole = "REQ-PARTICIPANT" // Required attendee
	RoleOptParticipant AttendeeRole = "OPT-PARTICIPANT" // Optional attendee
	RoleNonParticipant AttendeeRole = "NON-PARTICIPANT" // Informational only
	RoleChair          AttendeeRole = "CHAIR"           // Meeting organizer
)

// AttendeeStatus defines the participation status
type AttendeeStatus string

const (
	StatusNeedsAction AttendeeStatus = "NEEDS-ACTION" // Not yet responded
	StatusAccepted    AttendeeStatus = "ACCEPTED"     // Accepted
	StatusDeclined    AttendeeStatus = "DECLINED"     // Declined
	StatusTentative   AttendeeStatus = "TENTATIVE"    // Maybe
	StatusDelegated   AttendeeStatus = "DELEGATED"    // Delegated to someone else
)

// Attendee represents a person invited to the event
type Attendee struct {
	Name   string         // Display name (optional)
	Email  string         // Email address (required)
	Role   AttendeeRole   // Participation role
	Status AttendeeStatus // RSVP status
	RSVP   bool           // Whether RSVP is requested
}

// Organizer represents the event organizer
type Organizer struct {
	Name  string // Display name (optional)
	Email string // Email address (required)
}

// Reminder represents an alarm/reminder before the event
type Reminder struct {
	Duration time.Duration // How long before event to trigger
}

// EventStatus represents the event status
type EventStatus string

const (
	StatusConfirmed EventStatus = "CONFIRMED"
	StatusTentativeEvent EventStatus = "TENTATIVE"
	StatusCancelled EventStatus = "CANCELLED"
)
