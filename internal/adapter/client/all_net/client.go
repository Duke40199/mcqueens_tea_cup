package all_net

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"

	"github.com/PuerkitoBio/goquery"
)

type AllNetClient struct {
	config config.Config
}

func NewAllNetClient(config config.Config) port.AllNetClient {
	return &AllNetClient{
		config: config,
	}
}

func (c *AllNetClient) GetListStore(ctx context.Context, gameCode, languageCode, areaCode string) ([]entity.StoreLocation, error) {
	url := c.config.GetAllNetClientCfg().AllNetHost + c.config.GetAllNetClientCfg().GetListStoreLocationURLPath
	url = strings.Replace(url, ":gameCode", gameCode, 1)
	url = strings.Replace(url, ":languageCode", languageCode, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	fmt.Printf("=== AllNetClient: GetListStoreFromAllNet %s\n", url)

	// Use NewRequestWithContext to respect the provided ctx
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	// Parse the HTML response body directly with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var data []entity.StoreLocation

	// Iterate over each list item in the store list
	doc.Find("ul.store_list li").Each(func(i int, s *goquery.Selection) {
		// Extract Name
		name := strings.TrimSpace(s.Find("span.store_name").Text())

		// Extract Address (optional, but easily available in your HTML payload!)
		address := strings.TrimSpace(s.Find("span.store_address").Text())

		// Ensure we don't append empty entries
		if name != "" {
			store := entity.StoreLocation{
				Name:    name,    // Replace with your actual struct field name
				Address: address, // Replace or remove based on your struct
			}
			data = append(data, store)
		}
	})

	return data, nil
}
