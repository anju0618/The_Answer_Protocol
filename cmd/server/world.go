package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RoomID      string `json:"room_id"`
	Obtainable  bool   `json:"obtainable"`
}

type NPC struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Role        string   `json:"role"`
	RoomID      string   `json:"room_id"`
	HP          int      `json:"hp"`
	Dialogue    []string `json:"dialogue"`
}

type QuestObjective struct {
	Type     string `json:"type"`
	TargetID string `json:"target_id"`
	Count    int    `json:"count"`
}

type QuestReward struct {
	HP int `json:"hp"`
}

type Quest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	GiverNPCID  string         `json:"giver_npc_id"`
	Objective   QuestObjective `json:"objective"`
	Reward      QuestReward    `json:"reward"`
}

type World struct {
	StartRoomID string            `json:"start_room_id"`
	Rooms       map[string]*Room  `json:"rooms"`
	Items       map[string]*Item  `json:"items"`
	NPCs        map[string]*NPC   `json:"npcs"`
	Quests      map[string]*Quest `json:"quests"`
}

func loadWorld(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read world data: %w", err)
	}

	var world World
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("decode world data: %w", err)
	}
	if err := world.validate(); err != nil {
		return nil, err
	}
	return &world, nil
}

func (w *World) validate() error {
	if w == nil {
		return fmt.Errorf("world is null")
	}
	if w.Rooms[w.StartRoomID] == nil {
		return fmt.Errorf("start room %q does not exist", w.StartRoomID)
	}
	for id, room := range w.Rooms {
		if id == "" {
			return fmt.Errorf("room ID is empty")
		}
		if room == nil {
			return fmt.Errorf("room %q is null", id)
		}
		if room.ID != id {
			return fmt.Errorf("room %q has ID %q", id, room.ID)
		}
		for dir, dest := range room.Exits {
			if w.Rooms[dest] == nil {
				return fmt.Errorf("room %q exit %q points to unknown room %q", id, dir, dest)
			}
		}
	}
	for id, item := range w.Items {
		if id == "" {
			return fmt.Errorf("item ID is empty")
		}
		if item == nil {
			return fmt.Errorf("item %q is null", id)
		}
		if w.Rooms[item.RoomID] == nil {
			return fmt.Errorf("item %q points to unknown room %q", id, item.RoomID)
		}
	}
	for id, npc := range w.NPCs {
		if id == "" {
			return fmt.Errorf("NPC ID is empty")
		}
		if npc == nil {
			return fmt.Errorf("NPC %q is null", id)
		}
		if w.Rooms[npc.RoomID] == nil {
			return fmt.Errorf("NPC %q points to unknown room %q", id, npc.RoomID)
		}
	}
	for id, quest := range w.Quests {
		if id == "" {
			return fmt.Errorf("quest ID is empty")
		}
		if quest == nil {
			return fmt.Errorf("quest %q is null", id)
		}
		if w.NPCs[quest.GiverNPCID] == nil {
			return fmt.Errorf("quest %q points to unknown giver NPC %q", id, quest.GiverNPCID)
		}
		if quest.Objective.Count < 1 {
			return fmt.Errorf("quest %q has invalid objective count %d", id, quest.Objective.Count)
		}
		switch quest.Objective.Type {
		case "collect_item":
			if w.Items[quest.Objective.TargetID] == nil {
				return fmt.Errorf("quest %q points to unknown item %q", id, quest.Objective.TargetID)
			}
		case "defeat_npc":
			if w.NPCs[quest.Objective.TargetID] == nil {
				return fmt.Errorf("quest %q points to unknown NPC %q", id, quest.Objective.TargetID)
			}
		default:
			return fmt.Errorf("quest %q has unknown objective type %q", id, quest.Objective.Type)
		}
	}
	return nil
}
