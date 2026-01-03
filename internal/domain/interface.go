package domain

// IdacRepository defines WHAT data we need, not HOW we get it.
type IdacRepository interface {
	GetTimeAttack(courseID, area, car, spec string) ([]TimeAttackRecord, error)
	GetTeamRanking(round int, rankCode string) ([]TeamRecord, error)
	GetCurrentRound() (int, error)
}

// AliasRepository defines how we interact with player aliases.
type AliasRepository interface {
	GetByAliasKey(discordID string) (PlayerAlias, bool, error)
	GetByIgn(ign string) (PlayerAlias, bool)

	Set(discordID, ign, area string) error
	Load() error
}

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
