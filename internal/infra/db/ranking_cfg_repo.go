package db

import (
	"context"
	"database/sql"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"

	_ "github.com/lib/pq"
)

type RankingCfgRepository struct {
	DB *sql.DB
}

// NewRankingCfgRepo returns the struct that satisfies RankingCfgRepository
func NewRankingCfgRepo(db *sql.DB) database.RankingCfgRepository {
	return &RankingCfgRepository{
		DB: db,
	}
}

// GetListTimeAttackRankingCfg fetches time ranking from DB
func (r *RankingCfgRepository) GetListTimeAttackRankingCfg(ctx context.Context) ([]*entity.TimeAttackRankingCfg, error) {
	query := `SELECT id, name FROM cfg_player_ranking WHERE type = $1`
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
	return listCfg, nil
}

// GetListPlayerGradeCfg fetches player grade cfg from DB
func (r *RankingCfgRepository) GetListPlayerGradeCfg(ctx context.Context) ([]*entity.PlayerGradeCfg, error) {
	query := `SELECT id, type, name, sega_id FROM cfg_player_ranking WHERE type IN ($1, $2)`
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
	return listCfg, nil
}

// GetListPlayerGradeCfg fetches player grade cfg from DB
func (r *RankingCfgRepository) GetPlayerGradeBySegaIDs(ctx context.Context, gradeSegaID, gradeNumSegaID string) ([]*entity.PlayerGradeCfg, error) {
	query := `SELECT id, type, name, sega_id, emoji FROM cfg_player_ranking WHERE sega_id IN ($1, $2)`
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
	return listCfg, nil
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (r *RankingCfgRepository) Load() error {
	return nil
}
