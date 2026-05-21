package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"McQueens_Tea_Cup/internal/domain/entity"
)

// JSONAliasStore implements domain.AliasRepository
type JSONAliasStore struct {
	mu       sync.RWMutex
	data     map[string]entity.PlayerAlias
	filename string
}

// NewJSONAliasStore creates the store and loads existing data immediately
func NewJSONAliasStore(filename string) *JSONAliasStore {
	store := &JSONAliasStore{
		data:     make(map[string]entity.PlayerAlias),
		filename: filename,
	}

	// Try loading on startup
	if err := store.Load(); err != nil {
		fmt.Printf("⚠️ [Persistence] Could not load aliases from %s: %v\n", filename, err)
	} else {
		fmt.Printf("✅ [Persistence] Loaded aliases from %s\n", filename)
	}

	return store
}

// Get retrieves an alias (Thread-safe)
func (s *JSONAliasStore) Get(key string) (entity.PlayerAlias, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Set updates an alias and saves to disk (Thread-safe)
func (s *JSONAliasStore) Set(key, ign, area string) error {
	s.mu.Lock()
	s.data[key] = entity.PlayerAlias{Ign: ign, Area: area}
	s.mu.Unlock()

	return s.save()
}

// Load reads from the JSON file
func (s *JSONAliasStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.ReadFile(s.filename)
	if os.IsNotExist(err) {
		return nil // No file yet, acceptable
	}
	if err != nil {
		return err
	}
	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, &s.data)
}

// save is an internal helper to write to disk
func (s *JSONAliasStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}
