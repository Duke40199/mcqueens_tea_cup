package entity

// PlayerProfileView holds a player's resolved account grade and Online Battle
// rank, ready for the presentation layer to render.
type PlayerProfileView struct {
	GradeName   string
	GradeNum    string
	OBRankName  string
	OBStarCount string
}

// TournamentInfoView is the resolved, sorted set of players for the tournament
// info command, along with the source calculation date.
type TournamentInfoView struct {
	CalcDate string
	Players  []*PlayerTournamentInfo
}
