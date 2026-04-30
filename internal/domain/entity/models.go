package entity

import "time"

// Item represents a generic news item (decoupled from gofeed)
type Item struct {
	Title       string
	Content     string
	Link        string
	PublishedAt time.Time
	ImageURL    string
	GUID        string
}

type MieMeBell struct {
	Blocks []string `json:"blocks"`
}
