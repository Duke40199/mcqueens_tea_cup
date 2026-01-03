package db

import (
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/domain"

	_ "github.com/lib/pq"
)

type AliasRepo struct {
	DB *sql.DB
}

// NewPostgresAliasRepo returns the struct that satisfies AliasRepository
func NewPostgresAliasRepo(db *sql.DB) *AliasRepo {
	return &AliasRepo{DB: db}
}

// Get fetches alias from DB
func (a *AliasRepo) GetByAliasKey(key string) (domain.PlayerAlias, bool, error) {
	query := `SELECT ign, area FROM player_alias WHERE alias_key = $1`

	row := a.DB.QueryRow(query, key)

	var alias domain.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PlayerAlias{}, false, err // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return domain.PlayerAlias{}, false, err
	}

	return alias, true, nil
}

func (a *AliasRepo) GetByIgn(ign string) (domain.PlayerAlias, bool) {
	query := `SELECT ign, area FROM player_alias WHERE ign = $1`

	row := a.DB.QueryRow(query, ign)

	var alias domain.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PlayerAlias{}, false // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return domain.PlayerAlias{}, false
	}

	return alias, true
}

// Set inserts or updates alias
func (a *AliasRepo) Set(key, ign, area string) error {
	// UPSERT: Insert, but if conflict (key exists), update the existing row
	query := `
		INSERT INTO player_aliases (alias_key, ign, area, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (alias_key) 
		DO UPDATE SET ign = EXCLUDED.ign, area = EXCLUDED.area, updated_at = NOW();
	`
	_, err := a.DB.Exec(query, key, ign, area)
	return err
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (a *AliasRepo) Load() error {
	return nil
}
