package calendar

import (
	_ "embed"
	"encoding/json"
	"log"
)

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

// NOTE: Filename must match schema.ExamplesFilename constant
// but go:embed requires a literal string (can't use constants)
//
//go:embed data-examples.json
var examplesJSON []byte

// ShowcaseExamples provides examples for the Google Calendar showcase page
// Loaded from data-examples.json at compile time
var ShowcaseExamples []ShowcaseExample

func init() {
	// Parse examples from embedded JSON
	var data struct {
		Examples []ShowcaseExample `json:"examples"`
	}

	if err := json.Unmarshal(examplesJSON, &data); err != nil {
		log.Fatalf("Failed to parse data-examples.json: %v", err)
	}

	ShowcaseExamples = data.Examples
}
