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

type OBRankingCfgRepository struct {
	DB        *sql.DB
	tableName string
}

// NewOBRankingCfgRepository returns the struct that satisfies AliasRepository
func NewOBRankingCfgRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.OBRankingCfgRepository {
	return &OBRankingCfgRepository{
		DB:        db,
		tableName: tables.OBRankingCfg,
	}
}

// GetByAliasKey fetches alias from DB
func (o *OBRankingCfgRepository) GetBySegaID(ctx context.Context, key string) (*entity.OBRankingCfg, error) {
	// Explicit columns matching the scan below (id, sega_id, name). Using SELECT *
	// returned all 4 columns while Scan only consumed 3, erroring on every call.
	query := fmt.Sprintf(`SELECT id, sega_id, name FROM %s WHERE sega_id = $1`, o.tableName)

	row := o.DB.QueryRowContext(ctx, query, key)

	var cfg entity.OBRankingCfg
	err := row.Scan(&cfg.ID, &cfg.SegaID, &cfg.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Not found
		}
		logger.Error(ctx, "failed to get ob ranking cfg by sega id", err)
		return nil, err
	}

	return &cfg, nil
}

func (o *OBRankingCfgRepository) GetRankingCfgMap(ctx context.Context) (map[string]entity.OBRankingCfg, error) {
	// Explicit columns (in scan order) so an added/reordered column can't corrupt
	// the positional Scan below.
	query := fmt.Sprintf(`SELECT id, name, sega_id, emoji FROM %s`, o.tableName)

	rows, err := o.DB.QueryContext(ctx, query)
	if err != nil {
		logger.Error(ctx, "failed to query ob ranking cfg map", err)
		return nil, err
	}
	defer rows.Close()

	var cfgMap = make(map[string]entity.OBRankingCfg)
	for rows.Next() {
		var cfg entity.OBRankingCfg
		err = rows.Scan(&cfg.ID, &cfg.Name, &cfg.SegaID, &cfg.Emoji)
		if err != nil {
			logger.Error(ctx, "failed to scan ob ranking cfg row", err)
			return nil, err
		}
		cfgMap[cfg.SegaID] = cfg
	}
	return cfgMap, rows.Err()
}
