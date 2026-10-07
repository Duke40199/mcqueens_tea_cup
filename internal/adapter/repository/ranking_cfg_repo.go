package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"

	_ "github.com/lib/pq"
)

type RankingCfgRepository struct {
	DB        *sql.DB
	tableName string
}

// NewRankingCfgRepo returns the struct that satisfies RankingCfgRepository
func NewRankingCfgRepo(db *sql.DB, tables config.DatabaseTablesConfig) port.RankingCfgRepository {
	return &RankingCfgRepository{
		DB:        db,
		tableName: tables.CfgPlayerRanking,
	}
}

// GetListTimeAttackRankingCfg fetches time ranking from DB
func (r *RankingCfgRepository) GetListTimeAttackRankingCfg(ctx context.Context) ([]*entity.TimeAttackRankingCfg, error) {
	query := fmt.Sprintf(`SELECT id, name FROM %s WHERE type = $1`, r.tableName)
	rows, err := r.DB.QueryContext(ctx, query, "RANK_TIME_ATTACK")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listCfg []*entity.TimeAttackRankingCfg
	for rows.Next() {
		var cfg entity.TimeAttackRankingCfg
		if err := rows.Scan(&cfg.ID, &cfg.Name); err != nil {
			return nil, err
		}
		listCfg = append(listCfg, &cfg)
	}
	return listCfg, rows.Err()
}

// GetListPlayerGradeCfg fetches player grade cfg from DB
func (r *RankingCfgRepository) GetListPlayerGradeCfg(ctx context.Context) ([]*entity.PlayerGradeCfg, error) {
	query := fmt.Sprintf(`SELECT id, type, name, sega_id FROM %s WHERE type IN ($1, $2)`, r.tableName)
	rows, err := r.DB.QueryContext(ctx, query, "RANK_NUMBER", "GRADE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listCfg []*entity.PlayerGradeCfg
	for rows.Next() {
		var cfg entity.PlayerGradeCfg
		if err := rows.Scan(&cfg.ID, &cfg.Type, &cfg.Name, &cfg.SegaID); err != nil {
			return nil, err
		}
		listCfg = append(listCfg, &cfg)
	}
	return listCfg, rows.Err()
}

// GetListPlayerGradeCfg fetches player grade cfg from DB
func (r *RankingCfgRepository) GetPlayerGradeBySegaIDs(ctx context.Context, gradeSegaID, gradeNumSegaID string) ([]*entity.PlayerGradeCfg, error) {
	query := fmt.Sprintf(`SELECT id, type, name, sega_id, emoji FROM %s WHERE sega_id IN ($1, $2)`, r.tableName)
	rows, err := r.DB.QueryContext(ctx, query, gradeSegaID, gradeNumSegaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listCfg []*entity.PlayerGradeCfg
	for rows.Next() {
		var cfg entity.PlayerGradeCfg
		if err := rows.Scan(&cfg.ID, &cfg.Type, &cfg.Name, &cfg.SegaID, &cfg.Emoji); err != nil {
			return nil, err
		}
		listCfg = append(listCfg, &cfg)
	}
	return listCfg, rows.Err()
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (r *RankingCfgRepository) Load() error {
	return nil
}
