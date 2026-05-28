package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type AllNetStoreLocationsRepository struct {
	DB *sql.DB
}

func NewAllNetStoreLocationsRepository(db *sql.DB) database.AllNetStoreLocationsRepository {
	return &AllNetStoreLocationsRepository{
		DB: db,
	}
}

func (r *AllNetStoreLocationsRepository) UpsertStoreLocation(ctx context.Context, storeLocEntity entity.StoreLocation) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, address, sega_area_code, all_net_area_code, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id;`, storeLocEntity.TableName())
	var newID int64
	err := r.DB.QueryRow(query,
		storeLocEntity.Name,
		storeLocEntity.Address,
		storeLocEntity.SegaAreaCode,
		storeLocEntity.AllNetAreaCode).Scan(&newID)
	return newID, err
}

func (r *AllNetStoreLocationsRepository) BulkUpsertStoreLocation(ctx context.Context, stores []entity.StoreLocation) error {
	if len(stores) == 0 {
		return nil
	}
	tableName := stores[0].TableName()
	var valueStrings []string
	var valueArgs []interface{}
	for i, store := range stores {
		// We multiply the index by 5 since there are 5 columns per store
		n := i * 4
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())", n+1, n+2, n+3, n+4))
		valueArgs = append(valueArgs, store.Name, store.Address, store.SegaAreaCode, store.AllNetAreaCode)
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (name, address, sega_area_code, all_net_area_code, created_at)
		VALUES %s
		ON CONFLICT (name) DO UPDATE SET
			address = EXCLUDED.address,
			sega_area_code = EXCLUDED.sega_area_code,
			all_net_area_code = EXCLUDED.all_net_area_code;`, tableName, strings.Join(valueStrings, ","))
	_, err := r.DB.ExecContext(ctx, query, valueArgs...)
	return err
}
