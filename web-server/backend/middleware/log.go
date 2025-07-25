package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/obelisk/example-ctf/services"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// LoggingMiddleware adds logging functionality to HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		// Create logger entry with request context
		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"request_id": requestID,
			"method":     r.Method,
			"path":       r.URL.Path,
			"remote_ip":  getClientIP(r),
		})

		// Ensure the logger respects the global log level
		logger.Logger.SetLevel(log.GetLevel())
		ctx := context.WithValue(r.Context(), services.LoggerContextKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
