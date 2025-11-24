package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

// --- Structures ---

type Config struct {
	Token    string       `json:"token"`
	Interval int          `json:"check_interval_minutes"`
	Feeds    []FeedConfig `json:"feeds"`
}

type FeedConfig struct {
	Title                       string `json:"title"`
	URL                         string `json:"url"`
	SourceLanguage              string `json:"source_language"` // e.g., "JP", "ja", "es"
	DestinationDiscordChannelID string `json:"channel_id"`      // Custom channel for this feed
}

type FeedState map[string]string // Map of FeedURL -> LastSeenGUID

// --- Global Variables ---
var (
	config     Config
	feedState  FeedState
	stateFile  = "feed_state.json"
	stateMutex sync.Mutex
)

// --- Main Entry Point ---

func main() {
	if err := loadConfig("config.json"); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	loadState()

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	if err := dg.Open(); err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	log.Println("Bot is running. Press CTRL-C to exit.")

	ticker := time.NewTicker(time.Duration(config.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately, then tick
	go checkFeeds(dg)

	go func() {
		for range ticker.C {
			checkFeeds(dg)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down...")
}

// --- RSS Logic ---

func checkFeeds(s *discordgo.Session) {
	log.Println("Checking feeds...")
	fp := gofeed.NewParser()

	for _, feedConf := range config.Feeds {
		feed, err := fp.ParseURL(feedConf.URL)
		if err != nil {
			log.Printf("Error parsing %s: %v", feedConf.URL, err)
			continue
		}

		if len(feed.Items) == 0 {
			continue
		}

		// Get the last seen ID for this feed
		stateMutex.Lock()
		lastSeenID, exists := feedState[feedConf.URL]
		stateMutex.Unlock()

		// Determine display title
		displayTitle := feedConf.Title
		if displayTitle == "" {
			displayTitle = feed.Title
		}

		// Identify new items
		var newItems []*gofeed.Item

		// If this is the first time we see this feed, just take the latest one to avoid spam
		if !exists {
			newItems = append(newItems, feed.Items[0])
		} else {
			// Find all items newer than lastSeenID
			// RSS items are usually sorted Newest -> Oldest.
			// We scan from the beginning (Newest) until we find the lastSeenID.
			for _, item := range feed.Items {
				id := item.GUID
				if id == "" {
					id = item.Link
				}

				if id == lastSeenID {
					break // We found the checkpoint, stop gathering
				}

				newItems = append(newItems, item)
			}

			// Safety guard: If the feed rotated completely or we missed too much,
			// limit it to the 5 newest items to prevent spamming the channel.
			if len(newItems) > 5 {
				newItems = newItems[:5]
			}
		}

		// If no new items, continue to next feed
		if len(newItems) == 0 {
			continue
		}

		// Reverse newItems so we post Oldest -> Newest
		// We gathered them [Newest, 2nd Newest, ...].
		// We want to post them in the order they appeared in time.
		for i := len(newItems) - 1; i >= 0; i-- {
			item := newItems[i]
			log.Printf("New item found for %s: %s", displayTitle, item.Title)
			sendUpdate(s, feed, item, displayTitle, feedConf.DestinationDiscordChannelID, feedConf.SourceLanguage)
		}

		// Update state to the absolute newest item (index 0 of feed)
		latest := feed.Items[0]
		newLatestID := latest.GUID
		if newLatestID == "" {
			newLatestID = latest.Link
		}

		stateMutex.Lock()
		feedState[feedConf.URL] = newLatestID
		saveState()
		stateMutex.Unlock()
	}
}

func sendUpdate(s *discordgo.Session, feed *gofeed.Feed, item *gofeed.Item, customTitle string, channelID, sourceLang string) {
	// 1. Normalize the Link (Twitter fix)
	finalLink := normalizeLink(item.Link)

	// 2. Prepare Content
	rawContent := item.Description
	if rawContent == "" {
		rawContent = item.Content
	}

	cleanContent := cleanHTML(rawContent)

	// 3. Translation Logic
	var translationTextEn string
	var translationTextVn string

	// Default to "en" if sourceLang is missing but content exists
	if sourceLang == "" {
		sourceLang = "en"
	}

	// Basic mapping for common codes
	if strings.ToUpper(sourceLang) == "JP" {
		sourceLang = "ja"
	}

	// Logic: If source is NOT English, translate to English.
	if strings.ToLower(sourceLang) != "en" {
		if strings.TrimSpace(cleanContent) != "" {
			tr, err := translateText(cleanContent, sourceLang, "en")
			if err != nil {
				log.Printf("Translation error (EN): %v", err)
			} else {
				translationTextEn = tr
			}
		}
	}

	// Logic: If source is NOT Vietnamese, translate to Vietnamese.
	if strings.ToLower(sourceLang) != "vi" {
		if strings.TrimSpace(cleanContent) != "" {
			trVn, err := translateText(cleanContent, sourceLang, "vi")
			if err != nil {
				log.Printf("Translation error (VN): %v", err)
			} else {
				translationTextVn = trVn
			}
		}
	}

	// 4. Construct the Message
	msgBody := fmt.Sprintf("**%s**\n\n", customTitle)

	itemTitle := strings.TrimSpace(item.Title)
	contentBody := strings.TrimSpace(cleanContent)

	// Add Title if it's a Headline (not a Tweet duplicate)
	if itemTitle != "" && !strings.Contains(contentBody, itemTitle) && len(itemTitle) < 100 {
		msgBody += fmt.Sprintf("**%s**\n", itemTitle)
	}

	// Add Original Content (Truncated to ~800 to leave room for translation)
	msgBody += truncate(contentBody, 800)

	// Add English Translation if available
	if translationTextEn != "" {
		msgBody += fmt.Sprintf("\n\n🇬🇧 **Translation:**\n%s", truncate(translationTextEn, 800))
	}

	// Add Vietnamese Translation if available
	if translationTextVn != "" {
		msgBody += fmt.Sprintf("\n\n🇻🇳 **Bản dịch:**\n%s", truncate(translationTextVn, 800))
	}

	// Add Link
	msgBody += fmt.Sprintf("\n\n%s", finalLink)

	// 5. Handle Image
	if item.Image != nil {
		msgBody += "\n" + item.Image.URL
	}

	// 6. Send Standard Message to the Specific Channel ID
	_, err := s.ChannelMessageSend(channelID, msgBody)
	if err != nil {
		log.Printf("Error sending to Discord channel %s: %v", channelID, err)
	}
}

// --- Helper Functions ---

func translateText(text, source, target string) (string, error) {
	// Use Google Translate 'gtx' endpoint (free, unofficial)
	baseURL := "https://translate.googleapis.com/translate_a/single"
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("client", "gtx")
	q.Set("sl", source)
	q.Set("tl", target)
	q.Set("dt", "t")
	q.Set("q", text)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google translate api returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse JSON: [[[ "translated", "original", ...], ...], ...]
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result) > 0 {
		inner, ok := result[0].([]interface{})
		if ok {
			var fullTranslation strings.Builder
			for _, slice := range inner {
				s, ok := slice.([]interface{})
				if ok && len(s) > 0 {
					translatedPart, ok := s[0].(string)
					if ok {
						fullTranslation.WriteString(translatedPart)
					}
				}
			}
			return fullTranslation.String(), nil
		}
	}

	return "", fmt.Errorf("could not parse translation response")
}

func normalizeLink(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	re := regexp.MustCompile(`^/([^/]+)/status/(\d+)`)

	if re.MatchString(u.Path) {
		u.Scheme = "https"
		u.Host = "twitter.com"
		return u.String()
	}

	return link
}

func cleanHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	stripped := re.ReplaceAllString(input, "")
	return html.UnescapeString(stripped)
}

func loadConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

func loadState() {
	feedState = make(FeedState)
	file, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("Error reading state file: %v", err)
		return
	}
	json.Unmarshal(file, &feedState)
}

func saveState() {
	data, err := json.MarshalIndent(feedState, "", "  ")
	if err != nil {
		log.Printf("Error marshaling state: %v", err)
		return
	}
	err = os.WriteFile(stateFile, data, 0644)
	if err != nil {
		log.Printf("Error writing state file: %v", err)
	}
}

func truncate(text string, length int) string {
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length]) + "..."
}
