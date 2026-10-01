package main

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func journalSection(title string, content fyne.CanvasObject) fyne.CanvasObject {
	heading := canvas.NewText(title, gold)
	heading.TextSize = 14
	heading.TextStyle.Bold = true
	return textVBox(container.NewPadded(heading), content)
}

func journalCard(content fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(navy)
	background.CornerRadius = 4
	return container.NewStack(background, container.NewPadded(content))
}

func journalName(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}

func journalRow(name string, action *widget.Button) fyne.CanvasObject {
	if action == nil {
		return journalCard(journalName(name))
	}
	return journalCard(container.NewBorder(nil, nil, nil, action, journalName(name)))
}

func (ui *gui) showRoom(view lookView) {
	previous := ui.room
	if previous.Room.ID != view.Room.ID {
		ui.clearChoices()
	}
	ui.room = view
	if view.Room.ID != "" {
		if ui.visited == nil {
			ui.visited = map[string]bool{}
		}
		ui.visited[view.Room.ID] = true
	}
	defer ui.refreshMap()
	if ui.scene.Image == nil || previous.Room.ID != view.Room.ID || !slices.Equal(previous.NPCs, view.NPCs) {
		ui.scene.Resource = nil
		ui.scene.Image = composeScene(view.Room.ID, view.NPCs)
		ui.scene.Refresh()
	}
	if previous.Room.ID != view.Room.ID || !slices.Equal(previous.Items, view.Items) {
		ui.showItemPhotos(view.Items)
	}
	if view.Room.ID == "" {
		ui.roomTitle.SetText(ui.tr("Your journey", "冒険の旅"))
		ui.roomDesc.SetText(ui.tr("Connect to begin your journey.", "接続して冒険を始めましょう。"))
	} else {
		ui.roomTitle.SetText(view.Room.Name)
		ui.roomDesc.SetText(view.Room.Description)
		ui.roomCount.SetText(fmt.Sprintf(ui.tr("Here: %d", "部屋: %d 人"), len(view.Players)))
	}
	if view.Room.ID != "" && reflect.DeepEqual(previous, view) {
		return
	}
	directions := make([]string, 0, len(view.Room.Exits))
	for direction := range view.Room.Exits {
		directions = append(directions, direction)
	}
	sort.Strings(directions)
	var exits, players, items, npcs []fyne.CanvasObject
	for _, direction := range directions {
		exits = append(exits, journalRow(ui.exitLabel(direction, view.Room.Exits[direction]),
			ui.commandButton(ui.tr("Go", "進む"), func() { ui.send("MOVE " + direction) })))
	}
	for _, id := range view.NPCs {
		buttons := []fyne.CanvasObject{ui.commandButton(ui.tr("Talk", "話す"), func() { ui.send("TALK " + id) })}
		known := ui.catalog != nil && ui.catalog.NPCs[id].Role != ""
		if !known || ui.catalog.hasQuest(id) {
			buttons = append(buttons, ui.commandButton(ui.tr("Quest", "依頼"), func() { ui.send("QUEST " + id) }))
		}
		actions := container.NewGridWithColumns(len(buttons), buttons...)
		content := textVBox(journalName(ui.catalog.label("npc", id, ui.locale)), actions)
		if !known || ui.catalog.NPCs[id].Role == "enemy" {
			content.Add(container.NewGridWithColumns(2,
				ui.commandButton(ui.tr("Attack", "戦う"), func() { ui.send("ATTACK " + id) }),
				ui.commandButton(ui.tr("Flee", "逃げる"), func() { ui.send("FLEE") }),
			))
		}
		npcs = append(npcs, journalCard(content))
	}
	for _, id := range view.Items {
		items = append(items, ui.itemRow(id, "TAKE", ui.tr("Take", "取る")))
	}
	for _, name := range view.Players {
		var invite *widget.Button
		if name != strings.TrimSpace(ui.nameEntry.Text) {
			invite = ui.commandButton(ui.tr("Invite", "招待"), func() { ui.send("GROUP INVITE " + name) })
		}
		players = append(players, journalRow(name, invite))
	}
	ui.setJournalRows(ui.exitBox, exits, ui.tr("No paths from here.", "移動先がありません。"))
	ui.setJournalRows(ui.npcBox, npcs, ui.tr("No one to talk to here.", "話しかける相手はいません。"))
	ui.setJournalRows(ui.itemBox, items, ui.tr("No items here.", "道具は落ちていません。"))
	ui.setJournalRows(ui.playerBox, players, ui.tr("No players here.", "プレイヤーはいません。"))
}

func (catalog *worldCatalog) hasQuest(npcID string) bool {
	if catalog == nil {
		return false
	}
	for _, quest := range catalog.Quests {
		if quest.GiverNPCID == npcID {
			return true
		}
	}
	return false
}

func (ui *gui) itemRow(id, command, text string) fyne.CanvasObject {
	var picture fyne.CanvasObject
	if data, err := itemPhotoAssets.ReadFile("assets/items/" + id + ".png"); err == nil {
		image := canvas.NewImageFromResource(fyne.NewStaticResource(id+".png", data))
		image.FillMode = canvas.ImageFillContain
		picture = container.NewGridWrap(fyne.NewSize(44, 44), image)
	}
	button := ui.commandButton(text, func() { ui.send(command + " " + id) })
	return journalCard(container.NewBorder(nil, nil, picture, button, journalName(ui.catalog.label("item", id, ui.locale))))
}

func (ui *gui) showInventory(ids []string) {
	ui.inventory = append([]string(nil), ids...)
	var rows []fyne.CanvasObject
	if len(ids) > 0 {
		rows = append(rows, widget.NewLabel(fmt.Sprintf(ui.tr("%d items carried", "持ち物 %d 個"), len(ids))))
	}
	for _, id := range ids {
		rows = append(rows, ui.itemRow(id, "DROP", ui.tr("Drop", "置く")))
	}
	ui.setJournalRows(ui.inventoryBox, rows, ui.tr("Your bag is empty. Pick up items in Around.", "持ち物はありません。「まわり」から道具を拾えます。"))
	ui.showEndings()
}

func (ui *gui) showQuests(quests []questView) {
	ui.quests = append([]questView(nil), quests...)
	var rows []fyne.CanvasObject
	for _, quest := range quests {
		name := journalName(ui.catalog.label("quest", quest.QuestID, ui.locale))
		status := widget.NewLabel(ui.statusWord(quest.Status) + "  ·  " + quest.Progress)
		content := textVBox(name, status)
		if ui.catalog != nil {
			if description := ui.catalog.Quests[quest.QuestID].Description.get(ui.locale); description != "" {
				label := widget.NewLabel(description)
				label.Wrapping = fyne.TextWrapWord
				content.Add(label)
			}
		}
		current, goal, _ := strings.Cut(quest.Progress, "/")
		progress, progressErr := strconv.Atoi(current)
		target, targetErr := strconv.Atoi(goal)
		if progressErr == nil && targetErr == nil && target > 0 {
			bar := widget.NewProgressBar()
			bar.SetValue(float64(progress) / float64(target))
			content.Add(bar)
		}
		rows = append(rows, journalCard(content))
	}
	ui.setJournalRows(ui.questBox, rows, ui.tr("No quests yet. Ask people in Around for a quest.", "受けているクエストはありません。「まわり」の人物から依頼を受けられます。"))
}

func (ui *gui) setJournalRows(box *fyne.Container, rows []fyne.CanvasObject, empty string) {
	if len(rows) == 0 {
		label := widget.NewLabel(empty)
		label.Wrapping = fyne.TextWrapWord
		rows = []fyne.CanvasObject{container.NewPadded(label)}
	}
	box.Objects = rows
	box.Refresh()
}

func (ui *gui) showItemPhotos(ids []string) {
	photos := make([]fyne.CanvasObject, 0, len(ids))
	for _, id := range ids {
		if card := itemPhotoCard(id); card != nil {
			photos = append(photos, card)
		}
	}
	ui.itemPhotoBox.Objects = photos
	ui.itemPhotoBox.Refresh()
	if len(photos) == 0 {
		ui.itemPhotoScroll.Hide()
		ui.photoStrip.Hide()
	} else {
		ui.itemPhotoScroll.Show()
		ui.photoStrip.Show()
	}
	ui.itemPhotoScroll.ScrollToOffset(fyne.Position{})
}

func (ui *gui) scrollItemPhotos(direction float32) {
	if len(ui.itemPhotoBox.Objects) == 0 {
		return
	}
	step := ui.itemPhotoBox.Objects[0].MinSize().Width + theme.Padding()
	offset := ui.itemPhotoScroll.Offset
	offset.X = max(0, offset.X+direction*step)
	ui.itemPhotoScroll.ScrollToOffset(offset)
}

// startingCrew is the crew every player gets when entering the Odyssey (server side: 12).
const startingCrew = 12

func (ui *gui) showState(state stateView) {
	ui.state = state
	if state.CrewInitialized {
		ui.crewLabel.SetText(fmt.Sprintf(ui.tr("Crew: %d", "仲間: %d 人"), state.Crew))
		ui.crewBar.Set(state.Crew, max(startingCrew, state.Crew))
	} else {
		ui.crewLabel.SetText(ui.tr("Crew: -", "仲間: - 人"))
		ui.crewBar.Set(0, startingCrew)
	}
	group := "-"
	if state.Group != "" {
		group = state.Group
	}
	ui.groupLabel.SetText(ui.tr("Group: ", "グループ: ") + group)
}
