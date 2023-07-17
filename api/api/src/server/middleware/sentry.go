package middleware

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// InitSentry initializes and returns a Gin compatible Sentry middleware.
func Sentry() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{
		// By default SentryGin catches panics and returns a 200 anyway.
		// Avoid that behaviour, otherwise the healthcheck endpoint stops working.
		Repanic: true,

		// Do not block a request while waiting for an event to be sent to sentry.
		WaitForDelivery: false,
	})
}
