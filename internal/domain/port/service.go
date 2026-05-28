package port

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

type IDACStoreLocationService interface {
	GetListAllNextStore(ctx context.Context, areaCode string) ([]entity.StoreLocation, string, error)
	GetMapStoreFromTopPlayers(ctx context.Context, areaCode string) (map[string]entity.StoreLocation, error)
	BulkUpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) error
	// UpsertStoreLocation(ctx context.Context, listStoreLocations []entity.StoreLocation) (int64, error)
}

type IDACCarService interface {
	GetListCarDetailByTAFormat(ctx context.Context, listCarSegaFormat []string) (map[string]entity.CarSpecInfo, error)
	GetListTopTACarsWithPercentage(ctx context.Context, segaCourseID string, resultCount int64) ([]entity.IDACCarUsagePercentage, error)
}

type IDACTimeAttackService interface {
	GetMetadataBySegaCourseID(ctx context.Context, segaCourseID string) ([]*entity.TimeAttackRankingMetadata, error)
}

type IDACAreaService interface {
	GetAreaMetadata(ctx context.Context) ([]entity.IDACAreaMetadata, error)
}
type IDACPlayerInfoService interface {
}

type IDACTrackService interface{}
