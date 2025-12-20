package idac_domain

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"McQueens_Tea_Cup/internal/domain"
)

// PlayerAlias represents the saved data
type PlayerAlias struct {
	Ign  string `json:"ign"`
	Area string `json:"area"`
}

// AliasStore handles storage
type AliasStore struct {
	mu   sync.RWMutex
	Data map[string]PlayerAlias // Key is Discord User ID
	File string
}

// Global instance
var Aliases = &AliasStore{
	Data: make(map[string]PlayerAlias),
	File: "aliases.json",
}

// --- FIX: Auto-load on startup ---
func init() {
	// This runs automatically when the program starts
	if err := Aliases.Load(); err != nil {
		fmt.Println("⚠️ No aliases loaded (starting fresh).")
	} else {
		fmt.Printf("✅ Automatically loaded %d aliases from %s\n", len(Aliases.Data), Aliases.File)
	}
}

// Load reads from disk
func (s *AliasStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.ReadFile(s.File)
	if os.IsNotExist(err) {
		return nil // File doesn't exist yet, that's fine
	}
	if err != nil {
		return err
	}

	// If file is empty, stop here to avoid JSON error
	if len(file) == 0 {
		return nil
	}

	// Unmarshal merges data into the map
	return json.Unmarshal(file, &s.Data)
}

// Set updates a user's alias and saves immediately
func (s *AliasStore) Set(userID, ign, area string) error {
	s.mu.Lock()
	// Update the in-memory map (Appends if new, Updates if existing)
	s.Data[userID] = PlayerAlias{Ign: ign, Area: area}
	s.mu.Unlock()

	// Persist the FULL map to disk
	return s.Save()
}

// Save writes the entire map to disk
func (s *AliasStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert the whole map to JSON
	data, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return err
	}
	// Overwrite the file with the complete updated list
	return os.WriteFile(s.File, data, 0644)
}

// Get retrieves a user's alias
func (s *AliasStore) Get(userID string) (PlayerAlias, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Data[userID]
	return val, ok
}

// ResolvePlayerCredential parses input for Mention vs Custom Alias vs Literal IGN
func ResolvePlayerCredential(input, manualArea string) (string, string, bool, error) {
	// Regex for Mentions/IDs
	reMention := regexp.MustCompile(`^<@!?(\d+)>$`)
	reID := regexp.MustCompile(`^\d{17,20}$`)

	var lookupKey string

	// 1. Check for Discord Mention/ID (<@123...>)
	if match := reMention.FindStringSubmatch(input); len(match) > 1 {
		lookupKey = match[1]
	} else if reID.MatchString(input) {
		lookupKey = input
	}

	// A. IT IS A DISCORD ID -> Force Lookup
	if lookupKey != "" {
		alias, found := Aliases.Get(lookupKey)
		if !found {
			return "", "", false, fmt.Errorf("User <@%s> has no alias saved.", lookupKey)
		}
		return alias.Ign, alias.Area, true, nil
	}

	// 2. Check for Text Input

	// CASE A: No Area provided -> Treat as ALIAS
	if manualArea == "" {
		// Check our aliases.json (case-insensitive key)
		if alias, found := Aliases.Get(strings.ToLower(input)); found {
			return alias.Ign, alias.Area, true, nil
		}
		// Not found? Return error (we can't search server without an area)
		return "", "", false, nil // "Not found as alias"
	}

	// CASE B: Area provided -> Treat as LITERAL IGN
	// Resolve "world" -> "all", etc.
	finalArea := strings.ToLower(manualArea)
	if finalArea == "world" || finalArea == "global" {
		finalArea = "all"
	}
	if val, ok := domain.AreaAliases[finalArea]; ok {
		finalArea = val
	}

	return input, finalArea, false, nil
}
