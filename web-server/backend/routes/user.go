package routes

import (
	"encoding/json"
	"net/http"

	"github.com/obelisk/example-ctf/services"
	"github.com/obelisk/example-ctf/utility"
)

// SetAliasRequest represents the request body for setting an alias
type SetAliasRequest struct {
	Alias string `json:"alias"`
}

// SetAlias handles setting a user's alias
func SetAlias(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var req SetAliasRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set the alias
		err := container.UserClient.SetAlias(ctx, user.Email, req.Alias)
		if err != nil {
			if services.IsClientError(err) {
				// Show user-friendly error to client
				log.Errorf("alias denied: %v", err)
				utility.SendJSONError(w, err.Error(), http.StatusBadRequest)
			} else {
				// Log internal error, show generic message to client
				log.Errorf("Internal error setting alias: %v", err)
				utility.SendJSONError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Success response
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message": "Alias set successfully",
		}); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})
}

// RemoveAlias handles removing a user's alias
func RemoveAlias(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Remove the alias
		err := container.UserClient.RemoveAlias(ctx, user.Email)
		if err != nil {
			if services.IsClientError(err) {
				// Show user-friendly error to client
				log.Errorf("remove alias: %v", err)
				utility.SendJSONError(w, err.Error(), http.StatusBadRequest)
			} else {
				// Internal server error
				log.Errorf("Internal error removing alias: %v", err)
				utility.SendJSONError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Success response
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message": "Alias removed successfully",
		}); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})
}
