package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/obelisk/example-ctf/config"

	"github.com/TwiN/go-away"
)

// ClientError represents an error that should be shown to the client
type ClientError struct {
	Message string
}

func (e ClientError) Error() string {
	return e.Message
}

// IsClientError checks if an error is meant for client display
func IsClientError(err error) bool {
	var clientErr ClientError
	return errors.As(err, &clientErr)
}

// UserProfile represents a user's profile information
type UserProfile struct {
	UserEmail                        string    `json:"user_email"`
	Alias                            string    `json:"alias"`
	Tokens                           int       `json:"tokens"`
	Points                           int       `json:"points"`
	ExamChallengesSolved             int       `json:"exam_challenges_solved"`
	LastExamChallengeSolvedTimestamp time.Time `json:"last_exam_challenge_solved_timestamp"`
	LastChallengeSolvedTimestamp     time.Time `json:"last_challenge_solved_timestamp"`
}

// UserClient handles user-related operations
type UserClient struct {
	db     *sql.DB
	config *config.Config
	cache  map[string]*UserProfile
	mutex  sync.RWMutex
}

// NewUserClient creates a new user client
func NewUserClient(db *sql.DB, config *config.Config) *UserClient {
	return &UserClient{
		db:     db,
		config: config,
		cache:  make(map[string]*UserProfile),
	}
}

// GetUserProfile retrieves a user's profile with caching
func (uc *UserClient) GetUserProfile(ctx context.Context, slackService *SlackService, user *User) (*UserProfile, error) {
	userEmail := user.Email

	// Check cache first
	uc.mutex.RLock()
	if profile, exists := uc.cache[userEmail]; exists {
		uc.mutex.RUnlock()
		return profile, nil
	}
	uc.mutex.RUnlock()

	// Cache miss, fetch from database
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	// Double-check cache after acquiring lock
	if profile, exists := uc.cache[userEmail]; exists {
		return profile, nil
	}

	_, alias, tokens, points, examChallengesSolved, lastExamTimestamp, lastChallengeTimestamp, err := uc.getUserProfileData(userEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile data: %w", err)
	}

	profile := &UserProfile{
		UserEmail:                        userEmail,
		Alias:                            alias,
		Tokens:                           tokens,
		Points:                           points,
		ExamChallengesSolved:             examChallengesSolved,
		LastExamChallengeSolvedTimestamp: lastExamTimestamp,
		LastChallengeSolvedTimestamp:     lastChallengeTimestamp,
	}

	uc.cache[userEmail] = profile

	return profile, nil
}

// getUserProfileData fetches alias, tokens, points, and timestamp data for a user
// If the user doesn't exist, creates a new user row with default values
func (uc *UserClient) getUserProfileData(userEmail string) (bool, string, int, int, int, time.Time, time.Time, error) {
	var alias string
	var tokens, points, examChallengesSolved int
	var lastExamTimestamp, lastChallengeTimestamp time.Time

	query := `
		SELECT COALESCE(ua.alias, '') as alias,
		       COALESCE(u.tokens_available, 0), COALESCE(u.points_achieved, 0), 
		       COALESCE(u.exam_challenges_solved, 0), COALESCE(u.last_exam_challenge_solved_timestamp, NOW()),
		       COALESCE(u.last_challenge_solved_timestamp, NOW())
		FROM users u
		LEFT JOIN user_aliases ua ON u.user_email = ua.user_email AND ua.deleted_at IS NULL
		WHERE u.user_email = $1
	`

	err := uc.db.QueryRow(query, userEmail).Scan(&alias, &tokens, &points, &examChallengesSolved, &lastExamTimestamp, &lastChallengeTimestamp)
	if err == sql.ErrNoRows {
		// User doesn't exist, create them with default values
		_, err := uc.db.Exec(`
			INSERT INTO users (user_email, tokens_available, tokens_burned, points_achieved, exam_challenges_solved, last_exam_challenge_solved_timestamp, last_challenge_solved_timestamp) 
			VALUES ($1, 0, 0, 0, 0, NOW(), NOW())
		`, userEmail)
		if err != nil {
			return false, "", 0, 0, 0, time.Time{}, time.Time{}, fmt.Errorf("failed to create new user: %w", err)
		}
		// Return default values for newly created user
		return true, "", 0, 0, 0, time.Now(), time.Now(), nil
	}
	if err != nil {
		return false, "", 0, 0, 0, time.Time{}, time.Time{}, fmt.Errorf("failed to query user profile data: %w", err)
	}

	return false, alias, tokens, points, examChallengesSolved, lastExamTimestamp, lastChallengeTimestamp, nil
}

// SetAlias sets or updates a user's alias
func (uc *UserClient) SetAlias(ctx context.Context, userEmail string, alias string) error {
	// Validate alias input
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ClientError{Message: "Alias cannot be empty"}
	}

	// Check for reasonable length limits
	if len(alias) > 24 {
		return ClientError{Message: "Alias cannot be longer than 24 characters"}
	}

	// Validate alias format: only alphanumeric and [_-.] characters
	validAliasRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !validAliasRegex.MatchString(alias) {
		return ClientError{Message: "Alias can only contain letters, numbers, underscores, hyphens, and periods"}
	}

	if goaway.IsProfane(alias) {
		return ClientError{Message: fmt.Sprintf("Alias not allowed: %s", alias)}
	}

	// Start transaction for atomic operation
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check daily limit - only allow 1 alias set per day
	var lastSetTime sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT created_at 
		FROM user_aliases 
		WHERE user_email = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, userEmail).Scan(&lastSetTime)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check alias history: %w", err)
	}

	// If user has set an alias within the last 24 hours, deny the request
	if lastSetTime.Valid {
		timeSinceLastSet := time.Since(lastSetTime.Time)
		if timeSinceLastSet < 24*time.Hour {
			hoursRemaining := 24 - int(timeSinceLastSet.Hours())
			return ClientError{Message: fmt.Sprintf("You can only set an alias once per day. Try again in %d hours", hoursRemaining)}
		}
	}

	// Insert or update the alias with soft deletion support
	query := `
		INSERT INTO user_aliases (user_email, alias, created_at, deleted_at) 
		VALUES ($1, $2, NOW(), NULL)
		ON CONFLICT (user_email) 
		DO UPDATE SET 
			alias = EXCLUDED.alias, 
			created_at = NOW(),
			deleted_at = NULL
	`

	_, err = tx.ExecContext(ctx, query, userEmail, alias)
	if err != nil {
		// Check if this is a unique constraint violation on the alias column
		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ClientError{Message: "Alias already taken"}
		}
		// Internal database error - don't expose to client
		return fmt.Errorf("failed to set alias: %w", err)
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Set alias to '%s'", alias))
	if err != nil {
		return fmt.Errorf("failed to log alias change: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache since alias has changed
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	return nil
}

// RemoveAlias removes a user's alias using soft deletion
func (uc *UserClient) RemoveAlias(ctx context.Context, userEmail string) error {
	// Start transaction for atomic operation
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, get the current active alias for logging purposes
	var currentAlias string
	err = tx.QueryRowContext(ctx, `
		SELECT alias 
		FROM user_aliases 
		WHERE user_email = $1 AND deleted_at IS NULL
	`, userEmail).Scan(&currentAlias)
	if err != nil {
		if err == sql.ErrNoRows {
			return ClientError{Message: "No alias found to remove"}
		}
		return fmt.Errorf("failed to check existing alias: %w", err)
	}

	// Soft delete the alias by setting deleted_at timestamp
	query := `
		UPDATE user_aliases 
		SET deleted_at = NOW() 
		WHERE user_email = $1 AND deleted_at IS NULL
	`
	result, err := tx.ExecContext(ctx, query, userEmail)
	if err != nil {
		return fmt.Errorf("failed to remove alias: %w", err)
	}

	// Check if any rows were affected (should be 1 since we already confirmed it exists)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ClientError{Message: "No alias found to remove"}
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Removed alias '%s'", currentAlias))
	if err != nil {
		return fmt.Errorf("failed to log alias removal: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache since alias has been removed
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	return nil
}

// CompleteExamChallenge completes an exam challenge for a user
func (uc *UserClient) CompleteExamChallenge(ctx context.Context, userEmail string, challengeID int) error {
	// Start transaction
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Add token and increment exam challenges solved
	query := `
		INSERT INTO users (user_email, tokens_available, tokens_burned, points_achieved, exam_challenges_solved, last_exam_challenge_solved_timestamp, last_challenge_solved_timestamp) 
		VALUES ($1, 1, 0, 0, 1, NOW(), NOW())
		ON CONFLICT (user_email) 
		DO UPDATE SET 
			tokens_available = users.tokens_available + 1,
			exam_challenges_solved = users.exam_challenges_solved + 1,
			last_exam_challenge_solved_timestamp = NOW()
	`

	_, err = tx.ExecContext(ctx, query, userEmail)
	if err != nil {
		return fmt.Errorf("failed to complete exam challenge: %w", err)
	}

	// Record challenge completion
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) 
		VALUES ($1, $2, NOW())
	`, userEmail, challengeID)
	if err != nil {
		return fmt.Errorf("failed to record challenge completion: %w", err)
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Completed exam challenge %d: added 1 token", challengeID))
	if err != nil {
		return fmt.Errorf("failed to log exam challenge completion: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	return nil
}

// CompleteChallenge adds 1 token and points to a user's account in a single transaction
func (uc *UserClient) CompleteChallenge(ctx context.Context, userEmail string, pointAmount int, challengeID int) error {
	// Start transaction
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Add token and points in a single query
	query := `
		INSERT INTO users (user_email, tokens_available, tokens_burned, points_achieved, exam_challenges_solved, last_exam_challenge_solved_timestamp, last_challenge_solved_timestamp) 
		VALUES ($1, 1, 0, $2, 0, NOW(), NOW())
		ON CONFLICT (user_email) 
		DO UPDATE SET 
			tokens_available = users.tokens_available + 1,
			points_achieved = users.points_achieved + $2,
			last_challenge_solved_timestamp = NOW()
	`

	_, err = tx.ExecContext(ctx, query, userEmail, pointAmount)
	if err != nil {
		return fmt.Errorf("failed to add token and points to user: %w", err)
	}

	// Record challenge completion
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) 
		VALUES ($1, $2, NOW())
	`, userEmail, challengeID)
	if err != nil {
		return fmt.Errorf("failed to record challenge completion: %w", err)
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Completed challenge %d: added 1 token and %d points", challengeID, pointAmount))
	if err != nil {
		return fmt.Errorf("failed to log challenge completion: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	return nil
}

// BurnToken burns 1 token from a user's available balance
// Returns the number of tokens successfully burned (0 if not enough tokens available)
func (uc *UserClient) BurnToken(ctx context.Context, userEmail string, challengeID int) (int, error) {
	// Start transaction
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Atomic check and burn in a single query
	query := `
		UPDATE users 
		SET tokens_available = tokens_available - 1,
		    tokens_burned = tokens_burned + 1
		WHERE user_email = $1 AND tokens_available >= 1
		RETURNING tokens_available
	`

	var newAvailableTokens int
	err = tx.QueryRowContext(ctx, query, userEmail).Scan(&newAvailableTokens)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows affected means either user doesn't exist or not enough tokens
			return 0, nil
		}
		return 0, fmt.Errorf("failed to burn token from user: %w", err)
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Burned 1 token for exam challenge %d", challengeID))
	if err != nil {
		return 0, fmt.Errorf("failed to log token burn: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	// If the update succeeded, we burned 1 token
	return 1, nil
}

// RefundToken refunds 1 token by decreasing tokens_burned and increasing tokens_available
func (uc *UserClient) RefundToken(ctx context.Context, userEmail string, challengeID int) error {
	// Start transaction
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Refund token
	query := `
		UPDATE users 
		SET tokens_available = tokens_available + 1,
		    tokens_burned = tokens_burned - 1
		WHERE user_email = $1 AND tokens_burned >= 1
	`

	result, err := tx.ExecContext(ctx, query, userEmail)
	if err != nil {
		return fmt.Errorf("failed to refund token for user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for token refund: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no tokens to refund for user %s", userEmail)
	}

	// Log to user history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_history_log (user_email, log, date) 
		VALUES ($1, $2, NOW())
	`, userEmail, fmt.Sprintf("Refunded 1 token for exam challenge %d", challengeID))
	if err != nil {
		return fmt.Errorf("failed to log token refund: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	uc.mutex.Lock()
	delete(uc.cache, userEmail)
	uc.mutex.Unlock()

	return nil
}
