package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/obelisk/example-ctf/config"
	"github.com/obelisk/example-ctf/services"
	"github.com/obelisk/example-ctf/utility"
	"github.com/sirupsen/logrus"
)

// ClientInfo stores rate limiting information for a client
type ClientInfo struct {
	// Amount of requests left
	Tokens float64
	// Last time tokens were topped up
	LastRefill  time.Time
	BurstTokens int
	LastRequest time.Time
}

// RateLimiter implements a configurable token bucket rate limiter with LRU cache
type RateLimiter struct {
	config *config.RateLimitConfig
	cache  *lru.Cache[string, *ClientInfo]
	done   chan bool
}

const internalError = "Internal Error"

// NewRateLimiter creates a new rate limiter with the specified configuration
func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	// Create LRU cache with the specified capacity
	cache, err := lru.New[string, *ClientInfo](cfg.MaxClients)
	if err != nil {
		// This should never happen with valid capacity, but handle it gracefully
		panic("failed to create LRU cache: " + err.Error())
	}

	rl := &RateLimiter{
		config: cfg,
		cache:  cache,
		done:   make(chan bool),
	}

	return rl
}

// getClientIP extracts the real client IP from the request, handling proxy headers
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Check for X-Real-IP header
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	// Check for Cloudflare header
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}
	// Check for X-Client-IP header
	if clientIP := r.Header.Get("X-Client-IP"); clientIP != "" {
		return strings.TrimSpace(clientIP)
	}
	// Fallback: remove port from RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// getUserID extracts the user ID from the request context
func getUserID(a *services.AuthClient, r *http.Request) string {
	ctx := r.Context()
	user, ok := a.GetUserFromContext(ctx)
	if !ok {
		return ""
	}
	return user.Email
}

// refillTokens refills tokens for a client based on time elapsed
func (rl *RateLimiter) refillTokens(clientKey string) {
	now := time.Now()
	client, exists := rl.cache.Get(clientKey)

	if !exists {
		// First request from this client
		client = &ClientInfo{
			Tokens:      rl.config.RequestsPerSec,
			LastRefill:  now,
			BurstTokens: rl.config.BurstSize,
			LastRequest: now,
		}
		rl.cache.Add(clientKey, client)
		return
	}

	// Calculate time elapsed since last refill
	elapsed := now.Sub(client.LastRefill).Seconds()

	// Refill tokens based on elapsed time
	tokensToAdd := elapsed * rl.config.RequestsPerSec
	client.Tokens += tokensToAdd

	// Cap tokens at the burst size
	maxTokens := float64(rl.config.BurstSize)
	if client.Tokens > maxTokens {
		client.Tokens = maxTokens
	}

	client.LastRefill = now
	client.LastRequest = now

	// Update the cache (this will move the item to front)
	rl.cache.Add(clientKey, client)
}

// consumeToken attempts to consume a token for the client
// Returns true if a token was consumed successfully
func (rl *RateLimiter) consumeToken(clientKey string) bool {
	client, exists := rl.cache.Get(clientKey)
	if !exists || client.Tokens < 1.0 {
		return false
	}

	client.Tokens -= 1.0

	// Update the cache (this will move the item to front)
	rl.cache.Add(clientKey, client)

	return true
}

// RateLimitMiddleware creates a rate limiting middleware using the configuration
func RateLimitMiddleware(container *services.Container) func(http.Handler) http.Handler {
	// If rate limiting is disabled, return a no-op middleware
	if !container.Config.HTTP.RateLimit.Enabled {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
	}

	limiter := NewRateLimiter(&container.Config.HTTP.RateLimit)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := services.GetLogger(ctx)

			clientIP := getClientIP(r)

			// Refill tokens for this client
			limiter.refillTokens(clientIP)

			// Try to consume a token
			ok := limiter.consumeToken(clientIP)

			if !ok {
				log.WithFields(logrus.Fields{
					"client_ip": clientIP,
				}).Info("IP rate limit exceeded")
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded",
				}); err != nil {
					log.Errorf("write error: %v", err)
					utility.SendJSONError(w, internalError, http.StatusInternalServerError)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserRateLimitMiddleware creates a user-based rate limiting middleware
func UserRateLimitMiddleware(container *services.Container) func(http.Handler) http.Handler {
	// If rate limiting is disabled, return a no-op middleware
	if !container.Config.HTTP.RateLimit.Enabled {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
	}

	limiter := NewRateLimiter(&container.Config.HTTP.RateLimit)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := services.GetLogger(ctx)

			userID := getUserID(container.Auth, r)

			// If no user ID (unauthenticated), fall back to IP-based rate limiting
			if userID == "" {
				clientIP := getClientIP(r)
				userID = "ip:" + clientIP
			}

			// Refill tokens for this user
			limiter.refillTokens(userID)

			// Try to consume a token
			ok := limiter.consumeToken(userID)

			if !ok {
				log.WithFields(logrus.Fields{
					"user_id": userID,
				}).Info("user rate limit exceeded")
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded",
				}); err != nil {
					log.Errorf("write error: %v", err)
					utility.SendJSONError(w, internalError, http.StatusInternalServerError)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
