package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/obelisk/example-ctf/config"
	"github.com/sirupsen/logrus"
)

// SlackMessage represents a message to be sent to Slack
type SlackMessage struct {
	Text      string `json:"text"`
	IsPrivate bool   `json:"-"` // Not sent to Slack, used internally to determine channel
}

// SlackService handles sending messages to Slack webhook
type SlackService struct {
	privateWebhookURL string
	publicWebhookURL  string
	messageChan       chan SlackMessage
	client            *http.Client
	db                *sql.DB
	config            *config.Config
	cachedStats       *LeaderboardStats
	cacheTimestamp    time.Time
	cacheMutex        sync.RWMutex
}

// NewSlackService creates a new Slack service
func NewSlackService(db *sql.DB, cfg *config.Config) *SlackService {
	privateWebhookURL := os.Getenv("SLACK_PRIVATE_WEBHOOK")
	publicWebhookURL := os.Getenv("SLACK_PUBLIC_WEBHOOK")

	// Return nil if no webhooks configured
	if privateWebhookURL == "" && publicWebhookURL == "" {
		return nil
	}

	service := &SlackService{
		privateWebhookURL: privateWebhookURL,
		publicWebhookURL:  publicWebhookURL,
		messageChan:       make(chan SlackMessage, 128), // Buffer for 128 messages
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		db:     db,
		config: cfg,
	}

	// Start the message dispatcher
	go service.dispatcher()

	return service
}

// dispatcher runs in a single goroutine and sends messages to Slack
func (s *SlackService) dispatcher() {
	for msg := range s.messageChan {
		if err := s.sendMessage(msg); err != nil {
			logrus.WithError(err).Error("failed to send slack message")
		}
	}
}

// sendMessage sends a single message to Slack webhook
func (s *SlackService) sendMessage(msg SlackMessage) error {
	// Determine which webhook to use
	var webhookURL string
	if msg.IsPrivate {
		webhookURL = s.privateWebhookURL
	} else {
		webhookURL = s.publicWebhookURL
	}

	// Skip if webhook not configured
	if webhookURL == "" {
		logrus.Debug("webhook not configured, skipping message")
		return nil
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to post to slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status: %d", resp.StatusCode)
	}

	return nil
}

// SendMessageAsync sends a message to Slack asynchronously
func (s *SlackService) SendMessageAsync(text string, isPrivate bool) {
	if s == nil {
		return // No webhook configured
	}

	select {
	case s.messageChan <- SlackMessage{Text: text, IsPrivate: isPrivate}:
		// Message queued successfully
	default:
		// Channel is full, log but don't block
		logrus.Warn("slack message queue is full, dropping message")
	}
}

// SendChallengeCompletion sends a notification when a user completes a challenge
func (s *SlackService) SendChallengeCompletion(user *User, challengeName string) {
	ctx := context.Background()
	alias := s.getUserAlias(ctx, user.Email)

	var privateText, publicText string
	if alias != "" {
		privateText = fmt.Sprintf("🎉 *%s* (%s) solved challenge *%s*", user.Email, alias, challengeName)
		publicText = fmt.Sprintf("🎉 *%s* solved challenge *%s*", alias, challengeName)
	} else {
		privateText = fmt.Sprintf("🎉 *%s* solved challenge *%s*", user.Email, challengeName)
		publicText = fmt.Sprintf("🎉 *%s* solved challenge *%s*", user.Email, challengeName)
	}

	s.SendMessageAsync(privateText, true)
	s.SendMessageAsync(publicText, false)
}

// SendExamChallengeFailedAttempt sends a notification when a user fails an exam challenge
func (s *SlackService) SendExamChallengeFailedAttempt(user *User, challengeName string) {
	if s == nil {
		return
	}

	// Get user alias for display
	ctx := context.Background()
	alias := s.getUserAlias(ctx, user.Email)

	var privateText, publicText string
	if alias != "" {
		privateText = fmt.Sprintf("❌ *%s* (%s) submitted a wrong flag for exam challenge *%s*", user.Email, alias, challengeName)
		publicText = fmt.Sprintf("❌ *%s* submitted a wrong flag for exam challenge *%s*", alias, challengeName)
	} else {
		privateText = fmt.Sprintf("❌ *%s* submitted a wrong flag for exam challenge *%s*", user.Email, challengeName)
		publicText = fmt.Sprintf("❌ *%s* submitted a wrong flag for exam challenge *%s*", user.Email, challengeName)
	}

	s.SendMessageAsync(privateText, true)
	s.SendMessageAsync(publicText, false)
}

// SendExamChallengeCompletion sends a notification when a user completes an exam challenge
func (s *SlackService) SendExamChallengeCompletion(user *User, challengeName string) {
	if s == nil {
		return
	}

	// Get user alias for display
	ctx := context.Background()
	alias := s.getUserAlias(ctx, user.Email)

	var privateText, publicText string
	if alias != "" {
		privateText = fmt.Sprintf("🎉 *%s* (%s) solved exam challenge *%s*", user.Email, alias, challengeName)
		publicText = fmt.Sprintf("🎉 *%s* solved exam challenge *%s*", alias, challengeName)
	} else {
		privateText = fmt.Sprintf("🎉 *%s* solved exam challenge *%s*", user.Email, challengeName)
		publicText = fmt.Sprintf("🎉 *%s* solved exam challenge *%s*", user.Email, challengeName)
	}

	s.SendMessageAsync(privateText, true)
	s.SendMessageAsync(publicText, false)
}

// LeaderboardStats represents statistics for the leaderboard
type LeaderboardStats struct {
	TopScorers            []TopScorer
	TotalUsers            int
	TotalSubmissions      int
	SuccessfulSubmissions int
	WrongSubmissions      int
}

// TopScorer represents a user's score
type TopScorer struct {
	UserEmail            string `json:"user_email"`
	Alias                string `json:"alias"`
	Points               int    `json:"points"`
	ExamChallengesSolved int    `json:"exam_challenges_solved"`
}

// getUserAlias gets a user's alias from the database
func (s *SlackService) getUserAlias(ctx context.Context, userEmail string) string {
	var alias string
	err := s.db.QueryRowContext(ctx, `
		SELECT alias FROM user_aliases 
		WHERE user_email = $1 AND deleted_at IS NULL
	`, userEmail).Scan(&alias)
	if err != nil {
		return ""
	}
	return alias
}

// sendLeaderboardUpdate sends a leaderboard update to Slack only if data has changed
func (s *SlackService) sendLeaderboardUpdate(ctx context.Context) {
	if s == nil {
		return
	}

	stats, err := s.getLeaderboardStats(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to get leaderboard stats")
		return
	}

	// Check if stats have changed since last update
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if s.cachedStats != nil && s.statsEqual(*s.cachedStats, stats) {
		logrus.Debug("leaderboard stats unchanged, skipping slack update")
		return
	}

	// Stats have changed, update cache and send message
	s.cachedStats = &stats
	s.cacheTimestamp = time.Now()

	logrus.Info("leaderboard stats changed, sending slack update")
	s.SendPrivateLeaderboardUpdate(stats)
	s.SendPublicLeaderboardUpdate(stats)
}

// SendPrivateLeaderboardUpdate builds and sends a private leaderboard message to Slack
func (s *SlackService) SendPrivateLeaderboardUpdate(stats LeaderboardStats) {
	if s == nil {
		return
	}

	// Build the leaderboard text for private channel
	var text string
	text += "🏆 *Leaderboard Update*\n\n"

	// Add top scorers with email and alias
	text += "*Top 16 Scorers:*\n"
	for i, scorer := range stats.TopScorers {
		if scorer.Alias != "" {
			text += fmt.Sprintf("%d. `%s (%s)` - %d exam challenges - %d points\n", i+1, scorer.UserEmail, scorer.Alias, scorer.ExamChallengesSolved, scorer.Points)
		} else {
			text += fmt.Sprintf("%d. `%s` - %d exam challenges - %d points\n", i+1, scorer.UserEmail, scorer.ExamChallengesSolved, scorer.Points)
		}
	}

	text += "\n*Statistics:*\n"
	text += fmt.Sprintf("• Total Users: %d\n", stats.TotalUsers)
	text += fmt.Sprintf("• Total Submissions: %d\n", stats.TotalSubmissions)
	text += fmt.Sprintf("• Successful Submissions: %d\n", stats.SuccessfulSubmissions)
	text += fmt.Sprintf("• Wrong Submissions: %d\n", stats.WrongSubmissions)

	s.SendMessageAsync(text, true)
}

// SendPublicLeaderboardUpdate builds and sends a public leaderboard message to Slack
func (s *SlackService) SendPublicLeaderboardUpdate(stats LeaderboardStats) {
	if s == nil {
		return
	}

	// Build the leaderboard text for public channel (only users with aliases)
	var text string
	text += "🏆 *Leaderboard Update*\n\n"

	// Add top scorers with aliases, fallback to email
	text += "*Top 16 Scorers:*\n"
	for i, scorer := range stats.TopScorers {
		if scorer.Alias != "" {
			text += fmt.Sprintf("%d. `%s` - %d exam challenges - %d points\n", i+1, scorer.Alias, scorer.ExamChallengesSolved, scorer.Points)
		} else {
			text += fmt.Sprintf("%d. `%s` - %d exam challenges - %d points\n", i+1, scorer.UserEmail, scorer.ExamChallengesSolved, scorer.Points)
		}
	}

	text += "\n*Statistics:*\n"
	text += fmt.Sprintf("• Total Users: %d\n", stats.TotalUsers)
	text += fmt.Sprintf("• Total Submissions: %d\n", stats.TotalSubmissions)
	text += fmt.Sprintf("• Successful Submissions: %d\n", stats.SuccessfulSubmissions)
	text += fmt.Sprintf("• Wrong Submissions: %d\n", stats.WrongSubmissions)

	s.SendMessageAsync(text, false)
}

// statsEqual compares two LeaderboardStats to check if they are equal
func (s *SlackService) statsEqual(a, b LeaderboardStats) bool {
	// Compare basic stats
	if a.TotalUsers != b.TotalUsers ||
		a.TotalSubmissions != b.TotalSubmissions ||
		a.SuccessfulSubmissions != b.SuccessfulSubmissions ||
		a.WrongSubmissions != b.WrongSubmissions {
		return false
	}

	// Compare top scorers using deep equal
	return reflect.DeepEqual(a.TopScorers, b.TopScorers)
}

// getLeaderboardStats queries the database for leaderboard statistics
func (s *SlackService) getLeaderboardStats(ctx context.Context) (LeaderboardStats, error) {
	stats := LeaderboardStats{}

	// Get top 16 scorers
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.user_email, u.points_achieved, u.exam_challenges_solved, COALESCE(ua.alias, '') as alias
		FROM users u
		LEFT JOIN user_aliases ua ON u.user_email = ua.user_email AND ua.deleted_at IS NULL
		WHERE u.points_achieved > 0 
		ORDER BY 
			u.exam_challenges_solved DESC,
			CASE 
				WHEN u.exam_challenges_solved > 0 THEN u.last_exam_challenge_solved_timestamp 
				ELSE NULL 
			END ASC NULLS LAST,
			u.points_achieved DESC,
			u.last_challenge_solved_timestamp ASC 
		LIMIT 16
	`)
	if err != nil {
		return stats, fmt.Errorf("failed to query top scorers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var scorer TopScorer
		if err := rows.Scan(&scorer.UserEmail, &scorer.Points, &scorer.ExamChallengesSolved, &scorer.Alias); err != nil {
			return stats, fmt.Errorf("failed to scan scorer: %w", err)
		}
		stats.TopScorers = append(stats.TopScorers, scorer)
	}

	// Get total unique users
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT user_email) 
		FROM users 
	`).Scan(&stats.TotalUsers)
	if err != nil {
		return stats, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get submission statistics
	err = s.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(CASE WHEN log LIKE 'Completed challenge%' THEN 1 END) as successful_submissions,
			COUNT(CASE WHEN log LIKE 'Wrong flag attempt%' THEN 1 END) as wrong_submissions
		FROM user_history_log
	`).Scan(&stats.SuccessfulSubmissions, &stats.WrongSubmissions)
	if err != nil {
		return stats, fmt.Errorf("failed to get submission stats: %w", err)
	}
	stats.TotalSubmissions = stats.SuccessfulSubmissions + stats.WrongSubmissions

	return stats, nil
}

// StartLeaderboardUpdates starts a goroutine that periodically sends leaderboard updates
func (s *SlackService) StartLeaderboardUpdates(ctx context.Context) {
	if s == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(s.config.Slack.LeaderboardInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.sendLeaderboardUpdate(ctx)
			}
		}
	}()
}
