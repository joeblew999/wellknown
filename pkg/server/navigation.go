package server

import "fmt"

// ServiceConfig represents a registered service in the navigation
type ServiceConfig struct {
	Platform    string
	AppType     string
	Title       string
	HasCustom   bool
	HasExamples bool
}

// ServiceRegistry manages registered services (no more global state!)
type ServiceRegistry struct {
	services []ServiceConfig
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make([]ServiceConfig, 0),
	}
}

// Register adds a service to the registry
func (r *ServiceRegistry) Register(config ServiceConfig) {
	r.services = append(r.services, config)
}

// GetAll returns all registered services
func (r *ServiceRegistry) GetAll() []ServiceConfig {
	return r.services
}

// Clear removes all registered services (for testing)
func (r *ServiceRegistry) Clear() {
	r.services = nil
}

// GetNavigation returns navigation for the current request path
func (r *ServiceRegistry) GetNavigation(currentPath string) []NavSection {
	var sections []NavSection

	for _, service := range r.services {
		var links []NavLink

		if service.HasCustom {
			customURL := fmt.Sprintf("/%s/%s", service.Platform, service.AppType)
			links = append(links, NavLink{
				Label:    "Custom",
				URL:      customURL,
				IsActive: currentPath == customURL,
			})
		}

		if service.HasExamples {
			examplesURL := fmt.Sprintf("/%s/%s/examples", service.Platform, service.AppType)
			links = append(links, NavLink{
				Label:    "Examples",
				URL:      examplesURL,
				IsActive: currentPath == examplesURL,
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
