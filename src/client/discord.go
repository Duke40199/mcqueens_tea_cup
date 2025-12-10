package client

type FeedConfig struct {
	Title                       string `json:"title"`
	URL                         string `json:"url"`
	SourceLanguage              string `json:"source_language"` // e.g., "JP", "ja", "es"
	DestinationDiscordChannelID string `json:"channel_id"`      // Custom channel for this feed
}
