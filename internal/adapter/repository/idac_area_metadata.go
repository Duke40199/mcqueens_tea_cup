package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type IDACAreaMetadataRepository struct {
	DB        *sql.DB
	tableName string
}

func NewIDACAreaMetadataRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.IDACAreaMetadataRepository {
	return &IDACAreaMetadataRepository{
		DB:        db,
		tableName: tables.IDACAreaMetadata,
	}
}

func (r *IDACAreaMetadataRepository) GetAll(ctx context.Context) ([]entity.IDACAreaMetadata, error) {
	// Explicit column list (in scan order) so an added/reordered column can't
	// silently corrupt the positional Scan below.
	query := fmt.Sprintf(
		`SELECT id, sega_area_code, name, aliases, created_at, updated_at, all_net_area_code, area_type FROM %s`,
		r.tableName,
	)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []entity.IDACAreaMetadata
	for rows.Next() {
		var area entity.IDACAreaMetadata
		if err := rows.Scan(
			&area.ID,
			&area.SegaAreaCode,
			&area.Name,
			&area.Aliases,
			&area.CreatedAt,
			&area.UpdatedAt,
			&area.ALLNetCode,
			&area.AreaType); err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}
	return areas, rows.Err()
}
