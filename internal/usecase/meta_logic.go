package usecase

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"
)

type MetaLogicService struct {
	SegaClient entity.SegaClient
	CarRepo    postgres.CarRepository
}

func NewMetaLogicService(client entity.SegaClient, repo postgres.CarRepository) *MetaLogicService {
	return &MetaLogicService{
		SegaClient: client,
		CarRepo:    repo,
	}
}

// GetOBMetaPages fetches OBMeta data and returns formatted Discord pages
func (s *MetaLogicService) GetOBMetaPages(ctx context.Context, limit int, area string) ([]string, error) {
	areaInput := strings.ToLower(area)
	if areaInput == "" {
		areaInput = "all"
	}
	if val, ok := entity.AreaAliases[areaInput]; ok {
		area = val
	}
	areaName := area
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	finalArea := area

	currentRound, err := s.SegaClient.GetCurrentRound()
	if err != nil {
		return nil, fmt.Errorf("error fetching current round: %w", err)
	}

	records, err := s.SegaClient.GetListOBRanking(strconv.Itoa(currentRound), finalArea)
	if err != nil {
		return nil, fmt.Errorf("error fetching list ob ranking: %w", err)
	}
	if records == nil || len(records.Records) == 0 {
		return []string{fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName)}, nil
	}

	if len(records.Records) < limit {
		limit = len(records.Records)
	}

	baseSpecMap, err := s.CarRepo.GetBaseSpecMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching car spec data: %w", err)
	}

	type CarStat struct {
		Name  string
		Count int
	}

	specStyleMap := make(map[string]map[string]map[string]*CarStat)
	totalSampled := 0

	for j := 0; j < limit; j++ {
		r := records.Records[j]
		carName := strings.TrimSpace(r.CarName)
		if carName == "" {
			continue
		}
		bracketStart := strings.Index(carName, "[")
		bracketEnd := strings.Index(carName, "]")

		var modelCode, style string
		if bracketStart != -1 && bracketEnd != -1 && bracketEnd > bracketStart {
			modelCode = strings.TrimSpace(carName[:bracketStart])
			style = strings.ToUpper(strings.TrimSpace(carName[bracketStart+1 : bracketEnd]))
		} else {
			modelCode = carName
			style = "Unknown"
		}

		specInfo, ok := baseSpecMap[modelCode]
		if ok {
			modelCode = specInfo.ModelCode
		}

		baseSpec := "Unknown"
		if ok {
			baseSpec = specInfo.BaseSpec
		}

		if specStyleMap[baseSpec] == nil {
			specStyleMap[baseSpec] = make(map[string]map[string]*CarStat)
		}
		if specStyleMap[baseSpec][style] == nil {
			specStyleMap[baseSpec][style] = make(map[string]*CarStat)
		}
		if specStyleMap[baseSpec][style][modelCode] == nil {
			specStyleMap[baseSpec][style][modelCode] = &CarStat{Name: s.getCarStatName(&specInfo, modelCode), Count: 0}
		}
		specStyleMap[baseSpec][style][modelCode].Count++
		totalSampled++
	}

	var pages []string
	var currentMessage strings.Builder

	header := fmt.Sprintf(
		"# Initial DAC Most Used Cars (Online Battle)\n"+
			"### **Round: %d | Sample Size:** Top %d PRIDE Players\n"+
			"### Calculated at: %s (JST)\n"+
			"### ***(p.s. results are refreshed every 15 minutes.)***\n\n",
		currentRound, totalSampled, records.CalcDate)
	currentMessage.WriteString(header)

	specKeys := make([]string, 0, len(specStyleMap))
	for k := range specStyleMap {
		specKeys = append(specKeys, k)
	}
	sort.Strings(specKeys)

	for _, baseSpec := range specKeys {
		styleMap := specStyleMap[baseSpec]
		specEmoji := entity.SpecEmojis[baseSpec]
		if specEmoji == "" {
			specEmoji = "📦"
		}
		var section string
		styleKeys := make([]string, 0, len(styleMap))
		for k := range styleMap {
			styleKeys = append(styleKeys, k)
		}
		sort.Strings(styleKeys)

		for _, style := range styleKeys {
			styleEmoji := entity.SpecEmojis[strings.ToLower(style)]
			carMap := styleMap[style]
			cars := make([]CarStat, 0, len(carMap))
			for _, cs := range carMap {
				cars = append(cars, *cs)
			}
			sort.Slice(cars, func(i, j int) bool {
				return cars[i].Count > cars[j].Count
			})

			carParts := make([]string, 0, 5)
			for idx, c := range cars {
				if idx >= 5 {
					break
				}
				carParts = append(carParts, fmt.Sprintf("- %s (%d)", c.Name, c.Count))
			}
			section += fmt.Sprintf("%s %s:\n%s\n", specEmoji, styleEmoji, strings.Join(carParts, "\n"))
		}
		section += "\n"

		if currentMessage.Len()+len(section) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
		}
		currentMessage.WriteString(section)
	}

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	return pages, nil
}

func (s *MetaLogicService) getCarStatName(carStat *entity.CarSpecInfo, defaultCode string) string {
	if carStat.Maker == "" || carStat.CarName == "" {
		return defaultCode
	}
	return fmt.Sprintf("%s %s [%s]", strings.Title(carStat.Maker), strings.Title(carStat.CarName), strings.Title(carStat.ModelCode))
}

// SleepUntilNextSync calculates the wait time until the next Sega OB update (hh:02:02, 16:02, 31:02, 46:02)
// and sleeps with a 10-second safety buffer.
func (s *MetaLogicService) SleepUntilNextSync(ctx context.Context) {
	targets := []int{2, 16, 31, 46}
	buffer := 10 * time.Second

	for {
		now := time.Now().In(time.FixedZone("JST", 9*60*60)) // Use JST for Sega sync
		currentMin := now.Minute()
		currentSec := now.Second()

		var nextMin int
		found := false
		for _, m := range targets {
			if currentMin < m || (currentMin == m && currentSec < 0) {
				nextMin = m
				found = true
				break
			}
		}

		nextTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), nextMin, 0, 0, now.Location())
		if !found {
			// Next update is next hour at :02
			nextTime = nextTime.Add(time.Hour)
			nextTime = time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), nextTime.Hour(), 2, 0, 0, nextTime.Location())
		}

		// Apply safety buffer
		wakeTime := nextTime.Add(buffer)
		waitDuration := time.Until(wakeTime)

		if waitDuration > 0 {
			log.Printf("😴 Sleeping for %v until next Sega update (%s)...", waitDuration.Round(time.Second), wakeTime.Format("15:04:05"))
			timer := time.NewTimer(waitDuration)
			select {
			case <-timer.C:
				return
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		// If we missed it somehow, loop again immediately for next target
		return
	}
}

// IsDataFresh checks if the CalcDate from Sega is at or after the most recent expected update time.
func (s *MetaLogicService) IsDataFresh(calcDateStr string) bool {
	jstLoc := time.FixedZone("JST", 9*60*60)
	calcTime, err := time.ParseInLocation("2006/01/02 15:04:05", calcDateStr, jstLoc)
	if err != nil {
		return false
	}

	now := time.Now().In(jstLoc)
	targets := []int{2, 16, 31, 46}

	var latestExpectedMin int
	found := false
	for i := len(targets) - 1; i >= 0; i-- {
		m := targets[i]
		if now.Minute() >= m {
			latestExpectedMin = m
			found = true
			break
		}
	}

	var latestExpectedTime time.Time
	if !found {
		// Latest expected update was last hour at :46
		prevHour := now.Add(-time.Hour)
		latestExpectedTime = time.Date(prevHour.Year(), prevHour.Month(), prevHour.Day(), prevHour.Hour(), 46, 0, 0, jstLoc)
	} else {
		latestExpectedTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), latestExpectedMin, 0, 0, jstLoc)
	}

	// Data is fresh if CalcDate is >= latestExpectedTime
	return !calcTime.Before(latestExpectedTime)
}
