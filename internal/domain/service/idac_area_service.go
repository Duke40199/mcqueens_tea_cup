package service

import (
	"context"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type IDACAreaService struct {
	cfg config.Config
	// repos
	idacAreaMetadataRepo port.IDACAreaMetadataRepository
	// clients
	segaClient port.SegaIDACClient
}

func NewIDACAreaService(
	cfg config.Config,
	idacAreaMetadataRepo port.IDACAreaMetadataRepository,
	segaClient port.SegaIDACClient,
) port.IDACAreaService {
	return &IDACAreaService{
		cfg:                  cfg,
		idacAreaMetadataRepo: idacAreaMetadataRepo,
		segaClient:           segaClient,
	}
}

func (s *IDACAreaService) GetAreaMetadata(ctx context.Context) ([]entity.IDACAreaMetadata, error) {
	areaMetadata, err := s.idacAreaMetadataRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return areaMetadata, nil
}
