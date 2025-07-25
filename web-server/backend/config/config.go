package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config stores the main application configuration
type Config struct {
	HTTP        HTTPConfig        `validate:"required"`
	HealthCheck HealthCheckConfig `validate:"required"`
	Auth        AuthConfig        `validate:"required"`
	Database    DatabaseConfig    `validate:"required"`
	AwsConfig   AwsConfig         `validate:"required"`
	Slack       SlackConfig       `validate:"required"`
}

// HTTPConfig stores configuration for the public facing HTTP server.
type HTTPConfig struct {
	Hostname string
	Port     uint16 `validate:"required"`

	Timeout               time.Duration   `validate:"required"`
	RequestSizeLimitBytes uint64          `validate:"required"`
	RateLimit             RateLimitConfig `validate:"required"`
}

// HealthCheckConfig stores configuration for the health check endpoint
type HealthCheckConfig struct {
	Hostname string `validate:""`
	Port     uint16 `validate:"required"`
}

// RateLimitConfig stores configuration for rate limiting
type RateLimitConfig struct {
	Enabled         bool          `validate:"required"`
	RequestsPerSec  float64       `validate:"required,min=0.1"`
	BurstSize       int           `validate:"required,min=1"`
	WindowSize      time.Duration `validate:"required"`
	CleanupInterval time.Duration `validate:"required"`
	MaxClients      int           `validate:"required,min=1"`
}

// AuthConfig stores authentication configuration
type AuthConfig struct {
	TestMode *AuthTestMode `yaml:"testMode,omitempty"`

	ExpectedVerifiedAccessInstanceARN string        `validate:"required"`
	ExpectedIssuer                    string        `validate:"required"`
	AWSRegion                         string        `validate:"required"`
	PublicKeyCacheTTL                 time.Duration `yaml:"publicKeyCacheTTL,omitempty"`
}

// AuthTestMode stores test mode authentication configuration
type AuthTestMode struct {
	Enabled  bool
	TestUser string
}

// DatabaseConfig stores database configuration
type DatabaseConfig struct {
	Hostname string `validate:"required,hostname_rfc1123"`
	Port     uint16 `validate:"required"`
	User     string `validate:"required"`
	Password string
	Database string `validate:"required"`
	SslMode  string `validate:"required,oneof=disable allow prefer require verify-ca verify-full"`
}

// AwsConfig stores config to access S3
type AwsConfig struct {
	BucketName string `validate:"required"`
}

// SlackConfig stores configuration for Slack integration
type SlackConfig struct {
	LeaderboardInterval time.Duration `yaml:"leaderboardInterval,omitempty"`
}

// GetConfig loads and returns the application configuration
func GetConfig() (Config, error) {
	var c Config

	// Load the config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// Load env variables
	viper.SetEnvPrefix("ctfbackend")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	configFilePath, isSet := os.LookupEnv("CONFIG_FILE")
	if isSet {
		buf, err := os.ReadFile(configFilePath)
		if err != nil {
			return c, fmt.Errorf("failed to read config from file path: '%s': %w", configFilePath, err)
		}
		if err := viper.ReadConfig(bytes.NewReader(buf)); err != nil {
			return c, err
		}
	} else {
		if err := viper.ReadInConfig(); err != nil {
			return c, err
		}
	}
	if err := viper.Unmarshal(&c); err != nil {
		return c, err
	}

	// Apply env vars to the config if set.
	setValueFromEnvVar("POSTGRES_PASSWORD", &c.Database.Password)

	// Set default values for optional fields
	if c.Slack.LeaderboardInterval == 0 {
		c.Slack.LeaderboardInterval = 30 * time.Minute
	}

	// Validate configuration.
	if err := validator.New().Struct(c); err != nil {
		return c, fmt.Errorf("configuration file failed validation: %w", err)
	}
	return c, nil
}

func setValueFromEnvVar(envVar string, field any) {
	value, isSet := os.LookupEnv(envVar)
	if !isSet {
		return
	}
	fmt.Printf("loading config value from env var '%s'", envVar)
	v := reflect.ValueOf(field)
	// Check if the passed interface is a pointer to a string field in a struct
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.String {
		panic(fmt.Errorf("expected a pointer to a string field for env var '%s'", envVar))
	}
	v.Elem().SetString(value)
}
