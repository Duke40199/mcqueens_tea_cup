package db

import (
	"context"
	"database/sql"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type AreaRepository struct {
	DB *sql.DB
}

func NewAreaRepository(db *sql.DB) postgres.AreaRepository {
	return &AreaRepository{
		DB: db,
	}
}

func (r *AreaRepository) GetOBActiveAreas(ctx context.Context) ([]entity.AreaSyncInfo, error) {
	query := `SELECT sega_code, name, COALESCE(timezone, 'Asia/Tokyo') FROM sega_idac_area_codes_metadata WHERE is_cron_ob_active_status = true`
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
