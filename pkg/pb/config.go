package wellknown

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/tools/osutils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	OAuth    OAuthConfig
	Database DatabaseConfig
	AI       AIConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DataDir string
}

// AIConfig holds AI/LLM integration configuration
type AIConfig struct {
	Anthropic AnthropicConfig
}

// AnthropicConfig holds Anthropic Claude API configuration
type AnthropicConfig struct {
	// UseOAuth indicates whether to use Anthropic OAuth tokens (default: false)
	// When true, tokens would be fetched from 'anthropic_tokens' collection (NOT IMPLEMENTED YET)
	// For now, use APIKey mode (UseOAuth=false) which is simpler and recommended
	UseOAuth bool

	// APIKey is the Anthropic API key for direct API access (RECOMMENDED)
	// This is the primary/simple mode for Claude API integration
	// Get your key from: https://console.anthropic.com/
	APIKey string

	// Model specifies which Claude model to use
	// Default: claude-sonnet-4-5-20250929
	Model string
}

// LoadConfig loads configuration from environment variables
// In development (go run), it will try to load .env file first
func LoadConfig() (*Config, error) {
	// PocketBase best practice: Only load .env in development
	// In production, use environment variables or command-line flags
	if osutils.IsProbablyGoRun() {
		// Try to load .env file, but don't fail if it doesn't exist
		_ = godotenv.Load()
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnvOrDefault("SERVER_HOST", "127.0.0.1"),
			Port: getEnvIntOrDefault("SERVER_PORT", 8090),
		},
		OAuth: OAuthConfig{
			Google: GoogleOAuthConfig{
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			},
		},
		Database: DatabaseConfig{
			DataDir: getEnvOrDefault("PB_DATA_DIR", ".data/pb"),
		},
		AI: AIConfig{
			Anthropic: AnthropicConfig{
				UseOAuth: getBoolEnvOrDefault("AI_USE_OAUTH", false), // Default to API key mode (simpler)
				APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
				Model:    getEnvOrDefault("ANTHROPIC_MODEL", "claude-sonnet-4-5-20250929"),
			},
		},
	}

	// Check if Google OAuth is configured
	cfg.OAuth.Google.Enabled = cfg.OAuth.Google.ClientID != "" &&
		cfg.OAuth.Google.ClientSecret != "" &&
		cfg.OAuth.Google.RedirectURL != ""

	return cfg, nil
}

// ToOAuth2Config converts GoogleOAuthConfig to oauth2.Config
func (g *GoogleOAuthConfig) ToOAuth2Config() *oauth2.Config {
	if !g.Enabled {
		return nil
	}

	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}
}

// ServerAddress returns the full server address
func (s *ServerConfig) ServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// ServerURL returns the full server URL
func (s *ServerConfig) ServerURL() string {
	return fmt.Sprintf("http://%s:%d", s.Host, s.Port)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.DataDir == "" {
		return fmt.Errorf("database data directory cannot be empty")
	}

	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}
