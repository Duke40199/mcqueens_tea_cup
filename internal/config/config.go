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
	GetSegaClientCfg() string
}

// AppConfig holds all configuration
type AppConfig struct {
	DiscordCfg           DiscordConfig
	RSSCfg               RSSConfig
	DatabaseCfg          DatabaseConfig
	MetaSyncCfg          MetaSyncConfig
	ActivePlayersSyncCfg ActivePlayersSyncConfig
	SegaClientCfg        SegaClientConfig
}

type SegaClientConfig struct {
	SegaIDACHost string
	// sega cfgs
	GetListConstConfigURLPath string
	GetCurrentRoundUrlPath    string
	// time trail
	GetTimeTrailURLPath string
	// ob
	GetListOBRankingURLPath string
	// rankings
	GetTeamRankingUrlPath     string
	GetListPlayerGradeUrlPath string
}

type MetaSyncConfig struct {
	ChannelID       string `json:"channel_id"`
	IntervalMinutes int    `json:"sync_interval_minutes"`
	DowntimeStart   int    `json:"downtime_start"`
	DowntimeEnd     int    `json:"downtime_end"`
	DowntimeTZ      string `json:"downtime_tz"`
}

type ActivePlayersSyncConfig struct {
	ChannelID       string `json:"channel_id"`
	IntervalMinutes int    `json:"sync_interval_minutes"`
	DowntimeStart   int    `json:"downtime_start"`
	DowntimeEnd     int    `json:"downtime_end"`
	DowntimeTZ      string `json:"downtime_tz"`
}

// DiscordConfig holds Discord Integration configuration
type DiscordConfig struct {
	Token                        string `json:"token"`
	IDACOBMetaCarsChannelID      string `json:"idac_ob_meta_cars_channel_id"`
	IDACOBActivePlayersChannelID string `json:"idac_ob_active_players_channel_id"`
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

// env getter funcs
func getEnv(key string, defaultValue string) string {
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

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	_ = godotenv.Load()
	cfg := &AppConfig{
		DiscordCfg: DiscordConfig{
			Token:                        getEnv("DISCORD_BOT_TOKEN", ""),
			IDACOBMetaCarsChannelID:      getEnv("DISCORD_OB_META_CARS_CHANNEL_ID", ""),
			IDACOBActivePlayersChannelID: getEnv("DISCORD_OB_ACTIVE_PLAYERS_CHANNEL_ID", ""),
		},
		DatabaseCfg: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnvAsInt("DB_PORT", 0),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
		},
		SegaClientCfg: SegaClientConfig{
			SegaIDACHost:              getEnv("SEGA_IDAC_HOST", ""),
			GetListConstConfigURLPath: getEnv("SEGA_IDAC_GET_LIST_CONST_CFG_URL_PATH", ""),
			GetTimeTrailURLPath:       getEnv("SEGA_IDAC_GET_TIME_TRAIL_URL_PATH", ""),
			GetListOBRankingURLPath:   getEnv("SEGA_IDAC_GET_LIST_OB_RANKING_URL_PATH", ""),
			GetCurrentRoundUrlPath:    getEnv("SEGA_IDAC_GET_CURRENT_ROUND_URL_PATH", ""),
			GetTeamRankingUrlPath:     getEnv("SEGA_IDAC_GET_TEAM_RANKING_URL_PATH", ""),
			GetListPlayerGradeUrlPath: getEnv("SEGA_IDAC_GET_LIST_PLAYER_GRADE_URL_PATH", ""),
		},
		MetaSyncCfg: MetaSyncConfig{
			ChannelID:       getEnv("DISCORD_OB_META_CARS_CHANNEL_ID", ""),
			IntervalMinutes: getEnvAsInt("OB_META_CARS_SYNC_INTERVAL_MINUTES", 15),
			DowntimeStart:   getEnvAsInt("OB_META_CARS_DOWNTIME_START_HOUR", 0),
			DowntimeEnd:     getEnvAsInt("OB_META_CARS_DOWNTIME_END_HOUR", 0),
			DowntimeTZ:      getEnv("OB_META_CARS_DOWNTIME_TZ", ""),
		},
		ActivePlayersSyncCfg: ActivePlayersSyncConfig{
			ChannelID:       getEnv("DISCORD_OB_ACTIVE_PLAYERS_CHANNEL_ID", ""),
			IntervalMinutes: getEnvAsInt("OB_ACTIVE_PLAYERS_SYNC_INTERVAL_MINUTES", 15),
			DowntimeStart:   getEnvAsInt("OB_ACTIVE_PLAYERS_DOWNTIME_START_HOUR", 0),
			DowntimeEnd:     getEnvAsInt("OB_ACTIVE_PLAYERS_DOWNTIME_END_HOUR", 0),
			DowntimeTZ:      getEnv("OB_ACTIVE_PLAYERS_DOWNTIME_TZ", ""),
		},
	}
	cfg.RSSCfg.Interval, _ = strconv.Atoi(os.Getenv("RSS_FETCH_INTERVAL"))
	cfg.RSSCfg.Feeds, _ = LoadRSSConfig()

	// Validations
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
