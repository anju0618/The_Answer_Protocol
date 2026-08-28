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
	if _, ok := w.Rooms[w.StartRoomID]; !ok {
		return fmt.Errorf("start room %q does not exist", w.StartRoomID)
	}
	for id, room := range w.Rooms {
		for dir, dest := range room.Exits {
			if _, ok := w.Rooms[dest]; !ok {
				return fmt.Errorf("room %q exit %q points to unknown room %q", id, dir, dest)
			}
		}
	}
	return nil
}
