package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config interface {
	GetDiscordCfg() DiscordConfig
	GetRSSCfg() RSSConfig
}

// AppConfig holds all configuration
type AppConfig struct {
	DiscordCfg DiscordConfig
	RSSCfg     RSSConfig
}

// DiscordConfig holds Discord Integration configuration
type DiscordConfig struct {
	Token string
}

type RSSConfig struct {
	Interval int `json:"check_interval_minutes"`
}

// GetDiscordCfg returns Discord configuration
func (c *AppConfig) GetDiscordCfg() DiscordConfig {
	return c.DiscordCfg
}

// GetRSSCfg returns RSS-related configuration
func (c *AppConfig) GetRSSCfg() RSSConfig {
	return c.RSSCfg
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	// Just load config from source file for dev environment
	if strings.EqualFold(getEnv("CAS_ENV", "local"), "EnvLocal") {
		candidates := []string{
			"example.env",
			"../example.env",
			"../../example.env",
			"../../../example.env",
			"../../../../example.env",
			"../../../../../example.env",
		}
		if explicit := strings.TrimSpace(os.Getenv("ENV_FILE")); explicit != "" {
			if _, err := os.Stat(explicit); err != nil {
				return &AppConfig{}, errors.New("env file not found")
			}
		} else {
			for _, p := range candidates {
				if _, err := os.Stat(p); err == nil {
					if err := godotenv.Load(p); err == nil {
						break
					}
				}
			}
		}
	}

	cfg := &AppConfig{
		DiscordCfg: DiscordConfig{
			Token: getEnv("DISCORD_TOKEN", ""),
		},
		RSSCfg: RSSConfig{
			Interval: getEnvAsInt("RSS_FETCH_INTERVAL", -1),
		},
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
