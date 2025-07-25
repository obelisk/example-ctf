-- Create challenges table
CREATE TABLE IF NOT EXISTS challenges (
    id                      INTEGER PRIMARY KEY,
    nested_id               INTEGER NOT NULL,
    name                    TEXT    NOT NULL,
    description             TEXT    NOT NULL,
    category                TEXT    NOT NULL,
    point_reward_amount     INTEGER NOT NULL,  -- Renamed from token_reward_amount
    file_asset              TEXT,
    text_asset              TEXT
);

CREATE TABLE IF NOT EXISTS flags (
    challenge_id            INTEGER PRIMARY KEY,
    flag_value              TEXT    NOT NULL,
    validation_handler      TEXT    NOT NULL,
    FOREIGN KEY (challenge_id) REFERENCES challenges(id)
);

-- Create users table to track tokens, points, and exam progress
CREATE TABLE IF NOT EXISTS users (
    user_email                            TEXT      PRIMARY KEY,
    tokens_available                      INTEGER   NOT NULL DEFAULT 0,
    tokens_burned                         INTEGER   NOT NULL DEFAULT 0,
    points_achieved                       INTEGER   NOT NULL DEFAULT 0,
    exam_challenges_solved                INTEGER   NOT NULL DEFAULT 0,
    last_exam_challenge_solved_timestamp  TIMESTAMP NOT NULL DEFAULT NOW(),
    last_challenge_solved_timestamp       TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create user_challenges_completed table
CREATE TABLE IF NOT EXISTS user_challenges_completed (
    user_email   TEXT    NOT NULL,
    challenge_id INTEGER NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    completed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_email, challenge_id)
);

-- Create user_history_log table
CREATE TABLE IF NOT EXISTS user_history_log (
    user_email  TEXT      NOT NULL,
    log         TEXT      NOT NULL,
    date        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create user_aliases table to track user aliases (one per user)
CREATE TABLE IF NOT EXISTS user_aliases (
    user_email  TEXT PRIMARY KEY,
    alias       TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

-- Create unique constraint on alias only for non-deleted aliases
-- This ensures no two users can have the same active alias at the same time
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_aliases_unique_active 
ON user_aliases (alias) WHERE deleted_at IS NULL;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_points_achieved ON users(points_achieved);
CREATE INDEX IF NOT EXISTS idx_users_exam_challenges_solved ON users(exam_challenges_solved);
CREATE INDEX IF NOT EXISTS idx_users_last_challenge_timestamp ON users(last_challenge_solved_timestamp);
CREATE INDEX IF NOT EXISTS idx_users_last_exam_timestamp ON users(last_exam_challenge_solved_timestamp);
CREATE INDEX IF NOT EXISTS idx_challenges_category ON challenges(category);
CREATE INDEX IF NOT EXISTS idx_challenges_nested_id_category ON challenges(nested_id, category);
CREATE INDEX IF NOT EXISTS idx_user_challenges_completed_user_email ON user_challenges_completed(user_email);
CREATE INDEX IF NOT EXISTS idx_user_challenges_completed_challenge ON user_challenges_completed(challenge_id);
CREATE INDEX IF NOT EXISTS idx_user_challenges_completed_completed_at ON user_challenges_completed(completed_at);
CREATE INDEX IF NOT EXISTS idx_user_history_log_user_email ON user_history_log(user_email);