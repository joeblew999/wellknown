package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// saveCookies serializes cookies for future reuse
func saveCookies(browser *rod.Browser) {
	cookies, err := browser.GetCookies()
	if err != nil {
		log.Printf("⚠️ Unable to get cookies: %v", err)
		return
	}
	f, err := os.Create(cookieFile)
	if err != nil {
		log.Printf("⚠️ Unable to create cookie file: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(cookies)
	log.Printf("💾 Cookies saved to %s\n", cookieFile)
}

// loadCookies restores session cookies from previous run
func loadCookies(browser *rod.Browser) {
	f, err := os.Open(cookieFile)
	if err != nil {
		log.Printf("ℹ️ No previous cookies found (%v). New login may be required.", err)
		return
	}
	defer f.Close()

	var cookies []*proto.NetworkCookie
	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		log.Printf("⚠️ Error decoding cookie file: %v", err)
		return
	}

	browser.MustSetCookies(cookies...)
	log.Printf("🍪 Cookies loaded from %s", cookieFile)
}

// cookiesExist checks if we have saved cookies from a previous run
func cookiesExist() bool {
	_, err := os.Stat(cookieFile)
	return err == nil
}
