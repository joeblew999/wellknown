package handlers

import (
	"html/template"
)

var Templates *template.Template
var LocalURL string
var MobileURL string

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
	LocalURL     string
	MobileURL    string
}
