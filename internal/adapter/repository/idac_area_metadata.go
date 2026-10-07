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
	query := fmt.Sprintf(`SELECT * FROM %s`, r.tableName)
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
	return areas, nil
}
