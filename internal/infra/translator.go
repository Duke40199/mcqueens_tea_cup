package infra

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GoogleTranslator struct{}

func NewGoogleTranslator() *GoogleTranslator {
	return &GoogleTranslator{}
}

func (t *GoogleTranslator) Translate(text, source, target string) (string, error) {
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
