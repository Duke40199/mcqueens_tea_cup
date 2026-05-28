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
	segaIDACClient    port.SegaIDACClient
	storeLocationRepo database.AllNetStoreLocationsRepository
}

func NewIDACStoreLocationService(
	cfg config.Config,
	allNetClient port.AllNetClient,
	segaIDACCLient port.SegaIDACClient,
	storeLocationRepo database.AllNetStoreLocationsRepository,
) port.IDACStoreLocationService {
	return &StoreLocationService{
		cfg:               cfg,
		allNetClient:      allNetClient,
		segaIDACClient:    segaIDACCLient,
		storeLocationRepo: storeLocationRepo,
	}
}

func (s *StoreLocationService) BulkUpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) error {
	return s.storeLocationRepo.BulkUpsertStoreLocation(ctx, listStoreLocations)
}

func (s *StoreLocationService) GetListAllNextStore(ctx context.Context, areaCode string) ([]entity.StoreLocation, string, error) {
	return s.allNetClient.GetListStore(ctx, s.cfg.GetAllNetClientCfg().IDACGameCode, s.cfg.GetAllNetClientCfg().EnglishLanguageCode, areaCode)
}

func (s *StoreLocationService) GetMapStoreFromTopPlayers(ctx context.Context, areaCode string) (map[string]entity.StoreLocation, error) {
	listTopPlayers, err := s.segaIDACClient.GetListPlayerGrade(areaCode)
	if err != nil {
		return map[string]entity.StoreLocation{}, err
	}
	if len(listTopPlayers.Records) == 0 {
		return map[string]entity.StoreLocation{}, nil
	}
	// key: shopName
	mapFoundStore := map[string]entity.StoreLocation{}
	for _, player := range listTopPlayers.Records {
		if _, ok := mapFoundStore[player.ShopName]; !ok {
			mapFoundStore[player.ShopName] = entity.StoreLocation{
				Name: player.ShopName,
			}
			continue
		}
	}
	return mapFoundStore, nil

}
