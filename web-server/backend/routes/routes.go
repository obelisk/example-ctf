package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"

	"github.com/obelisk/example-ctf/services"
	"github.com/obelisk/example-ctf/utility"
)

const internalError = "Internal Error"
const notFoundError = "Not Found"
const invalidRequestError = "Invalid Request"

// validateChallengeID validates and sanitizes challenge ID input
func validateChallengeID(id string) (int, error) {
	// Remove any whitespace
	id = strings.TrimSpace(id)

	// Parse as integer
	challengeID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid challenge ID format")
	}

	// Validate range (reasonable bounds)
	if challengeID <= 0 || challengeID > 128 {
		return 0, fmt.Errorf("challenge ID out of valid range")
	}

	return challengeID, nil
}

// HealthCheckHandler checks if the services are online, including the DB
func HealthCheckHandler(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Check database connection
		err := container.DB.Ping()
		if err != nil {
			log.Errorf("Health check failed: Database ping failed: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Errorf("write error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// GetProfile returns the current user's profile
func GetProfile(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		log.Info("user requested profile")

		// Get user profile with caching
		profile, err := container.UserClient.GetUserProfile(ctx, container.SlackService, user)
		if err != nil {
			log.Errorf("unable to get user profile: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// ListChallenges returns a list of all available challenges
func ListChallenges(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		log.Info("user requested challenge list")

		// Query challenges with completion status for the user
		rows, err := container.DB.Query(`
			SELECT c.id, c.nested_id, c.name, c.description, c.category, c.point_reward_amount,
			       CASE WHEN ucc.challenge_id IS NOT NULL THEN true ELSE false END as completed
			FROM challenges c
			LEFT JOIN user_challenges_completed ucc ON c.id = ucc.challenge_id AND ucc.user_email = $1
			WHERE c.category != 'exam'
			ORDER BY c.id
		`, user.Email)

		if err != nil {
			log.Errorf("unable to query challenges from database: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		challenges := make([]services.Challenge, 0)
		for rows.Next() {
			var challenge services.Challenge
			if err := rows.Scan(&challenge.ID, &challenge.NestedID, &challenge.Name, &challenge.Description, &challenge.Category, &challenge.PointRewardAmount, &challenge.Completed); err != nil {
				log.Errorf("unable to scan rows queried from database: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
				return
			}
			challenges = append(challenges, challenge)
		}

		if err := rows.Err(); err != nil {
			log.Errorf("rows error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(challenges); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// GetChallenge returns detailed information about a specific challenge
func GetChallenge(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Get challenge ID to fetch and validate
		id := mux.Vars(r)["id"]
		challengeID, err := validateChallengeID(id)
		if err != nil {
			log.Errorf("invalid challenge ID: %v", err)
			utility.SendJSONError(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		log = log.WithFields(logrus.Fields{
			"challenge_id": challengeID,
		})
		log.Info("user requested challenge details")

		var challenge services.DetailedChallenge
		err = container.DB.QueryRow(`
			SELECT c.id, c.nested_id, c.name, c.description, c.category, c.point_reward_amount, c.file_asset, c.text_asset,
			       CASE WHEN ucc.challenge_id IS NOT NULL THEN true ELSE false END as completed
			FROM challenges c
			LEFT JOIN user_challenges_completed ucc ON c.id = ucc.challenge_id AND ucc.user_email = $2
			WHERE c.id = $1 AND c.category != 'exam'
			`, challengeID, user.Email).Scan(
			&challenge.ID,
			&challenge.NestedID,
			&challenge.Name,
			&challenge.Description,
			&challenge.Category,
			&challenge.PointRewardAmount,
			&challenge.FileAsset,
			&challenge.TextAsset,
			&challenge.Completed,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, notFoundError, http.StatusNotFound)
			} else {
				log.Errorf("unable to query challenge from database: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Treat exam challenges as if they don't exist
		if challenge.Category == "exam" {
			http.Error(w, notFoundError, http.StatusNotFound)
			return
		}

		// Transform file_asset to presigned URL if not empty
		if challenge.FileAsset != nil && *challenge.FileAsset != "" {
			presignedURL, err := container.AssetService.GetAsset(ctx, *challenge.FileAsset)
			if err != nil {
				log.Errorf("failed to get presigned URL for asset %s: %v", challenge.FileAsset, err)
				http.Error(w, internalError, http.StatusInternalServerError)
				return
			}
			challenge.FileAsset = &presignedURL
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(challenge); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// SubmitChallenge handles challenge flag submissions
func SubmitChallenge(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Get challenge ID to submit flag for and validate
		challengeIDStr := mux.Vars(r)["id"]
		challengeID, err := validateChallengeID(challengeIDStr)
		if err != nil {
			log.Errorf("invalid challenge ID: %v", err)
			http.Error(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		log = log.WithFields(logrus.Fields{
			"challenge_id": challengeID,
		})

		type submission struct {
			Flag string `json:"flag"`
		}

		var sub submission
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			log.Errorf("failed to decode flag submission: %v", err)
			utility.SendJSONError(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		// Sanitize and validate flag input
		// Error messages from this function can be returned to the client
		sanitizedFlag, err := container.ChallengeClient.SanitizeFlag(sub.Flag)
		if err != nil {
			log.Errorf("flag validation failed: %v", err)
			utility.SendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		sub.Flag = sanitizedFlag

		completed, completedAt, err := container.ChallengeClient.CheckUserCompletedChallenge(user.Email, challengeID)
		if err != nil {
			log.Errorf("Database error checking challenge completion: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		if completed {
			log.Info("flag submission rejected - challenge already completed")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"message":      "Challenge already completed",
				"completed_at": completedAt,
			}); err != nil {
				log.Errorf("encode error: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Get the challenge to validate the flag
		flagValue, pointRewardAmount, validationHandler, challengeName, category, err := container.ChallengeClient.GetChallengeFlagAndReward(challengeID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, notFoundError, http.StatusNotFound)
			} else {
				log.Errorf("database error fetching challenge: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Treat exam challenges as if they don't exist
		if category == "exam" {
			http.Error(w, notFoundError, http.StatusNotFound)
			return
		}

		// Validate the flag
		isValid, customIncorrectMessage := container.ChallengeClient.ValidateFlag(sub.Flag, flagValue, validationHandler, user.Email)
		if !isValid {
			log.Info("flag submission failed - incorrect flag")

			// Log the wrong flag attempt to the database
			_, err = container.DB.Exec(`
				INSERT INTO user_history_log (user_email, log, date) 
				VALUES ($1, $2, NOW())
			`, user.Email, fmt.Sprintf("Wrong flag attempt for challenge %d: %s", challengeID, sub.Flag))
			if err != nil {
				log.Errorf("failed to log wrong flag attempt: %v", err)
				// Don't fail the request, just log the error
			}

			// Use custom message if provided, otherwise use default
			errorMessage := "Incorrect flag"
			if customIncorrectMessage != "" {
				errorMessage = customIncorrectMessage
			}

			if err := json.NewEncoder(w).Encode(map[string]any{
				"message": errorMessage,
			}); err != nil {
				log.Errorf("encode error: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Complete challenge (awards token and points, records completion)
		err = container.UserClient.CompleteChallenge(ctx, user.Email, pointRewardAmount, challengeID)
		if err != nil {
			log.Errorf("failed to complete challenge: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		log.WithFields(logrus.Fields{
			"tokens_earned": 1,
			"points_earned": pointRewardAmount,
		}).Info("challenge completed successfully")

		// Send Slack notification
		container.SlackService.SendChallengeCompletion(user, challengeName)

		// Return success response
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message":       "Challenge completed successfully!",
			"tokens_earned": 1,
			"points_earned": pointRewardAmount,
		}); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// ListExamChallenges returns a list of exam challenges up to the next unsolved challenge
func ListExamChallenges(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Get user profile to check ExamChallengesSolved
		profile, err := container.UserClient.GetUserProfile(ctx, container.SlackService, user)
		if err != nil {
			log.Errorf("unable to get user profile: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Query exam challenges up to the next unsolved challenge (ExamChallengesSolved + 1)
		maxNestedID := profile.ExamChallengesSolved + 1
		rows, err := container.DB.Query(`
			SELECT c.nested_id, c.name, c.description, c.category, c.point_reward_amount,
			       CASE WHEN ucc.challenge_id IS NOT NULL THEN true ELSE false END as completed
			FROM challenges c
			LEFT JOIN user_challenges_completed ucc ON c.id = ucc.challenge_id AND ucc.user_email = $1
			WHERE c.category = 'exam' AND c.nested_id <= $2
			ORDER BY c.nested_id
		`, user.Email, maxNestedID)

		if err != nil {
			log.Errorf("unable to query exam challenges from database: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		challenges := make([]services.Challenge, 0)
		for rows.Next() {
			var challenge services.Challenge
			if err := rows.Scan(&challenge.NestedID, &challenge.Name, &challenge.Description, &challenge.Category, &challenge.PointRewardAmount, &challenge.Completed); err != nil {
				log.Errorf("unable to scan rows queried from database: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
				return
			}
			// Use nested_id as the exposed ID for exam challenges
			challenge.ID = challenge.NestedID
			challenges = append(challenges, challenge)
		}

		if err := rows.Err(); err != nil {
			log.Errorf("rows error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(challenges); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// GetExamChallenge returns detailed information about a specific exam challenge
func GetExamChallenge(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		// Get user from context (set by auth middleware)
		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Get nested ID (exposed as ID for exam challenges)
		id := mux.Vars(r)["id"]
		nestedID, err := validateChallengeID(id)
		if err != nil {
			log.Errorf("invalid challenge ID: %v", err)
			utility.SendJSONError(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		log = log.WithFields(logrus.Fields{
			"exam_nested_id": nestedID,
		})

		// Get user profile to check ExamChallengesSolved
		profile, err := container.UserClient.GetUserProfile(ctx, container.SlackService, user)
		if err != nil {
			log.Errorf("unable to get user profile: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Check if user has access to this exam challenge (sequential access)
		maxAllowedNestedID := profile.ExamChallengesSolved + 1
		if nestedID > maxAllowedNestedID {
			log.Infof("user attempted to access exam challenge beyond their progress: nested_id=%d, max_allowed=%d", nestedID, maxAllowedNestedID)
			http.Error(w, notFoundError, http.StatusNotFound)
			return
		}

		var challenge services.DetailedChallenge
		err = container.DB.QueryRow(`
			SELECT c.id, c.nested_id, c.name, c.description, c.category, c.point_reward_amount, c.file_asset, c.text_asset,
			       CASE WHEN ucc.challenge_id IS NOT NULL THEN true ELSE false END as completed
			FROM challenges c
			LEFT JOIN user_challenges_completed ucc ON c.id = ucc.challenge_id AND ucc.user_email = $2
			WHERE c.category = 'exam' AND c.nested_id = $1
			`, nestedID, user.Email).Scan(
			&challenge.ID,
			&challenge.NestedID,
			&challenge.Name,
			&challenge.Description,
			&challenge.Category,
			&challenge.PointRewardAmount,
			&challenge.FileAsset,
			&challenge.TextAsset,
			&challenge.Completed,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, notFoundError, http.StatusNotFound)
			} else {
				log.Errorf("unable to query exam challenge from database: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Use nested_id as the exposed ID for exam challenges
		challenge.ID = challenge.NestedID

		// Transform file_asset to presigned URL if not empty
		if challenge.FileAsset != nil && *challenge.FileAsset != "" {
			presignedURL, err := container.AssetService.GetAsset(ctx, *challenge.FileAsset)
			if err != nil {
				log.Errorf("failed to get presigned URL for asset %s: %v", challenge.FileAsset, err)
				http.Error(w, internalError, http.StatusInternalServerError)
				return
			}
			challenge.FileAsset = &presignedURL
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(challenge); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}

// SubmitExamChallenge handles exam challenge flag submissions
func SubmitExamChallenge(container *services.Container) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := services.GetLogger(ctx)

		user, ok := container.Auth.GetUserFromContext(ctx)
		if !ok {
			log.Errorf("missing user context after authenticated middleware")
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Get nested ID (exposed as ID for exam challenges)
		challengeIDStr := mux.Vars(r)["id"]
		nestedID, err := validateChallengeID(challengeIDStr)
		if err != nil {
			log.Errorf("invalid challenge ID: %v", err)
			http.Error(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		log = log.WithFields(logrus.Fields{
			"exam_nested_id": nestedID,
		})

		type submission struct {
			Flag string `json:"flag"`
		}

		var sub submission
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			log.Errorf("failed to decode flag submission: %v", err)
			utility.SendJSONError(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		// Sanitize and validate flag input
		sanitizedFlag, err := container.ChallengeClient.SanitizeFlag(sub.Flag)
		if err != nil {
			log.Errorf("flag validation failed: %v", err)
			utility.SendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		sub.Flag = sanitizedFlag

		// Get user profile to check ExamChallengesSolved
		profile, err := container.UserClient.GetUserProfile(ctx, container.SlackService, user)
		if err != nil {
			log.Errorf("unable to get user profile: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		// Check if user has access to this exam challenge (sequential access)
		maxAllowedNestedID := profile.ExamChallengesSolved + 1
		if nestedID > maxAllowedNestedID {
			log.Infof("user attempted to submit exam challenge beyond their progress: nested_id=%d, max_allowed=%d", nestedID, maxAllowedNestedID)
			http.Error(w, notFoundError, http.StatusNotFound)
			return
		}

		// Get the global challenge ID for this nested ID
		var globalChallengeID int
		err = container.DB.QueryRow(`
			SELECT id FROM challenges WHERE category = 'exam' AND nested_id = $1
		`, nestedID).Scan(&globalChallengeID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, notFoundError, http.StatusNotFound)
			} else {
				log.Errorf("unable to find exam challenge: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		completed, completedAt, err := container.ChallengeClient.CheckUserCompletedChallenge(user.Email, globalChallengeID)
		if err != nil {
			log.Errorf("Database error checking challenge completion: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		if completed {
			log.Info("exam flag submission rejected - challenge already completed")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"message":       "Challenge already completed",
				"completed_at":  completedAt,
				"tokens_burned": 0,
			}); err != nil {
				log.Errorf("encode error: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Burn 1 token for every submission (whether correct or not)
		tokensBurned, err := container.UserClient.BurnToken(ctx, user.Email, globalChallengeID)
		if err != nil {
			log.Errorf("failed to burn token: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}
		if tokensBurned == 0 {
			log.Info("exam flag submission rejected - insufficient tokens")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"message": "Insufficient tokens. Each exam submission costs 1 token.",
			}); err != nil {
				log.Errorf("encode error: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Get the challenge to validate the flag
		flagValue, _, validationHandler, challengeName, _, err := container.ChallengeClient.GetChallengeFlagAndReward(globalChallengeID)
		if err != nil {
			// Refund the token on unexpected error
			if refundErr := container.UserClient.RefundToken(ctx, user.Email, globalChallengeID); refundErr != nil {
				log.Errorf("failed to refund token after challenge fetch error: %v", refundErr)
			} else {
				log.Info("refunded 1 token due to challenge fetch error")
			}

			if err == sql.ErrNoRows {
				http.Error(w, notFoundError, http.StatusNotFound)
			} else {
				log.Errorf("database error fetching challenge: %v", err)
				http.Error(w, internalError, http.StatusInternalServerError)
			}
			return
		}

		// Validate the flag
		isValid, customIncorrectMessage := container.ChallengeClient.ValidateFlag(sub.Flag, flagValue, validationHandler, user.Email)
		if !isValid {
			log.Info("exam flag submission failed - incorrect flag")

			// Log the wrong flag attempt to the database
			_, err = container.DB.Exec(`
				INSERT INTO user_history_log (user_email, log, date) 
				VALUES ($1, $2, NOW())
			`, user.Email, fmt.Sprintf("Wrong flag attempt for exam challenge %d: %s", nestedID, sub.Flag))
			if err != nil {
				log.Errorf("failed to log wrong flag attempt: %v", err)
			}

			// Use custom message if provided, otherwise use default
			errorMessage := "Incorrect flag"
			if customIncorrectMessage != "" {
				errorMessage = customIncorrectMessage
			}

			container.SlackService.SendExamChallengeFailedAttempt(user, challengeName)
			json.NewEncoder(w).Encode(map[string]any{
				"message":       errorMessage,
				"tokens_burned": 1,
			})
			return
		}

		// Complete exam challenge (awards token, increments counter, records completion)
		err = container.UserClient.CompleteExamChallenge(ctx, user.Email, globalChallengeID)
		if err != nil {
			// Refund the token on unexpected error
			if refundErr := container.UserClient.RefundToken(ctx, user.Email, globalChallengeID); refundErr != nil {
				log.Errorf("failed to refund token after challenge completion error: %v", refundErr)
			} else {
				log.Info("refunded 1 token due to challenge completion error")
			}
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		log.WithFields(logrus.Fields{
			"tokens_earned": 1,
			"tokens_burned": 1,
			"points_earned": 0,
		}).Info("exam challenge completed successfully")

		// Send Slack notification
		container.SlackService.SendExamChallengeCompletion(user, challengeName)

		// Return success response
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message":       "Exam challenge completed successfully!",
			"tokens_earned": 1,
			"tokens_burned": 1,
			"points_earned": 0,
		}); err != nil {
			log.Errorf("encode error: %v", err)
			http.Error(w, internalError, http.StatusInternalServerError)
		}
	})
}
