package main

const japaneseLanguageOption = "日本語"

func (ui *gui) tr(english, japanese string) string {
	if ui.locale == "ja" {
		return japanese
	}
	return english
}

func (ui *gui) switchLocale(locale string) {
	if locale != "ja" {
		locale = "en"
	}
	if locale == ui.locale {
		return
	}
	address := ui.hostEntry.Text
	name := ui.nameEntry.Text
	chatMessage := ui.chatEntry.Text
	chatScope := ui.chatScope.Selected
	messageTab := ui.messages.SelectedIndex()
	journalTab := ui.journal.SelectedIndex()
	room, inventory, quests, state := ui.room, ui.inventory, ui.quests, ui.state
	ui.clearChoices()
	ui.locale = locale
	ui.build()
	ui.hostEntry.SetText(address)
	ui.nameEntry.SetText(name)
	ui.chatEntry.SetText(chatMessage)
	ui.chatScope.SetSelected(chatScope)
	ui.messages.SelectIndex(messageTab)
	ui.journal.SelectIndex(journalTab)
	ui.showRoom(room)
	ui.showInventory(inventory)
	ui.showQuests(quests)
	ui.showState(state)
	ui.showFight(ui.fight)
	if ui.client != nil {
		ui.connectButton.Disable()
		ui.settingsButton.Disable()
		ui.languageSelect.Disable()
		ui.statusLabel.SetText(ui.tr("Connecting...", "接続中..."))
	}
	ui.window.Canvas().Unfocus()
}

var directionNames = map[string]string{"north": "北", "south": "南", "east": "東", "west": "西"}

func (ui *gui) exitLabel(direction, roomID string) string {
	name := direction
	if ui.locale == "ja" {
		if japanese, ok := directionNames[direction]; ok {
			name = japanese
		}
	}
	return name + ": " + ui.catalog.label("room", roomID, ui.locale)
}

var statusWords = map[string]string{"healthy": "健康", "combat": "戦闘中", "active": "進行中", "completed": "達成"}

func (ui *gui) statusWord(status string) string {
	if ui.locale == "ja" {
		if japanese, ok := statusWords[status]; ok {
			return japanese
		}
	}
	return status
}
