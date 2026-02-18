package db

import (
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"

	_ "github.com/lib/pq"
)

type AliasRepository struct {
	DB *sql.DB
}

// NewPostgresAliasRepo returns the struct that satisfies AliasRepository
func NewPostgresAliasRepo(db *sql.DB) postgres.AliasRepository {
	return &AliasRepository{
		DB: db,
	}
}

// GetByAliasKey fetches alias from DB
func (a *AliasRepository) GetByAliasKey(key string) (entity.PlayerAlias, bool, error) {
	query := `SELECT ign, area FROM player_alias WHERE alias_key = $1`

	row := a.DB.QueryRow(query, key)

	var alias entity.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.PlayerAlias{}, false, err // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return entity.PlayerAlias{}, false, err
	}

	return alias, true, nil
}

func (a *AliasRepository) GetByIgnAndAreaCode(ign, areaCode string) (entity.PlayerAlias, bool, error) {
	query := `SELECT ign, area FROM player_alias 
              WHERE lower(normalize(ign, NFKC)) = lower(normalize($1, NFKC))
              AND area = $2
              LIMIT 1`
	row := a.DB.QueryRow(query, ign, areaCode)

	var alias entity.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.PlayerAlias{}, false, err // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return entity.PlayerAlias{}, false, err
	}

	return alias, true, nil
}

// SetPlayerAlias inserts or updates alias
func (a *AliasRepository) SetPlayerAlias(key, ign, area string) error {
	// UPSERT: Insert, but if conflict (key exists), update the existing row
	query := `
		INSERT INTO player_alias (alias_key, ign, area, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (alias_key) 
		DO UPDATE SET ign = EXCLUDED.ign, area = EXCLUDED.area, updated_at = NOW();
	`
	_, err := a.DB.Exec(query, key, ign, area)
	return err
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (a *AliasRepository) Load() error {
	return nil
}
