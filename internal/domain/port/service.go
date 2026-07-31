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
	GetTimeTrail(ctx context.Context, courseID, area, carID, spec string) ([]entity.TimeAttackRecord, error)
}

type IDACAreaService interface {
	GetAreaMetadata(ctx context.Context) ([]entity.IDACAreaMetadata, error)
}

type IDACOBRankingService interface {
	GetRanking(ctx context.Context, round, area string, limit int) (*entity.OBRankingView, error)
}

type IDACTeamService interface {
	GetSortedTeamRankings(ctx context.Context, rankCodes []string) (int, []entity.TeamRecord, error)
}

type IDACPlayerService interface {
	ResolvePlayer(ctx context.Context, input, manualArea string) (string, string, bool, error)
	GetPlayerProfile(ctx context.Context, ign, area string) (*entity.PlayerProfileView, error)
	GetTournamentInfo(ctx context.Context, areaCode string) (*entity.TournamentInfoView, error)
}

type IDACOBMetaService interface {
	GetMeta(ctx context.Context, area string, limit int) (*entity.OBMetaView, error)
}
type IDACPlayerInfoService interface {
}

type IDACTrackService interface{}
