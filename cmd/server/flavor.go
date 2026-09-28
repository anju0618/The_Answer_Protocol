package main

import (
	"strconv"
	"strings"
)

type flavor struct {
	key    string
	player string
	npc    *NPC
	room   *Room
	n      int
	m      int
}

var flavorTexts = map[string]LocalizedText{
	"attack_unwinnable": {
		"en": "{player} attacks {npc} and is driven back, losing {n} crew.",
		"ja": "{player}は{npc}に挑んだが押し返され、仲間を{n}人失った。",
	},
	"attack_unprepared": {
		"en": "{player} attacks {npc} unprepared and is killed.",
		"ja": "{player}は備えのないまま{npc}に挑み、殺された。",
	},
	"attack_defeat": {
		"en": "{player} defeats {npc}.",
		"ja": "{player}は{npc}を打ち倒した。",
	},
	"attack_struck_down": {
		"en": "{player} is struck down by {npc}.",
		"ja": "{player}は{npc}に打ち倒された。",
	},
	"attack_hit": {
		"en": "{player} attacks {npc} for {n} damage and takes {m} in return.",
		"ja": "{player}は{npc}に{n}のダメージを与え、{m}の反撃を受けた。",
	},
	"flee_success": {
		"en": "{player} flees from {npc}.",
		"ja": "{player}は{npc}から逃げ出した。",
	},
	"flee_dead": {
		"en": "{player} tries to flee {npc} and is cut down.",
		"ja": "{player}は{npc}から逃げようとして斬り伏せられた。",
	},
	"flee_hit": {
		"en": "{player} tries to flee {npc} and is struck for {n}.",
		"ja": "{player}は{npc}から逃げようとして{n}のダメージを受けた。",
	},
	"slip_past": {
		"en": "{player} tries to slip past {npc} and doesn't make it.",
		"ja": "{player}は{npc}の脇をすり抜けようとしたが、逃げ切れなかった。",
	},
	"talk_unprepared": {
		"en": "{player} speaks with {npc} unprepared, and does not survive it.",
		"ja": "{player}は備えのないまま{npc}と言葉を交わし、生きて帰れなかった。",
	},
	"hazard_lethal": {
		"en": "{player} is lost to {room}.",
		"ja": "{player}は{room}に飲み込まれた。",
	},
	"hazard_item": {
		"en": "{player}, unprepared, does not survive {room}.",
		"ja": "{player}は備えのないまま{room}に踏み込み、生きて抜けられなかった。",
	},
	"hazard_crew_dead": {
		"en": "{player} and the remaining crew are lost passing {room}.",
		"ja": "{player}と残りの仲間は{room}を通り抜けようとして全滅した。",
	},
	"hazard_crew_loss": {
		"en": "{player} passes {room}, losing {n} crew.",
		"ja": "{player}は{room}を通り抜けたが、仲間を{n}人失った。",
	},
	"lotus": {
		"en": "{player} tastes the lotus, and forgets there ever was a home to return to.",
		"ja": "{player}はロトスの実を口にし、帰るべき故郷があったことを忘れてしまった。",
	},
	"cattle": {
		"en": "{player} lays a hand on the cattle of Helios, and the sky answers.",
		"ja": "{player}がヘリオスの牛に手をかけると、天が応えた。",
	},
	"winds_opened": {
		"en": "{player} opens the bag of winds too soon, and a storm drives the ship back, costing {n} crew.",
		"ja": "{player}は風の革袋を早まって開けてしまい、嵐に船を押し戻されて仲間を{n}人失った。",
	},
}

func (f flavor) text(locale string) string {
	npcName, roomName := "", ""
	if f.npc != nil {
		npcName = f.npc.Name.Get(locale)
	}
	if f.room != nil {
		roomName = f.room.Name.Get(locale)
	}
	return strings.NewReplacer(
		"{player}", f.player,
		"{npc}", npcName,
		"{room}", roomName,
		"{n}", strconv.Itoa(f.n),
		"{m}", strconv.Itoa(f.m),
	).Replace(flavorTexts[f.key].Get(locale))
}

func (s *Server) broadcastFlavorLocked(roomID string, f flavor) {
	for playerName, other := range s.players {
		if other.RoomID != roomID {
			continue
		}
		if recipient := s.clients[playerName]; recipient != nil {
			recipient.enqueueEvent("EVT ROOM COMBAT " + f.text(s.localeOfLocked(playerName)))
		}
	}
}
