package main

const japaneseLanguageOption = "日本語 (独自拡張)"

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
	rawCommand := ui.rawEntry.Text
	chatMessage := ui.chatEntry.Text
	chatScope := ui.chatScope.Selected
	messageTab := ui.messages.SelectedIndex()
	ui.locale = locale
	ui.build()
	ui.hostEntry.SetText(address)
	ui.nameEntry.SetText(name)
	ui.rawEntry.SetText(rawCommand)
	ui.chatEntry.SetText(chatMessage)
	ui.chatScope.SetSelected(chatScope)
	ui.messages.SelectIndex(messageTab)
	if ui.client != nil {
		ui.connectButton.Disable()
		ui.languageSelect.Disable()
		ui.statusLabel.SetText(ui.tr("Connecting...", "接続中..."))
	}
	ui.window.Canvas().Focus(ui.nameEntry)
}
