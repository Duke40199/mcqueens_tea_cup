package discord

import (
	"McQueens_Tea_Cup/internal/domain/entity"
	"context"
	"strings"
)

func (h *Handler) GetListCarDetailByTAFormat(ctx context.Context, listCarSegaFormat []string) (map[string]entity.CarSpecInfo, error) {
	aliasSpecMap := make(map[string]string, 0)
	// sega format: FD3S[DH]
	for _, segaFormat := range listCarSegaFormat {
		// 0: FD3S, 1: DH
		splitStr := strings.Split(segaFormat, "[")
		start := strings.LastIndex(segaFormat, "[")
		end := strings.LastIndex(segaFormat, "]")
		if start == -1 || end == -1 || end <= start {
			aliasSpecMap[splitStr[0]] = "" // No valid brackets found
		}
		aliasSpecMap[splitStr[0]] = segaFormat[start+1 : end]
	}
	specInfo, err := h.CarRepo.GetCarWithSpecsByAliases(ctx, aliasSpecMap)
	if err != nil {
		return nil, err
	}
	return specInfo, nil
}
