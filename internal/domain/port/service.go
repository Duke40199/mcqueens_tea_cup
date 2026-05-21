package port

import (
	"McQueens_Tea_Cup/internal/domain/entity"
	"context"
)

type StoreLocationService interface {
	// Define the methods for the StoreLocationService
	UpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) (int64, error)
}
