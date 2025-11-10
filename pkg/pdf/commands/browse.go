package commands

import (
	"fmt"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

// BrowseOptions contains options for browsing the forms catalog
type BrowseOptions struct {
	CatalogPath string
	State       string
}

// BrowseResult contains the results of browsing the forms catalog
type BrowseResult struct {
	States []string
	Forms  []pdfform.TransferForm
}

// Browse loads the forms catalog and returns either all states or forms for a specific state
// Emits events: browse.started, browse.completed, browse.error
func Browse(opts BrowseOptions) (*BrowseResult, error) {
	// Emit started event
	Emit(EventBrowseStarted, map[string]interface{}{
		"catalog_path": opts.CatalogPath,
		"state":        opts.State,
	})

	catalog, err := pdfform.LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		EmitError(EventBrowseError, err, map[string]interface{}{
			"catalog_path": opts.CatalogPath,
		})
		return nil, fmt.Errorf("failed to load forms catalog: %w", err)
	}

	result := &BrowseResult{}

	if opts.State == "" {
		// Return all states
		result.States = catalog.ListStates()
		Emit(EventBrowseCompleted, map[string]interface{}{
			"state_count": len(result.States),
			"states":      result.States,
		})
	} else {
		// Return forms for specific state
		result.Forms = catalog.GetFormsByState(opts.State)
		if len(result.Forms) == 0 {
			err := fmt.Errorf("no forms found for state: %s", opts.State)
			EmitError(EventBrowseError, err, map[string]interface{}{
				"state": opts.State,
			})
			return nil, err
		}
		Emit(EventBrowseCompleted, map[string]interface{}{
			"state":      opts.State,
			"form_count": len(result.Forms),
		})
	}

	return result, nil
}
