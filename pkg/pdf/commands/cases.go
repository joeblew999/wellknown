package commands

import (
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

// CreateCaseOptions contains options for creating a case
type CreateCaseOptions struct {
	FormCode   string
	CaseName   string
	EntityName string
	DataDir    string
}

// CreateCaseResult contains the result of creating a case
type CreateCaseResult struct {
	Case     *pdfform.Case
	CasePath string
}

// CreateCase creates a new case
// Emits events: case.created, case.error
func CreateCase(formCode, caseName, entityName, dataDir string) (*pdfform.Case, string, error) {
	// Emit started event (using case.created type since there's no case.started)
	// We could add a case.creating event if needed

	c, casePath, err := pdfform.CreateCase(formCode, caseName, entityName, dataDir)
	if err != nil {
		EmitError(EventCaseError, err, map[string]interface{}{
			"form_code":   formCode,
			"case_name":   caseName,
			"entity_name": entityName,
			"stage":       "create",
		})
		return nil, "", err
	}

	// Emit created event
	Emit(EventCaseCreated, map[string]interface{}{
		"case_id":     c.Metadata.CaseID,
		"case_name":   caseName,
		"form_code":   formCode,
		"entity_name": entityName,
		"case_path":   casePath,
	})

	return c, casePath, nil
}

// ListCasesOptions contains options for listing cases
type ListCasesOptions struct {
	DataDir    string
	EntityName string
}

// ListCases lists all available cases, optionally filtered by entity
// Does not emit events (read-only operation)
func ListCases(dataDir, entityName string) ([]string, error) {
	return pdfform.ListCases(dataDir, entityName)
}

// LoadCaseOptions contains options for loading a case
type LoadCaseOptions struct {
	CasePath string
}

// LoadCase loads a case from a file
// Emits events: case.loaded, case.error
func LoadCase(casePath string) (*pdfform.Case, error) {
	c, err := pdfform.LoadCase(casePath)
	if err != nil {
		EmitError(EventCaseError, err, map[string]interface{}{
			"case_path": casePath,
			"stage":     "load",
		})
		return nil, err
	}

	// Emit loaded event
	Emit(EventCaseLoaded, map[string]interface{}{
		"case_id":   c.Metadata.CaseID,
		"case_name": c.Metadata.CaseName,
		"form_code": c.FormReference.FormCode,
		"case_path": casePath,
	})

	return c, nil
}

// SaveCaseOptions contains options for saving a case
type SaveCaseOptions struct {
	Case     *pdfform.Case
	CasePath string
}

// SaveCase saves a case to a file
// Emits events: case.updated, case.error
func SaveCase(c *pdfform.Case, casePath string) error {
	err := pdfform.SaveCase(c, casePath)
	if err != nil {
		EmitError(EventCaseError, err, map[string]interface{}{
			"case_id": c.Metadata.CaseID,
			"stage":   "save",
		})
		return err
	}

	// Emit updated event
	Emit(EventCaseUpdated, map[string]interface{}{
		"case_id":   c.Metadata.CaseID,
		"case_name": c.Metadata.CaseName,
		"form_code": c.FormReference.FormCode,
		"case_path": casePath,
	})

	return nil
}

// FindCaseByID finds a case file by its ID
// Does not emit events (read-only helper)
func FindCaseByID(caseID, dataDir string) (string, error) {
	cases, err := ListCases(dataDir, "")
	if err != nil {
		return "", err
	}

	for _, c := range cases {
		if filepath.Base(c) == caseID+".json" {
			return c, nil
		}
	}

	return "", nil
}
