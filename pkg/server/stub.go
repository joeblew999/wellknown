package server

import (
	"log"
	"net/http"
	"strings"
)

func Stub(platform, appType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s (stub)", r.Method, r.URL.Path)

		currentPage := "custom"
		if strings.HasSuffix(r.URL.Path, "/showcase") {
			currentPage = "showcase"
		}

		Templates.ExecuteTemplate(w, "base", PageData{
			Platform:    platform,
			AppType:     appType,
			CurrentPage: currentPage,
			IsStub:      true,
			LocalURL:    LocalURL,
			MobileURL:   MobileURL,
			Navigation:  GetNavigation(r.URL.Path),
		})
	}
}

// RegisterMapsRoutes registers all Maps routes (stubs for now) with the given mux
func RegisterMapsRoutes(mux *http.ServeMux) {
	// Google Maps
	mux.HandleFunc("/google/maps", Stub("google", "maps"))
	mux.HandleFunc("/google/maps/showcase", Stub("google", "maps"))
	registerService(ServiceConfig{
		Platform:    "google",
		AppType:     "maps",
		Title:       "Google Maps",
		HasCustom:   true,
		HasShowcase: true,
	})

	// Apple Maps
	mux.HandleFunc("/apple/maps", Stub("apple", "maps"))
	mux.HandleFunc("/apple/maps/showcase", Stub("apple", "maps"))
	registerService(ServiceConfig{
		Platform:    "apple",
		AppType:     "maps",
		Title:       "Apple Maps",
		HasCustom:   true,
		HasShowcase: true,
	})
}
