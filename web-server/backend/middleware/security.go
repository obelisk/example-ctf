package middleware

import (
	"net/http"
	"strconv"

	"github.com/obelisk/example-ctf/services"
	"github.com/obelisk/example-ctf/utility"
	"github.com/sirupsen/logrus"
)

const (
	requestTooLarge = "Request too large"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self';")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimitMiddleware creates a middleware that limits request body size
func RequestSizeLimitMiddleware(container *services.Container) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := services.GetLogger(ctx)

			// Check Content-Length header if present
			if contentLength := r.Header.Get("Content-Length"); contentLength != "" {
				size, err := strconv.ParseUint(contentLength, 10, 64)
				if err != nil {
					log.WithFields(logrus.Fields{
						"content_length": contentLength,
					}).Errorf("invalid content-length header")
					utility.SendJSONError(w, requestTooLarge, http.StatusBadRequest)
					return
				}

				if size > container.Config.HTTP.RequestSizeLimitBytes {
					log.WithFields(logrus.Fields{
						"content_length": size,
					}).Info("request size limit exceeded")
					utility.SendJSONError(w, requestTooLarge, http.StatusRequestEntityTooLarge)
					return
				}
			}

			// Wrap the request body with a size-limited reader
			limitedBody := http.MaxBytesReader(w, r.Body, int64(container.Config.HTTP.RequestSizeLimitBytes))
			r.Body = limitedBody

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
