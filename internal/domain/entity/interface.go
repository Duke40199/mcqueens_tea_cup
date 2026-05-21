package entity

// RSSFetcher defines how we get feeds (abstracts gofeed)
type RSSFetcher interface {
	Fetch(url string) ([]Item, error)
}

// Notifier defines how we send updates (abstracts discordgo)
type Notifier interface {
	Send(channelID string, item Item, feedTitle string, translationEN, translationVN string) error
}

// StateStore defines how we save/load "last seen" items (abstracts file/db)
type StateStore interface {
	GetLastSeen(feedURL string) (string, bool)
	SetLastSeen(feedURL, guid string) error
}

// Translator defines how we translate text
type Translator interface {
	Translate(text, sourceLang, targetLang string) (string, error)
}
