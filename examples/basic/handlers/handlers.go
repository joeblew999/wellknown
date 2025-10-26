package handlers

import (
	"html/template"
)

var Templates *template.Template

type PageData struct {
	Platform     string
	AppType      string
	CurrentPage  string
	TemplateName string
	IsStub       bool
	GeneratedURL string
	Error        string
	Event        interface{}
	TestCases    interface{}
}
