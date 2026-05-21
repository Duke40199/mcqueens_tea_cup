package port

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

type AllNetClient interface {
	GetListStore(ctx context.Context, gameCode, languageCode, areaCode string) ([]entity.StoreLocation, error)
}

type SegaIDACClient interface {
	GetListTimeTrail(courseID, area, car, spec string) ([]entity.TimeAttackRecord, error)
	GetTeamRanking(round int, rankCode string) ([]entity.TeamRecord, error)
	GetListOBRanking(roundNum string, areaCode string) (*entity.IdacOBRankingResponse, error)
	GetCurrentRound() (int, error)
	FetchConst() (*entity.IdacConstResponse, error)
	GetListPlayerGrade(areaCode string) (*entity.IdacPlayerRankingResponse, error)
	GetPlayerGradeByIGN(ign, areaCode string) (*entity.PlayerRankingRecord, error)
	GetOBRankingByIGN(ign, roundNum, areaCode string) (*entity.OBRankingRecord, error)
}
