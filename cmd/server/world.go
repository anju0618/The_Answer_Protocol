// data/world.json をGoの構造体に読み込むデータモデルと、そのロード・検証処理。
// ワールドは「部屋(Room, room.go)」「アイテム(Item)」「NPC」「クエスト(Quest)」
// の4種類のマスタデータで構成される。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Item はワールド内に一意に存在するアイテム。Obtainable が true のものだけ
// TAKE できる(false のものは「取れない設定物」用)。
type Item struct {
	Name        LocalizedText `json:"name"`
	Description LocalizedText `json:"description"`
	RoomID      string        `json:"room_id"`
	Obtainable  bool          `json:"obtainable"`
}

// NPC はワールド内のNPC1体。Role は "dialogue"(会話専用)・
// "quest_giver"(クエスト付与)・"enemy"(戦闘可能)のいずれか。
// MythRequirementItem/Quest 以降は combat.go の神話ゲート判定
// (Player.meetsMythRequirement)で使う。
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
}

// QuestObjective はクエスト達成条件。Type は "collect_item"(アイテム所持)
// または "defeat_npc"(NPC撃破)。
type QuestObjective struct {
	Type     string `json:"type"`
	TargetID string `json:"target_id"`
	Count    int    `json:"count"`
}

// QuestReward はクエスト達成時にプレイヤーへ付与される報酬(HP回復量)。
type QuestReward struct {
	HP int `json:"hp"`
}

// Quest はNPCから受注できるクエスト1件。進行状況自体はプレイヤーごとに
// Player.Quests(player.go)で管理し、こちらは不変のマスタデータ。
type Quest struct {
	Name        LocalizedText  `json:"name"`
	Description LocalizedText  `json:"description"`
	GiverNPCID  string         `json:"giver_npc_id"`
	Objective   QuestObjective `json:"objective"`
	Reward      QuestReward    `json:"reward"`
}

// World はワールド全体(全部屋・全アイテム・全NPC・全クエスト)。
// data/world.json をそのままアンマーシャルしたもので、サーバー起動中は
// 読み取り専用のマスタデータとして扱う(可変なのはNPC.HP・Item.RoomIDなど、
// 各ハンドラがs.muの下で直接書き換える一部フィールドのみ)。
type World struct {
	StartRoomID string            `json:"start_room_id"`
	Rooms       map[string]*Room  `json:"rooms"`
	Items       map[string]*Item  `json:"items"`
	NPCs        map[string]*NPC   `json:"npcs"`
	Quests      map[string]*Quest `json:"quests"`
}

// loadWorld は path からワールドデータを読み込み、validate で整合性を
// 検証してから返す。
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

// validate はワールドデータの参照整合性(出口・アイテム・NPC・クエストが
// 指すIDが実在するか等)を検証する。ロード時に一度だけ呼ばれ、壊れた
// world.json でサーバーが起動しないようにするためのもの。
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
	return nil
}

// resolveNPCInRoom は roomID 内のNPCを、まずID完全一致、次に指定
// localeでの表示名の大文字小文字を無視した一致で探す。同名NPCが複数
// いる場合はID辞書順で最小のものを返す(TAKE/DROPのアイテム名解決と
// 同じ決定的なタイブレーク方式)。見つからなければ空文字列を返す。
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

// questByGiver は npcID がgiver_npc_idとして設定されているクエストを
// 返す。複数あった場合はID辞書順で最小のものを、無ければ ("", nil) を返す。
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
