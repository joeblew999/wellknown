package server

import "fmt"

// ServiceConfig represents a registered service in the navigation
type ServiceConfig struct {
	Platform    string
	AppType     string
	Title       string
	HasCustom   bool
	HasShowcase bool
}

// registeredServices tracks all services for navigation generation
var registeredServices []ServiceConfig

// registerService adds a service to the navigation registry
// This is called by RegisterRoutes() to build the navigation automatically
func registerService(config ServiceConfig) {
	registeredServices = append(registeredServices, config)
}

// GetNavigation returns navigation for the current request path
// This is the public API used by templates and handlers
func GetNavigation(currentPath string) []NavSection {
	return buildNavigation(currentPath)
}

// buildNavigation generates navigation structure from registered services
// currentPath is the current request path (e.g., "/google/calendar" or "/apple/calendar/showcase")
// This dynamically builds the navigation menu based on which services have been registered
func buildNavigation(currentPath string) []NavSection {
	var sections []NavSection

	for _, service := range registeredServices {
		var links []NavLink

		if service.HasCustom {
			customURL := fmt.Sprintf("/%s/%s", service.Platform, service.AppType)
			links = append(links, NavLink{
				Label:    "Custom",
				URL:      customURL,
				IsActive: currentPath == customURL,
			})
		}

		if service.HasShowcase {
			showcaseURL := fmt.Sprintf("/%s/%s/showcase", service.Platform, service.AppType)
			links = append(links, NavLink{
				Label:    "Showcase",
				URL:      showcaseURL,
				IsActive: currentPath == showcaseURL,
			})
		}

		if len(links) > 0 {
			sections = append(sections, NavSection{
				Title: service.Title,
				Links: links,
			})
		}
	}

	// Add Tools section at the end
	sections = append(sections, NavSection{
		Title: "Tools",
		Links: []NavLink{
			{
				Label:    "GCP OAuth Setup",
				URL:      "/tools/gcp-setup",
				IsActive: currentPath == "/tools/gcp-setup",
			},
		},
	})

	return sections
}

// GetRegisteredServices returns all registered services (for testing)
func GetRegisteredServices() []ServiceConfig {
	return registeredServices
}

// ClearRegisteredServices clears all registered services (for testing)
func ClearRegisteredServices() {
	registeredServices = nil
}
