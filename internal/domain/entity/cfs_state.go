package entity

type CfsState struct {
	ID        int64  `json:"id"`
	DiscordID string `json:"discord_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}
