package discord

import (
	"McQueens_Tea_Cup/internal/domain/entity"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleStoreLocation(i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	// 1. Validate & Parse Inputs
	allNetAreaCode, _ := optMap["area-select"]
	listAllNetStore, foundAreaName, err := h.IDACStoreLocationService.GetListAllNextStore(context.Background(), allNetAreaCode)
	if err != nil {
		h.SendDeferredError(i, "Error finding list store.")
		return
	}
	if listAllNetStore == nil {
		h.SendDeferredError(i, "Cannot find list store.")
		return
	}
	var segaAreaCode string
	for _, area := range h.AreaMetadata {
		if area.ALLNetCode == allNetAreaCode {
			segaAreaCode = area.SegaAreaCode
			break
		}
	}
	if segaAreaCode == "" {
		h.SendDeferredError(i, "Cannot find area code.")
		return
	}
	mapSegaStore, err := h.IDACStoreLocationService.GetMapStoreFromTopPlayers(context.Background(), segaAreaCode)
	if err != nil {
		h.SendDeferredError(i, "Error finding list store.")
		return
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
	err = h.IDACStoreLocationService.BulkUpsertStoreLocation(context.Background(), listAllNetStore)
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
		h.SendPagination(i, pages)
		return
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
	h.SendPagination(i, pages)
}
