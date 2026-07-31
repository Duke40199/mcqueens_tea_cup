package entity

// OBRankingEntry is a single Online Battle ranking row with its rank name and
// points already resolved against the ranking config. It carries no Discord
// formatting so the presentation layer decides how to render it.
type OBRankingEntry struct {
	Rank      string
	Name      string
	RankName  string
	Point     string
	StarCount string
	IsPride   bool
}

// OBRankingView is the resolved, limit-applied result of an Online Battle
// ranking lookup, ready for a presenter to paginate and format.
type OBRankingView struct {
	CalcDate string
	Entries  []OBRankingEntry
}
