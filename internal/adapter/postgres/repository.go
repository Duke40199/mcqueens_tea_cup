package postgres

import "McQueens_Tea_Cup/internal/domain/entity"

// AliasRepository defines how we interact with player aliases.
type AliasRepository interface {
	GetByAliasKey(discordID string) (entity.PlayerAlias, bool, error)
	GetByIgn(ign string) (entity.PlayerAlias, bool)

	SetPlayerAlias(discordID, ign, area string) error
	Load() error
}
