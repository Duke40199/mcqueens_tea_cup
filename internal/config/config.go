package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	DatabaseTablesCfg    DatabaseTablesConfig
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
	BotOwnerID                   string
	IDACOBMetaCarsChannelID      string
	IDACOBActivePlayersChannelID string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DatabaseTablesConfig holds the physical table names, so they can be
// overridden per environment without touching the repository code.
type DatabaseTablesConfig struct {
	IDACCarsMetadata      string
	IDACCarStylesMetadata string
	IDACTATimeMetadata    string
	IDACAreaMetadata      string
	IDACOBSyncAreaCfg     string
	IDACStores            string
	PlayerAlias           string
	OBRankingCfg          string
	CfgPlayerRanking      string
	CfsState              string
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
			BotOwnerID:                   getEnv("DISCORD_BOT_OWNER_ID", ""),
			IDACOBMetaCarsChannelID:      getEnv("DISCORD_OB_META_CARS_CHANNEL_ID", ""),
			IDACOBActivePlayersChannelID: getEnv("DISCORD_OB_ACTIVE_PLAYERS_CHANNEL_ID", ""),
		},
		DatabaseCfg: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnvAsInt("DB_PORT", 0),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
		},
		DatabaseTablesCfg: DatabaseTablesConfig{
			IDACCarsMetadata:      getEnv("DB_TABLE_IDAC_CARS_METADATA", "sega_idac_cars_metadata"),
			IDACCarStylesMetadata: getEnv("DB_TABLE_IDAC_CAR_STYLES_METADATA", "sega_idac_car_styles_metadata"),
			IDACTATimeMetadata:    getEnv("DB_TABLE_IDAC_TA_TIME_METADATA", "sega_idac_ta_time_metadata"),
			IDACAreaMetadata:      getEnv("DB_TABLE_IDAC_AREA_METADATA", "idac_area_metadata"),
			IDACOBSyncAreaCfg:     getEnv("DB_TABLE_IDAC_OB_SYNC_AREA_CFG", "idac_ob_sync_area_cfg"),
			IDACStores:            getEnv("DB_TABLE_IDAC_STORES", "idac_stores"),
			PlayerAlias:           getEnv("DB_TABLE_PLAYER_ALIAS", "player_alias"),
			OBRankingCfg:          getEnv("DB_TABLE_OB_RANKING_CFG", "ob_ranking_cfg"),
			CfgPlayerRanking:      getEnv("DB_TABLE_CFG_PLAYER_RANKING", "cfg_player_ranking"),
			CfsState:              getEnv("DB_TABLE_CFS_STATE", "cfs_state"),
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

	// Validations: fail fast with a clear message naming the actual env var.
	missing := []string{}
	if cfg.DiscordCfg.Token == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
	if cfg.DatabaseCfg.Host == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.DatabaseCfg.Port == 0 {
		missing = append(missing, "DB_PORT")
	}
	if cfg.DatabaseCfg.User == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.DatabaseCfg.Name == "" {
		missing = append(missing, "DB_NAME")
	}
	if cfg.SegaClientCfg.SegaIDACHost == "" {
		missing = append(missing, "SEGA_IDAC_HOST")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
