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
	ui.quitButton.Importance = widget.LowImportance
	ui.connectButton.Importance = widget.HighImportance
	ui.quitButton.Disable()
	ui.statusLabel = dimLabel(ui.tr("Not connected", "未接続"))
	ui.roomCount = dimLabel(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount = dimLabel(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel = boldLabel("HP: -")
	ui.crewLabel = boldLabel(ui.tr("Crew: -", "仲間: - 人"))
	ui.hpBar, ui.crewBar = newStatBar(), newStatBar()
	ui.groupLabel = dimLabel(ui.tr("Group: -", "グループ: -"))

	ui.roomTitle = widget.NewLabelWithStyle(ui.tr("Your journey", "冒険の旅"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.roomTitle.Truncation = fyne.TextTruncateEllipsis
	ui.roomDesc = widget.NewLabel(ui.tr("Connect to begin your journey.", "接続して冒険を始めましょう。"))
	ui.roomDesc.Wrapping = fyne.TextWrapWord
	ui.exitBox = textVBox()
	ui.playerBox = textVBox()
	ui.itemBox = textVBox()
	ui.npcBox = textVBox()
	ui.inventoryBox = textVBox()
	ui.questBox = textVBox()
	ui.endingBox = newEndingBox()
	ui.choiceTitle = widget.NewLabel("")
	ui.choiceBox = textVBox()
	ui.itemPhotoBox = container.NewHBox()
	ui.itemPhotoScroll = container.NewHScroll(ui.itemPhotoBox)
	ui.itemPhotoScroll.Hide()
	ui.scene = sceneImage()
	photos := container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { ui.scrollItemPhotos(-1) }),
		widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { ui.scrollItemPhotos(1) }),
		ui.itemPhotoScroll,
	)
	ui.photoStrip = photos
	ui.photoStrip.Hide()
	ui.flash = newFlashLayer()
	ui.sceneMapBox = container.New(&mapLayout{})
	ui.sceneMapPanel = framed(ui.tr("Map", "地図"), ui.mapView(ui.sceneMapBox))
	ui.sceneMapPanel.Hide()
	visualLayout := &sceneVisualLayout{photos: ui.itemPhotoBox}
	sceneVisual := container.NewScroll(container.New(visualLayout, ui.scene, ui.flash, ui.photoStrip, ui.sceneMapPanel))
	sceneVisual.Direction = container.ScrollNone
	sceneFrame := canvas.NewRectangle(ink)
	sceneFrame.StrokeColor = gold
	sceneFrame.StrokeWidth = 1
	visual := container.NewStack(sceneFrame, sceneVisual)
	ui.scenePanel = container.NewBorder(ui.roomTitle, nil, nil, nil,
		container.New(&sceneLayout{}, visual, container.NewVScroll(ui.roomDesc)))

	surroundings := textVBox(
		journalSection(ui.tr("Paths", "移動先"), ui.exitBox),
		journalSection(ui.tr("People & creatures", "人物・生きもの"), ui.npcBox),
		journalSection(ui.tr("Items here", "落ちている道具"), ui.itemBox),
		journalSection(ui.tr("Players here", "この部屋のプレイヤー"), ui.playerBox),
	)
	mapTab := container.NewTabItem(ui.tr("Map", "地図"), ui.buildMapTab())
	ui.journal = container.NewAppTabs(
		container.NewTabItem(ui.tr("Around", "まわり"), container.NewVScroll(surroundings)),
		container.NewTabItem(ui.tr("Inventory", "持ち物"), container.NewVScroll(ui.inventoryBox)),
		container.NewTabItem(ui.tr("Quests", "クエスト"), container.NewVScroll(textVBox(ui.questBox, ui.endingBox))),
		mapTab,
	)
	journal := ui.journal
	visualLayout.onMapVisibility = func(visible bool) {
		if visible && len(journal.Items) == 4 {
			if journal.Selected() == mapTab {
				journal.SelectIndex(0)
			}
			journal.Remove(mapTab)
		} else if !visible && len(journal.Items) == 3 {
			journal.Append(mapTab)
		}
	}
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
	ui.commandButtons = container.NewGridWithColumns(3,
		ui.commandButton(ui.tr("Refresh", "更新"), func() { ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO", "STATE") }),
		ui.commandButton(ui.tr("Group", "グループ"), func() { ui.chooseAction("GROUP") }),
		ui.commandButton(ui.tr("Chat", "チャット"), ui.focusChat),
	)
	ui.buildCombatPanel()
	ui.detailPanel = container.NewBorder(ui.combatPanel, ui.commandButtons, nil, nil, ui.journal)
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
	header := container.New(&headerLayout{}, brand,
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
	ui.showEndings()
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

// dimLabel is for secondary info (counts, status): same size, quieter colour.
func dimLabel(text string) *widget.Label {
	label := compactLabel(text)
	label.Importance = widget.LowImportance
	return label
}

// boldLabel is for the numbers the player watches most (HP, crew).
func boldLabel(text string) *widget.Label {
	label := compactLabel(text)
	label.TextStyle.Bold = true
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

func (*screenLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(3) * theme.Padding()
	for _, object := range objects {
		height += object.MinSize().Height
	}
	return fyne.NewSize(560, max(680, height))
}

func (*screenLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	headerHeight := heightForWidth(objects[0], size.Width)
	statsHeight := heightForWidth(objects[1], size.Width)
	messagesHeight := min(float32(230), size.Height*0.26)
	if size.Width < 960 {
		messagesHeight = min(float32(170), size.Height*0.22)
	}
	messagesHeight = max(objects[3].MinSize().Height, min(messagesHeight, size.Height-headerHeight-statsHeight-objects[2].MinSize().Height-3*gap))
	playHeight := max(float32(0), size.Height-headerHeight-statsHeight-messagesHeight-3*gap)
	y := float32(0)
	for index, height := range []float32{headerHeight, statsHeight, playHeight, messagesHeight} {
		objects[index].Move(fyne.NewPos(0, y))
		objects[index].Resize(fyne.NewSize(size.Width, height))
		y += height + gap
	}
}

func heightForWidth(object fyne.CanvasObject, width float32) float32 {
	object.Resize(fyne.NewSize(width, object.MinSize().Height))
	return object.MinSize().Height
}

type headerLayout struct{ stacked bool }

func (l *headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	brand, controls := objects[0].MinSize(), objects[1].MinSize()
	if l.stacked {
		return fyne.NewSize(max(brand.Width, controls.Width), brand.Height+controls.Height+theme.Padding())
	}
	return fyne.NewSize(brand.Width+controls.Width+theme.Padding(), max(brand.Height, controls.Height))
}

func (l *headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	brand, controls := objects[0].MinSize(), objects[1].MinSize()
	l.stacked = brand.Width+controls.Width+theme.Padding() > size.Width
	objects[0].Resize(brand)
	objects[1].Resize(controls)
	if l.stacked {
		objects[0].Move(fyne.NewPos(0, 0))
		objects[1].Move(fyne.NewPos(max(0, size.Width-controls.Width), brand.Height+theme.Padding()))
		return
	}
	height := max(brand.Height, controls.Height)
	objects[0].Move(fyne.NewPos(0, (height-brand.Height)/2))
	objects[1].Move(fyne.NewPos(size.Width-controls.Width, (height-controls.Height)/2))
}

type statsLayout struct{ columns int }

func (l *statsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return layout.NewGridLayoutWithColumns(max(3, l.columns)).MinSize(objects)
}

func (l *statsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.columns = 6
	if size.Width < 1000 {
		l.columns = 3
	}
	layout.NewGridLayoutWithColumns(l.columns).Layout(objects, size)
}

func textVBox(objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&textVBoxLayout{}, objects...)
}

type textVBoxLayout struct{}

func (*textVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return layout.NewVBoxLayout().MinSize(objects)
}

func (*textVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, object := range objects {
		if !object.Visible() {
			continue
		}
		height := heightForWidth(object, size.Width)
		object.Move(fyne.NewPos(0, y))
		object.Resize(fyne.NewSize(size.Width, height))
		y += height + theme.Padding()
	}
}

type adventureLayout struct{ wide bool }

func (l *adventureLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := max(objects[0].MinSize().Height, objects[1].MinSize().Height)
	if !l.wide {
		height = objects[0].MinSize().Height + objects[1].MinSize().Height + theme.Padding()
	}
	return fyne.NewSize(0, max(400, height))
}

func (l *adventureLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	l.wide = size.Width >= 960
	if l.wide {
		sideWidth := max(float32(330), objects[1].MinSize().Width)
		objects[0].Move(fyne.NewPos(0, 0))
		objects[0].Resize(fyne.NewSize(size.Width-sideWidth-gap, size.Height))
		objects[1].Move(fyne.NewPos(size.Width-sideWidth, 0))
		objects[1].Resize(fyne.NewSize(sideWidth, size.Height))
		return
	}
	sceneHeight := min(size.Width*0.6+80, max(100, size.Height-190))
	sceneHeight = min(max(objects[0].MinSize().Height, sceneHeight), max(0, size.Height-objects[1].MinSize().Height-gap))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, sceneHeight))
	objects[1].Move(fyne.NewPos(0, sceneHeight+gap))
	objects[1].Resize(fyne.NewSize(size.Width, max(0, size.Height-sceneHeight-gap)))
}

type sceneLayout struct{}

func (*sceneLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	art, description := objects[0].MinSize(), objects[1].MinSize()
	return fyne.NewSize(max(art.Width, description.Width), art.Height+max(description.Height, textLineHeight())+theme.Padding())
}

func (*sceneLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	descriptionHeight := min(float32(62), max(textLineHeight(), size.Height*0.2))
	descriptionHeight = min(descriptionHeight, max(0, size.Height-objects[0].MinSize().Height-theme.Padding()))
	artHeight := max(float32(0), size.Height-descriptionHeight-theme.Padding())
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, artHeight))
	objects[1].Move(fyne.NewPos(0, artHeight+theme.Padding()))
	objects[1].Resize(fyne.NewSize(size.Width, descriptionHeight))
}

func textLineHeight() float32 {
	textTheme := theme.Current()
	style := fyne.TextStyle{}
	size, _ := fyne.CurrentApp().Driver().RenderedTextSize("Ag国", textTheme.Size(theme.SizeNameText), style, textTheme.Font(style))
	return size.Height + 2*textTheme.Size(theme.SizeNameInnerPadding)
}

type sceneVisualLayout struct {
	photos          *fyne.Container
	onMapVisibility func(bool)
}

func (*sceneVisualLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(100, 100) }

func (l *sceneVisualLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	inside := fyne.NewSize(max(0, size.Width-2*gap), max(0, size.Height-2*gap))
	for _, layer := range objects[:2] {
		layer.Move(fyne.NewPos(gap, gap))
		layer.Resize(inside)
	}
	mapPanel := objects[3]
	renderedWidth := min(inside.Width, inside.Height*float32(artWidth)/float32(artHeight))
	leftSpace := (inside.Width - renderedWidth) / 2
	mapSize := mapPanel.MinSize()
	if leftSpace >= mapSize.Width+2*gap && inside.Height >= mapSize.Height {
		mapPanel.Show()
		mapPanel.Move(fyne.NewPos(gap+(leftSpace-mapSize.Width)/2, gap+(inside.Height-mapSize.Height)/2))
		mapPanel.Resize(mapSize)
	} else {
		mapPanel.Hide()
	}
	if l.onMapVisibility != nil {
		l.onMapVisibility(mapPanel.Visible())
	}
	if objects[2].Visible() {
		thumbnail := fyne.NewSquareSize(min(104, max(48, inside.Height*0.3)))
		for _, photo := range l.photos.Objects {
			card := photo.(*fyne.Container)
			if card.MinSize() != thumbnail {
				card.Layout = layout.NewGridWrapLayout(thumbnail)
				card.Resize(thumbnail)
			}
		}
		strip := fyne.NewSize(min(290, inside.Width), objects[2].MinSize().Height)
		strip.Height = min(strip.Height, inside.Height)
		objects[2].Move(fyne.NewPos(size.Width-gap-strip.Width, size.Height-gap-strip.Height))
		objects[2].Resize(strip)
	}
}
