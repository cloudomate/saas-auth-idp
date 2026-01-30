package config

import (
	"os"
)

// Config holds all configuration values
type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret string

	// OAuth - Google
	GoogleClientID     string
	GoogleClientSecret string

	// OAuth - GitHub
	GitHubClientID     string
	GitHubClientSecret string

	// Email
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string

	// App
	AppURL     string
	FrontendURL string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		Port: getEnv("PORT", "8000"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/saas_starter?sslmode=disable"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "development-jwt-secret-change-in-production"),

		// OAuth - Google
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),

		// OAuth - GitHub
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),

		// App
		AppURL:      getEnv("APP_URL", "http://localhost:8000"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetJWTSecret returns the JWT signing secret as bytes
func (c *Config) GetJWTSecret() []byte {
	return []byte(c.JWTSecret)
}

// HasGoogleOAuth returns true if Google OAuth is configured
func (c *Config) HasGoogleOAuth() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

// HasGitHubOAuth returns true if GitHub OAuth is configured
func (c *Config) HasGitHubOAuth() bool {
	return c.GitHubClientID != "" && c.GitHubClientSecret != ""
}

// HasSMTP returns true if SMTP is configured
func (c *Config) HasSMTP() bool {
	return c.SMTPHost != "" && c.SMTPUser != ""
}
