package service

import (
	"context"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type StoreLocationService struct {
	cfg               config.Config
	storeLocationRepo database.AllNetStoreLocationsRepository
}

func NewStoreLocationService(
	cfg config.Config,
	storeLocationRepo database.AllNetStoreLocationsRepository,
) port.StoreLocationService {
	return &StoreLocationService{
		cfg:               cfg,
		storeLocationRepo: storeLocationRepo,
	}
}

func (s *StoreLocationService) UpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) (int64, error) {
	return s.storeLocationRepo.UpsertStoreLocation(ctx, listStoreLocations[0])
}

func (s *StoreLocationService) GetListStoreFromAllNet(ctx context.Context, areaCode string) ([]entity.StoreLocation, error) {
	return []entity.StoreLocation{}, nil
}
