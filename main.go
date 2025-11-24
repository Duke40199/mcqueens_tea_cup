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
	DestinationDiscordChannelID string `json:"channel_id"`
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

		latestItem := feed.Items[0]

		uniqueID := latestItem.GUID
		if uniqueID == "" {
			uniqueID = latestItem.Link
		}

		stateMutex.Lock()
		lastSeenID, exists := feedState[feedConf.URL]

		if !exists || lastSeenID != uniqueID {
			// Determine display title (Config > Feed Title)
			displayTitle := feedConf.Title
			if displayTitle == "" {
				displayTitle = feed.Title
			}

			log.Printf("New item found for %s", displayTitle)

			feedState[feedConf.URL] = uniqueID
			saveState()
			stateMutex.Unlock()

			sendUpdate(s, feed, latestItem, displayTitle, feedConf.DestinationDiscordChannelID, feedConf.SourceLanguage)
		} else {
			stateMutex.Unlock()
		}
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
	// Basic mapping for common codes if needed, though Google is smart
	if strings.ToUpper(sourceLang) == "JP" {
		sourceLang = "ja"
	}
	if sourceLang != "" && strings.ToLower(sourceLang) != "en" {
		// en translation

		// Don't translate if it's empty
		if strings.TrimSpace(cleanContent) != "" {
			tr, err := translateText(cleanContent, sourceLang, "en")
			if err != nil {
				log.Printf("Translation error: %v", err)
			} else {
				translationTextEn = tr
			}
		}
	}
	// vn translation
	if strings.ToLower(sourceLang) != "vi" {
		// Don't translate if it's empty
		if strings.TrimSpace(cleanContent) != "" {
			if sourceLang == "" {
				sourceLang = "en"
			}
			trVn, err := translateText(cleanContent, sourceLang, "vi")
			if err != nil {
				log.Printf("Translation error: %v", err)
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

	// Add Translation if available
	if translationTextEn != "" {
		msgBody += fmt.Sprintf("\n\n🇬🇧 **Translation:**\n%s", truncate(translationTextEn, 800))
	}
	if translationTextVn != "" {
		msgBody += fmt.Sprintf("\n\n🇻🇳 **Bản dịch:**\n%s", truncate(translationTextVn, 800))
	}

	// Add Link
	msgBody += fmt.Sprintf("\n\n%s", finalLink)

	// 5. Handle Image
	if item.Image != nil {
		msgBody += "\n" + item.Image.URL
	}

	// 6. Send Standard Message
	_, err := s.ChannelMessageSend(channelID, msgBody)
	if err != nil {
		log.Printf("Error sending to Discord: %v", err)
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
	// It's a messy nested array structure.
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
