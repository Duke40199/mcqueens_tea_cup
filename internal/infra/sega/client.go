package sega

import (
	"McQueens_Tea_Cup/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	TimeTrailUrl    string
	CurrentRoundUrl string
	TeamRankingUrl  string
}

// GetCurrentRound implements [domain.IdacRepository].
func (c *Client) GetCurrentRound() (int, error) {
	panic("unimplemented")
}

// GetTeamRanking implements [domain.IdacRepository].
func (c *Client) GetTeamRanking(round int, rankCode string) ([]domain.TeamRecord, error) {
	panic("unimplemented")
}

func NewClient() *Client {
	return &Client{
		TimeTrailUrl:    "https://initiald.sega.jp/inidac/json/ranking/v1",
		CurrentRoundUrl: "https://initiald.sega.jp/inidac/json/ranking/v1/currentRoundInfo.json",
		TeamRankingUrl:  "https://initiald.sega.jp/inidac/json/ranking/v1/leaguePoint",
	}
}

// GetTimeAttack implements domain.IdacRepository
func (c *Client) GetTimeAttack(courseID, area, carID, spec string) ([]domain.TimeAttackRecord, error) {
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
	var data domain.IdacResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Records, nil
}

func (c *Client) FetchCurrentRoundNum(roundNum int, rankCode string) (int, error) {
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
	roundNum, _ = strconv.Atoi(roundStr)
	return roundNum, nil
}

func (c *Client) FetchTeamRankings(roundNum int, rankCode string) ([]domain.TeamRecord, error) {
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

	var data domain.TeamRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	foundTeams := make([]domain.TeamRecord, 0)
	// set league emoji values for each team
	for _, foundTeam := range data.Records {
		foundTeam.LeagueEmoji = domain.TeamLeagueEmojis[rankCode]
		foundTeams = append(foundTeams, foundTeam)
	}
	return foundTeams, nil
}
