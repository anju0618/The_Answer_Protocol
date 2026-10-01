package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type roomView struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

type lookView struct {
	Room    roomView `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
}

type statusView struct {
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Status string `json:"status"`
}

type stateView struct {
	Crew            int      `json:"crew"`
	CrewInitialized bool     `json:"crew_initialized"`
	Players         []string `json:"players"`
	Group           string   `json:"group"`
	Invitations     []string `json:"invitations"`
}

type questView struct {
	QuestID  string `json:"quest_id"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

type localizedName map[string]string

func (name localizedName) get(locale string) string {
	if value := name[locale]; value != "" {
		return value
	}
	return name["en"]
}

type catalogEntry struct {
	Name        localizedName     `json:"name"`
	Description localizedName     `json:"description"`
	Role        string            `json:"role"`
	GiverNPCID  string            `json:"giver_npc_id"`
	HP          int               `json:"hp"`
	Exits       map[string]string `json:"exits"`
	Ending      *struct {
		ID         string        `json:"id"`
		Name       localizedName `json:"name"`
		RewardItem string        `json:"reward_item"`
	} `json:"ending"`
	Hazard *struct {
		Type string `json:"type"`
	} `json:"hazard"`
}

type worldCatalog struct {
	Rooms  map[string]catalogEntry `json:"rooms"`
	Items  map[string]catalogEntry `json:"items"`
	NPCs   map[string]catalogEntry `json:"npcs"`
	Quests map[string]catalogEntry `json:"quests"`
}

func loadCatalog() (*worldCatalog, error) {
	var candidates []string
	for _, base := range []string{workingDirectory(), executableDirectory()} {
		for range 4 {
			candidates = append(candidates, filepath.Join(base, "data", "world.json"))
			base = filepath.Dir(base)
		}
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var catalog worldCatalog
		if err := json.Unmarshal(data, &catalog); err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return &catalog, nil
	}
	return nil, fmt.Errorf("data/world.json was not found")
}

func workingDirectory() string {
	path, _ := os.Getwd()
	return path
}

func executableDirectory() string {
	path, _ := os.Executable()
	return filepath.Dir(path)
}

func (catalog *worldCatalog) label(kind, id, locale string) string {
	if catalog == nil {
		return id
	}
	var entry catalogEntry
	switch kind {
	case "room":
		entry = catalog.Rooms[id]
	case "item":
		entry = catalog.Items[id]
	case "npc":
		entry = catalog.NPCs[id]
	case "quest":
		entry = catalog.Quests[id]
	}
	if name := entry.Name.get(locale); name != "" {
		return name
	}
	return id
}

func decodeOK(line string, target any) error {
	if !strings.HasPrefix(line, "OK ") {
		return fmt.Errorf("unexpected response %q", line)
	}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "OK ")), target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (catalog *worldCatalog) gameOverRoom(roomID string) (catalogEntry, bool) {
	if catalog == nil {
		return catalogEntry{}, false
	}
	entry, ok := catalog.Rooms[roomID]
	if !ok || entry.Hazard == nil || entry.Hazard.Type != "lethal" {
		return catalogEntry{}, false
	}
	return entry, true
}
