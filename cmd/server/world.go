package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Item struct {
	Name        LocalizedText `json:"name"`
	Description LocalizedText `json:"description"`
	RoomID      string        `json:"room_id"`
	Obtainable  bool          `json:"obtainable"`
	Renewable   bool          `json:"renewable,omitempty"`
	RewardOnly  bool          `json:"reward_only,omitempty"`

	HomeRoomID string `json:"-"`
}

type NPC struct {
	Name                 LocalizedText   `json:"name"`
	Description          LocalizedText   `json:"description"`
	Role                 string          `json:"role"`
	RoomID               string          `json:"room_id"`
	HP                   int             `json:"hp"`
	Dialogue             []LocalizedText `json:"dialogue"`
	MythRequirementItem  string          `json:"myth_requirement_item,omitempty"`
	MythRequirementQuest string          `json:"myth_requirement_quest,omitempty"`
	FleeAccurate         bool            `json:"flee_accurate,omitempty"`
	FleeSucceedsOnce     bool            `json:"flee_succeeds_once,omitempty"`
	Unwinnable           bool            `json:"unwinnable,omitempty"`
	CrewLossOnAttack     int             `json:"crew_loss_on_attack,omitempty"`
	Guide                bool            `json:"guide,omitempty"`
	Ending               *Ending         `json:"ending,omitempty"`
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
	Name        LocalizedText  `json:"name"`
	Description LocalizedText  `json:"description"`
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
	// Hints maps what killed a player (an NPC, room or item ID) to the Moirai's hint about it.
	Hints map[string]LocalizedText `json:"hints,omitempty"`
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
	for _, item := range world.Items {
		if item != nil {
			item.HomeRoomID = item.RoomID
		}
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
		if h := room.Hazard; h != nil {
			switch h.Type {
			case "lethal":
			case "item_gate":
				if w.Items[h.RequiredItemID] == nil {
					return fmt.Errorf("room %q hazard points to unknown item %q", id, h.RequiredItemID)
				}
			case "crew_gate":
				if h.CrewLoss < 1 {
					return fmt.Errorf("room %q crew_gate hazard has invalid crew_loss %d", id, h.CrewLoss)
				}
			case "crew_cost":
				if h.CrewLoss < 1 {
					return fmt.Errorf("room %q crew_cost hazard has invalid crew_loss %d", id, h.CrewLoss)
				}
			default:
				return fmt.Errorf("room %q has unknown hazard type %q", id, h.Type)
			}
		}
	}
	for subject := range w.Hints {
		if w.Rooms[subject] == nil && w.Items[subject] == nil && w.NPCs[subject] == nil {
			return fmt.Errorf("hint is about unknown room, item or NPC %q", subject)
		}
	}
	for id, item := range w.Items {
		if id == "" {
			return fmt.Errorf("item ID is empty")
		}
		if item == nil {
			return fmt.Errorf("item %q is null", id)
		}
		if item.RewardOnly {
			if item.RoomID != "" || item.Obtainable {
				return fmt.Errorf("reward-only item %q must have no room and must not be obtainable", id)
			}
		} else if w.Rooms[item.RoomID] == nil {
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
		if npc.MythRequirementItem != "" && w.Items[npc.MythRequirementItem] == nil {
			return fmt.Errorf("NPC %q myth requirement points to unknown item %q", id, npc.MythRequirementItem)
		}
		if npc.MythRequirementQuest != "" && w.Quests[npc.MythRequirementQuest] == nil {
			return fmt.Errorf("NPC %q myth requirement points to unknown quest %q", id, npc.MythRequirementQuest)
		}
		if npc.Unwinnable && npc.CrewLossOnAttack < 1 {
			return fmt.Errorf("NPC %q is unwinnable but has invalid crew_loss_on_attack %d", id, npc.CrewLossOnAttack)
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
	return w.validateEndings()
}

func (w *World) resolveNPCInRoom(roomID, query, locale string) string {
	if npc := w.NPCs[query]; npc != nil && npc.RoomID == roomID {
		return query
	}
	npcID := ""
	for id, npc := range w.NPCs {
		if npc != nil && npc.RoomID == roomID && strings.EqualFold(npc.Name.Get(locale), query) && (npcID == "" || id < npcID) {
			npcID = id
		}
	}
	return npcID
}

func (w *World) questByGiver(npcID string) (string, *Quest) {
	questID := ""
	for id, quest := range w.Quests {
		if quest != nil && quest.GiverNPCID == npcID && (questID == "" || id < questID) {
			questID = id
		}
	}
	if questID == "" {
		return "", nil
	}
	return questID, w.Quests[questID]
}

func (item *Item) availableTo(player *Player, id string) bool {
	return item.Obtainable && item.visibleTo(player, id, player.RoomID)
}

func (item *Item) visibleTo(player *Player, id, roomID string) bool {
	if item.RoomID != roomID {
		return false
	}
	return !(item.Renewable && player.hasItem(id))
}
