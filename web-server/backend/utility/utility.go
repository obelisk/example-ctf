package utility

import (
	"encoding/json"
	"net/http"
)

// ConstantTimeStringEqual performs a constant-time comparison of two strings
// to prevent timing attacks
func ConstantTimeStringEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte = 0
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}

	return diff == 0
}

// SendJSONError sends a JSON error response
func SendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
