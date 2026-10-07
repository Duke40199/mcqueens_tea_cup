package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/pkg/utils"
)

type AllNetStoreLocationsRepository struct {
	DB        *sql.DB
	tableName string
}

func NewAllNetStoreLocationsRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.AllNetStoreLocationsRepository {
	return &AllNetStoreLocationsRepository{
		DB:        db,
		tableName: tables.IDACStores,
	}
}

func (r *AllNetStoreLocationsRepository) UpsertStoreLocation(ctx context.Context, storeLocEntity entity.StoreLocation) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, address, sega_area_code, all_net_area_code, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id;`, r.tableName)
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

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op once Commit succeeds

	// Chunk so we never exceed Postgres's 65535 bind-parameter limit (4 params/row).
	for _, batch := range utils.ChunkSlice(stores, pgMaxBulkRows) {
		var valueStrings []string
		var valueArgs []interface{}
		for i, store := range batch {
			n := i * 4 // 4 bound columns per store (created_at uses NOW())
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())", n+1, n+2, n+3, n+4))
			valueArgs = append(valueArgs, store.Name, store.Address, store.SegaAreaCode, store.AllNetAreaCode)
		}
		query := fmt.Sprintf(`
			INSERT INTO %s (name, address, sega_area_code, all_net_area_code, created_at)
			VALUES %s
			ON CONFLICT (name) DO UPDATE SET
				address = EXCLUDED.address,
				sega_area_code = EXCLUDED.sega_area_code,
				all_net_area_code = EXCLUDED.all_net_area_code;`, r.tableName, strings.Join(valueStrings, ","))
		if _, err := tx.ExecContext(ctx, query, valueArgs...); err != nil {
			return err
		}
	}
	return tx.Commit()
}
