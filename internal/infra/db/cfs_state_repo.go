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

func (r *CfsStateRepo) CreateCfsState(discordID, content string) (int64, error) {
	// Let PostgreSQL automatically generate the next `id` using SERIAL,
	// and then immediately return that new `id` back to us.
	query := `
		INSERT INTO cfs_state (discord_id, content, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id;
	`
	var newID int64
	err := r.DB.QueryRow(query, discordID, content).Scan(&newID)

	return newID, err
}
