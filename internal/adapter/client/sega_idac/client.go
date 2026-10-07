package sega_idac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
)

// segaHTTPTimeout bounds every Sega request so a stalled endpoint can't hang a
// command handler or sync goroutine indefinitely. A caller-supplied ctx with an
// earlier deadline still wins.
const segaHTTPTimeout = 15 * time.Second

type SegaIDACClient struct {
	config     config.Config
	httpClient *http.Client
}

func NewSegaIDACClient(cfg config.Config) port.SegaIDACClient {
	return &SegaIDACClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: segaHTTPTimeout},
	}
}

// makeHttpRequest performs a GET against url honoring ctx and decodes a JSON body into out.
func (c *SegaIDACClient) makeHttpRequest(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// GetTimeAttack
func (c *SegaIDACClient) GetListTimeTrail(ctx context.Context, courseID, areaCode, carID, spec string) ([]entity.TimeAttackRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTimeTrailURLPath
	url = strings.Replace(url, ":courseID", courseID, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	url = strings.Replace(url, ":carID", carID, 1)
	logger.Debug(ctx, fmt.Sprintf("request url: %s", url))
	// 2. Fetch + parse
	var data entity.IdacTimeAttackRecordResponse
	if err := c.makeHttpRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetListTimeTrail: %w", err)
	}
	return data.Records, nil
}

func (c *SegaIDACClient) GetOBRankingByIGN(ctx context.Context, ign, roundNum, areaCode string) (*entity.OBRankingRecord, error) {
	normalizedIGN := entity.NormalizeTextWidth(ign)
	listOBRankings, err := c.GetListOBRanking(ctx, roundNum, areaCode)
	if err != nil {
		return nil, err
	}
	if len(listOBRankings.Records) == 0 {
		return nil, nil
	}
	for _, obRanking := range listOBRankings.Records {
		if entity.NormalizeTextWidth(obRanking.Name) == normalizedIGN {
			return &obRanking, nil
		}
	}
	return nil, nil
}

func (c *SegaIDACClient) GetListOBRanking(ctx context.Context, roundNum string, areaCode string) (*entity.IdacOBRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListOBRankingURLPath
	url = strings.Replace(url, ":roundNum", roundNum, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	logger.Debug(ctx, fmt.Sprintf("request url: %s", url))
	// 2. Fetch + parse
	var data entity.IdacOBRankingResponse
	if err := c.makeHttpRequest(ctx, url, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetTeamRanking
func (c *SegaIDACClient) GetTeamRanking(ctx context.Context, roundNum int, rankCode string) ([]entity.TeamRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTeamRankingUrlPath
	url = strings.Replace(url, ":roundCount", strconv.Itoa(roundNum), 1)
	url = strings.Replace(url, ":rankType", rankCode, 1)
	logger.Debug(ctx, fmt.Sprintf("request url: %s", url))
	// 2. Fetch + parse
	var data entity.IdacTeamRankingResponse
	if err := c.makeHttpRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetTeamRanking: %w", err)
	}
	foundTeams := make([]entity.TeamRecord, 0)
	// set league emoji values for each team
	for _, foundTeam := range data.Records {
		foundTeam.LeagueEmoji = entity.TeamLeagueEmojis[rankCode]
		foundTeams = append(foundTeams, foundTeam)
	}
	return foundTeams, nil
}

func (c *SegaIDACClient) GetCurrentRound(ctx context.Context) (int, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetCurrentRoundUrlPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return -1, fmt.Errorf("creating request: %w", err)
	}
	respRound, err := c.httpClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer respRound.Body.Close()
	if respRound.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("sega api status: %d", respRound.StatusCode)
	}
	bodyBytes, err := io.ReadAll(respRound.Body)
	if err != nil {
		return -1, fmt.Errorf("error while reading response: %w", err)
	}

	roundStr := strings.TrimSpace(string(bodyBytes))
	roundNum, err := strconv.Atoi(roundStr)
	if err != nil {
		return -1, fmt.Errorf("parsing round number %q: %w", roundStr, err)
	}
	return roundNum, nil
}

func (c *SegaIDACClient) GetListPlayerGrade(ctx context.Context, areaCode string) (*entity.IdacPlayerRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListPlayerGradeUrlPath
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	logger.Debug(ctx, fmt.Sprintf("request url: %s", url))
	// 2. Fetch + parse
	var data entity.IdacPlayerRankingResponse
	if err := c.makeHttpRequest(ctx, url, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *SegaIDACClient) GetPlayerGradeByIGN(ctx context.Context, ign, areaCode string) (*entity.PlayerRankingRecord, error) {
	listAreaGrade, err := c.GetListPlayerGrade(ctx, areaCode)
	normalizedIgn := entity.NormalizeTextWidth(ign)
	if err != nil {
		return nil, err
	}
	if len(listAreaGrade.Records) == 0 {
		return nil, nil
	}
	for _, record := range listAreaGrade.Records {
		if entity.NormalizeTextWidth(record.Name) == normalizedIgn {
			return &record, nil
		}
	}

	return nil, nil
}

func (c *SegaIDACClient) FetchConst(ctx context.Context) (*entity.IdacConstResponse, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListConstConfigURLPath
	logger.Debug(ctx, fmt.Sprintf("request url: %s", url))
	var data entity.IdacConstResponse
	if err := c.makeHttpRequest(ctx, url, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
