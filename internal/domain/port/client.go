package port

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

type AllNetClient interface {
	GetListStore(ctx context.Context, gameCode, languageCode, areaCode string) ([]entity.StoreLocation, string, error)
}

type SegaIDACClient interface {
	GetListTimeTrail(ctx context.Context, courseID, area, car, spec string) ([]entity.TimeAttackRecord, error)
	GetTeamRanking(ctx context.Context, round int, rankCode string) ([]entity.TeamRecord, error)
	GetListOBRanking(ctx context.Context, roundNum string, areaCode string) (*entity.IdacOBRankingResponse, error)
	GetCurrentRound(ctx context.Context) (int, error)
	FetchConst(ctx context.Context) (*entity.IdacConstResponse, error)
	GetListPlayerGrade(ctx context.Context, areaCode string) (*entity.IdacPlayerRankingResponse, error)
	GetPlayerGradeByIGN(ctx context.Context, ign, areaCode string) (*entity.PlayerRankingRecord, error)
	GetOBRankingByIGN(ctx context.Context, ign, roundNum, areaCode string) (*entity.OBRankingRecord, error)
}
