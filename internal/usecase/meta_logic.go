package usecase

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

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
