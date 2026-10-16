package entity

import "time"

// CommandGuide is an authored help article for a bot command, stored in the
// bot_command_guides table and surfaced by /idac help. CommandKey is the stable
// lookup key (e.g. "idac time-attack"); Title is the display heading; Content is
// the guide body written in Discord-flavoured markdown; Type groups the guide by
// command family (e.g. "IDAC", "COMMON").
type CommandGuide struct {
	ID         int64
	CommandKey string
	Title      string
	Content    string
	Type       string
	SortOrder  int
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
