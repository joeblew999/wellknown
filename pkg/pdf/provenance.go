package pdfform

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Provenance tracks the origin and processing history of a PDF form
type Provenance struct {
	CatalogFormCode string    `json:"catalog_form_code,omitempty"`
	CatalogState    string    `json:"catalog_state,omitempty"`
	SourceURL       string    `json:"source_url,omitempty"`
	DownloadedAt    time.Time `json:"downloaded_at,omitempty"`
	InspectedAt     time.Time `json:"inspected_at,omitempty"`
}

// SaveProvenanceMetadata saves provenance metadata to a .meta.json file alongside the PDF
func SaveProvenanceMetadata(pdfPath, formCode, state, sourceURL string) error {
	prov := &Provenance{
		CatalogFormCode: formCode,
		CatalogState:    state,
		SourceURL:       sourceURL,
		DownloadedAt:    time.Now(),
	}

	metaPath := pdfPath + ".meta.json"
	data, err := json.MarshalIndent(prov, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provenance: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write provenance metadata: %w", err)
	}

	return nil
}

// LoadProvenanceMetadata loads provenance metadata from a .meta.json file
func LoadProvenanceMetadata(pdfPath string) (*Provenance, error) {
	metaPath := pdfPath + ".meta.json"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata file, not an error
		}
		return nil, fmt.Errorf("failed to read provenance metadata: %w", err)
	}

	var prov Provenance
	if err := json.Unmarshal(data, &prov); err != nil {
		return nil, fmt.Errorf("failed to parse provenance metadata: %w", err)
	}

	return &prov, nil
}

// AddInspectedTimestamp updates the provenance metadata to mark when inspection happened
func AddInspectedTimestamp(pdfPath string) error {
	prov, err := LoadProvenanceMetadata(pdfPath)
	if err != nil || prov == nil {
		return err // Nothing to update
	}

	prov.InspectedAt = time.Now()

	// Save back to metadata file
	metaPath := pdfPath + ".meta.json"
	data, err := json.MarshalIndent(prov, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provenance: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write provenance metadata: %w", err)
	}

	return nil
}
