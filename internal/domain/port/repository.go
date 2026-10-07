package port

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

// AliasRepository defines how we interact with player aliases.
type AliasRepository interface {
	GetByAliasKey(ctx context.Context, discordID string) (entity.PlayerAlias, bool, error)
	GetByIgnAndAreaCode(ctx context.Context, ign, areaCode string) (entity.PlayerAlias, bool, error)

	SetPlayerAlias(discordID, ign, area string) error
	Load() error
}

// OBRankingCfgRepository defines how we interact with Online Battle ranking config.
type OBRankingCfgRepository interface {
	GetRankingCfgMap(ctx context.Context) (map[string]entity.OBRankingCfg, error)
	GetBySegaID(ctx context.Context, segaID string) (*entity.OBRankingCfg, error)
}

type CarRepository interface {
	UpsertCars(ctx context.Context, cars []entity.CarMetadata) error
	UpsertCarStyles(ctx context.Context, styles []entity.CarStyleMetadata) error
	GetBaseSpecMap(ctx context.Context) (map[string]entity.CarSpecInfo, error)
	GetSegaIDToUUIDMap(ctx context.Context) (map[int64]string, error)
	GetCarWithSpecsByAliases(ctx context.Context, aliasSpecMap map[string]string) (map[string]entity.CarSpecInfo, error)
	GetListCarWithAggregatedSpecs(ctx context.Context) ([]*entity.CarMetadata, error)
}

type AreaRepository interface {
	GetOBActiveAreas(ctx context.Context) ([]entity.AreaSyncInfo, error)
}

type IDACAreaMetadataRepository interface {
	GetAll(ctx context.Context) ([]entity.IDACAreaMetadata, error)
}

type RankingCfgRepository interface {
	GetListTimeAttackRankingCfg(ctx context.Context) ([]*entity.TimeAttackRankingCfg, error)
	GetListPlayerGradeCfg(ctx context.Context) ([]*entity.PlayerGradeCfg, error)
	GetPlayerGradeBySegaIDs(ctx context.Context, gradeSegaID, gradeNumSegaID string) ([]*entity.PlayerGradeCfg, error)
}

type TATimeMetadataRepository interface {
	GetByCourseID(ctx context.Context, courseID string) ([]*entity.TimeAttackRankingMetadata, error)
}

type CfsStateRepository interface {
	GetLatestCfsState(ctx context.Context) (*entity.CfsState, error)
	CreateCfsState(ctx context.Context, discordID, content string) (int64, error)
}

type AllNetStoreLocationsRepository interface {
	BulkUpsertStoreLocation(ctx context.Context, stores []entity.StoreLocation) error
}
