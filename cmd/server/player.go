// プレイヤーの永続状態(Player)と、それに関する判定用の小さなヘルパー群。
package main

// PlayerQuest はプレイヤー1人・クエスト1件ぶんの進行状況。
// Status は "active" または "completed"。
type PlayerQuest struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

// Player は1プレイヤーぶんのセーブデータ兼サーバー内部状態。
// playerdata.json にこの構造体がそのままJSONとして永続化される
// (player_store.go)。Crew以降のフィールドはオデュッセイア編の
// 神話ゲート/戦闘システム用で、RFC規定のレスポンスJSONには含めない。
type Player struct {
	Name            string                  `json:"name"`
	HP              int                     `json:"hp"`
	RoomID          string                  `json:"room_id"`
	Inventory       []string                `json:"inventory"`
	Crew            int                     `json:"crew,omitempty"`
	CrewInitialized bool                    `json:"crew_initialized,omitempty"`
	CombatTargetID  string                  `json:"combat_target_id,omitempty"`
	FledFrom        map[string]bool         `json:"fled_from,omitempty"`
	Quests          map[string]*PlayerQuest `json:"quests,omitempty"`
	exiting         bool
}

// hasItem は itemID を所持しているかを返す。
func (p *Player) hasItem(itemID string) bool {
	for _, id := range p.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}

// meetsMythRequirement は p が npc の神話ゲート(必要アイテム所持・
// 必要クエスト達成)を満たしているかを返す。どちらの条件も設定されて
// いなければ常にtrue(ゲート無し)。
func (p *Player) meetsMythRequirement(npc *NPC) bool {
	if npc.MythRequirementItem != "" && !p.hasItem(npc.MythRequirementItem) {
		return false
	}
	if npc.MythRequirementQuest != "" {
		state := p.Quests[npc.MythRequirementQuest]
		if state == nil || state.Status != "completed" {
			return false
		}
	}
	return true
}

// hasMythRequirement は npc に何らかの神話ゲート(アイテムまたは
// クエスト)が設定されているかを返す。
func (npc *NPC) hasMythRequirement() bool {
	return npc.MythRequirementItem != "" || npc.MythRequirementQuest != ""
}
