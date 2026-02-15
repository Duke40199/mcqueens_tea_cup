package entity

// IdacRepository defines WHAT data we need, not HOW we get it.
type SegaClient interface {
	GetTimeAttack(courseID, area, car, spec string) ([]TimeAttackRecord, error)
	GetTeamRanking(round int, rankCode string) ([]TeamRecord, error)
	GetListOBRanking(roundNum string, areaCode string) (*IdacOBRankingResponse, error)
	GetCurrentRound() (int, error)
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
