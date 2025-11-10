// Package webui provides a reusable web GUI for environment variable inspection and debugging.
//
// Usage:
//
//	handler := webui.NewHandler(MyRegistry)
//	mux := http.NewServeMux()
//	handler.RegisterRoutes(mux)
//	http.ListenAndServe(":8080", mux)
package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/joeblew999/wellknown/pkg/env"
)

// Handler provides HTTP handlers for environment variable inspection.
type Handler struct {
	registry  *env.Registry
	baseURL   string
	startTime time.Time
}

// NewHandler creates a new webui handler for the given registry.
func NewHandler(registry *env.Registry) *Handler {
	return &Handler{
		registry:  registry,
		startTime: time.Now(),
	}
}

// RegisterRoutes registers all webui routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/env", h.handleEnv)
	mux.HandleFunc("/health", h.handleHealth)
}

// handleHealth returns health check information including environment detection.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":         "ok",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"environment":    env.DetectEnvironment(),
		"uptime":         time.Since(h.startTime).String(),
		"go_version":     runtime.Version(),
		"num_goroutines": runtime.NumGoroutine(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleEnv displays all environment variables from the registry.
// Supports dual format: HTML (default) and JSON (?format=json).
func (h *Handler) handleEnv(w http.ResponseWriter, r *http.Request) {
	vars := h.registry.All()
	grouped := groupVariables(vars)

	// Build JSON response
	response := map[string]interface{}{
		"total_variables": len(vars),
		"groups":          grouped,
		"environment":     env.DetectEnvironment(),
		"variables":       buildVariableStatus(vars),
	}

	// Check format preference
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Default: HTML output
	h.renderEnvHTML(w, grouped, vars)
}

// renderEnvHTML renders the HTML view of environment variables.
func (h *Handler) renderEnvHTML(w http.ResponseWriter, grouped map[string][]env.EnvVar, allVars []env.EnvVar) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	environment := env.DetectEnvironment()
	configured := countConfigured(allVars)
	missing := countMissingRequired(allVars)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>env | %s</title>
    %s
    %s
</head>
<body>
    <main class="container">
        <header>
            <h2>env</h2>
            <div class="stats">
                <span><strong>%d</strong> set</span>
                <span><strong>%d</strong> missing</span>
                <span>%s</span>
            </div>
        </header>

        <input type="search" id="filter" placeholder="Filter variables..." autocomplete="off">

        <div class="export-bar">
            <button onclick="copyAsExport()">Copy export commands</button>
            <button onclick="copyAsDotenv()">Copy .env format</button>
            <button onclick="copyAsJSON()">Copy JSON</button>
            <a href="/env?format=json" role="button" class="outline">View JSON</a>
        </div>

        <table id="envTable">
            <thead>
                <tr>
                    <th></th>
                    <th>Variable</th>
                    <th>Value</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>`,
		environment,
		picoCSSLink,
		customStyles,
		configured,
		missing,
		environment,
	)

	// Render ALL variables in a single table (no grouping - simpler!)
	for _, v := range allVars {
		html += renderVariableRow(v)
	}

	html += `
            </tbody>
        </table>
    </main>
    <script>
// Filter functionality
document.getElementById('filter').addEventListener('input', (e) => {
    const filter = e.target.value.toLowerCase();
    document.querySelectorAll('#envTable tbody tr').forEach(row => {
        const varName = row.getAttribute('data-var');
        row.classList.toggle('hidden', !varName.toLowerCase().includes(filter));
    });
});

// Copy functionality
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        console.log('Copied to clipboard');
    });
}

function copyValue(name, value) {
    copyToClipboard(value);
}

function copyAsExport() {
    const lines = [];
    document.querySelectorAll('#envTable tbody tr').forEach(row => {
        if (row.classList.contains('hidden')) return;
        const name = row.getAttribute('data-var');
        const value = row.getAttribute('data-value');
        if (value) {
            lines.push('export ' + name + '="' + value + '"');
        }
    });
    copyToClipboard(lines.join('\n'));
}

function copyAsDotenv() {
    const lines = [];
    document.querySelectorAll('#envTable tbody tr').forEach(row => {
        if (row.classList.contains('hidden')) return;
        const name = row.getAttribute('data-var');
        const value = row.getAttribute('data-value');
        if (value) {
            lines.push(name + '=' + value);
        }
    });
    copyToClipboard(lines.join('\n'));
}

function copyAsJSON() {
    const obj = {};
    document.querySelectorAll('#envTable tbody tr').forEach(row => {
        if (row.classList.contains('hidden')) return;
        const name = row.getAttribute('data-var');
        const value = row.getAttribute('data-value');
        if (value) {
            obj[name] = value;
        }
    });
    copyToClipboard(JSON.stringify(obj, null, 2));
}
    </script>
</body>
</html>`

	fmt.Fprint(w, html)
}

// Helper functions

// groupVariables groups variables by their Group field.
func groupVariables(vars []env.EnvVar) map[string][]env.EnvVar {
	grouped := make(map[string][]env.EnvVar)
	for _, v := range vars {
		group := v.Group
		if group == "" {
			group = "General"
		}
		grouped[group] = append(grouped[group], v)
	}
	return grouped
}

// buildVariableStatus creates the variable status map for JSON responses.
func buildVariableStatus(vars []env.EnvVar) map[string]interface{} {
	varStatus := make(map[string]interface{})
	for _, v := range vars {
		value := os.Getenv(v.Name)
		status := map[string]interface{}{
			"configured":  value != "",
			"required":    v.Required,
			"secret":      v.Secret,
			"has_default": v.Default != "",
		}
		if v.Secret && value != "" {
			status["value"] = "••••••••"
		} else if value != "" {
			status["value"] = value
		}
		varStatus[strings.ToLower(v.Name)+"_configured"] = status["configured"]
	}
	return varStatus
}

// wantsJSON checks if the client wants JSON response.
func wantsJSON(r *http.Request) bool {
	format := r.URL.Query().Get("format")
	acceptHeader := r.Header.Get("Accept")
	return format == "json" || strings.Contains(acceptHeader, "application/json")
}

// renderVariableRow renders a single variable as a table row - ultra-simple developer format
func renderVariableRow(v env.EnvVar) string {
	value := os.Getenv(v.Name)
	configured := value != ""

	// Row class for highlighting missing required vars
	rowClass := ""
	if v.Required && !configured {
		rowClass = " class=\"missing-required\""
	}

	// Status dot
	statusClass := "status-empty"
	if v.Required && !configured {
		statusClass = "status-missing"
	} else if configured {
		statusClass = "status-set"
	}

	// Value display with copy button
	var valueHTML string
	if configured {
		if v.Secret {
			valueHTML = fmt.Sprintf(`<div class="value-cell"><span class="secret">••••••••</span></div>`)
		} else {
			// Escape value for HTML attribute
			escapedValue := strings.ReplaceAll(value, `"`, `&quot;`)
			valueHTML = fmt.Sprintf(`<div class="value-cell"><code>%s</code><button class="copy-btn" onclick="copyValue('%s', '%s')" title="Copy value">📋</button></div>`,
				value, v.Name, escapedValue)
		}
	} else if v.Default != "" {
		valueHTML = fmt.Sprintf(`<span class="empty">default: %s</span>`, v.Default)
	} else {
		valueHTML = `<span class="empty">—</span>`
	}

	// Tags - minimal badges
	var tags []string
	if v.Secret {
		tags = append(tags, `<span class="tag tag-secret">SECRET</span>`)
	}
	if v.Required {
		tags = append(tags, `<span class="tag tag-required">REQ</span>`)
	}
	tagsHTML := strings.Join(tags, " ")

	// Escape value for data attribute
	dataValue := ""
	if configured && !v.Secret {
		dataValue = strings.ReplaceAll(value, `"`, `&quot;`)
	}

	return fmt.Sprintf(`
                <tr%s data-var="%s" data-value="%s">
                    <td><span class="status %s"></span></td>
                    <td><span class="var-name">%s</span> %s</td>
                    <td>%s</td>
                    <td></td>
                </tr>`,
		rowClass, v.Name, dataValue,
		statusClass,
		v.Name, tagsHTML,
		valueHTML)
}

// countMissingRequired counts how many required variables are not configured
func countMissingRequired(vars []env.EnvVar) int {
	count := 0
	for _, v := range vars {
		if v.Required && os.Getenv(v.Name) == "" {
			count++
		}
	}
	return count
}

// countConfigured counts how many variables are configured.
func countConfigured(vars []env.EnvVar) int {
	count := 0
	for _, v := range vars {
		if os.Getenv(v.Name) != "" {
			count++
		}
	}
	return count
}

// countSecrets counts how many variables are marked as secrets.
func countSecrets(vars []env.EnvVar) int {
	count := 0
	for _, v := range vars {
		if v.Secret {
			count++
		}
	}
	return count
}
