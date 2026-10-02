package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *gui) commandButton(text string, action func()) *widget.Button {
	return widget.NewButton(text, func() {
		action()
		if ui.window.Canvas().Focused() != ui.chatEntry {
			ui.window.Canvas().Unfocus()
		}
	})
}

func (ui *gui) focusChat() {
	ui.messages.SelectIndex(1)
	ui.window.Canvas().Focus(ui.chatEntry)
}

func (ui *gui) chooseAction(action string) {
	if !ui.connected {
		ui.addLog(ui.tr("Connect before sending commands", "接続後に操作してください"))
		return
	}
	ui.clearChoices()
	var choices []menuChoice
	title := ui.tr("Choose a target", "対象を選ぶ")
	switch action {
	case "COMMANDS":
		// Rarely used commands live on their own page so the main screen stays uncluttered.
		title = ui.tr("Other commands", "その他の操作")
		choices = []menuChoice{
			{label: ui.tr("Look around (LOOK)", "まわりを見る (LOOK)"), command: "LOOK"},
			{label: ui.tr("My status (STATUS)", "自分の状態 (STATUS)"), command: "STATUS"},
			{label: ui.tr("My quests (QUESTS)", "クエスト一覧 (QUESTS)"), command: "QUESTS"},
			{label: ui.tr("Players online (WHO)", "接続人数 (WHO)"), command: "WHO"},
		}
	case "GROUP":
		title = ui.tr("Player group", "プレイヤーのグループ")
		choices = []menuChoice{
			{label: ui.tr("Create a group", "グループを作る"), command: "GROUP CREATE"},
			{label: ui.tr("Invite a player", "プレイヤーを招待"), action: "INVITE"},
			{label: ui.tr("Accept an invitation", "招待を受ける"), action: "JOIN"},
			{label: ui.tr("Leave the group", "グループから抜ける"), command: "GROUP LEAVE"},
		}
	case "INVITE":
		title = ui.tr("Invite a player", "招待するプレイヤー")
		players := ui.state.Players
		if ui.stateUnavailable || players == nil {
			players = ui.room.Players
		}
		for _, name := range players {
			if name != strings.TrimSpace(ui.nameEntry.Text) {
				choices = append(choices, menuChoice{label: name, command: "GROUP INVITE " + name})
			}
		}
	case "JOIN":
		title = ui.tr("Accept an invitation", "参加するグループのリーダー")
		for _, leader := range ui.state.Invitations {
			choices = append(choices, menuChoice{label: leader, command: "GROUP JOIN " + leader})
		}
	}
	ui.choices = choices
	ui.choiceTitle.SetText(title)
	var rows []fyne.CanvasObject
	for index, choice := range choices {
		rows = append(rows, ui.commandButton(choice.label, func() { ui.runChoice(index) }))
	}
	ui.setJournalRows(ui.choiceBox, rows, ui.tr("No available targets.", "選べる対象がありません。"))
	cancel := widget.NewButton(ui.tr("Back", "戻る"), ui.clearChoices)
	content := container.NewBorder(ui.choiceTitle, cancel, nil, nil, container.NewVScroll(ui.choiceBox))
	ui.choicePopup = widget.NewModalPopUp(container.NewPadded(content), ui.window.Canvas())
	ui.choicePopup.Resize(fyne.NewSize(min(480, ui.window.Canvas().Size().Width-40), min(440, ui.window.Canvas().Size().Height-60)))
	ui.choicePopup.Show()
}

func (ui *gui) runChoice(index int) {
	if index < 0 || index >= len(ui.choices) {
		return
	}
	choice := ui.choices[index]
	if choice.action != "" {
		ui.chooseAction(choice.action)
		return
	}
	if ui.send(choice.command) {
		ui.clearChoices()
	}
}

func (ui *gui) clearChoices() {
	if ui.choicePopup != nil {
		ui.choicePopup.Hide()
		ui.choicePopup = nil
		ui.window.Canvas().Unfocus()
	}
	ui.choices = nil
}
