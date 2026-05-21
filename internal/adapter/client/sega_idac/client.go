package sega_idac

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type SegaIDACClient struct {
	config config.Config
}

func NewSegaIDACClient(cfg config.Config) port.SegaIDACClient {
	return &SegaIDACClient{
		config: cfg,
	}
}

// GetTimeAttack
func (c *SegaIDACClient) GetListTimeTrail(courseID, areaCode, carID, spec string) ([]entity.TimeAttackRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTimeTrailURLPath
	url = strings.Replace(url, ":courseID", courseID, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	url = strings.Replace(url, ":carID", carID, 1)
	fmt.Printf("=== SegaIDACClient: GetListTimeTrailResults full path: %s\n", url)
	// 2. Make request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("=== SegaIDACClient: GetListTimeTrailResults: status code %d", resp.StatusCode)
	}
	// 3. Parse response
	var data entity.IdacTimeAttackRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("=== SegaIDACClient: GetListTimeTrailResults: cannot parse res %d, err %s\n", resp.StatusCode, error.Error)
		return nil, err
	}
	return data.Records, nil
}

func (c *SegaIDACClient) GetOBRankingByIGN(ign, roundNum, areaCode string) (*entity.OBRankingRecord, error) {
	normalizedIGN := entity.NormalizeTextWidth(ign)
	listOBRankings, err := c.GetListOBRanking(roundNum, areaCode)
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

func (c *SegaIDACClient) GetListOBRanking(roundNum string, areaCode string) (*entity.IdacOBRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListOBRankingURLPath
	url = strings.Replace(url, ":roundNum", roundNum, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	fmt.Printf("=== SegaIDACClient.GetListOBRanking full URL: %s\n", url)
	// 2. Make request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	// 3. Parse response
	var data entity.IdacOBRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetTeamRanking
func (c *SegaIDACClient) GetTeamRanking(roundNum int, rankCode string) ([]entity.TeamRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTeamRankingUrlPath
	url = strings.Replace(url, ":roundCount", strconv.Itoa(roundNum), 1)
	url = strings.Replace(url, ":rankType", rankCode, 1)
	fmt.Printf("=== SegaIDACClient: GetTeamRanking full path: %s\n", url)
	// 2. Make request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("=== SegaIDACClient: GetTeamRanking: status code %d", resp.StatusCode)
	}
	// 3. Parse response
	var data entity.IdacTeamRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("=== SegaIDACClient: GetTeamRanking: cannot parse res %d, err %s\n", resp.StatusCode, error.Error)
		return nil, err
	}
	foundTeams := make([]entity.TeamRecord, 0)
	// set league emoji values for each team
	for _, foundTeam := range data.Records {
		foundTeam.LeagueEmoji = entity.TeamLeagueEmojis[rankCode]
		foundTeams = append(foundTeams, foundTeam)
	}
	return foundTeams, nil
}

func (c *SegaIDACClient) GetCurrentRound() (int, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetCurrentRoundUrlPath
	respRound, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	defer respRound.Body.Close()
	if respRound.StatusCode != 200 {
		return -1, fmt.Errorf("sega api status: %d", respRound.StatusCode)
	}
	bodyBytes, err := io.ReadAll(respRound.Body)
	if err != nil {
		return -1, fmt.Errorf("error while reading response: %d", err)
	}

	roundStr := strings.TrimSpace(string(bodyBytes))
	roundNum, _ := strconv.Atoi(roundStr)
	return roundNum, nil
}

func (c *SegaIDACClient) GetListPlayerGrade(areaCode string) (*entity.IdacPlayerRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListPlayerGradeUrlPath
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	fmt.Printf("=== SegaIDACClient.GetListPlayerGrade url: %s\n", url)
	// 2. Make request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	// 3. Parse response
	var data entity.IdacPlayerRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *SegaIDACClient) GetPlayerGradeByIGN(ign, areaCode string) (*entity.PlayerRankingRecord, error) {
	listAreaGrade, err := c.GetListPlayerGrade(areaCode)
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

func (c *SegaIDACClient) FetchConst() (*entity.IdacConstResponse, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListConstConfigURLPath
	fmt.Printf("=== SegaIDACClient: FetchConst %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data entity.IdacConstResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
