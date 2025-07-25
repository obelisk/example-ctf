package services

import (
	"database/sql"
	"fmt"
	"github.com/obelisk/example-ctf/config"
)

// Container holds all application dependencies
type Container struct {
	DB              *sql.DB
	Config          *config.Config
	ChallengeClient *ChallengeClient
	Auth            *AuthClient
	UserClient      *UserClient
	AssetService    *AssetService
	SlackService    *SlackService
}

// NewContainer creates a new dependency container
func NewContainer(db *sql.DB, cfg *config.Config) *Container {
	assetService, err := NewAssetService(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize asset service: %v", err))
	}

	return &Container{
		DB:              db,
		Config:          cfg,
		ChallengeClient: NewChallengeClient(db, cfg),
		Auth:            NewAuthClient(db, cfg),
		UserClient:      NewUserClient(db, cfg),
		AssetService:    assetService,
		SlackService:    NewSlackService(db, cfg),
	}
}
