package port

import "McQueens_Tea_Cup/internal/domain/entity"

type AllNetClient interface {
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
