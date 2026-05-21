package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config interface {
	GetDiscordCfg() DiscordConfig
	GetSegaClientCfg() SegaClientConfig
	GetAllNetClientCfg() AllNetClientConfig
	GetMetaSyncCfg() MetaSyncConfig
	GetActivePlayersSyncCfg() ActivePlayersSyncConfig
}

func (c *AppConfig) GetDiscordCfg() DiscordConfig {
	return c.DiscordCfg
}
func (c *AppConfig) GetSegaClientCfg() SegaClientConfig {
	return c.SegaClientCfg
}
func (c *AppConfig) GetAllNetClientCfg() AllNetClientConfig {
	return c.AllNetClientConfig
}
func (c *AppConfig) GetMetaSyncCfg() MetaSyncConfig {
	return c.MetaSyncCfg
}
func (c *AppConfig) GetActivePlayersSyncCfg() ActivePlayersSyncConfig {
	return c.ActivePlayersSyncCfg
}

// AppConfig holds all configuration
type AppConfig struct {
	DiscordCfg           DiscordConfig
	DatabaseCfg          DatabaseConfig
	MetaSyncCfg          MetaSyncConfig
	ActivePlayersSyncCfg ActivePlayersSyncConfig
	SegaClientCfg        SegaClientConfig
	AllNetClientConfig   AllNetClientConfig
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

type AllNetClientConfig struct {
	AllNetHost                  string
	GetListStoreLocationURLPath string
	// consts
	IDACGameCode        string
	EnglishLanguageCode string
}

type MetaSyncConfig struct {
	ChannelID       string
	IntervalMinutes int
	DowntimeStart   int
	DowntimeEnd     int
	DowntimeTZ      string
}

type ActivePlayersSyncConfig struct {
	ChannelID       string
	IntervalMinutes int
	DowntimeStart   int
	DowntimeEnd     int
	DowntimeTZ      string
}

// DiscordConfig holds Discord Integration configuration
type DiscordConfig struct {
	Token                        string
	IDACOBMetaCarsChannelID      string
	IDACOBActivePlayersChannelID string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
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
		AllNetClientConfig: AllNetClientConfig{
			AllNetHost:                  getEnv("ALLNET_HOST", ""),
			GetListStoreLocationURLPath: getEnv("ALLNET_GET_LIST_STORE_LOCATION_URL_PATH", ""),
			IDACGameCode:                getEnv("ALLNET_IDAC_GAME_CODE", ""),
			EnglishLanguageCode:         getEnv("ALLNET_EN_LANGUAGE_CODE", ""),
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

	// Validations
	if cfg.DiscordCfg.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is missing")
	}

	return cfg, nil
}
