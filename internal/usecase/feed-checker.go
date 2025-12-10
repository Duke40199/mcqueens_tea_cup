package usecase

import (
	"log"

	"McQueens_Tea_Cup/internal/domain"
)

type FeedChecker struct {
	rss        domain.RSSFetcher
	notifier   domain.Notifier
	state      domain.StateStore
	translator domain.Translator
}

func NewFeedChecker(rss domain.RSSFetcher, not domain.Notifier, state domain.StateStore, trans domain.Translator) *FeedChecker {
	return &FeedChecker{
		rss:        rss,
		notifier:   not,
		state:      state,
		translator: trans,
	}
}

func (fc *FeedChecker) Check(feeds []domain.FeedConfig) {
	for _, feedConf := range feeds {
		items, err := fc.rss.Fetch(feedConf.URL)
		if err != nil {
			log.Printf("Error fetching %s: %v", feedConf.URL, err)
			continue
		}

		if len(items) == 0 {
			continue
		}

		lastSeenID, exists := fc.state.GetLastSeen(feedConf.URL)

		// Logic to filter new items vs old items
		var newItems []domain.Item

		if !exists {
			// Cold start logic: just mark the latest as seen without posting
			log.Printf("Cold start for %s", feedConf.Title)
			fc.state.SetLastSeen(feedConf.URL, items[0].GUID)
			continue
		}

		// Find new items
		for _, item := range items {
			if item.GUID == lastSeenID {
				break
			}

			newItems = append(newItems, item)
		}
		// Spam protection
		if len(newItems) > 5 {
			newItems = newItems[:5]
		}
		// Process and Send (Reverse order)
		for i := len(newItems) - 1; i >= 0; i-- {
			item := newItems[i]
			var transEN, transVN string
			// Basic mapping for common codes
			if feedConf.SourceLanguage == "ja" {
				transEN, _ = fc.translator.Translate(item.Content, feedConf.SourceLanguage, "en")
			}
			if feedConf.SourceLanguage != "en" {
				transEN, _ = fc.translator.Translate(item.Content, feedConf.SourceLanguage, "en")
			}
			if feedConf.SourceLanguage != "vi" {
				transVN, _ = fc.translator.Translate(item.Content, feedConf.SourceLanguage, "vi")
			}
			// Send
			fc.notifier.Send(feedConf.DestinationDiscordChannelID, item, feedConf.Title, transEN, transVN)
		}

		// Update State
		if len(items) > 0 {
			fc.state.SetLastSeen(feedConf.URL, items[0].GUID)
		}
	}
}
