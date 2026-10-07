package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type CfsStateRepo struct {
	DB        *sql.DB
	tableName string
}

func NewCfsStateRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.CfsStateRepository {
	return &CfsStateRepo{DB: db, tableName: tables.CfsState}
}

func (r *CfsStateRepo) GetLatestCfsState(ctx context.Context) (*entity.CfsState, error) {
	query := fmt.Sprintf(`SELECT id FROM %s ORDER BY created_at DESC LIMIT 1;`, r.tableName)
	row := r.DB.QueryRowContext(ctx, query)

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

func (r *CfsStateRepo) CreateCfsState(ctx context.Context, discordID, content string) (int64, error) {
	// Let databaseQL automatically generate the next `id` using SERIAL,
	// and then immediately return that new `id` back to us.
	query := fmt.Sprintf(`
		INSERT INTO %s (discord_id, content, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id;
	`, r.tableName)
	var newID int64
	err := r.DB.QueryRowContext(ctx, query, discordID, content).Scan(&newID)

	return newID, err
}
