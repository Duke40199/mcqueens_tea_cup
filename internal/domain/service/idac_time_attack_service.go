package service

import (
	"context"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type IDACTimeAttackService struct {
	cfg config.Config
	// repos
	carRepo          database.CarRepository
	timeMetadataRepo database.TATimeMetadataRepository
	// clients
	segaClient port.SegaIDACClient
}

func NewIDACTimeAttackService(
	cfg config.Config,
	timeMetadataRepo database.TATimeMetadataRepository,
	segaClient port.SegaIDACClient,
) port.IDACTimeAttackService {
	return &IDACTimeAttackService{
		cfg:              cfg,
		timeMetadataRepo: timeMetadataRepo,
		segaClient:       segaClient,
	}
}

func (s *IDACTimeAttackService) GetMetadataBySegaCourseID(ctx context.Context, segaCourseID string) ([]*entity.TimeAttackRankingMetadata, error) {
	timeAttackMetadata, err := s.timeMetadataRepo.GetByCourseID(ctx, segaCourseID)
	if err != nil {
		return nil, err
	}
	return timeAttackMetadata, nil
}

// GetTimeTrail returns the raw Time Trial records for a course/area/car from the
// Sega client, keeping the transport concern out of the presentation layer.
func (s *IDACTimeAttackService) GetTimeTrail(ctx context.Context, courseID, area, carID, spec string) ([]entity.TimeAttackRecord, error) {
	return s.segaClient.GetListTimeTrail(courseID, area, carID, spec)
}
