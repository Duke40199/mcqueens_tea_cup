package database

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

// AliasRepository defines how we interact with player aliases.
type AliasRepository interface {
	GetByAliasKey(discordID string) (entity.PlayerAlias, bool, error)
	GetByIgnAndAreaCode(ign, areaCode string) (entity.PlayerAlias, bool, error)

	SetPlayerAlias(discordID, ign, area string) error
	Load() error
}

// AliasRepository defines how we interact with player aliases.
type OBRankingCfgRepository interface {
	GetRankingCfgMap() (map[string]entity.OBRankingCfg, error)
	GetBySegaID(segaID string) (*entity.OBRankingCfg, error)
}

type CarRepository interface {
	UpsertCars(ctx context.Context, cars []entity.CarMetadata) error
	UpsertCarStyles(ctx context.Context, styles []entity.CarStyleMetadata) error
	GetBaseSpecMap(ctx context.Context) (map[string]entity.CarSpecInfo, error)
	GetSegaIDToUUIDMap(ctx context.Context) (map[int64]string, error)
	GetCarWithSpecsByAliases(ctx context.Context, aliasSpecMap map[string]string) (map[string]entity.CarSpecInfo, error)
}

type AreaRepository interface {
	GetOBActiveAreas(ctx context.Context) ([]entity.AreaSyncInfo, error)
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
	CreateCfsState(discordID, content string) (int64, error)
}
