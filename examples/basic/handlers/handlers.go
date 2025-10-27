package handlers

import (
	"html/template"

	"github.com/joeblew999/wellknown/pkg/types"
)

var Templates *template.Template
var LocalURL string
var MobileURL string

// ServiceExample is the interface that all service examples must implement
type ServiceExample interface {
	GetName() string
	GetDescription() string
}

type PageData struct {
	Platform     string
	AppType      string
	CurrentPage  string
	TemplateName string
	IsStub       bool
	GeneratedURL string
	AppURL       string // Native app deep link
	Error        string
	Event        *types.CalendarEvent
	TestCases    interface{} // Keep as interface{} for flexibility, but document expected types
	LocalURL     string
	MobileURL    string
}
