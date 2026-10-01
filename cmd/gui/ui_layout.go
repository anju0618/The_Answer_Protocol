package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *gui) build() {
	ui.hostEntry = widget.NewEntry()
	ui.hostEntry.SetText("127.0.0.1:4242")
	ui.hostEntry.SetPlaceHolder("host:port")
	ui.nameEntry = widget.NewEntry()
	ui.nameEntry.SetPlaceHolder(ui.tr("player name", "プレイヤー名"))
	ui.languageSelect = widget.NewSelect([]string{"English", japaneseLanguageOption}, nil)
	ui.languageSelect.SetSelected(ui.tr("English", japaneseLanguageOption))
	ui.connectButton = widget.NewButton(ui.tr("Connect", "接続"), func() {
		if strings.TrimSpace(ui.nameEntry.Text) == "" {
			ui.showConnectionSettings()
		} else {
			ui.connect()
		}
	})
	ui.settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), ui.showConnectionSettings)
	ui.quitButton = widget.NewButton(ui.tr("Disconnect", "切断"), func() { ui.send("QUIT") })
	ui.quitButton.Disable()
	ui.statusLabel = compactLabel(ui.tr("Not connected", "未接続"))
	ui.roomCount = compactLabel(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount = compactLabel(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel = compactLabel("HP: -")
	ui.crewLabel = compactLabel(ui.tr("Crew: -", "仲間: - 人"))
	ui.hpBar, ui.crewBar = newStatBar(), newStatBar()
	ui.groupLabel = compactLabel(ui.tr("Group: -", "グループ: -"))

	ui.roomTitle = widget.NewLabelWithStyle(ui.tr("Your journey", "冒険の旅"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.roomTitle.Truncation = fyne.TextTruncateEllipsis
	ui.roomDesc = widget.NewLabel(ui.tr("Connect to begin your journey.", "接続して冒険を始めましょう。"))
	ui.roomDesc.Wrapping = fyne.TextWrapWord
	ui.exitBox = container.NewVBox()
	ui.playerBox = container.NewVBox()
	ui.itemBox = container.NewVBox()
	ui.npcBox = container.NewVBox()
	ui.inventoryBox = container.NewVBox()
	ui.questBox = container.NewVBox()
	ui.choiceTitle = widget.NewLabel("")
	ui.choiceBox = container.NewVBox()
	ui.itemPhotoBox = container.NewHBox()
	ui.itemPhotoScroll = container.NewHScroll(ui.itemPhotoBox)
	ui.itemPhotoScroll.Hide()
	ui.scene = sceneImage()
	photos := container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { ui.scrollItemPhotos(-110) }),
		widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { ui.scrollItemPhotos(110) }),
		ui.itemPhotoScroll,
	)
	ui.photoStrip = container.NewGridWrap(fyne.NewSize(290, 112), photos)
	ui.photoStrip.Hide()
	photoOverlay := container.NewHBox(layout.NewSpacer(), ui.photoStrip)
	sceneVisual := container.NewStack(ui.scene, container.NewBorder(nil, photoOverlay, nil, nil))
	sceneFrame := canvas.NewRectangle(ink)
	sceneFrame.StrokeColor = gold
	sceneFrame.StrokeWidth = 1
	visual := container.NewStack(sceneFrame, sceneVisual)
	ui.scenePanel = container.NewBorder(ui.roomTitle, nil, nil, nil,
		container.New(&sceneLayout{}, visual, container.NewVScroll(ui.roomDesc)))

	surroundings := container.NewVBox(
		journalSection(ui.tr("Paths", "移動先"), ui.exitBox),
		journalSection(ui.tr("People & creatures", "人物・生きもの"), ui.npcBox),
		journalSection(ui.tr("Items here", "落ちている道具"), ui.itemBox),
		journalSection(ui.tr("Players here", "この部屋のプレイヤー"), ui.playerBox),
	)
	ui.journal = container.NewAppTabs(
		container.NewTabItem(ui.tr("Around", "まわり"), container.NewVScroll(surroundings)),
		container.NewTabItem(ui.tr("Inventory", "持ち物"), container.NewVScroll(ui.inventoryBox)),
		container.NewTabItem(ui.tr("Quests", "クエスト"), container.NewVScroll(ui.questBox)),
	)
	ui.journal.OnSelected = func(item *container.TabItem) {
		if !ui.connected {
			return
		}
		switch item {
		case ui.journal.Items[1]:
			ui.send("INVENTORY")
		case ui.journal.Items[2]:
			ui.send("QUESTS")
		}
	}
	ui.commandButtons = container.NewGridWithColumns(4,
		ui.commandButton(ui.tr("Refresh", "更新"), func() { ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO", "STATE") }),
		ui.commandButton(ui.tr("Group", "グループ"), func() { ui.chooseAction("GROUP") }),
		ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") }),
		ui.commandButton(ui.tr("Chat", "チャット"), ui.focusChat),
	)
	ui.detailPanel = container.NewBorder(nil, ui.commandButtons, nil, nil, ui.journal)
	ui.playArea = container.New(&adventureLayout{}, ui.scenePanel, ui.detailPanel)

	ui.chatScope = widget.NewSelect([]string{"GLOBAL", "ROOM", "GROUP"}, nil)
	ui.chatScope.SetSelected("ROOM")
	ui.chatEntry = widget.NewEntry()
	ui.chatEntry.SetPlaceHolder(ui.tr("Write a message…", "メッセージを入力…"))
	ui.chatEntry.OnSubmitted = func(string) { ui.sendChat() }
	ui.chatLabel = widget.NewLabel("")
	ui.chatLabel.Wrapping = fyne.TextWrapWord
	ui.chatScroll = container.NewVScroll(ui.chatLabel)
	chatInput := container.NewBorder(nil, nil, ui.chatScope, widget.NewButton(ui.tr("Send", "送信"), ui.sendChat), ui.chatEntry)
	chatPane := container.NewBorder(nil, chatInput, nil, nil, ui.chatScroll)
	ui.logLabel = widget.NewLabel("")
	ui.logLabel.Wrapping = fyne.TextWrapWord
	ui.logScroll = container.NewVScroll(ui.logLabel)
	ui.storyText = newStoryText()
	ui.storyScroll = container.NewVScroll(ui.storyText)
	ui.messages = container.NewAppTabs(
		container.NewTabItem(ui.tr("Adventure", "ぼうけん"), ui.storyScroll),
		container.NewTabItem(ui.tr("Chat", "チャット"), chatPane),
		container.NewTabItem(ui.tr("Log", "ログ"), ui.logScroll),
	)
	brand := canvas.NewText("THE ANSWER PROTOCOL", gold)
	brand.TextSize = 17
	brand.TextStyle.Bold = true
	header := container.NewBorder(nil, nil, brand,
		container.NewHBox(ui.languageSelect, ui.settingsButton, ui.connectButton, ui.quitButton))
	stats := container.New(&statsLayout{}, container.NewStack(ui.hpBar.box, ui.hpLabel), container.NewStack(ui.crewBar.box, ui.crewLabel), ui.groupLabel, ui.roomCount, ui.totalCount, ui.statusLabel)
	ui.window.SetContent(container.NewPadded(container.New(&screenLayout{}, header, stats, ui.playArea, ui.messages)))
	ui.languageSelect.OnChanged = func(selection string) {
		if ui.client != nil || ui.connected || ui.dialing {
			return
		}
		if selection == japaneseLanguageOption {
			ui.switchLocale("ja")
		} else {
			ui.switchLocale("en")
		}
	}
	ui.showRoom(lookView{})
	ui.showInventory(nil)
	ui.showQuests(nil)
	for _, saved := range []struct {
		lines []string
		label *widget.Label
	}{
		{ui.chatLines, ui.chatLabel}, {ui.logLines, ui.logLabel},
	} {
		if len(saved.lines) > 0 {
			saved.label.SetText(strings.Join(saved.lines, "\n"))
		}
	}
	if len(ui.storyLines) > 0 {
		ui.storyText.Segments = storySegments(ui.storyLines)
		ui.storyText.Refresh()
	} else {
		ui.addStory(ui.tr("The gods of Greece await you.", "ギリシアの神々があなたを待っている。"))
	}
}

func compactLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}

func (ui *gui) showConnectionSettings() {
	if ui.client != nil || ui.dialing {
		return
	}
	var popup *widget.PopUp
	closePopup := func() { popup.Hide(); ui.window.Canvas().Unfocus() }
	connect := widget.NewButton(ui.tr("Connect", "接続"), func() { closePopup(); ui.connect() })
	ui.nameEntry.OnSubmitted = func(string) { closePopup(); ui.connect() }
	ui.hostEntry.OnSubmitted = func(string) { ui.window.Canvas().Focus(ui.nameEntry) }
	content := container.NewVBox(
		widget.NewLabelWithStyle(ui.tr("Connection settings", "接続設定"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(ui.tr("Server", "接続先")), ui.hostEntry,
		widget.NewLabel(ui.tr("Player name", "名前")), ui.nameEntry,
		container.NewGridWithColumns(2, widget.NewButton(ui.tr("Cancel", "戻る"), closePopup), connect),
	)
	popup = widget.NewModalPopUp(container.NewPadded(content), ui.window.Canvas())
	popup.Resize(fyne.NewSize(min(440, ui.window.Canvas().Size().Width-30), content.MinSize().Height+12))
	popup.Show()
	ui.window.Canvas().Focus(ui.nameEntry)
}

type screenLayout struct{}

func (*screenLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(560, 680) }

func (*screenLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	headerHeight := objects[0].MinSize().Height
	statsHeight := float32(30)
	if size.Width < 1000 {
		statsHeight = 60 + gap
	}
	messagesHeight := min(float32(230), size.Height*0.26)
	if size.Width < 960 {
		messagesHeight = min(float32(170), size.Height*0.22)
	}
	playHeight := max(float32(0), size.Height-headerHeight-statsHeight-messagesHeight-3*gap)
	y := float32(0)
	for index, height := range []float32{headerHeight, statsHeight, playHeight, messagesHeight} {
		objects[index].Move(fyne.NewPos(0, y))
		objects[index].Resize(fyne.NewSize(size.Width, height))
		y += height + gap
	}
}

type statsLayout struct{}

func (*statsLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 30) }

func (*statsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	columns := 6
	if size.Width < 1000 {
		columns = 3
	}
	layout.NewGridLayoutWithColumns(columns).Layout(objects, size)
}

type adventureLayout struct{}

func (*adventureLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 400) }

func (*adventureLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	if size.Width >= 960 {
		sideWidth := float32(330)
		objects[0].Move(fyne.NewPos(0, 0))
		objects[0].Resize(fyne.NewSize(size.Width-sideWidth-gap, size.Height))
		objects[1].Move(fyne.NewPos(size.Width-sideWidth, 0))
		objects[1].Resize(fyne.NewSize(sideWidth, size.Height))
		return
	}
	sceneHeight := min(size.Width*0.6+80, max(100, size.Height-190))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, sceneHeight))
	objects[1].Move(fyne.NewPos(0, sceneHeight+gap))
	objects[1].Resize(fyne.NewSize(size.Width, max(0, size.Height-sceneHeight-gap)))
}

type sceneLayout struct{}

func (*sceneLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 100) }

func (*sceneLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	descriptionHeight := min(float32(62), size.Height*0.2)
	artHeight := max(float32(0), size.Height-descriptionHeight-theme.Padding())
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, artHeight))
	objects[1].Move(fyne.NewPos(0, artHeight+theme.Padding()))
	objects[1].Resize(fyne.NewSize(size.Width, descriptionHeight))
}
