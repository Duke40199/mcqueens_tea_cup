package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type IDACOBMetaService struct {
	// clients
	segaClient port.SegaIDACClient
	// repos
	carRepo database.CarRepository
}

func NewIDACOBMetaService(segaClient port.SegaIDACClient, carRepo database.CarRepository) port.IDACOBMetaService {
	return &IDACOBMetaService{segaClient: segaClient, carRepo: carRepo}
}

// GetMeta samples the top `limit` Online Battle players for an area, groups their
// cars by base spec -> style -> model code, and returns the counts sorted for
// stable rendering. An empty view (TotalSampled == 0) is a valid, non-error
// "no data" result.
func (s *IDACOBMetaService) GetMeta(ctx context.Context, area string, limit int) (*entity.OBMetaView, error) {
	currentRound, err := s.segaClient.GetCurrentRound()
	if err != nil {
		return nil, err
	}

	resp, err := s.segaClient.GetListOBRanking(strconv.Itoa(currentRound), area)
	if err != nil {
		return nil, err
	}
	view := &entity.OBMetaView{Round: currentRound}
	if resp == nil || len(resp.Records) == 0 {
		return view, nil
	}
	view.CalcDate = resp.CalcDate

	baseSpecMap, err := s.carRepo.GetBaseSpecMap(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(resp.Records) {
		limit = len(resp.Records)
	}

	// specStyleMap: baseSpec -> style -> modelCode -> *carStat
	type carStat struct {
		Name  string
		Count int
	}
	specStyleMap := make(map[string]map[string]map[string]*carStat)
	totalSampled := 0

	for j := 0; j < limit; j++ {
		carName := strings.TrimSpace(resp.Records[j].CarName)
		if carName == "" {
			continue
		}
		// Parse "FL5[DH]" -> modelCode="FL5", style="DH"
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
			modelCode = specInfo.ModelCode // normalize aliases -> canonical code
		}
		baseSpec := "Unknown"
		if ok {
			baseSpec = specInfo.BaseSpec
		}

		if specStyleMap[baseSpec] == nil {
			specStyleMap[baseSpec] = make(map[string]map[string]*carStat)
		}
		if specStyleMap[baseSpec][style] == nil {
			specStyleMap[baseSpec][style] = make(map[string]*carStat)
		}
		if specStyleMap[baseSpec][style][modelCode] == nil {
			specStyleMap[baseSpec][style][modelCode] = &carStat{Name: obMetaCarStatName(&specInfo, modelCode)}
		}
		specStyleMap[baseSpec][style][modelCode].Count++
		totalSampled++
	}
	view.TotalSampled = totalSampled

	// Flatten into sorted slices: base specs, then styles, then cars by count desc.
	specKeys := make([]string, 0, len(specStyleMap))
	for k := range specStyleMap {
		specKeys = append(specKeys, k)
	}
	sort.Strings(specKeys)

	for _, baseSpec := range specKeys {
		styleMap := specStyleMap[baseSpec]
		styleKeys := make([]string, 0, len(styleMap))
		for k := range styleMap {
			styleKeys = append(styleKeys, k)
		}
		sort.Strings(styleKeys)

		specGroup := entity.OBMetaSpecGroup{BaseSpec: baseSpec}
		for _, style := range styleKeys {
			carMap := styleMap[style]
			cars := make([]entity.OBMetaCarStat, 0, len(carMap))
			for _, cs := range carMap {
				cars = append(cars, entity.OBMetaCarStat{Name: cs.Name, Count: cs.Count})
			}
			sort.Slice(cars, func(i, j int) bool { return cars[i].Count > cars[j].Count })
			specGroup.Styles = append(specGroup.Styles, entity.OBMetaStyleGroup{Style: style, Cars: cars})
		}
		view.Specs = append(view.Specs, specGroup)
	}
	return view, nil
}

func obMetaCarStatName(carStat *entity.CarSpecInfo, defaultCode string) string {
	if carStat.Maker == "" || carStat.CarName == "" {
		return defaultCode
	}
	return fmt.Sprintf("%s %s [%s]", strings.Title(carStat.Maker), strings.Title(carStat.CarName), strings.Title(carStat.ModelCode))
}

// FormatOBMetaPages renders an aggregated OB meta view into Discord message
// pages (top 5 cars per style, paginated at ~1900 characters). It is shared by
// the /idac ob-meta command and the OB meta sync cron so both emit identical
// output.
func FormatOBMetaPages(view *entity.OBMetaView) []string {
	header := fmt.Sprintf(
		"# Initial DAC Most Used Cars (Online Battle)\n"+
			"### **Round: %d | Sample Size:** Top %d PRIDE Players\n"+
			"### Calculated at: %s (JST)\n"+
			"### ***(p.s. results are refreshed every 15 minutes.)***\n\n",
		view.Round, view.TotalSampled, view.CalcDate)

	var pages []string
	var currentMessage strings.Builder
	currentMessage.WriteString(header)

	for _, specGroup := range view.Specs {
		specEmoji := entity.SpecEmojis[specGroup.BaseSpec]
		if specEmoji == "" {
			specEmoji = "📦"
		}
		var section string
		for _, styleGroup := range specGroup.Styles {
			styleEmoji := entity.SpecEmojis[strings.ToLower(styleGroup.Style)]
			// Limit to top 5 cars per style.
			carParts := make([]string, 0, 5)
			for idx, c := range styleGroup.Cars {
				if idx >= 5 {
					break
				}
				carParts = append(carParts, fmt.Sprintf("- %s (%d)", c.Name, c.Count))
			}
			section += fmt.Sprintf("%s %s:\n%s\n", specEmoji, styleEmoji, strings.Join(carParts, "\n"))
		}
		section += "\n"

		// Pagination check
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
	return pages
}
