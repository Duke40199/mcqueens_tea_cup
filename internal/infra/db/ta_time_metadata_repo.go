package db

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"

	_ "github.com/lib/pq"
)

const taTimeMetadataTableName = "sega_idac_ta_time_metadata"

type TATimeMetadataRepository struct {
	DB *sql.DB
}

// NewPostgresTATimeMetadataRepository returns the struct that satisfies TATimeMetadataRepository
func NewPostgresTATimeMetadataRepository(db *sql.DB) postgres.TATimeMetadataRepository {
	return &TATimeMetadataRepository{
		DB: db,
	}
}

// GetByCourseID fetches time ranking from DB
func (r *TATimeMetadataRepository) GetByCourseID(ctx context.Context, courseID string) ([]*entity.TimeAttackRankingMetadata, error) {
	// 1. Debug exactly what is being passed
	query := `
        SELECT m.id, m.required_time, r.name
        FROM sega_idac_ta_time_metadata m
        LEFT JOIN cfg_player_ranking r ON m.rank_id::uuid = r.id
        WHERE m.course_id = $1
        ORDER BY m.required_time ASC`

	rows, err := r.DB.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var listCfg []*entity.TimeAttackRankingMetadata
	var rowCount int

	for rows.Next() {
		rowCount++
		var cfg entity.TimeAttackRankingMetadata
		if err := rows.Scan(&cfg.ID, &cfg.RequiredTime, &cfg.RankName); err != nil {
			return nil, fmt.Errorf("failed to scan row %d: %w", rowCount, err)
		}
		listCfg = append(listCfg, &cfg)
	}
	// 2. Catch errors that happened DURING the loop
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}
	// 3. Debug how many rows were actually found
	fmt.Printf("[DEBUG] Query successful. Found %d rows.\n", rowCount)

	return listCfg, nil
}

// Load is not needed for DB (Query on demand), so we leave it empty to satisfy interface
func (r *TATimeMetadataRepository) Load() error {
	return nil
}
