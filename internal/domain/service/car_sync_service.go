package service

import (
	"context"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"

	"github.com/google/uuid"
)

type CarSyncService struct {
	SegaClient port.SegaIDACClient
	CarRepo    database.CarRepository
}

func NewCarSyncService(client port.SegaIDACClient, repo database.CarRepository) *CarSyncService {
	return &CarSyncService{
		SegaClient: client,
		CarRepo:    repo,
	}
}

func (s *CarSyncService) SyncData(ctx context.Context) error {
	logger.Info(ctx, "starting car/style sync")
	// 0. Fetch existing car mappings to reuse UUIDs
	existingCarMap, err := s.CarRepo.GetSegaIDToUUIDMap(ctx)
	if err != nil {
		logger.Warn(ctx, fmt.Sprintf("could not fetch existing car map, proceeding with new UUIDs: %v", err))
		existingCarMap = make(map[int64]string)
	}

	// 1. Fetch data
	data, err := s.SegaClient.FetchConst(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch const data: %w", err)
	}
	logger.Info(ctx, fmt.Sprintf("fetched %d cars and %d styles, saving to DB", len(data.Cars), len(data.Styles)))
	// 2. Normalize Sega's car data
	foundCars := make([]entity.CarMetadata, 0)
	carStyleIDsMap := make(map[int64]string)
	for _, car := range data.Cars {
		// Reuse existing UUID if available
		if existingID, ok := existingCarMap[car.SegaCarID]; ok {
			car.ID = existingID
		} else {
			car.ID = uuid.NewString()
		}
		car.Maker = car.GetNormalizedMakerName()
		car.BaseStyleName = car.GetNormalizedBaseStyle()
		car.ModelCode = car.GetCarModelCode()
		car.Name = car.GetNormalizedCarName()
		logger.Debug(ctx, fmt.Sprintf("car: %s, model code: %s, maker: %s, style: %s", car.Name, car.ModelCode, car.Maker, car.BaseStyleName))
		foundCars = append(foundCars, car)
		for _, styleID := range car.CarStyleIDs {
			carStyleIDsMap[styleID] = car.ID
		}
	}
	// 3. Upsert car data
	if err := s.CarRepo.UpsertCars(ctx, foundCars); err != nil {
		return fmt.Errorf("failed to save cars: %w", err)
	}
	foundStyles := make([]entity.CarStyleMetadata, 0)
	// 4. Normalize Sega's style data
	for _, segaStyle := range data.Styles {
		segaStyle.ID = uuid.NewString()
		segaStyle.CarID = carStyleIDsMap[segaStyle.StyleCarID]
		logger.Debug(ctx, fmt.Sprintf("style: %s, car id: %s, style name: %s", segaStyle.Name, segaStyle.CarID, segaStyle.RouteStyleName))
		foundStyles = append(foundStyles, segaStyle)
	}
	// 5. Upsert style data
	if err := s.CarRepo.UpsertCarStyles(ctx, foundStyles); err != nil {
		return fmt.Errorf("failed to upsert styles: %w", err)
	}
	logger.Info(ctx, "car/style sync completed")
	return nil
}
