package infra

import (
	"html"
	"net/url"
	"regexp"

	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/mmcdole/gofeed"
)

type GoFeedFetcher struct {
	parser *gofeed.Parser
}

func NewGoFeedFetcher() *GoFeedFetcher {
	return &GoFeedFetcher{parser: gofeed.NewParser()}
}

func (g *GoFeedFetcher) Fetch(url string) ([]entity.Item, error) {
	feed, err := g.parser.ParseURL(url)
	if err != nil {
		return nil, err
	}

	var items []entity.Item
	for _, i := range feed.Items {
		guid := i.GUID
		if guid == "" {
			guid = i.Link
		}

		rawContent := i.Description
		if rawContent == "" {
			rawContent = i.Content
		}

		imgURL := ""
		if i.Image != nil {
			imgURL = i.Image.URL
		}

		// CLEANING HAPPENS HERE:
		items = append(items, entity.Item{
			Title:    i.Title,
			Content:  cleanHTML(rawContent), // <--- HTML stripped here
			Link:     normalizeLink(i.Link), // <--- Twitter links fixed here
			ImageURL: imgURL,
			GUID:     guid,
		})
	}
	return items, nil
}

// --- Helpers Implementation ---

func cleanHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	stripped := re.ReplaceAllString(input, "")
	return html.UnescapeString(stripped)
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
