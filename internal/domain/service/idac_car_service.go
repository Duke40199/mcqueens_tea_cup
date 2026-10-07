package service

import (
	"context"
	"sort"
	"strings"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
)

type IDACCarService struct {
	cfg config.Config
	// repos
	carRepo           port.CarRepository
	storeLocationRepo port.AllNetStoreLocationsRepository
	// clients
	segaClient port.SegaIDACClient
}

func NewIDACCarService(
	cfg config.Config,
	carRepo port.CarRepository,
	segaClient port.SegaIDACClient,
) port.IDACCarService {
	return &IDACCarService{
		cfg:        cfg,
		carRepo:    carRepo,
		segaClient: segaClient,
	}
}

func (s *IDACCarService) GetListCarDetailByTAFormat(ctx context.Context, listCarSegaFormat []string) (map[string]entity.CarSpecInfo, error) {
	aliasSpecMap := make(map[string]string, 0)
	// sega format: FD3S[DH]
	for _, segaFormat := range listCarSegaFormat {
		// 0: FD3S, 1: DH
		splitStr := strings.Split(segaFormat, "[")
		start := strings.LastIndex(segaFormat, "[")
		end := strings.LastIndex(segaFormat, "]")
		if start == -1 || end == -1 || end <= start {
			aliasSpecMap[splitStr[0]] = "" // No valid brackets found
			continue                       // avoid slicing with start/end == -1
		}
		aliasSpecMap[splitStr[0]] = segaFormat[start+1 : end]
	}
	specInfo, err := s.carRepo.GetCarWithSpecsByAliases(ctx, aliasSpecMap)
	if err != nil {
		return nil, err
	}
	return specInfo, nil
}

func (s *IDACCarService) GetListTopTACarsWithPercentage(ctx context.Context, segaCourseID string, resultCount int64) ([]entity.IDACCarUsagePercentage, error) {
	records, err := s.segaClient.GetListTimeTrail(ctx, segaCourseID, "area-all", "car-all", "")
	if err != nil {
		logger.Error(ctx, "failed to fetch time attack data from Sega", err)
		return nil, err
	}
	if len(records) == 0 {
		logger.Info(ctx, "no time attack records found")
		return []entity.IDACCarUsagePercentage{}, nil
	}
	return calculateCarUsagePercentage(records), nil
}

func calculateCarUsagePercentage(records []entity.TimeAttackRecord) []entity.IDACCarUsagePercentage {
	carUsage := make(map[string]entity.IDACCarUsagePercentage)
	for _, r := range records {
		if _, ok := carUsage[r.CarName]; ok {
			car := carUsage[r.CarName]
			car.Count++
			carUsage[r.CarName] = car
		} else {
			carPercentage := entity.IDACCarUsagePercentage{
				SegaCarName: r.CarName,
				Count:       1,
				Percentage:  0,
			}
			splitName := strings.Split(r.CarName, "[")
			carPercentage.CarName = splitName[0]
			carUsage[r.CarName] = carPercentage
		}
	}
	total := len(records)
	var sortedCars []entity.IDACCarUsagePercentage
	for _, car := range carUsage {
		car.Percentage = (float64(car.Count) / float64(total)) * 100
		sortedCars = append(sortedCars, car)
	}
	// Sort slice
	sort.Slice(sortedCars, func(i, j int) bool {
		return sortedCars[i].Percentage > sortedCars[j].Percentage
	})
	return sortedCars
}
