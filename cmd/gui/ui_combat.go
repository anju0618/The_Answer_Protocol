package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type attackView struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

type fightState struct {
	npcID string
	hp    int
	maxHP int
}

func (ui *gui) buildCombatPanel() {
	ui.combatName = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.combatName.Truncation = fyne.TextTruncateEllipsis
	ui.combatHP = compactLabel("")
	ui.combatBar = newStatBar()
	attack := ui.commandButton(ui.tr("Attack", "戦う"), func() {
		if ui.fight != nil {
			ui.send("ATTACK " + ui.fight.npcID)
		}
	})
	flee := ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") })
	defend := ui.commandButton(ui.tr("Brace", "構える"), func() { ui.send("DEFEND") })
	frame := canvas.NewRectangle(ink)
	frame.StrokeColor, _ = storyColor(colorStoryCombat)
	frame.StrokeWidth = 2
	inside := textVBox(
		ui.combatName,
		container.NewStack(ui.combatBar.box, ui.combatHP),
		container.NewGridWithColumns(3, attack, defend, flee),
	)
	ui.combatPanel = container.NewStack(frame, container.NewPadded(inside))
	ui.combatPanel.Hide()
}

func (ui *gui) showFight(fight *fightState) {
	ui.fight = fight
	if fight == nil {
		ui.combatPanel.Hide()
		return
	}
	ui.combatName.SetText(ui.tr("In combat: ", "戦闘中: ") + ui.catalog.label("npc", fight.npcID, ui.locale))
	ui.combatHP.SetText(fmt.Sprintf(ui.tr("Enemy HP: %d/%d", "敵のHP: %d/%d"), max(fight.hp, 0), fight.maxHP))
	ui.combatBar.Set(fight.hp, fight.maxHP)
	ui.combatPanel.Show()
}

func (ui *gui) handleAttack(request, line string) bool {
	var result attackView
	if decodeOK(line, &result) != nil || result.Status == "" {
		return false
	}
	npcID := strings.TrimSpace(strings.TrimPrefix(request, "ATTACK"))
	enemy := ui.catalog.label("npc", npcID, ui.locale)
	switch result.Status {
	case "combat":
		maxHP := ui.catalog.maxHP(npcID, result.TargetHP)
		if ui.fight != nil && ui.fight.npcID == npcID {
			maxHP = max(maxHP, ui.fight.maxHP)
		}
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("You hit %s for %d damage. (%d HP left)", "%s に %d ダメージ!(残りHP %d)"), enemy, result.Damage, result.TargetHP))
		ui.showFight(&fightState{npcID: npcID, hp: result.TargetHP, maxHP: maxHP})
		ui.flashScene(flashHurt, 350*time.Millisecond)
	case "victory":
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("Victory! %s is defeated.", "勝利!%s を倒した。"), enemy))
		ui.showFight(nil)
	case "dead":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("%s struck you down.", "%s にやられた。"), enemy))
		ui.showFight(nil)
	case "wounded":
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("You strike %s for %d damage. (%d HP left)", "%s に %d ダメージを与えた。(残りHP %d)"), enemy, result.Damage, result.TargetHP))
	case "murder":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("You killed %s. The Fates cut your thread.", "%s を手にかけた。運命の女神たちがあなたの糸を断ち切った。"), enemy))
		ui.showFight(nil)
	case "smitten":
		ui.addStoryKind(storyDeath, fmt.Sprintf(ui.tr("You raised your hand against %s, and were struck dead before the blow landed.", "%s に手を上げた。一撃が届く前に打ち殺された。"), enemy))
		ui.showFight(nil)
	case "overwhelmed":
		ui.addStoryKind(storyCombat, fmt.Sprintf(ui.tr("%s is too strong to beat by force.", "%s は力では敵わない。"), enemy))
		ui.showFight(nil)
	default:
		return false
	}
	return true
}

func (ui *gui) handleFlee(line string) bool {
	var result struct {
		Result string `json:"result"`
	}
	if decodeOK(line, &result) != nil || result.Result == "" {
		return false
	}
	if result.Result == "success" {
		ui.addStoryKind(storyCombat, ui.tr("You got away.", "うまく逃げ切った。"))
		ui.showFight(nil)
	} else {
		ui.addStoryKind(storyCombat, ui.tr("You could not get away!", "逃げられなかった!"))
	}
	return true
}

func (catalog *worldCatalog) maxHP(npcID string, fallback int) int {
	if catalog != nil {
		if hp := catalog.NPCs[npcID].HP; hp > 0 {
			return hp
		}
	}
	return fallback
}

func (ui *gui) handleDefend(line string) bool {
	var result struct {
		Result string `json:"result"`
	}
	if decodeOK(line, &result) != nil || result.Result != "braced" {
		return false
	}
	ui.addStoryKind(storyCombat, ui.tr("You brace yourself. The next counter-attack will hurt half as much.", "身構えた。次の反撃のダメージは半分になる。"))
	return true
}
