package sega

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/domain/entity"
)

type SegaClient struct {
	TimeTrailUrl    string
	CurrentRoundUrl string
	TeamRankingUrl  string
}

// GetTeamRanking implements [domain.IdacRepository].
func (c *SegaClient) GetTeamRanking(roundNum int, rankCode string) ([]entity.TeamRecord, error) {
	rankURL := fmt.Sprintf("https://initiald.sega.jp/inidac/json/ranking/v1/leaguePoint/lp-round-%d_rank-%s.json", roundNum, rankCode)
	fmt.Printf("Fetching ranking data from %s\n", rankURL)
	resp, err := http.Get(rankURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data entity.IdacTeamRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
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

func NewClient() *SegaClient {
	return &SegaClient{
		TimeTrailUrl:    "https://initiald.sega.jp/inidac/json/ranking/v1",
		CurrentRoundUrl: "https://initiald.sega.jp/inidac/json/ranking/v1/currentRoundInfo.json",
		TeamRankingUrl:  "https://initiald.sega.jp/inidac/json/ranking/v1/leaguePoint",
	}
}

// GetTimeAttack implements domain.IdacRepository
func (c *SegaClient) GetTimeAttack(courseID, area, carID, spec string) ([]entity.TimeAttackRecord, error) {
	// 1. Construct URL (Logic moved from handler)
	filename := fmt.Sprintf("ta_%s_%s_%s.json", courseID, area, carID)
	fullURL := fmt.Sprintf("%s/timeTrial/%s", c.TimeTrailUrl, filename)
	// 2. Fetch
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sega api status: %d", resp.StatusCode)
	}

	// 3. Parse
	var data entity.IdacTimeAttackRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Records, nil
}

func (c *SegaClient) GetCurrentRound() (int, error) {
	roundURL := "https://initiald.sega.jp/inidac/json/ranking/v1/currentRoundInfo.json"
	respRound, err := http.Get(roundURL)
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

func (c *SegaClient) FetchTeamRankings(roundNum int, rankCode string) ([]entity.TeamRecord, error) {
	rankURL := fmt.Sprintf("https://initiald.sega.jp/inidac/json/ranking/v1/leaguePoint/lp-round-%d_rank-%s.json", roundNum, rankCode)
	fmt.Printf("Fetching ranking data from %s\n", rankURL)
	resp, err := http.Get(rankURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data entity.IdacTeamRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
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

func (c *SegaClient) GetListOBRanking(roundNum string, areaCode string) (*entity.IdacOBRankingResponse, error) {
	rankURL := fmt.Sprintf("https://initiald.sega.jp/inidac/json/ranking/v1/roundPoint/rp_round-%s_%s.json", roundNum, areaCode)
	fmt.Printf("Fetching ob ranking data from %s\n", rankURL)
	resp, err := http.Get(rankURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data entity.IdacOBRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	// foundTeams := make([]entity.TeamRecord, 0)
	// set league emoji values for each team
	// for _, foundTeam := range data.Records {
	// 	// foundTeam.LeagueEmoji = entity.TeamLeagueEmojis[rankCode]
	// 	// foundTeams = append(foundTeams, foundTeam)
	// }
	return &data, nil
}

func (c *SegaClient) GetListPlayerGrade(areaCode string) (*entity.IdacPlayerRankingResponse, error) {
	rankURL := fmt.Sprintf("https://initiald.sega.jp/inidac/json/ranking/v1/grade/gr_%s.json", areaCode)
	fmt.Printf("Fetching player ranking data from %s\n", rankURL)
	resp, err := http.Get(rankURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data entity.IdacPlayerRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	// foundTeams := make([]entity.TeamRecord, 0)
	// set league emoji values for each team
	// for _, foundTeam := range data.Records {
	// 	// foundTeam.LeagueEmoji = entity.TeamLeagueEmojis[rankCode]
	// 	// foundTeams = append(foundTeams, foundTeam)
	// }
	return &data, nil
}

func (c *SegaClient) FetchConst() (*entity.IdacConstResponse, error) {
	constURL := "https://initiald.sega.jp/inidac/json/ranking/v1/const.json"
	fmt.Printf("Fetching const data from %s\n", constURL)
	resp, err := http.Get(constURL)
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
