package discord

import (
	"fmt"
	"sort"
	"strings"

	"McQueens_Tea_Cup/internal/domain/entity"
)

func (h *Handler) HandleStoreLocation(cc *CommandContext) error {
	// 1. Validate & Parse Inputs
	allNetAreaCode := cc.OptMap["area-select"]
	listAllNetStore, foundAreaName, err := h.IDACStoreLocationService.GetListAllNextStore(cc.Ctx, allNetAreaCode)
	if err != nil {
		return fmt.Errorf("fetching AllNet store list: %w", err)
	}
	if listAllNetStore == nil {
		return NewUserError("Cannot find list store.")
	}
	var segaAreaCode string
	for _, area := range h.AreaMetadata {
		if area.ALLNetCode == allNetAreaCode {
			segaAreaCode = area.SegaAreaCode
			break
		}
	}
	if segaAreaCode == "" {
		return NewUserError("Cannot find area code.")
	}
	mapSegaStore, err := h.IDACStoreLocationService.GetMapStoreFromTopPlayers(cc.Ctx, segaAreaCode)
	if err != nil {
		return fmt.Errorf("fetching Sega store map: %w", err)
	}
	if len(mapSegaStore) == 0 {
		fmt.Printf("No area found from Sega API.")
	}
	if len(mapSegaStore) > 0 {
		allNetMap := make(map[string]bool)
		for i := range listAllNetStore {
			allNetMap[listAllNetStore[i].Name] = true
			listAllNetStore[i].SegaAreaCode = segaAreaCode
			listAllNetStore[i].AllNetAreaCode = allNetAreaCode
		}

		for key, segaStore := range mapSegaStore {
			if !allNetMap[key] {
				listAllNetStore = append(listAllNetStore, entity.StoreLocation{
					Name:           segaStore.Name,
					Address:        "Imported from Sega IDAC",
					SegaAreaCode:   segaAreaCode,
					AllNetAreaCode: allNetAreaCode,
				})
			}
		}
	}
	sort.Slice(listAllNetStore, func(i, j int) bool {
		return listAllNetStore[i].Name < listAllNetStore[j].Name
	})
	err = h.IDACStoreLocationService.BulkUpsertStoreLocation(cc.Ctx, listAllNetStore)
	if err != nil {
		fmt.Printf("error bulk upsert store location: %v\n", err)
	}
	// 4. Build Response
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	header := fmt.Sprintf("# (WIP) List Store Location (%s)\n", foundAreaName)

	// Initialize first page with header
	currentMessage.WriteString(header)
	// write not found if list is empty
	if len(listAllNetStore) == 0 {
		currentMessage.WriteString("No stores found for the specified area.")
		pages = append(pages, currentMessage.String())
		cc.SendPages(pages)
		return nil
	}
	for j := 0; j < len(listAllNetStore); j++ {
		var entry string
		entry = fmt.Sprintf("%d. **%s** — %s \n", j+1, listAllNetStore[j].Name, listAllNetStore[j].Address)

		// Split if 10 items OR length > 1900
		if itemsInChunk >= 10 || currentMessage.Len()+len(entry) > 1900 {
			pages = append(pages, currentMessage.String())

			currentMessage.Reset()
			currentMessage.WriteString(header) // Add header to every page for clarity
			itemsInChunk = 0
		}

		currentMessage.WriteString(entry)
		itemsInChunk++
	}
	// Append final page
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	// 5. Hand over to Pagination Helper
	cc.SendPages(pages)
	return nil
}
