// プレイヤー1人だけに届く通知(EVT PLAYER ...)と、死亡理由の文言、
// 運命の間の案内(ガイド)をまとめたファイル。
//
// これまでの「死」はEVT ROOM COMBATを死んだ部屋へ流すだけだったが、
// 死んだ本人はその時点でもう運命の間へ送られているため、自分の死因が
// 一切分からなかった。ここでは本人宛てに「何が起きたか・次にどうすれば
// よいか」を届ける。RFCが定義するJSONレスポンスの形は一切変えず、
// 新しいEVTのタイプ(PLAYER)を足すだけにしている(READMEの
// 「Protocol Implementation」のEVT ROOM COMBATと同じ考え方)。
//
//	EVT PLAYER DEATH <text>  死亡理由とリスポーン先の説明
//	EVT PLAYER GUIDE <text>  運命の間のチュートリアル(初回接続時/モイライへのTALK)
//	EVT PLAYER QUEST <text>  クエストの案内・進捗・達成通知
package main

import (
	"fmt"
	"sort"
	"strings"
)

// Format は t の locale 版テキストを fmt.Sprintf の書式として展開する。
func (t LocalizedText) Format(locale string, args ...any) string {
	return fmt.Sprintf(t.Get(locale), args...)
}

// localeOfLocked は接続中プレイヤー name の言語を返す(未接続なら既定言語)。
func (s *Server) localeOfLocked(name string) string {
	if client := s.clients[name]; client != nil && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}

// sendPlayerEventLocked は name 1人だけに "EVT PLAYER <kind> <text>" を送る。
// 呼び出し側はs.muを保持していること。1行=1メッセージのため、textに
// 改行が混ざらないようここで空白に置き換える。
func (s *Server) sendPlayerEventLocked(name, kind, text string) {
	client := s.clients[name]
	if client == nil {
		return
	}
	text = strings.NewReplacer("\r", " ", "\n", " ").Replace(text)
	client.enqueueEvent("EVT PLAYER " + kind + " " + text)
}

// deathTexts は死亡理由の文言(fmt書式)。キーは respawnPlayerLocked の
// cause に渡す。各文言が受け取る引数はコメントのとおり。
var deathTexts = map[string]LocalizedText{
	// 引数: 敵の名前
	"attack_counter": {
		"en": "You were struck down by %s and your HP ran out. Tip: check STATUS often, and use FLEE to escape before your HP gets low.",
		"ja": "%sに打ち倒され、HPが尽きた。ヒント: STATUSでHPをこまめに確認し、危なくなる前にFLEEで逃げよう。",
	},
	// 引数: 敵の名前, 必要条件の説明
	"attack_unprepared": {
		"en": "%s is protected by an old myth, and you attacked it unprepared. You were killed instantly. %s",
		"ja": "%sは古い神話に守られており、備えのないまま挑んだあなたは即座に殺された。%s",
	},
	// 引数: 敵の名前
	"flee_failed": {
		"en": "You tried to flee from %s, but failed and were cut down. Some enemies can only be escaped from once, so think of another way.",
		"ja": "%sから逃げようとしたが失敗し、斬り伏せられた。一度しか逃げられない敵もいる。別の手を考えよう。",
	},
	// 引数: 敵の名前
	"slip_past": {
		"en": "You tried to leave while %s still blocked your way, and it cut you down. Defeat it with ATTACK, or escape it with FLEE first.",
		"ja": "%sが道をふさいでいる間に立ち去ろうとして、斬り伏せられた。ATTACKで倒すか、先にFLEEで振り切ろう。",
	},
	// 引数: NPCの名前, 必要条件の説明
	"talk_unprepared": {
		"en": "%s cannot be faced unprepared, and speaking with them was fatal. %s",
		"ja": "%sは備えなしに向き合える相手ではなく、話しかけたことが命取りになった。%s",
	},
	// 引数: 部屋の名前
	"hazard_lethal": {
		"en": "There is no surviving %s. Some places must be avoided altogether: use LOOK to find another exit.",
		"ja": "%sを生きて抜けることはできない。避けるべき場所もある。LOOKで別の出口を探そう。",
	},
	// 引数: 部屋の名前, 必要アイテム名
	"hazard_item": {
		"en": "You entered %s without the protection you needed. Next time, bring \"%s\" with you (TAKE it before you go).",
		"ja": "%sに、必要な備えなしで踏み込んでしまった。次は「%s」を持ってから来よう(先にTAKEしておく)。",
	},
	// 引数: 部屋の名前, 必要人数, 現在の人数
	"hazard_crew": {
		"en": "Your crew was too small to get through %s (at least %d people were needed, but you had %d). Do not waste crew earlier on the voyage.",
		"ja": "%sを抜けるには仲間が足りなかった(最低%d人必要だが、%d人しかいなかった)。航海の序盤で仲間を無駄にしないこと。",
	},
	"lotus": {
		"en": "You tasted the lotus and forgot the way home. Tempting items can be traps: think before you TAKE.",
		"ja": "ロトスの実を口にして、故郷への道を忘れてしまった。魅力的なアイテムは罠かもしれない。TAKEの前によく考えよう。",
	},
	"cattle": {
		"en": "You laid hands on the sacred cattle of Helios, and the gods answered with death. Some things must not be taken, however hungry you are.",
		"ja": "ヘリオスの聖なる牛に手をかけ、神々は死をもって応えた。どれほど飢えていても、手を出してはいけないものがある。",
	},
}

// respawnTail は死亡メッセージの末尾に付ける「どこで復活したか」の説明。
// 引数: 復活先の部屋名, 復活後のHP。
var respawnTail = LocalizedText{
	"en": "You awaken in %s with %d HP.",
	"ja": "%sで目を覚ました。HPは%dに減っている。",
}

// outcomeTexts は死亡時の持ち物の扱い(hardcore.go)の説明文。
var outcomeTexts = map[deathOutcome]LocalizedText{
	outcomeLost: {
		"en": "You lost everything you were carrying except your trophies, so fetch what you need again.",
		"ja": "記念品を除いて、持ち物はすべて失われた。必要なものは取り直そう。",
	},
	outcomeKept: {
		"en": "Your companions dragged your belongings to safety, so you lost nothing.",
		"ja": "仲間が持ち物を守ってくれたので、何も失わずに済んだ。",
	},
}

// enemiesRecoverTail は死亡すると傷つけた敵が全快することの説明。
var enemiesRecoverTail = LocalizedText{
	"en": "Every enemy you had wounded has recovered.",
	"ja": "傷つけた敵はすべて全快した。",
}

// notifyDeathLocked は name 本人へEVT PLAYER DEATHで死亡理由を伝える。
// cause は deathTexts のキー、args はその文言の引数。呼び出し側は、
// args の中の名前などを name の言語(s.localeOfLocked(name))で解決して
// 渡すこと(通常は同じ接続のclientLocaleがそのまま使える)。
func (s *Server) notifyDeathLocked(name, cause string, outcome deathOutcome, args ...any) {
	locale := s.localeOfLocked(name)
	text, ok := deathTexts[cause]
	if !ok {
		return
	}
	message := text.Format(locale, args...)
	if s.world != nil {
		if room := s.world.Rooms[s.world.StartRoomID]; room != nil {
			message += " " + respawnTail.Format(locale, room.Name.Get(locale), respawnHP)
		}
	}
	if tail, ok := outcomeTexts[outcome]; ok {
		message += " " + tail.Get(locale)
	}
	message += " " + enemiesRecoverTail.Get(locale)
	s.sendPlayerEventLocked(name, "DEATH", message)
}

// mythNeedText は npc の神話ゲートを突破するために必要だったものを、
// locale の言語で説明する(死亡メッセージのヒント用)。
func (s *Server) mythNeedText(npc *NPC, locale string) string {
	var needs []string
	if npc.MythRequirementItem != "" && s.world != nil {
		if item := s.world.Items[npc.MythRequirementItem]; item != nil {
			needs = append(needs, LocalizedText{
				"en": "obtain \"%s\"",
				"ja": "「%s」を手に入れる",
			}.Format(locale, item.Name.Get(locale)))
		}
	}
	if npc.MythRequirementQuest != "" && s.world != nil {
		if quest := s.world.Quests[npc.MythRequirementQuest]; quest != nil {
			needs = append(needs, LocalizedText{
				"en": "complete the quest \"%s\"",
				"ja": "クエスト「%s」を達成する",
			}.Format(locale, quest.Name.Get(locale)))
		}
	}
	if len(needs) == 0 {
		return ""
	}
	if locale == "ja" {
		return "先にやるべきこと: " + strings.Join(needs, "、") + "。"
	}
	return "What you needed first: " + strings.Join(needs, " and ") + "."
}

// guideNPC は開始部屋にいるガイドNPC(world.jsonで "guide": true が付いた
// NPC。運命の間のモイライ)を返す。複数いればID辞書順で最小のもの、
// いなければnil。
func (s *Server) guideNPC() *NPC {
	if s.world == nil {
		return nil
	}
	ids := make([]string, 0)
	for id, npc := range s.world.NPCs {
		if npc != nil && npc.Guide && npc.RoomID == s.world.StartRoomID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)
	return s.world.NPCs[ids[0]]
}

// sendGuideLocked はガイドNPCのセリフのうち index 番目以降を、すべて
// EVT PLAYER GUIDEとして name に送る。初回接続時は0番目から全部
// (handleConnect)、モイライへのTALKでは0番目がOKレスポンスに乗るため
// 1番目から(handleTalk)送る。
func (s *Server) sendGuideLocked(name string, index int) {
	guide := s.guideNPC()
	if guide == nil {
		return
	}
	locale := s.localeOfLocked(name)
	for i := index; i < len(guide.Dialogue); i++ {
		if line := guide.Dialogue[i].Get(locale); strings.TrimSpace(line) != "" {
			s.sendPlayerEventLocked(name, "GUIDE", line)
		}
	}
}
