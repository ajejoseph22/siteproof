// Package config handles application configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OpenAI   OpenAIConfig
	Postmark PostmarkConfig
	Twilio   TwilioConfig
	Storage  StorageConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// OpenAIConfig holds OpenAI API configuration.
type OpenAIConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
}

// PostmarkConfig holds Postmark email API configuration.
type PostmarkConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
	Timeout   time.Duration
}

// TwilioConfig holds Twilio SMS API configuration.
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	Timeout    time.Duration
}

// StorageConfig holds file storage configuration.
type StorageConfig struct {
	BasePath     string
	MaxFileSize  int64
	AllowedTypes []string
}

// Load reads configuration from environment variables.
// It returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnvRequired("DATABASE_URL"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnvRequired("JWT_SECRET"),
			AccessTokenTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
			Issuer:          getEnv("JWT_ISSUER", "siteproof"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  getEnv("OPENAI_API_KEY", ""),
			Model:   getEnv("OPENAI_MODEL", "gpt-4o-mini"),
			Timeout: getEnvDuration("OPENAI_TIMEOUT", 10*time.Second),
		},
		Postmark: PostmarkConfig{
			APIKey:    getEnv("POSTMARK_API_KEY", ""),
			FromEmail: getEnv("POSTMARK_FROM_EMAIL", "tickets@siteproof.app"),
			FromName:  getEnv("POSTMARK_FROM_NAME", "SiteProof"),
			Timeout:   getEnvDuration("POSTMARK_TIMEOUT", 5*time.Second),
		},
		Twilio: TwilioConfig{
			AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
			FromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
			Timeout:    getEnvDuration("TWILIO_TIMEOUT", 5*time.Second),
		},
		Storage: StorageConfig{
			BasePath:    getEnv("STORAGE_BASE_PATH", "/data/uploads"),
			MaxFileSize: getEnvInt64("STORAGE_MAX_FILE_SIZE", 10*1024*1024), // 10MB
			AllowedTypes: []string{
				"image/jpeg",
				"image/png",
				"image/webp",
				"image/heic",
				"image/heif",
			},
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

// Addr returns the server address in host:port format.
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Helper functions for environment variable parsing.

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	return os.Getenv(key)
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
