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
	allNetClient      port.AllNetClient
	storeLocationRepo database.AllNetStoreLocationsRepository
}

func NewIDACStoreLocationService(
	cfg config.Config,
	allNetClient port.AllNetClient,
	storeLocationRepo database.AllNetStoreLocationsRepository,
) port.IDACStoreLocationService {
	return &StoreLocationService{
		cfg:               cfg,
		allNetClient:      allNetClient,
		storeLocationRepo: storeLocationRepo,
	}
}

func (s *StoreLocationService) UpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) (int64, error) {
	return s.storeLocationRepo.UpsertStoreLocation(ctx, listStoreLocations[0])
}

func (s *StoreLocationService) GetListAllNextStore(ctx context.Context, areaCode string) ([]entity.StoreLocation, string, error) {
	return s.allNetClient.GetListStore(ctx, s.cfg.GetAllNetClientCfg().IDACGameCode, s.cfg.GetAllNetClientCfg().EnglishLanguageCode, areaCode)
}
