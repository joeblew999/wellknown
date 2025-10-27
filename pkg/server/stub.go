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
