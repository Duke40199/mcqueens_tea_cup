package entity

type OBRankingCfg struct {
	ID     string  `json:"id"`
	SegaID string  `json:"sega_id"`
	Name   string  `json:"name"`
	Emoji  *string `json:"emoji,omitempty"`
}
