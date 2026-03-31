package db

import (
	"context"
	"database/sql"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"

	_ "github.com/lib/pq"
)

type RankingCfgRepository struct {
	DB *sql.DB
}

// NewPostgresRankingCfgRepo returns the struct that satisfies RankingCfgRepository
func NewPostgresRankingCfgRepo(db *sql.DB) postgres.RankingCfgRepository {
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

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (r *RankingCfgRepository) Load() error {
	return nil
}
