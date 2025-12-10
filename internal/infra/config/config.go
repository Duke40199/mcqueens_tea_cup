package config

import (
	"encoding/json"
	"log"
	"os"

	"McQueens_Tea_Cup/internal/domain"
)

// --- Global Variables ---
var (
	stateFile = "feed_state.json"
)

func LoadConfig(filename string) (*domain.Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	newCfg := &domain.Config{}
	if err = json.Unmarshal(file, newCfg); err != nil {
		return nil, err
	}
	return newCfg, err
}

func LoadState() (*domain.FeedState, error) {
	newFeedState := &domain.FeedState{}
	file, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		log.Printf("Error reading state file: %v", err)
		return nil, err
	}
	if err = json.Unmarshal(file, &newFeedState); err != nil {
		return nil, err
	}
	return newFeedState, nil
}
