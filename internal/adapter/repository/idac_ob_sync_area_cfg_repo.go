package repository

import (
	"context"
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type AreaRepository struct {
	DB        *sql.DB
	tableName string
}

func NewAreaRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.AreaRepository {
	return &AreaRepository{
		DB:        db,
		tableName: tables.IDACOBSyncAreaCfg,
	}
}

func (r *AreaRepository) GetOBActiveAreas(ctx context.Context) ([]entity.AreaSyncInfo, error) {
	query := fmt.Sprintf(`SELECT sega_code, name, COALESCE(timezone, 'Asia/Tokyo') FROM %s WHERE is_cron_ob_active_status = true`, r.tableName)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []entity.AreaSyncInfo
	for rows.Next() {
		var area entity.AreaSyncInfo
		if err := rows.Scan(&area.AreaCode, &area.AreaName, &area.Timezone); err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}
	return areas, nil
}
