package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Sentry reports panics and request context to Sentry when SENTRY_DSN is configured.
func Sentry(next http.Handler) http.Handler {
	if sentry.CurrentHub().Client() == nil {
		return next
	}
	handler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: false,
	})
	return handler.Handle(next)
}
