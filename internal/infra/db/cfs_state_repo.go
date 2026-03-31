package db

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type CfsStateRepo struct {
	DB *sql.DB
}

func NewCfsStateRepository(db *sql.DB) postgres.CfsStateRepository {
	return &CfsStateRepo{DB: db}
}

func (r *CfsStateRepo) GetLatestCfsState(ctx context.Context) (*entity.CfsState, error) {
	query := `SELECT id FROM cfs_state ORDER BY created_at DESC LIMIT 1;`
	row := r.DB.QueryRow(query)

	var cfsState entity.CfsState
	err := row.Scan(&cfsState.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return nil, err
	}
	return &cfsState, nil
}

func (r *CfsStateRepo) CreateCfsState(id int64, discordID, content string) error {
	// UPSERT: Insert, but if conflict (key exists), update the existing row
	query := `
		INSERT INTO cfs_state (id, discord_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (id) 
		DO NOTHING;
	`
	_, err := r.DB.Exec(query, id, discordID, content)
	return err
}
