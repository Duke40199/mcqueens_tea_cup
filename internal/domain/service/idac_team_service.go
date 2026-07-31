package service

import (
	"context"
	"sort"
	"strconv"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type IDACTeamService struct {
	segaClient port.SegaIDACClient
}

func NewIDACTeamService(segaClient port.SegaIDACClient) port.IDACTeamService {
	return &IDACTeamService{segaClient: segaClient}
}

// GetSortedTeamRankings resolves the current round, fetches team rankings for
// each requested rank code, and returns them aggregated and sorted by points
// descending. When a single rank code is requested a fetch failure is returned
// as an error; when several are requested, failing codes are skipped so a
// partial result is still produced.
func (s *IDACTeamService) GetSortedTeamRankings(ctx context.Context, rankCodes []string) (int, []entity.TeamRecord, error) {
	round, err := s.segaClient.GetCurrentRound()
	if err != nil {
		return 0, nil, err
	}

	var allRecords []entity.TeamRecord
	for _, code := range rankCodes {
		records, err := s.segaClient.GetTeamRanking(round, code)
		if err != nil {
			if len(rankCodes) == 1 {
				return round, nil, err
			}
			continue
		}
		allRecords = append(allRecords, records...)
	}

	sort.Slice(allRecords, func(i, j int) bool {
		p1, _ := strconv.Atoi(allRecords[i].Point)
		p2, _ := strconv.Atoi(allRecords[j].Point)
		return p1 > p2
	})
	return round, allRecords, nil
}
