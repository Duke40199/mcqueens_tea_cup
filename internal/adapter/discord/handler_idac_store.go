package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleStoreLocation(i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	// 1. Validate & Parse Inputs
	areaCode, _ := optMap["area"]
	listStore, foundAreaName, err := h.StoreLocationService.GetListAllNextStore(context.Background(), areaCode)
	if err != nil {
		h.SendDeferredError(i, "Error finding list store.")
		return
	}
	if listStore == nil {
		h.SendDeferredError(i, "Cannot find list store.")
		return
	}
	// 4. Build Response
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	header := fmt.Sprintf("# (WIP) All.net Store Location (%s)\n", foundAreaName)

	// Initialize first page with header
	currentMessage.WriteString(header)
	// write not found if list is empty
	if len(listStore) == 0 {
		currentMessage.WriteString("No stores found for the specified area.")
		pages = append(pages, currentMessage.String())
		h.SendPagination(i, pages)
		return
	}
	for j := 0; j < len(listStore); j++ {
		var entry string
		entry = fmt.Sprintf("%d. **%s** — %s \n", j+1, listStore[j].Name, listStore[j].Address)

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
