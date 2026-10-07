package entity

type MieMeBell struct {
	Blocks []string `json:"blocks"`
}

// ChannelMessage is a minimal, transport-agnostic view of a message the bot posts
// and manages in a channel. It lets the domain reconcile pages to a channel without
// depending on discordgo.
type ChannelMessage struct {
	ID      string
	Content string
}
