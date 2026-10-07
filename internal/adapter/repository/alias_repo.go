package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/pkg/logger"

	_ "github.com/lib/pq"
)

type AliasRepository struct {
	DB        *sql.DB
	tableName string
}

// NewAliasRepo returns the struct that satisfies AliasRepository
func NewAliasRepo(db *sql.DB, tables config.DatabaseTablesConfig) database.AliasRepository {
	return &AliasRepository{
		DB:        db,
		tableName: tables.PlayerAlias,
	}
}

// GetByAliasKey fetches alias from DB
func (a *AliasRepository) GetByAliasKey(ctx context.Context, key string) (entity.PlayerAlias, bool, error) {
	query := fmt.Sprintf(`SELECT ign, area FROM %s WHERE alias_key = $1`, a.tableName)

	row := a.DB.QueryRowContext(ctx, query, key)

	var alias entity.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.PlayerAlias{}, false, err // Not found
		}
		logger.Error(ctx, "failed to get alias by key", err)
		return entity.PlayerAlias{}, false, err
	}

	return alias, true, nil
}

func (a *AliasRepository) GetByIgnAndAreaCode(ctx context.Context, ign, areaCode string) (entity.PlayerAlias, bool, error) {
	query := fmt.Sprintf(`SELECT ign, area FROM %s
              WHERE lower(normalize(ign, NFKC)) = lower(normalize($1, NFKC))
              AND area = $2
              LIMIT 1`, a.tableName)
	row := a.DB.QueryRowContext(ctx, query, ign, areaCode)

	var alias entity.PlayerAlias
	err := row.Scan(&alias.Ign, &alias.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.PlayerAlias{}, false, err // Not found
		}
		logger.Error(ctx, "failed to get alias by ign and area", err)
		return entity.PlayerAlias{}, false, err
	}

	return alias, true, nil
}

// SetPlayerAlias inserts or updates alias
func (a *AliasRepository) SetPlayerAlias(key, ign, area string) error {
	// UPSERT: Insert, but if conflict (key exists), update the existing row
	query := fmt.Sprintf(`
		INSERT INTO %s (alias_key, ign, area, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (alias_key)
		DO UPDATE SET ign = EXCLUDED.ign, area = EXCLUDED.area, updated_at = NOW();
	`, a.tableName)
	_, err := a.DB.Exec(query, key, ign, area)
	return err
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (a *AliasRepository) Load() error {
	return nil
}
