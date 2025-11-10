package commands

import (
	"encoding/json"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Browse events
	EventBrowseStarted   EventType = "browse.started"
	EventBrowseCompleted EventType = "browse.completed"
	EventBrowseError     EventType = "browse.error"

	// Download events
	EventDownloadStarted   EventType = "download.started"
	EventDownloadProgress  EventType = "download.progress"
	EventDownloadCompleted EventType = "download.completed"
	EventDownloadError     EventType = "download.error"

	// Inspect events
	EventInspectStarted   EventType = "inspect.started"
	EventInspectCompleted EventType = "inspect.completed"
	EventInspectError     EventType = "inspect.error"

	// Fill events
	EventFillStarted   EventType = "fill.started"
	EventFillCompleted EventType = "fill.completed"
	EventFillError     EventType = "fill.error"

	// Case events
	EventCaseCreated EventType = "case.created"
	EventCaseLoaded  EventType = "case.loaded"
	EventCaseUpdated EventType = "case.updated"
	EventCaseError   EventType = "case.error"

	// Test events
	EventTestStarted   EventType = "test.started"
	EventTestCompleted EventType = "test.completed"
	EventTestError     EventType = "test.error"
)

// Event represents a system event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Error     error                  `json:"error,omitempty"`
}

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewErrorEvent creates a new error event
func NewErrorEvent(eventType EventType, err error, data map[string]interface{}) *Event {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}
}

// ToJSON converts the event to JSON string
func (e *Event) ToJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// Event Data Schemas
// These document the expected data fields for each event type.
// Use these as a reference when emitting events to ensure consistency.

// BrowseStartedData contains fields for browse.started event
type BrowseStartedData struct {
	CatalogPath string `json:"catalog_path"`
	State       string `json:"state,omitempty"`
}

// BrowseCompletedData contains fields for browse.completed event
type BrowseCompletedData struct {
	CatalogPath string `json:"catalog_path"`
	StateCount  int    `json:"state_count"`
	FormCount   int    `json:"form_count"`
}

// BrowseErrorData contains fields for browse.error event
type BrowseErrorData struct {
	CatalogPath string `json:"catalog_path"`
	Stage       string `json:"stage"` // load_catalog, filter_forms
}

// DownloadStartedData contains fields for download.started event
type DownloadStartedData struct {
	FormCode  string `json:"form_code"`
	OutputDir string `json:"output_dir"`
}

// DownloadProgressData contains fields for download.progress event
type DownloadProgressData struct {
	FormCode string  `json:"form_code"`
	FormName string  `json:"form_name,omitempty"`
	State    string  `json:"state,omitempty"`
	PDFPath  string  `json:"pdf_path,omitempty"`
	Stage    string  `json:"stage"`    // found_form, downloading, saving_metadata
	Progress float64 `json:"progress"` // 0.0 - 1.0
}

// DownloadCompletedData contains fields for download.completed event
type DownloadCompletedData struct {
	FormCode string  `json:"form_code"`
	PDFPath  string  `json:"pdf_path"`
	FormName string  `json:"form_name"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"` // Should be 1.0
}

// DownloadErrorData contains fields for download.error event
type DownloadErrorData struct {
	FormCode string `json:"form_code"`
	Stage    string `json:"stage"` // load_catalog, find_form, check_url, create_dir, download_pdf
}

// InspectStartedData contains fields for inspect.started event
type InspectStartedData struct {
	PDFPath   string `json:"pdf_path"`
	OutputDir string `json:"output_dir"`
}

// InspectCompletedData contains fields for inspect.completed event
type InspectCompletedData struct {
	PDFPath      string `json:"pdf_path"`
	TemplatePath string `json:"template_path"`
	FieldCount   int    `json:"field_count"`
}

// InspectErrorData contains fields for inspect.error event
type InspectErrorData struct {
	PDFPath string `json:"pdf_path"`
	Stage   string `json:"stage"` // list_fields, export_json
}

// FillStartedData contains fields for fill.started event
type FillStartedData struct {
	DataPath  string `json:"data_path"`
	OutputDir string `json:"output_dir"`
	Flatten   bool   `json:"flatten"`
	CasePath  string `json:"case_path,omitempty"` // If filling from case
}

// FillCompletedData contains fields for fill.completed event
type FillCompletedData struct {
	DataPath   string `json:"data_path,omitempty"`
	CasePath   string `json:"case_path,omitempty"`
	OutputPath string `json:"output_path"`
	InputPDF   string `json:"input_pdf"`
	Flattened  bool   `json:"flattened"`
}

// FillErrorData contains fields for fill.error event
type FillErrorData struct {
	DataPath string `json:"data_path,omitempty"`
	CasePath string `json:"case_path,omitempty"`
	Stage    string `json:"stage"` // create_dir, fill_pdf, flatten, fill_from_case
}

// CaseCreatedData contains fields for case.created event
type CaseCreatedData struct {
	CaseID     string `json:"case_id"`
	CaseName   string `json:"case_name"`
	EntityName string `json:"entity_name"`
	FormCode   string `json:"form_code"`
	CasePath   string `json:"case_path"`
}

// CaseLoadedData contains fields for case.loaded event
type CaseLoadedData struct {
	CasePath   string `json:"case_path"`
	CaseID     string `json:"case_id"`
	CaseName   string `json:"case_name"`
	EntityName string `json:"entity_name"`
	FormCode   string `json:"form_code"`
}

// CaseUpdatedData contains fields for case.updated event
type CaseUpdatedData struct {
	CasePath string `json:"case_path"`
	CaseID   string `json:"case_id"`
}

// CaseErrorData contains fields for case.error event
type CaseErrorData struct {
	CasePath string `json:"case_path,omitempty"`
	Stage    string `json:"stage"` // load_case, save_case, create
}

// TestStartedData contains fields for test.started event
type TestStartedData struct {
	TestName string `json:"test_name"`
	CasePath string `json:"case_path"`
}

// TestCompletedData contains fields for test.completed event
type TestCompletedData struct {
	TestName string `json:"test_name"`
	CasePath string `json:"case_path"`
	Success  bool   `json:"success"`
}

// TestErrorData contains fields for test.error event
type TestErrorData struct {
	TestName string `json:"test_name"`
	CasePath string `json:"case_path,omitempty"`
	Stage    string `json:"stage"`
}

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *Event
	bufferSize  int
}

// DefaultEventBus is the global event bus instance
var DefaultEventBus = NewEventBus(100)

// NewEventBus creates a new event bus with the specified buffer size
func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan *Event),
		bufferSize:  bufferSize,
	}
}

// Subscribe subscribes to events matching the pattern
// Pattern can be:
//   - Exact match: "browse.started"
//   - Wildcard: "browse.*" matches all browse events
//   - All events: "*"
func (eb *EventBus) Subscribe(pattern string) chan *Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan *Event, eb.bufferSize)
	eb.subscribers[pattern] = append(eb.subscribers[pattern], ch)
	return ch
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(ch chan *Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for pattern, channels := range eb.subscribers {
		for i, c := range channels {
			if c == ch {
				// Remove channel from slice
				eb.subscribers[pattern] = append(channels[:i], channels[i+1:]...)
				close(ch)
				return
			}
		}
	}
}

// Publish publishes an event to all matching subscribers
func (eb *EventBus) Publish(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Find all matching subscribers
	for pattern, channels := range eb.subscribers {
		if eb.matchPattern(pattern, string(event.Type)) {
			for _, ch := range channels {
				select {
				case ch <- event:
				default:
					// Channel full, skip this subscriber
				}
			}
		}
	}
}

// matchPattern checks if an event type matches a subscription pattern
func (eb *EventBus) matchPattern(pattern, eventType string) bool {
	if pattern == "*" {
		return true
	}

	// Check for wildcard pattern (e.g., "browse.*")
	if len(pattern) > 2 && pattern[len(pattern)-2:] == ".*" {
		prefix := pattern[:len(pattern)-2]
		return len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix
	}

	// Exact match
	return pattern == eventType
}

// Clear removes all subscribers (useful for testing)
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, channels := range eb.subscribers {
		for _, ch := range channels {
			close(ch)
		}
	}
	eb.subscribers = make(map[string][]chan *Event)
}

// Helper functions for publishing events

// Emit publishes an event to the default event bus
func Emit(eventType EventType, data map[string]interface{}) {
	event := NewEvent(eventType, data)
	DefaultEventBus.Publish(event)
}

// EmitError publishes an error event to the default event bus
func EmitError(eventType EventType, err error, data map[string]interface{}) {
	event := NewErrorEvent(eventType, err, data)
	DefaultEventBus.Publish(event)
}

// Subscribe subscribes to events on the default event bus
func Subscribe(pattern string) chan *Event {
	return DefaultEventBus.Subscribe(pattern)
}

// Unsubscribe removes a subscription from the default event bus
func Unsubscribe(ch chan *Event) {
	DefaultEventBus.Unsubscribe(ch)
}
