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
