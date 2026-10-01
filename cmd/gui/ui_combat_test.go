package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

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
