package main

import (
	"crypto/sha256"
	"image"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestDefeatedEnemyImageUpdatesWithoutLeavingRoom(t *testing.T) {
	ui := newCombatTestUI(t)
	look := lookView{Room: roomView{ID: "loc.argo_salmydessus", Name: "Salmydessus"}, NPCs: []string{"npc.harpy"}}
	ui.showRoom(look)
	standing := sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix)
	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":90,"target_hp":0,"damage":28,"status":"victory"}`)
	look.Defeated = []string{"npc.harpy"}
	ui.showRoom(look)
	defeated := sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix)
	if defeated == standing {
		t.Fatal("victory did not update the room image to the defeated enemy")
	}
	ui.showRoom(look)
	if sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix) != defeated {
		t.Fatal("an unchanged LOOK changed the defeated enemy image")
	}
	look.Defeated = nil
	ui.showRoom(look)
	if sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix) != standing {
		t.Fatal("the living enemy image was not restored when the defeated state cleared")
	}
}

func newCombatTestUI(t *testing.T) *gui {
	t.Helper()
	application := test.NewApp()
	t.Cleanup(application.Quit)
	window := application.NewWindow("test")
	t.Cleanup(window.Close)
	ui := &gui{window: window, locale: "en", catalog: &worldCatalog{NPCs: map[string]catalogEntry{
		"npc.harpy": {Name: localizedName{"en": "Harpy"}, Role: "enemy", HP: 40},
	}}}
	ui.build()
	return ui
}

func TestAttackShowsAndHidesCombatPanel(t *testing.T) {
	ui := newCombatTestUI(t)
	if ui.combatPanel.Visible() {
		t.Fatal("combat panel must start hidden")
	}

	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":90,"target_hp":28,"damage":12,"status":"combat"}`)
	if !ui.combatPanel.Visible() || ui.fight == nil || ui.fight.hp != 28 || ui.fight.maxHP != 40 {
		t.Fatalf("fight after hit = %+v, visible=%v", ui.fight, ui.combatPanel.Visible())
	}
	if ui.combatHP.Text != "Enemy HP: 28/40" {
		t.Fatalf("enemy HP label = %q", ui.combatHP.Text)
	}

	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":90,"target_hp":0,"damage":28,"status":"victory"}`)
	if ui.combatPanel.Visible() || ui.fight != nil {
		t.Fatal("victory must close the combat panel")
	}
}

func TestSuccessfulFleeAndMoveCloseCombatPanel(t *testing.T) {
	ui := newCombatTestUI(t)
	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":90,"target_hp":28,"damage":12,"status":"combat"}`)

	ui.handleResponse("FLEE", "FLEE", `OK {"result":"failed","hp":80}`)
	if !ui.combatPanel.Visible() {
		t.Fatal("a failed flee must keep the fight going")
	}
	ui.handleResponse("FLEE", "FLEE", `OK {"result":"success","hp":80}`)
	if ui.combatPanel.Visible() {
		t.Fatal("a successful flee must close the combat panel")
	}
}

func TestUnparsableAttackFallsBackToRawText(t *testing.T) {
	ui := newCombatTestUI(t)
	ui.handleResponse("ATTACK", "ATTACK npc.harpy", "OK something unexpected")
	last := ui.storyLines[len(ui.storyLines)-1]
	if last.text != "something unexpected" || ui.fight != nil {
		t.Fatalf("fallback story = %q, fight = %+v", last.text, ui.fight)
	}
}

func TestDefendResponseIsExplainedAndRefreshesStatus(t *testing.T) {
	ui := newCombatTestUI(t)
	ui.handleResponse("DEFEND", "DEFEND", `OK {"hp":80,"result":"braced"}`)
	last := ui.storyLines[len(ui.storyLines)-1]
	if last.kind != storyCombat || last.text != "You brace yourself. The next counter-attack will hurt half as much." {
		t.Fatalf("story = %+v", last)
	}
}

func TestAttackingPeopleHasNoFightPanelAndDeathsAreMarked(t *testing.T) {
	ui := newCombatTestUI(t)

	ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":100,"target_hp":9,"damage":11,"status":"wounded"}`)
	last := ui.storyLines[len(ui.storyLines)-1]
	if ui.combatPanel.Visible() || last.kind != storyCombat || last.text != "You strike Harpy for 11 damage. (9 HP left)" {
		t.Fatalf("after wounding: story = %+v, panel visible = %v", last, ui.combatPanel.Visible())
	}

	for status, want := range map[string]string{
		"murder":  "You killed Harpy. The Fates cut your thread.",
		"smitten": "You raised your hand against Harpy, and were struck dead before the blow landed.",
	} {
		ui.handleResponse("ATTACK", "ATTACK npc.harpy", `OK {"attacker_hp":0,"target_hp":0,"damage":0,"status":"`+status+`"}`)
		last = ui.storyLines[len(ui.storyLines)-1]
		if last.kind != storyDeath || last.text != want {
			t.Errorf("%s: story = %+v, want %q as a death line", status, last, want)
		}
	}
}

func TestCombatOutcomeAppearsOnceForTheActingPlayer(t *testing.T) {
	cases := []struct {
		name     string
		locale   string
		response serverMessage
		events   []string
		want     []string
	}{
		{
			name: "victory", locale: "en",
			response: serverMessage{kind: messageResponse, command: "ATTACK", request: "ATTACK npc.harpy", line: `OK {"attacker_hp":90,"target_hp":0,"damage":12,"status":"victory"}`},
			events:   []string{"EVT ROOM COMBAT alice defeats Harpy."},
			want:     []string{"Victory! Harpy is defeated."},
		},
		{
			name: "flee", locale: "ja",
			response: serverMessage{kind: messageResponse, command: "FLEE", request: "FLEE", line: `OK {"hp":90,"result":"success"}`},
			events:   []string{"EVT ROOM COMBAT aliceはハルピュイアから逃げ出した。"},
			want:     []string{"うまく逃げ切った。"},
		},
		{
			name: "defend", locale: "en",
			response: serverMessage{kind: messageResponse, command: "DEFEND", request: "DEFEND", line: `OK {"hp":90,"result":"braced"}`},
			events:   []string{"EVT ROOM COMBAT alice braces against Harpy."},
			want:     []string{"You brace yourself. The next counter-attack will hurt half as much."},
		},
		{
			name: "death outside hub", locale: "ja",
			response: serverMessage{kind: messageResponse, command: "ATTACK", request: "ATTACK npc.harpy", line: `OK {"attacker_hp":0,"target_hp":10,"damage":12,"status":"dead"}`},
			events:   []string{"EVT PLAYER DEATH ハルピュイアに打ち倒され、HPが尽きた。運命の間で目を覚ました。HPは20だ。"},
			want:     []string{"ハルピュイアに打ち倒され、HPが尽きた。運命の間で目を覚ました。HPは20だ。"},
		},
		{
			name: "death in hub", locale: "ja",
			response: serverMessage{kind: messageResponse, command: "ATTACK", request: "ATTACK npc.moirai", line: `OK {"attacker_hp":0,"target_hp":100,"damage":0,"status":"smitten"}`},
			events: []string{
				"EVT PLAYER DEATH モイライに手を上げた。一撃が届く前に、あなたは打ち殺された。運命の間で目を覚ました。HPは20だ。",
				"EVT ROOM COMBAT aliceはモイライに手を上げ、一撃が届く前に打ち殺された。",
			},
			want: []string{"モイライに手を上げた。一撃が届く前に、あなたは打ち殺された。運命の間で目を覚ました。HPは20だ。"},
		},
		{
			name: "death while fleeing", locale: "en",
			response: serverMessage{kind: messageResponse, command: "FLEE", request: "FLEE", line: `OK {"hp":20,"result":"failure_dead"}`},
			events:   []string{"EVT PLAYER DEATH You tried to flee from Harpy, but failed and were cut down. You awaken in the Hall with 20 HP."},
			want:     []string{"You tried to flee from Harpy, but failed and were cut down. You awaken in the Hall with 20 HP."},
		},
		{
			name: "external server without death event", locale: "en",
			response: serverMessage{kind: messageResponse, command: "ATTACK", request: "ATTACK npc.harpy", line: `OK {"attacker_hp":0,"target_hp":10,"damage":12,"status":"dead"}`},
			want:     []string{"Harpy struck you down."},
		},
		{
			name: "another player with a matching prefix", locale: "ja",
			response: serverMessage{kind: messageResponse, command: "ATTACK", request: "ATTACK npc.harpy", line: `OK {"attacker_hp":90,"target_hp":28,"damage":12,"status":"combat"}`},
			events: []string{
				"EVT ROOM COMBAT aliceは太郎はハルピュイアに5のダメージを与え、反撃で8のダメージを受けた。",
				"EVT ROOM COMBAT aliceはハルピュイアに12のダメージを与え、反撃で10のダメージを受けた。",
			},
			want: []string{
				"ハルピュイアに12ダメージ。敵の残りHP28、あなたのHP90。",
				"aliceは太郎はハルピュイアに5のダメージを与え、反撃で8のダメージを受けた。",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ui := newCombatTestUI(t)
			ui.locale = tc.locale
			ui.nameEntry.SetText("alice")
			ui.catalog.NPCs["npc.harpy"] = catalogEntry{Name: localizedName{"en": "Harpy", "ja": "ハルピュイア"}, Role: "enemy", HP: 40}
			ui.catalog.NPCs["npc.moirai"] = catalogEntry{Name: localizedName{"en": "Moirai", "ja": "モイライ"}}
			ui.room.Players = []string{"alice", "aliceは太郎"}
			start := len(ui.storyLines)
			for _, line := range tc.events {
				if len(line) > len("EVT PLAYER DEATH ") && line[:len("EVT PLAYER DEATH ")] == "EVT PLAYER DEATH " {
					ui.handleMessage(serverMessage{kind: messageEvent, line: line})
				}
			}
			ui.handleMessage(tc.response)
			for _, line := range tc.events {
				if len(line) > len("EVT ROOM COMBAT ") && line[:len("EVT ROOM COMBAT ")] == "EVT ROOM COMBAT " {
					ui.handleMessage(serverMessage{kind: messageEvent, line: line})
				}
			}
			got := ui.storyLines[start:]
			if len(got) != len(tc.want) {
				t.Fatalf("story entries = %v; want %v", got, tc.want)
			}
			for i, entry := range got {
				if entry.text != tc.want[i] {
					t.Errorf("entry %d = %q; want %q", i, entry.text, tc.want[i])
				}
			}
		})
	}
}
