package discord

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/domain/entity"
)

// trackBarWidth is the character width of the threshold timeline bar. Wider reads
// better on desktop; on very narrow mobile it may wrap (misaligning the rows) —
// tune here if that becomes a problem.
const trackBarWidth = 60

// HandleTrackDetails shows a track's per-rank time thresholds as a legend
// (number → rank badge → exact time), fastest first.
func (h *Handler) HandleTrackDetails(cc *CommandContext) error {
	courseID := cc.OptMap["track"]
	courseName := entity.CourseDisplayNameByCode[courseID]
	if courseName == "" {
		return cc.Edit("⚠️ **Track not found based on input.**")
	}

	thresholds, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(cc.Ctx, courseID)
	if err != nil {
		return fmt.Errorf("fetching track thresholds: %w", err)
	}
	if len(thresholds) == 0 {
		return cc.Edit(fmt.Sprintf("ℹ️ No rank time thresholds found for **%s** yet.", courseName))
	}

	// Top 3 cars sits on top. Too little car data isn't fatal here — we still show the
	// thresholds (the section is just empty).
	topCars, _, err := h.renderTopUsedCars(cc.Ctx, courseID)
	if err != nil {
		return err
	}

	// Thresholds can be long, so present them as a paginated list (the Top 3 section
	// repeats on each page for context) with the course image as a thumbnail.
	cc.SendPagesWithThumbnail(buildTrackDetailPages(courseName, topCars, thresholds), courseImageURL(courseID))
	return nil
}

// courseImageURL returns the Sega course image URL for a course. Track variants of
// the same mountain share one image keyed to the mountain's base course number,
// which is the course number rounded down to a multiple of 4 (e.g. Myogi DH course-4
// and UH course-6 both use 4_0.jpg). Returns "" if the course id can't be parsed.
func courseImageURL(courseID string) string {
	n, err := strconv.Atoi(strings.TrimPrefix(courseID, "course-"))
	if err != nil {
		return ""
	}
	base := n - (n % 4)
	return fmt.Sprintf("https://initiald.sega.jp/inidac/$site/images/course/%d_0.jpg", base)
}

// buildTrackDetailPages lays out the track-details response as paginated pages: each
// page repeats the Top-3 cars section and the thresholds header, then holds a chunk
// of the rank legend (number → rank badge → exact time) so every page stays under
// Discord's message limit. Emoji badges render because they sit outside any code
// block. The timeline bar is intentionally omitted from the response;
// renderTrackThresholdBar still builds it and is kept for reuse.
func buildTrackDetailPages(courseName, topCars string, thresholds []*entity.TimeAttackRankingMetadata) []string {
	const maxPageLen = 1900

	//prefix := fmt.Sprintf("## ⏱️ **%s** · Rank Time Thresholds\n", courseName)

	prefix := fmt.Sprintf("## ⏱️ Time Attack Grade Thresholds\n ## 📍: **%s**\n", courseName)
	if topCars != "" {
		prefix += topCars + "\n"
	}

	var pages []string
	var b strings.Builder
	b.WriteString(prefix)
	for i, t := range thresholds {
		line := fmt.Sprintf("%d. %s — `%s`\n", i+1, t.RankName, entity.FormatRaceTime(t.RequiredTime))
		// Start a new page when the current one would overflow (always keep ≥1 line).
		if b.Len() > len(prefix) && b.Len()+len(line) > maxPageLen {
			pages = append(pages, b.String())
			b.Reset()
			b.WriteString(prefix)
		}
		b.WriteString(line)
	}
	pages = append(pages, b.String())
	return pages
}

// renderTrackThresholdBar builds a numbered timeline bar (fastest → slowest) as a
// code block, with each rank's number anchored at its time position. It is not
// currently included in the /idac track-details response, but is kept for reuse.
// Thresholds are expected sorted fastest → slowest (as the repository returns them).
func renderTrackThresholdBar(thresholds []*entity.TimeAttackRankingMetadata) string {
	n := len(thresholds)
	minT := thresholds[0].RequiredTime
	maxT := thresholds[n-1].RequiredTime
	span := maxT.Sub(minT)

	// Build the number-label line and the marker line, aligned by time position.
	labelLine := []rune(strings.Repeat(" ", trackBarWidth))
	markerLine := []rune(strings.Repeat(" ", trackBarWidth))
	grow := func(minLen int) {
		for len(labelLine) < minLen {
			labelLine = append(labelLine, ' ')
		}
		for len(markerLine) < minLen {
			markerLine = append(markerLine, ' ')
		}
	}
	// Pad all numbers to a uniform width so 1- and multi-digit labels align the same
	// way; the marker (and the number's last digit) anchor at the time position.
	labelWidth := len(strconv.Itoa(n))
	lastLabelEnd := -1
	for i, t := range thresholds {
		pos := 0
		if span > 0 {
			frac := float64(t.RequiredTime.Sub(minT)) / float64(span)
			pos = int(math.Round(frac * float64(trackBarWidth-1)))
		}

		// Right-align the number in a fixed-width field so its last digit sits at the
		// time position, with the marker directly under that last digit. Nudge right
		// only to avoid overlapping the previous label.
		label := []rune(fmt.Sprintf("%*d", labelWidth, i+1))
		start := pos - (labelWidth - 1)
		if start < 0 {
			start = 0
		}
		if start <= lastLabelEnd+1 {
			start = lastLabelEnd + 2 // keep at least one space between labels
		}
		grow(start + labelWidth)
		copy(labelLine[start:], label)
		markerLine[start+labelWidth-1] = 'v'
		lastLabelEnd = start + labelWidth - 1
	}

	// Bar width tracks the furthest label (dense clusters can push past the base
	// width); endpoints align fastest on the left, slowest on the right.
	barLen := len(markerLine)
	minStr := entity.FormatRaceTime(minT)
	maxStr := entity.FormatRaceTime(maxT)
	gap := barLen - len(minStr) - len(maxStr)
	if gap < 1 {
		gap = 1
	}
	endpoints := minStr + strings.Repeat(" ", gap) + maxStr

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(strings.TrimRight(string(labelLine), " ") + "\n")
	sb.WriteString(strings.TrimRight(string(markerLine), " ") + "\n")
	sb.WriteString(strings.Repeat("=", barLen) + "\n")
	sb.WriteString(endpoints + "\n")
	sb.WriteString("```")
	return sb.String()
}
