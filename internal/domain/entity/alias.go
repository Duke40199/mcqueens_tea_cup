package entity

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PlayerAlias represents the saved data
type PlayerAlias struct {
	ID        uuid.UUID `json:"id"`
	Ign       string    `json:"ign"`
	AliasKey  string    `json:"alias_key"`
	Area      string    `json:"area"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

//// --- FIX: Auto-load on startup ---
//func init() {
//	// This runs automatically when the program starts
//	if err := Aliases.Load(); err != nil {
//		fmt.Println("⚠️ No aliases loaded (starting fresh).")
//	} else {
//		fmt.Printf("✅ Automatically loaded %d aliases from %s\n", len(Aliases.Data), Aliases.File)
//	}
//}

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

// ResolvePlayerCredential prioritizes Aliases first, then falls back to literal input.
// Returns: (FinalIGN, FinalArea, FoundAlias, Error)
func ResolvePlayerCredential(input, manualArea string) (string, string, bool, error) {
	// 1. Sanitize Inputs
	cleanInput := strings.TrimSpace(input)
	cleanArea := strings.TrimSpace(manualArea)

	// 2. Regex for Discord Mention/ID
	reMention := regexp.MustCompile(`^<@!?(\d+)>$`)
	reID := regexp.MustCompile(`^\d{17,20}$`)

	var lookupKey string
	if match := reMention.FindStringSubmatch(cleanInput); len(match) > 1 {
		lookupKey = match[1] // Extract ID from <@123>
	} else if reID.MatchString(cleanInput) {
		lookupKey = cleanInput // Raw ID
	}

	// 3. Check Alias Store (The "Check First" Logic)
	var aliasIgn, aliasArea string
	var foundAlias bool

	if lookupKey != "" {
		// Case A: Discord ID -> Strict Lookup
		if val, ok := Aliases.Get(lookupKey); ok {
			aliasIgn = val.Ign
			aliasArea = val.Area
			foundAlias = true
		} else {
			return "", "", false, fmt.Errorf("User <@%s> has no alias saved.", lookupKey)
		}
	} else {
		// Case B: Text Input -> Try Custom Tag Lookup
		// Use lowercase key for case-insensitive matching
		if val, ok := Aliases.Get(strings.ToLower(cleanInput)); ok {
			aliasIgn = val.Ign
			aliasArea = val.Area
			foundAlias = true
		}
	}

	// 4. Merge Logic (Alias vs Manual)
	finalIgn := cleanInput
	finalArea := cleanArea

	if foundAlias {
		finalIgn = aliasIgn
		// If Manual Area is provided, it OVERRIDES the alias area.
		// If not, we fall back to the alias area.
		if finalArea == "" {
			finalArea = aliasArea
		}
	}

	// 5. Normalize Final Area (Handle "world", "vn", etc.)
	if finalArea != "" {
		norm := strings.ToLower(finalArea)
		if norm == "world" || norm == "global" {
			norm = "all"
		}
		if val, ok := AreaAliases[norm]; ok {
			finalArea = val
		}
	}

	// 6. Final check: Do we have an area?
	// If no alias was found AND no manual area provided, we can't search.
	if finalArea == "" {
		return "", "", false, nil // Return empty to signal "Missing Area"
	}

	return finalIgn, finalArea, foundAlias, nil
}
