package config

import (
	"log"

	"github.com/getsentry/sentry-go"
)

// InitSentry initializes Sentry error reporting
func InitSentry(dsn string) {
	if dsn == "" {
		log.Println("SENTRY_DSN not set, skipping Sentry initialization")
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      "production",
		TracesSampleRate: 0.1, // Sample 10% of transactions for performance monitoring
	})
	if err != nil {
		log.Printf("Failed to initialize Sentry: %v", err)
		return
	}

	log.Println("Sentry initialized successfully")
}

// CaptureError captures an error and sends it to Sentry
func CaptureError(err error) {
	if err != nil {
		sentry.CaptureException(err)
	}
}

// CaptureMessage captures a message and sends it to Sentry
func CaptureMessage(msg string) {
	sentry.CaptureMessage(msg)
}

// FlushSentry flushes any pending events to Sentry
func FlushSentry() {
	sentry.Flush(2 * 1000) // 2 second timeout
}
