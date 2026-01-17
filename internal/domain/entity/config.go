package entity

type FeedState map[string]string // Map of FeedURL -> LastSeenGUID

type Config struct {
	Token    string       `json:"token"`
	Interval int          `json:"check_interval_minutes"`
	Feeds    []FeedConfig `json:"feeds"`
}

// FeedConfig represents a configured RSS source
type FeedConfig struct {
	Title                       string `json:"title"`
	URL                         string `json:"url"`
	SourceLanguage              string `json:"source_language"` // e.g., "JP", "ja", "es"
	DestinationDiscordChannelID string `json:"channel_id"`      // Custom channel for this feed
}

// AppConfig represents global settings
type AppConfig struct {
	Token    string
	Interval int
	Feeds    []FeedConfig
}
