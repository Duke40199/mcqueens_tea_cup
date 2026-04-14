package entity

import (
	"fmt"
	"time"
)

type TimeAttackRankingCfg struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type PlayerGradeCfg struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	SegaID string `json:"sega_id"`
}
type TimeAttackRankingMetadata struct {
	ID           string    `json:"id"`
	CourseID     string    `json:"course_id"`
	RequiredTime time.Time `json:"required_time"`
	RankID       string    `json:"rank_id"`
	RankName     string    `json:"rank_name"`
}

// ParseRaceTime converts a Sega IDAC time string (e.g., "3'14\"765") into a time.Time object.
func ParseRaceTime(rawTime string) (time.Time, error) {
	var min, sec, ms int

	// Extract the integers based on the exact format: [min]'[sec]"[ms]
	_, err := fmt.Sscanf(rawTime, "%d'%d\"%d", &min, &sec, &ms)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format, expected M'SS\"MMM: %w", err)
	}

	// Reconstruct the time into a standard PostgreSQL TIME string (00:MM:SS.MMM)
	// We hardcode the hour to 00 since time attack tracks don't take hours to finish.
	standardTimeStr := fmt.Sprintf("00:%02d:%02d.%03d", min, sec, ms)

	// Parse it using Go's standard time layout for this exact format
	parsedTime, err := time.Parse("15:04:05.000", standardTimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to build time object: %w", err)
	}

	return parsedTime, nil
}
