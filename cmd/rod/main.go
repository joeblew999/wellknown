// main.go
package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

// Configuration
const (
	googleEmail = "gedw99@gmail.com" // Google account to use for authentication
	stateDir    = ".rod"              // Local directory for state (cookies, logs, etc.)
)

// State file paths
var (
	cookieFile = filepath.Join(stateDir, "cookies.json")
	logFile    = filepath.Join(stateDir, "automation.log")
	consoleLog = filepath.Join(stateDir, "console.log")
	networkLog = filepath.Join(stateDir, "network.log")
)

// Logger handles for browser logging
var (
	consoleLogger *log.Logger
	networkLogger *log.Logger
)

func main() {
	// Create state directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create state directory: %v", err)
	}

	// Set up logging to both stdout and file
	logFileHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("⚠️  Failed to open log file: %v", err)
	} else {
		defer logFileHandle.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFileHandle))
	}

	log.Println("🚀 AuthOps Automation starting...")

	// Parse CLI flags: -show to display browser GUI, -clear-cookies to delete saved session
	headless := true
	clearCookies := false
	for _, arg := range os.Args[1:] {
		if arg == "-show" {
			headless = false
			log.Println("🖥️  Running in GUI mode (-show flag detected)")
		}
		if arg == "-clear-cookies" {
			clearCookies = true
		}
	}

	// Clear cookies if requested
	if clearCookies {
		if err := os.Remove(cookieFile); err == nil {
			log.Println("🗑️  Cleared saved cookies")
		} else if !os.IsNotExist(err) {
			log.Printf("⚠️  Failed to clear cookies: %v", err)
		}
	}

	// 1️⃣ Ensure Chromium is available (auto-downloads on first run)
	chromePath := launcher.NewBrowser().MustGet()
	log.Printf("✅ Chromium ready at: %s\n", chromePath)

	// 2️⃣ Launch Chromium with flags to appear as a normal browser
	// Google blocks automation, so we need to hide automation indicators
	launch := launcher.New().
		Bin(chromePath).
		Headless(headless).
		NoSandbox(true).
		Set("disable-blink-features", "AutomationControlled"). // Hide automation
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"). // Normal Chrome UA
		MustLaunch()

	// 3️⃣ Connect Rod browser instance
	browser := rod.New().ControlURL(launch).MustConnect()
	defer browser.MustClose()

	// 4️⃣ Set up browser console logging
	consoleFileHandle, err := os.OpenFile(consoleLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("⚠️  Failed to open console log file: %v", err)
		consoleLogger = log.New(io.Discard, "", 0)
	} else {
		defer consoleFileHandle.Close()
		consoleLogger = log.New(io.MultiWriter(os.Stdout, consoleFileHandle), "[CONSOLE] ", log.LstdFlags)
	}

	// Set up network logging
	networkFileHandle, err := os.OpenFile(networkLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("⚠️  Failed to open network log file: %v", err)
		networkLogger = log.New(io.Discard, "", 0)
	} else {
		defer networkFileHandle.Close()
		networkLogger = log.New(io.MultiWriter(os.Stdout, networkFileHandle), "[NETWORK] ", log.LstdFlags)
	}

	// 5️⃣ Load persisted cookies if we have any (to skip login)
	loadCookies(browser)

	// 6️⃣ Open target page (Google OAuth credentials setup)
	page := browser.MustPage()

	// Enable browser-level logging BEFORE navigation
	setupBrowserLogging(page)

	page.MustNavigate("https://console.cloud.google.com/apis/credentials")
	page.MustWaitLoad()

	log.Println("🌐 Opened Google Cloud Credentials page successfully.")

	// If no cookies, try to help with login by filling email
	if !cookiesExist() {
		log.Println("🔐 Attempting to fill in Google email...")

		// Wait for email input field to appear (Google may redirect to login)
		time.Sleep(2 * time.Second)
		emailInput, err := page.Element(`input[type="email"]`)
		if err != nil {
			log.Printf("⚠️  Email input field not found: %v", err)
			log.Println("   Google may have already logged you in, or the page structure changed")
		} else if emailInput != nil {
			emailInput.MustInput(googleEmail)
			log.Printf("✅ Filled email: %s", googleEmail)

			// Click Next button (try multiple selectors)
			time.Sleep(1 * time.Second)
			nextBtn, err := page.Element(`#identifierNext button`)
			if err == nil && nextBtn != nil {
				nextBtn.MustClick()
				log.Println("✅ Clicked Next button")

				// Wait for password page to load
				time.Sleep(3 * time.Second)

				// Try to dismiss any passkey dialogs by pressing Escape
				log.Println("🔍 Attempting to dismiss passkey dialog with Escape key...")
				page.KeyActions().Press(input.Escape).MustDo()
				log.Println("✅ Pressed Escape to dismiss dialog")
				time.Sleep(1 * time.Second)
			}
		}
	}

	// TODO:
	// Here you can add your page.MustElement().MustClick(), .MustInput() steps
	// to automate credential creation later.

	// Wait for user to complete authentication (if needed)
	// Check if cookies were loaded - if not, wait longer for manual login
	if !cookiesExist() {
		log.Println("⏳ Waiting 120 seconds for you to complete authentication...")
		log.Println("   ⚠️  NOTE: Passkeys won't work in automated browser!")
		log.Println("   ✅ Click 'Try another way' → 'Use your password' instead")
		log.Println("   💡 TIP: Use -show flag to see the browser: make show")
		time.Sleep(120 * time.Second)
	} else {
		log.Println("⏱️  Session restored from cookies. Waiting 10 seconds...")
		time.Sleep(10 * time.Second)
	}

	// 7️⃣ Save session cookies so next run skips login
	saveCookies(browser)

	log.Println("✅ Done. Chromium closed cleanly.")
}

