package main

import (
	"fmt"
	"sort"
	"strings"
)

func (t LocalizedText) Format(locale string, args ...any) string {
	return fmt.Sprintf(t.Get(locale), args...)
}

func (s *Server) localeOfLocked(name string) string {
	if client := s.clients[name]; client != nil && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}

func (s *Server) sendPlayerEventLocked(name, kind, text string) {
	client := s.clients[name]
	if client == nil {
		return
	}
	text = strings.NewReplacer("\r", " ", "\n", " ").Replace(text)
	client.enqueueEvent("EVT PLAYER " + kind + " " + text)
}

var deathTexts = map[string]LocalizedText{

	"attack_counter": {
		"en": "You were struck down by %s and your HP ran out.",
		"ja": "%sに打ち倒され、HPが尽きた。",
	},

	"attack_unprepared": {
		"en": "You attacked %s and were killed instantly.",
		"ja": "%sに挑み、即座に殺された。",
	},

	"attack_mighty": {
		"en": "You raised your hand against %s, and were struck dead before the blow landed.",
		"ja": "%sに手を上げた。一撃が届く前に、あなたは打ち殺された。",
	},

	"attack_murder": {
		"en": "You killed %s, and the Fates cut your thread for it.",
		"ja": "%sを手にかけた。運命の女神たちは、その報いにあなたの糸を断ち切った。",
	},

	"flee_failed": {
		"en": "You tried to flee from %s, but failed and were cut down.",
		"ja": "%sから逃げようとしたが失敗し、斬り伏せられた。",
	},

	"slip_past": {
		"en": "You tried to leave while %s still blocked your way, and it cut you down.",
		"ja": "%sが道をふさいでいる間に立ち去ろうとして、斬り伏せられた。",
	},

	"talk_unprepared": {
		"en": "You spoke with %s, and it was fatal.",
		"ja": "%sと言葉を交わし、命を落とした。",
	},

	"hazard_lethal": {
		"en": "You did not survive %s.",
		"ja": "%sを生きて抜けることはできなかった。",
	},

	"hazard_item": {
		"en": "You did not survive %s.",
		"ja": "%sを生きて抜けることはできなかった。",
	},

	"hazard_crew": {
		"en": "Your crew was lost passing %s.",
		"ja": "%sを通り抜けようとして、仲間もろとも全滅した。",
	},
	"lotus": {
		"en": "You tasted the lotus and forgot the way home.",
		"ja": "ロトスの実を口にして、故郷への道を忘れてしまった。",
	},
	"cattle": {
		"en": "You laid hands on the sacred cattle of Helios, and the gods answered with death.",
		"ja": "ヘリオスの聖なる牛に手をかけ、神々は死をもって応えた。",
	},
}

var respawnTail = LocalizedText{
	"en": "You awaken in %s with %d HP.",
	"ja": "%sで目を覚ました。HPは%dに減っている。",
}

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

var enemiesRecoverTail = LocalizedText{
	"en": "Every enemy you had wounded has recovered.",
	"ja": "傷つけた敵はすべて全快した。",
}

func (s *Server) notifyDeathLocked(name, cause string, outcome deathOutcome, args ...any) {
	locale := s.localeOfLocked(name)
	text, ok := deathTexts[cause]
	if !ok {
		return
	}
	message := strings.TrimSpace(text.Format(locale, args...))
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

func (s *Server) sendDeathHintLocked(player *Player, name string) {
	subject := player.LastDeathSubject
	if subject == "" || s.world == nil {
		return
	}
	player.LastDeathSubject = ""
	hint, ok := s.world.Hints[subject]
	if !ok {
		return
	}
	s.sendPlayerEventLocked(name, "HINT", hint.Get(s.localeOfLocked(name)))
}

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
