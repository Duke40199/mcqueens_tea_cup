package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type MetaLogicService struct {
	obMetaService port.IDACOBMetaService
}

func NewMetaLogicService(obMetaService port.IDACOBMetaService) *MetaLogicService {
	return &MetaLogicService{
		obMetaService: obMetaService,
	}
}

// GetOBMetaPages fetches OB meta data via the OB meta service and returns
// formatted Discord pages.
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

	view, err := s.obMetaService.GetMeta(ctx, area, limit)
	if err != nil {
		return nil, err
	}
	if view.TotalSampled == 0 {
		return []string{fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName)}, nil
	}
	return FormatOBMetaPages(view), nil
}

// SleepUntilNextSync calculates the wait time until the next Sega OB update (hh:02:02, 16:02, 31:02, 46:02)
// and sleeps with a 10-second safety buffer. It also respects scheduled downtime.
func (s *MetaLogicService) SleepUntilNextSync(ctx context.Context, downtimeStart, downtimeEnd int, downtimeTZ string) {
	targets := []int{2, 16, 31, 46}
	buffer := 10 * time.Second
	jstLoc := time.FixedZone("JST", 9*60*60)

	// Determine Downtime Location
	downtimeLoc := jstLoc
	if downtimeTZ != "" {
		if loc, err := time.LoadLocation(downtimeTZ); err == nil {
			downtimeLoc = loc
		}
	}

	for {
		nowJST := time.Now().In(jstLoc)
		nowDT := time.Now().In(downtimeLoc)

		// 1. Check for Downtime
		if downtimeStart != 0 && downtimeEnd != 0 {
			if s.IsInDowntime(nowDT, strconv.Itoa(downtimeStart), strconv.Itoa(downtimeEnd)) {
				s.sleepUntilDowntimeEnd(ctx, nowDT, strconv.Itoa(downtimeEnd))
				// After waking up from downtime, we should calculate the next sync target
				nowJST = time.Now().In(jstLoc)
			}
		}

		currentMin := nowJST.Minute()
		currentSec := nowJST.Second()

		var nextMin int
		found := false
		for _, m := range targets {
			if currentMin < m || (currentMin == m && currentSec < 0) {
				nextMin = m
				found = true
				break
			}
		}

		nextTime := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), nowJST.Hour(), nextMin, 0, 0, nowJST.Location())
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
	}
}

// IsInDowntime checks if the current time falls within the [start, end] window.
// Supports both "HH:MM" and "HH" formats.
func (s *MetaLogicService) IsInDowntime(now time.Time, startStr, endStr string) bool {
	parseTime := func(str string) (int, int, error) {
		if strings.Contains(str, ":") {
			parts := strings.Split(str, ":")
			if len(parts) != 2 {
				return 0, 0, fmt.Errorf("invalid format")
			}
			h, _ := strconv.Atoi(parts[0])
			m, _ := strconv.Atoi(parts[1])
			return h, m, nil
		}
		// Hour only
		h, err := strconv.Atoi(str)
		return h, 0, err
	}

	startH, startM, err := parseTime(startStr)
	if err != nil {
		return false
	}
	endH, endM, err := parseTime(endStr)
	if err != nil {
		return false
	}

	currentSeconds := now.Hour()*3600 + now.Minute()*60 + now.Second()
	startSeconds := startH*3600 + startM*60
	endSeconds := endH*3600 + endM*60

	if startSeconds <= endSeconds {
		// Standard: 10:00 to 22:00
		return currentSeconds >= startSeconds && currentSeconds < endSeconds
	} else {
		// Wrap around: 22:00 to 08:00
		return currentSeconds >= startSeconds || currentSeconds < endSeconds
	}
}

func (s *MetaLogicService) sleepUntilDowntimeEnd(ctx context.Context, now time.Time, endStr string) {
	var h, m int
	if strings.Contains(endStr, ":") {
		parts := strings.Split(endStr, ":")
		h, _ = strconv.Atoi(parts[0])
		m, _ = strconv.Atoi(parts[1])
	} else {
		h, _ = strconv.Atoi(endStr)
		m = 0
	}

	wakeTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	if !wakeTime.After(now) {
		wakeTime = wakeTime.Add(24 * time.Hour)
	}

	waitDuration := time.Until(wakeTime)
	log.Printf("😴 Scheduled Downtime: Sleeping for %v until %s...", waitDuration.Round(time.Second), wakeTime.Format("15:04:05"))

	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		log.Printf("☀️ Downtime ended. Resuming sync...")
	case <-ctx.Done():
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
