package db

import (
	"context"
	"database/sql"
	"fmt"

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
		INSERT INTO %s (name, address, country_code, sega_area_code, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id;`, storeLocEntity.TableName())
	var newID int64
	err := r.DB.QueryRow(query,
		storeLocEntity.Name,
		storeLocEntity.Address,
		storeLocEntity.CountryCode,
		storeLocEntity.SegaAreaCode).Scan(&newID)
	return newID, err
}
