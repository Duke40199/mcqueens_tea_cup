package usecase

import (
	"context"
	"fmt"
	"log"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/google/uuid"
)

type CarSyncService struct {
	SegaClient entity.SegaClient
	CarRepo    postgres.CarRepository
}

func NewCarSyncService(client entity.SegaClient, repo postgres.CarRepository) *CarSyncService {
	return &CarSyncService{
		SegaClient: client,
		CarRepo:    repo,
	}
}

func (s *CarSyncService) SyncData(ctx context.Context) error {
	log.Println("🚗 Starting Car/Style Sync...")
	// 0. Fetch existing car mappings to reuse UUIDs
	existingCarMap, err := s.CarRepo.GetSegaIDToUUIDMap(ctx)
	if err != nil {
		log.Printf("⚠️ Warning: Could not fetch existing car map: %v. Proceeding with new UUIDs.", err)
		existingCarMap = make(map[int64]string)
	}

	// 1. Fetch data
	data, err := s.SegaClient.FetchConst()
	if err != nil {
		return fmt.Errorf("failed to fetch const data: %w", err)
	}
	log.Printf("Fetched %d cars and %d styles. Saving to DB...", len(data.Cars), len(data.Styles))
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
		log.Printf("Car: %s, Model Code: %s, Maker: %s, Style: %s", car.Name, car.ModelCode, car.Maker, car.BaseStyleName)
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
		log.Printf("Style: %s, Car ID: %s, Style Name: %s", segaStyle.Name, segaStyle.CarID, segaStyle.RouteStyleName)
		foundStyles = append(foundStyles, segaStyle)
	}
	// 5. Upsert style data
	if err := s.CarRepo.UpsertCarStyles(ctx, foundStyles); err != nil {
		return fmt.Errorf("failed to upsert styles: %w", err)
	}
	log.Printf("✅ Car/Style Sync Completed")
	return nil
}
