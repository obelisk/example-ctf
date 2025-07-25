package middleware

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/obelisk/example-ctf/services"
	"github.com/obelisk/example-ctf/utility"
)

const unauthorized = "Unauthorized"

// LoadAuthenticatedUser loads the authenticated user into the request context
func LoadAuthenticatedUser(container *services.Container) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := services.GetLogger(ctx)

			// If test mode is enabled, set the user context to the test email
			if container.Config.Auth.TestMode != nil && container.Config.Auth.TestMode.Enabled && container.Config.Auth.TestMode.TestUser != "" {
				ctx = container.Auth.SetAuthenticatedFlag(ctx, true)
				ctx = container.Auth.SetUserContext(ctx, &services.User{Email: container.Config.Auth.TestMode.TestUser})
				log.Infoln("test mode: user authenticated")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			accessToken := container.Auth.GetAccessToken(r)
			user, err := container.Auth.ValidateAccessToken(accessToken)

			if err != nil {
				ctx = container.Auth.SetAuthenticatedFlag(ctx, false)
				log.Debugf("user unauthenticated: %v", err)
			} else {
				ctx = container.Auth.SetAuthenticatedFlag(ctx, true)
				ctx = container.Auth.SetUserContext(ctx, user)
				logger := log.WithFields(logrus.Fields{
					"user": user.Email,
				})
				logger.Debugln("user authenticated")
				ctx = context.WithValue(ctx, services.LoggerContextKey, logger)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthenticated middleware ensures the user is authenticated
func RequireAuthenticated(container *services.Container) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := services.GetLogger(ctx)

			// Check if user is authenticated
			authenticated, ok := container.Auth.GetAuthenticatedFlag(ctx)
			if !ok || !authenticated {
				log.Errorf("unauthorized access attempt")
				utility.SendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			// Call the next middleware function or final handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireUnauthenticated middleware ensures the user is not authenticated
func RequireUnauthenticated(container *services.Container) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// log := services.GetLogger(ctx)

			// Check if user is authenticated
			authenticated, ok := container.Auth.GetAuthenticatedFlag(ctx)
			if ok && authenticated {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// Call the next middleware function or final handler
			next.ServeHTTP(w, r)
		})
	}
}
