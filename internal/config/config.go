package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config interface {
	GetDiscordCfg() DiscordConfig
	GetRSSCfg() RSSConfig
}

// AppConfig holds all configuration
type AppConfig struct {
	DiscordCfg  DiscordConfig
	RSSCfg      RSSConfig
	DatabaseCfg DatabaseConfig
	MetaSyncCfg MetaSyncConfig
}

type MetaSyncConfig struct {
	ChannelID string `json:"channel_id"`
	Interval  int    `json:"sync_interval_minutes"`
}

// DiscordConfig holds Discord Integration configuration
type DiscordConfig struct {
	Token string `json:"token"`
}

type RSSConfig struct {
	Interval int          `json:"check_interval_minutes"`
	Feeds    []FeedConfig `json:"feeds"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
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
	_ = godotenv.Load()

	cfg := &AppConfig{}

	// discord
	cfg.DiscordCfg.Token = os.Getenv("DISCORD_BOT_TOKEN")
	// database
	cfg.DatabaseCfg.Host = os.Getenv("DB_HOST")
	cfg.DatabaseCfg.User = os.Getenv("DB_USER")
	cfg.DatabaseCfg.Password = os.Getenv("DB_PASSWORD")
	cfg.DatabaseCfg.Name = os.Getenv("DB_NAME")

	portStr := os.Getenv("DB_PORT")
	if port, err := strconv.Atoi(portStr); err == nil {
		cfg.DatabaseCfg.Port = port
	} else {
		cfg.DatabaseCfg.Port = 5432 // Default fallback
	}
	cfg.RSSCfg.Interval, _ = strconv.Atoi(os.Getenv("RSS_FETCH_INTERVAL"))
	cfg.RSSCfg.Feeds, _ = LoadRSSConfig()

	// Meta Sync
	cfg.MetaSyncCfg.ChannelID = os.Getenv("DISCORD_OB_META_CARS_CHANNEL_ID")
	metaInterval, _ := strconv.Atoi(os.Getenv("OB_META_CARS_SYNC_INTERVAL_MINUTES"))
	if metaInterval == 0 {
		metaInterval = 15 // Default 15 mins
	}
	cfg.MetaSyncCfg.Interval = metaInterval

	// Validation
	if cfg.DiscordCfg.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is missing")
	}

	return cfg, nil
}

// --- Global Variables ---
var (
	feedsPath = "./config.json"
)

func LoadRSSConfig() ([]FeedConfig, error) {
	file, err := os.ReadFile(feedsPath)
	if err != nil {
		return nil, err
	}
	newCfg := make([]FeedConfig, 0)
	if err = json.Unmarshal(file, &newCfg); err != nil {
		return nil, err
	}
	return newCfg, err
}

type FeedConfig struct {
	Title                       string `json:"title"`
	URL                         string `json:"url"`
	SourceLanguage              string `json:"source_language"` // e.g., "JP", "ja", "es"
	DestinationDiscordChannelID string `json:"channel_id"`      // Custom channel for this feed
}
