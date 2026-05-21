package port

import (
	"McQueens_Tea_Cup/internal/domain/entity"
	"context"
)

type StoreLocationService interface {
	// Define the methods for the StoreLocationService
	GetListAllNextStore(ctx context.Context, areaCode string) ([]entity.StoreLocation, string, error)
	// UpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) (int64, error)
}
