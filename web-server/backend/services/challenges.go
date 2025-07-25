package services

import (
	"crypto/ed25519"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/obelisk/example-ctf/config"
	"github.com/obelisk/example-ctf/utility"
	"golang.org/x/crypto/sha3"
)

// Challenge represents a challenge in the list view
type Challenge struct {
	ID                int    `json:"id"`
	NestedID          int    `json:"nested_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Category          string `json:"category"`
	PointRewardAmount int    `json:"point_reward_amount"`
	Completed         bool   `json:"completed"`
}

// DetailedChallenge represents a challenge with full details
type DetailedChallenge struct {
	ID          int    `json:"id"`
	NestedID    int    `json:"nested_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	// PointRewardAmount is the number of points awarded for completing this challenge
	PointRewardAmount int     `json:"point_reward_amount"`
	FileAsset         *string `json:"file_asset,omitempty"`
	TextAsset         *string `json:"text_asset,omitempty"`
	Completed         bool    `json:"completed"`
}

// ChallengeClient handles challenge-related operations
type ChallengeClient struct {
	config   *config.Config
	database *sql.DB
}

// NewChallengeClient creates a new challenge client
func NewChallengeClient(db *sql.DB, cfg *config.Config) *ChallengeClient {
	return &ChallengeClient{
		config:   cfg,
		database: db,
	}
}

// CheckUserCompletedChallenge checks if a user has completed a specific challenge
// Returns (completed, completed_at, err)
func (cc *ChallengeClient) CheckUserCompletedChallenge(userEmail string, challengeID int) (bool, time.Time, error) {
	var completedAt time.Time
	err := cc.database.QueryRow(`
		SELECT completed_at 
		FROM user_challenges_completed 
		WHERE user_email = $1 AND challenge_id = $2
	`, userEmail, challengeID).Scan(&completedAt)

	if err == nil {
		// User has already completed this challenge
		return true, completedAt, err
	}

	if err == sql.ErrNoRows {
		// User has not completed this challenge
		return false, time.Unix(0, 0), nil
	}

	return false, time.Unix(0, 0), err
}

// GetChallengeFlagAndReward retrieves the flag, reward, name, and category for a challenge
// Returns (flag, point_reward, validation_handler, challenge_name, category, err)
func (cc *ChallengeClient) GetChallengeFlagAndReward(challengeID int) (string, int, string, string, string, error) {
	var flagValue string
	var pointRewardAmount int
	var validationHandler string
	var challengeName string
	var category string
	err := cc.database.QueryRow(`
		SELECT f.flag_value, c.point_reward_amount, f.validation_handler, c.name, c.category
		FROM challenges c
		JOIN flags f ON c.id = f.challenge_id
		WHERE c.id = $1
	`, challengeID).Scan(&flagValue, &pointRewardAmount, &validationHandler, &challengeName, &category)

	if err == nil {
		return flagValue, pointRewardAmount, validationHandler, challengeName, category, err
	}

	return "", 0, "", "", "", err
}

// ValidateFlag validates a submitted flag based on the validation handler type
// Returns (isValid, customErrorMessage). If customErrorMessage is empty, use default "Incorrect flag" message.
func (cc *ChallengeClient) ValidateFlag(submittedFlag, expectedFlagValue, validationHandler, userEmail string) (bool, string) {
	switch validationHandler {
	case "StringEqual":
		return utility.ConstantTimeStringEqual(strings.ToLower(submittedFlag), strings.ToLower(expectedFlagValue)), ""
	case "Md5HashOfUsername":
		return cc.validateHashOfUsername(submittedFlag, expectedFlagValue, "md5", userEmail)
	case "Sha1HashOfUsername":
		return cc.validateHashOfUsername(submittedFlag, expectedFlagValue, "sha1", userEmail)
	case "Sha256HashOfUsername":
		return cc.validateHashOfUsername(submittedFlag, expectedFlagValue, "sha256", userEmail)
	case "Keccak256HashOfUsername":
		return cc.validateHashOfUsername(submittedFlag, expectedFlagValue, "keccak256", userEmail)
	case "ClientSideWasm":
		return cc.validateClientSideWasm(submittedFlag, expectedFlagValue, userEmail), ""
	default:
		// Default to string comparison (case-insensitive)
		return utility.ConstantTimeStringEqual(strings.ToLower(submittedFlag), strings.ToLower(expectedFlagValue)), ""
	}
}

// validateHashOfUsername validates hash-based PoW and returns custom messages
func (cc *ChallengeClient) validateHashOfUsername(submittedFlag, difficulty, hashType, userEmail string) (bool, string) {
	// Parse difficulty as integer
	requiredZeros, err := strconv.Atoi(difficulty)
	if err != nil || requiredZeros < 0 {
		return false, ""
	}

	// Check that the submitted flag contains the user email
	if !strings.Contains(submittedFlag, userEmail) {
		return false, ""
	}

	var hashHex string
	switch hashType {
	case "md5":
		hash := md5.Sum([]byte(submittedFlag))
		hashHex = fmt.Sprintf("%x", hash)
	case "sha1":
		hash := sha1.Sum([]byte(submittedFlag))
		hashHex = fmt.Sprintf("%x", hash)
	case "sha256":
		hash := sha256.Sum256([]byte(submittedFlag))
		hashHex = fmt.Sprintf("%x", hash)
	case "keccak256":
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write([]byte(submittedFlag))
		hash := hasher.Sum(nil)
		hashHex = fmt.Sprintf("%x", hash)
	default:
		return false, ""
	}

	// Check if the hash starts with the required number of leading zeros
	if requiredZeros == 0 {
		return true, ""
	}

	// Create a prefix string with the required number of zeros
	requiredPrefix := strings.Repeat("0", requiredZeros)

	// First, check if the difficulty requirement is met (correctness check first)
	if strings.HasPrefix(hashHex, requiredPrefix) {
		return true, ""
	}

	// Count actual leading zeros in the hash
	actualZeros := len(hashHex) - len(strings.TrimLeft(hashHex, "0"))

	// Custom messages based on progress toward the goal
	if actualZeros >= 2 {
		// Calculate 75% threshold of required zeros
		threshold75 := int(float64(requiredZeros) * 0.75)
		threshold50 := int(float64(requiredZeros) * 0.50)

		if actualZeros >= threshold75 {
			return false, "Need a little more work"
		} else if actualZeros >= threshold50 {
			return false, "NEED EVEN MORE WORK"
		}
		return false, "Need more work"
	}

	// Default case: less than 2 zeros, return empty message (use default "Incorrect flag")
	return false, ""
}

// validateClientSideWasm validates a JWT token signed with an Ed25519 key and checks for admin claims
func (cc *ChallengeClient) validateClientSideWasm(submittedFlag, publicKeyPEM, userEmail string) bool {
	// Parse the Ed25519 public key from PEM format
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false
	}

	// Parse the public key
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	// Ensure it's an Ed25519 public key
	publicKey, ok := publicKeyInterface.(ed25519.PublicKey)
	if !ok {
		return false
	}

	// Parse and validate the JWT
	token, err := jwt.ParseWithClaims(submittedFlag, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is EdDSA
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return false
	}

	// Check if token is valid and extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verify that is_admin is true
		if isAdmin, exists := claims["is_admin"]; exists {
			if adminBool, ok := isAdmin.(bool); ok {
				return adminBool
			}
		}
	}

	return false
}

// SanitizeFlag cleans and validates flag input
// Error messages from this function can be returned to the client
func (cc *ChallengeClient) SanitizeFlag(flag string) (string, error) {
	// Remove leading/trailing whitespace
	flag = strings.TrimSpace(flag)

	// Check for empty flag
	if flag == "" {
		return "", fmt.Errorf("Flag cannot be empty")
	}

	// No need for length check, it's handled by the body request size limit
	// Check that flag contains only ASCII printable characters (32-126)
	for _, r := range flag {
		if r < 32 || r > 126 {
			return "", fmt.Errorf("Only ASCII printable characters are allowed")
		}
	}

	return flag, nil
}
