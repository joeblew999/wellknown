package main

import (
	"fmt"
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// setupBrowserLogging enables capture of console logs and network requests
func setupBrowserLogging(page *rod.Page) {
	log.Println("📝 Setting up browser-level logging...")

	// Enable console API so we can capture console.log, console.error, etc.
	go page.EachEvent(func(e *proto.RuntimeConsoleAPICalled) {
		logConsoleEvent(e)
	})()

	// Enable page domain for console messages
	go page.EachEvent(func(e *proto.RuntimeExceptionThrown) {
		consoleLogger.Printf("❌ Exception: %s", e.ExceptionDetails.Text)
		if e.ExceptionDetails.Exception != nil {
			consoleLogger.Printf("   Details: %v", e.ExceptionDetails.Exception)
		}
	})()

	// Enable network logging for all requests
	go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		networkLogger.Printf("→ %s %s", e.Request.Method, e.Request.URL)
	})()

	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		networkLogger.Printf("← %d %s", e.Response.Status, e.Response.URL)
	})()

	go page.EachEvent(func(e *proto.NetworkLoadingFailed) {
		networkLogger.Printf("❌ Failed to load: %s (Error: %s)", e.RequestID, e.ErrorText)
	})()

	log.Println("✅ Browser logging enabled")
}

// logConsoleEvent formats and logs browser console events
func logConsoleEvent(e *proto.RuntimeConsoleAPICalled) {
	// Format console type (log, warn, error, etc.)
	logType := string(e.Type)

	// Extract all arguments - Rod provides Description as the string representation
	var args []string
	for _, arg := range e.Args {
		// Description contains the human-readable string representation
		if arg.Description != "" {
			args = append(args, arg.Description)
		} else {
			// Fallback: unmarshal the JSON Value
			var val interface{}
			if err := arg.Value.Unmarshal(&val); err == nil {
				args = append(args, fmt.Sprintf("%v", val))
			} else {
				// Last resort: show the type
				args = append(args, string(arg.Type))
			}
		}
	}

	// Log with appropriate emoji based on type
	var emoji string
	switch logType {
	case "log":
		emoji = "💬"
	case "warn", "warning":
		emoji = "⚠️"
	case "error":
		emoji = "❌"
	case "info":
		emoji = "ℹ️"
	case "debug":
		emoji = "🐛"
	default:
		emoji = "📝"
	}

	if len(args) > 0 {
		consoleLogger.Printf("%s [%s] %s", emoji, logType, args)
	}
}
