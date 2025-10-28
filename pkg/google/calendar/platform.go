package calendar

const (
	PlatformName = "google"
	ServiceName  = "calendar"
	OutputType   = "url"

	BaseURL     = "https://calendar.google.com/calendar/render"
	ActionParam = "TEMPLATE"
	TimeFormat  = "20060102T150405Z"

	QueryParamAction = "action"
	QueryParamDates  = "dates"
)

var FieldMapping = map[string]string{
	"title":       "text",
	"location":    "location",
	"description": "details",
}
