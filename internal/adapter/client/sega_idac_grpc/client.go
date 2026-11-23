// Package sega_idac_grpc is an alternative SegaIDACClient that talks to a
// sega-idac-client gRPC server instead of calling SEGA's HTTP endpoints directly.
// It implements the same port.SegaIDACClient interface as the HTTP client, so it
// is a drop-in replacement selected in main by SEGA_IDAC_GRPC_ADDR.
package sega_idac_grpc

import (
	"context"

	segaidacv1 "github.com/Duke40199/sega-idac-client/gen/segaidacv1"
	"google.golang.org/grpc"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type SegaIDACClient struct {
	client segaidacv1.SegaIDACServiceClient
}

// NewSegaIDACClient builds a gRPC-backed SegaIDACClient over an existing connection.
func NewSegaIDACClient(conn grpc.ClientConnInterface) port.SegaIDACClient {
	return &SegaIDACClient{client: segaidacv1.NewSegaIDACServiceClient(conn)}
}

func (c *SegaIDACClient) GetListTimeTrail(ctx context.Context, courseID, area, car, spec string) ([]entity.TimeAttackRecord, error) {
	resp, err := c.client.GetListTimeTrail(ctx, &segaidacv1.GetListTimeTrailRequest{
		CourseId: courseID,
		AreaCode: area,
		CarId:    car,
		Spec:     spec,
	})
	if err != nil {
		return nil, err
	}
	return fromProtoTimeAttackRecords(resp.GetRecords()), nil
}

func (c *SegaIDACClient) GetTeamRanking(ctx context.Context, round int, rankCode string) ([]entity.TeamRecord, error) {
	resp, err := c.client.GetTeamRanking(ctx, &segaidacv1.GetTeamRankingRequest{
		RoundNum: int32(round),
		RankCode: rankCode,
	})
	if err != nil {
		return nil, err
	}
	return fromProtoTeamRecords(resp.GetRecords()), nil
}

func (c *SegaIDACClient) GetListOBRanking(ctx context.Context, roundNum, areaCode string) (*entity.IdacOBRankingResponse, error) {
	resp, err := c.client.GetListOBRanking(ctx, &segaidacv1.GetListOBRankingRequest{
		RoundNum: roundNum,
		AreaCode: areaCode,
	})
	if err != nil {
		return nil, err
	}
	return &entity.IdacOBRankingResponse{
		CalcDate: resp.GetCalcDate(),
		Records:  fromProtoOBRankingRecords(resp.GetRecords()),
	}, nil
}

func (c *SegaIDACClient) GetCurrentRound(ctx context.Context) (int, error) {
	resp, err := c.client.GetCurrentRound(ctx, &segaidacv1.GetCurrentRoundRequest{})
	if err != nil {
		return -1, err
	}
	return int(resp.GetRound()), nil
}

func (c *SegaIDACClient) FetchConst(ctx context.Context) (*entity.IdacConstResponse, error) {
	resp, err := c.client.FetchConst(ctx, &segaidacv1.FetchConstRequest{})
	if err != nil {
		return nil, err
	}
	return &entity.IdacConstResponse{
		Cars:   fromProtoCarMetadataList(resp.GetCars()),
		Styles: fromProtoCarStyleMetadataList(resp.GetStyles()),
	}, nil
}

func (c *SegaIDACClient) GetListPlayerGrade(ctx context.Context, areaCode string) (*entity.IdacPlayerRankingResponse, error) {
	resp, err := c.client.GetListPlayerGrade(ctx, &segaidacv1.GetListPlayerGradeRequest{
		AreaCode: areaCode,
	})
	if err != nil {
		return nil, err
	}
	return &entity.IdacPlayerRankingResponse{
		CalcDate: resp.GetCalcDate(),
		Records:  fromProtoPlayerRankingRecords(resp.GetRecords()),
	}, nil
}

func (c *SegaIDACClient) GetPlayerGradeByIGN(ctx context.Context, ign, areaCode string) (*entity.PlayerRankingRecord, error) {
	resp, err := c.client.GetPlayerGradeByIGN(ctx, &segaidacv1.GetPlayerGradeByIGNRequest{
		Ign:      ign,
		AreaCode: areaCode,
	})
	if err != nil {
		return nil, err
	}
	if !resp.GetFound() {
		return nil, nil
	}
	return fromProtoPlayerRankingRecord(resp.GetRecord()), nil
}

func (c *SegaIDACClient) GetOBRankingByIGN(ctx context.Context, ign, roundNum, areaCode string) (*entity.OBRankingRecord, error) {
	resp, err := c.client.GetOBRankingByIGN(ctx, &segaidacv1.GetOBRankingByIGNRequest{
		Ign:      ign,
		RoundNum: roundNum,
		AreaCode: areaCode,
	})
	if err != nil {
		return nil, err
	}
	if !resp.GetFound() {
		return nil, nil
	}
	return fromProtoOBRankingRecord(resp.GetRecord()), nil
}

// ----- proto -> entity mappers -----

func fromProtoTimeAttackRecords(in []*segaidacv1.TimeAttackRecord) []entity.TimeAttackRecord {
	out := make([]entity.TimeAttackRecord, 0, len(in))
	for _, r := range in {
		out = append(out, entity.TimeAttackRecord{
			Rank:       r.GetRank(),
			Name:       r.GetName(),
			ShopName:   r.GetShopName(),
			Record:     r.GetRecord(),
			CarName:    r.GetCarName(),
			MyTitleID:  r.GetMyTitleId(),
			UpdateDate: r.GetUpdateDate(),
		})
	}
	return out
}

func fromProtoTeamRecords(in []*segaidacv1.TeamRecord) []entity.TeamRecord {
	out := make([]entity.TeamRecord, 0, len(in))
	for _, r := range in {
		out = append(out, entity.TeamRecord{
			Rank:           r.GetRank(),
			TeamName:       r.GetTeamName(),
			ShopName:       r.GetShopName(),
			Country:        int(r.GetCountry()),
			AceUserName:    r.GetAceUserName(),
			LeaderUserName: r.GetLeaderUserName(),
			Point:          r.GetPoint(),
			LeagueEmoji:    r.GetLeagueEmoji(),
		})
	}
	return out
}

func fromProtoOBRankingRecord(r *segaidacv1.OBRankingRecord) *entity.OBRankingRecord {
	if r == nil {
		return nil
	}
	return &entity.OBRankingRecord{
		Rank:               r.GetRank(),
		Name:               r.GetName(),
		ShopName:           r.GetShopName(),
		UpdateDate:         r.GetUpdateDate(),
		CarName:            r.GetCarName(),
		Point:              r.GetPoint(),
		PrideId:            r.GetPrideId(),
		PridePoint:         int(r.GetPridePoint()),
		OnlineBattleRankId: r.GetOnlineBattleRankId(),
		StarCnt:            int(r.GetStarCnt()),
	}
}

func fromProtoOBRankingRecords(in []*segaidacv1.OBRankingRecord) []entity.OBRankingRecord {
	out := make([]entity.OBRankingRecord, 0, len(in))
	for _, r := range in {
		if rec := fromProtoOBRankingRecord(r); rec != nil {
			out = append(out, *rec)
		}
	}
	return out
}

func fromProtoPlayerRankingRecord(r *segaidacv1.PlayerRankingRecord) *entity.PlayerRankingRecord {
	if r == nil {
		return nil
	}
	return &entity.PlayerRankingRecord{
		ID:         int(r.GetId()),
		Rank:       int(r.GetRank()),
		Name:       r.GetName(),
		ShopName:   r.GetShopName(),
		GradeID:    r.GetGradeId(),
		GradeExp:   int(r.GetGradeExp()),
		MytitleID:  r.GetMyTitleId(),
		UpdateDate: r.GetUpdateDate(),
		NumberIcon: r.GetNumberIcon(),
	}
}

func fromProtoPlayerRankingRecords(in []*segaidacv1.PlayerRankingRecord) []entity.PlayerRankingRecord {
	out := make([]entity.PlayerRankingRecord, 0, len(in))
	for _, r := range in {
		if rec := fromProtoPlayerRankingRecord(r); rec != nil {
			out = append(out, *rec)
		}
	}
	return out
}

func fromProtoCarMetadataList(in []*segaidacv1.CarMetadata) []entity.CarMetadata {
	out := make([]entity.CarMetadata, 0, len(in))
	for _, c := range in {
		out = append(out, entity.CarMetadata{
			ID:                  c.GetId(),
			SegaCarID:           c.GetSegaCarId(),
			Name:                c.GetName(),
			ModelCode:           c.GetModelCode(),
			Maker:               c.GetMaker(),
			BaseStyleName:       c.GetBaseStyleName(),
			NormalizedBaseStyle: c.GetNormalizedBaseStyle(),
			CarStyleIDs:         c.GetCarStyleIds(),
			SpecIDs:             c.GetSpecIds(),
			SpecNames:           c.GetSpecNames(),
		})
	}
	return out
}

func fromProtoCarStyleMetadataList(in []*segaidacv1.CarStyleMetadata) []entity.CarStyleMetadata {
	out := make([]entity.CarStyleMetadata, 0, len(in))
	for _, s := range in {
		out = append(out, entity.CarStyleMetadata{
			ID:             s.GetId(),
			StyleCarID:     s.GetStyleCarId(),
			CarID:          s.GetCarId(),
			Name:           s.GetName(),
			RouteStyleName: s.GetRouteStyleName(),
		})
	}
	return out
}
