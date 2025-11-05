package wellknown

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

// initTemplates loads and parses all HTML templates
func initTemplates() error {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}
	log.Println("✅ Templates loaded")
	return nil
}

// handleHome serves the home/login page
func handleHome(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authenticated := false
		authCookie, err := e.Request.Cookie("pb_auth")
		if err == nil && authCookie.Value != "" {
			authenticated = true
		}

		data := map[string]interface{}{
			"Authenticated": authenticated,
		}

		e.Response.Header().Set("Content-Type", "text/html")
		return templates.ExecuteTemplate(e.Response, "home.html", data)
	}
}

// handleCalendarPage serves the calendar events page
func handleCalendarPage(wk *Wellknown) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Check authentication
		authCookie, err := e.Request.Cookie("pb_auth")
		if err != nil || authCookie.Value == "" {
			return e.Redirect(http.StatusTemporaryRedirect, "/")
		}

		data := map[string]interface{}{
			"Authenticated": true,
		}

		e.Response.Header().Set("Content-Type", "text/html")
		return templates.ExecuteTemplate(e.Response, "calendar.html", data)
	}
}
