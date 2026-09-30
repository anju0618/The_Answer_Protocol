// プレイヤーの永続状態(Player)と、それに関する判定用の小さなヘルパー群。
package main

import "time"

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
	IntroSeen       bool                    `json:"intro_seen,omitempty"`
	CombatTargetID  string                  `json:"combat_target_id,omitempty"`
	FledFrom        map[string]bool         `json:"fled_from,omitempty"`
	Quests          map[string]*PlayerQuest `json:"quests,omitempty"`
	Endings         map[string]bool         `json:"endings,omitempty"`
	EnemyHP         map[string]int          `json:"enemy_hp,omitempty"`
	lastRegen       time.Time
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

// enemyHP は player から見た npc の現在HPを返す。敵のHPは全員共通の1つ
// ではなく「プレイヤーごと」に持つ(EnemyHPに記録がなければ npc.HP=最大値)。
// 共通だと、最初に倒した1人以外はその敵と戦えず、defeat_npcクエストも
// 達成できなくなってしまうため。
func (p *Player) enemyHP(npcID string, npc *NPC) int {
	if hp, ok := p.EnemyHP[npcID]; ok {
		return hp
	}
	return npc.HP
}

// setEnemyHP は player から見た npc の現在HPを記録する。
func (p *Player) setEnemyHP(npcID string, hp int) {
	if p.EnemyHP == nil {
		p.EnemyHP = make(map[string]int)
	}
	p.EnemyHP[npcID] = hp
}

const (
	regenInterval = 2 * time.Second
	regenAmount   = 1
)

// regenLocked は前回からの経過時間ぶん、player のHPを自然回復させる
// (regenInterval ごとに regenAmount、上限maxPlayerHP)。HPが減ったままだと
// 戦闘が続けられず、死んだ後(HP20で復活)に詰んでしまうため。時計は
// 満タンのときは進めない。呼び出しごとに遅延評価するので、タイマーgoroutine
// は不要。
func (p *Player) regenLocked(now time.Time) {
	if p.HP >= maxPlayerHP || p.lastRegen.IsZero() {
		p.lastRegen = now
		return
	}
	ticks := int(now.Sub(p.lastRegen) / regenInterval)
	if ticks <= 0 {
		return
	}
	p.HP += ticks * regenAmount
	if p.HP >= maxPlayerHP {
		p.HP = maxPlayerHP
		p.lastRegen = now
		return
	}
	p.lastRegen = p.lastRegen.Add(time.Duration(ticks) * regenInterval)
}
