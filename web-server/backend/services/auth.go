package services

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/obelisk/example-ctf/config"
	log "github.com/sirupsen/logrus"
)

// Context key types for type safety
type contextKey string

const (
	// authSessionKey stores the name of the session key which contains the session token
	authSessionKey = "session_id"

	// userContextKey stores the user data after session token validation
	userContextKey contextKey = "user"

	// authenticatedContextKey stores the user data after session token validation
	authenticatedContextKey contextKey = "authenticated"
)

// User represents an authenticated user
type User struct {
	Email string
}

// cachedPublicKey represents a cached AWS Verified Access public key with expiration
type cachedPublicKey struct {
	publicKey *ecdsa.PublicKey
	expiresAt time.Time
}

// AuthClient handles authentication operations
type AuthClient struct {
	config         *config.Config
	database       *sql.DB
	publicKeyCache map[string]cachedPublicKey
	cacheMux       sync.RWMutex
}

// NewAuthClient creates a new authentication client
func NewAuthClient(db *sql.DB, cfg *config.Config) *AuthClient {
	if cfg.Auth.TestMode != nil && cfg.Auth.TestMode.Enabled {
		log.Infoln("Auth.TestMode enabled")
	}
	return &AuthClient{
		config:         cfg,
		database:       db,
		publicKeyCache: make(map[string]cachedPublicKey),
	}
}

// GetAccessToken retrieves access token from Authorization header or cookie
func (a *AuthClient) GetAccessToken(r *http.Request) string {
	// First try the header (for direct API calls)
	authHeader := r.Header.Get("x-amzn-ava-user-context")
	if authHeader != "" {
		return authHeader
	}

	// If no header, try the cookie (for browser requests)
	cookie, err := r.Cookie("AWSVAAuthSessionCookie")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// getPublicKey retrieves an AWS Verified Access public key from cache or fetches it from AWS
func (a *AuthClient) getPublicKey(region, kid string) (*ecdsa.PublicKey, error) {
	cacheKey := fmt.Sprintf("%s:%s", region, kid)

	// Check cache first
	a.cacheMux.RLock()
	if cached, exists := a.publicKeyCache[cacheKey]; exists && time.Now().Before(cached.expiresAt) {
		a.cacheMux.RUnlock()
		log.Debugf("Using cached AWS VA public key for %s", cacheKey)
		return cached.publicKey, nil
	}
	a.cacheMux.RUnlock()

	// Fetch from AWS
	log.Debugf("Fetching public key for %s", cacheKey)
	url := fmt.Sprintf("https://public-keys.prod.verified-access.%s.amazonaws.com/%s", region, kid)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS VA public key: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch public key: HTTP %d", resp.StatusCode)
	}

	publicKeyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %v", err)
	}

	// Parse the public key
	publicKey, err := jwt.ParseECPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Cache the AWS VA public key
	cacheTTL := a.config.Auth.PublicKeyCacheTTL
	if cacheTTL == 0 {
		// Default
		cacheTTL = 1 * time.Minute
	}

	a.cacheMux.Lock()
	a.publicKeyCache[cacheKey] = cachedPublicKey{
		publicKey: publicKey,
		expiresAt: time.Now().Add(cacheTTL),
	}
	a.cacheMux.Unlock()

	log.Debugf("Cached AWS VA public key for %s (expires in %v)", cacheKey, cacheTTL)
	return publicKey, nil
}

// ValidateAccessToken validates an AWS VA-signed JWT and returns the user if valid
func (a *AuthClient) ValidateAccessToken(accessToken string) (*User, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("blank access token")
	}

	// Parse the JWT header to get key ID and signer
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT header: %v", err)
	}

	// https://docs.aws.amazon.com/verified-access/latest/ug/user-claims-passing.html#oidc-sample
	var header struct {
		Alg    string `json:"alg"`
		Kid    string `json:"kid"`
		Signer string `json:"signer"`
		Iss    string `json:"iss"`
		Exp    int64  `json:"exp"`
	}

	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse JWT header: %v", err)
	}

	// Validate signer (Verified Access instance ARN)
	expectedSigner := a.config.Auth.ExpectedVerifiedAccessInstanceARN
	if expectedSigner == "" || header.Signer != expectedSigner {
		log.WithFields(log.Fields{
			"expected_signer": expectedSigner,
			"actual_signer":   header.Signer,
		}).Debug("Invalid signer in JWT token")
		return nil, fmt.Errorf("invalid signer")
	}

	// Validate issuer
	expectedIssuer := a.config.Auth.ExpectedIssuer
	if expectedIssuer == "" || header.Iss != expectedIssuer {
		log.WithFields(log.Fields{
			"expected_issuer": expectedIssuer,
			"actual_issuer":   header.Iss,
		}).Debug("Invalid issuer in JWT token")
		return nil, fmt.Errorf("invalid issuer")
	}

	// Get public key from AWS regional endpoint
	region := a.config.Auth.AWSRegion
	if region == "" {
		return nil, fmt.Errorf("invalid AWS region")
	}

	if header.Kid == "" {
		return nil, fmt.Errorf("invalid kid")
	}

	// Get public key from cache or fetch from AWS
	publicKey, err := a.getPublicKey(region, header.Kid)
	if err != nil {
		return nil, err
	}

	// Verify and decode the JWT
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		// Verify the algorithm
		if token.Method.Alg() != "ES384" {
			return nil, fmt.Errorf("unexpected signing algorithm")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Extract user information from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Extract user email from claims
	emailClaim, ok := claims["email"].(string)
	if !ok || emailClaim == "" {
		return nil, fmt.Errorf("no email found in JWT claims")
	}

	user := &User{
		Email: emailClaim,
	}

	return user, nil
}

// SetUserContext stores user information in the request context
func (a *AuthClient) SetUserContext(ctx context.Context, userContext *User) context.Context {
	return context.WithValue(ctx, userContextKey, userContext)
}

// GetUserFromContext retrieves user information from the request context
func (a *AuthClient) GetUserFromContext(ctx context.Context) (*User, bool) {
	userContext, ok := ctx.Value(userContextKey).(*User)
	return userContext, ok
}

// SetAuthenticatedFlag stores the authentication status in the request context
func (a *AuthClient) SetAuthenticatedFlag(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, authenticatedContextKey, value)
}

// GetAuthenticatedFlag retrieves the authentication status from the request context
func (a *AuthClient) GetAuthenticatedFlag(ctx context.Context) (bool, bool) {
	authenticated, ok := ctx.Value(authenticatedContextKey).(bool)
	return authenticated, ok
}
