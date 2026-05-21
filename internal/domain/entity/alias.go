package entity

import (
	"time"

	"github.com/google/uuid"
)

// PlayerAlias represents the saved data
type PlayerAlias struct {
	ID        uuid.UUID `json:"id"`
	Ign       string    `json:"ign"`
	AliasKey  string    `json:"alias_key"`
	Area      string    `json:"area"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
