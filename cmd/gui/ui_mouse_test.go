package main

import (
	"bufio"
	"fmt"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func findButton(t *testing.T, root fyne.CanvasObject, text string) *widget.Button {
	t.Helper()
	var search func(fyne.CanvasObject) *widget.Button
	search = func(object fyne.CanvasObject) *widget.Button {
		if button, ok := object.(*widget.Button); ok && button.Text == text {
			return button
		}
		if box, ok := object.(*fyne.Container); ok {
			for _, child := range box.Objects {
				if button := search(child); button != nil {
					return button
				}
			}
		}
		return nil
	}
	button := search(root)
	if button == nil {
		t.Fatalf("button %q not found", text)
	}
	return button
}

func mouseTestGUI(t *testing.T) (*gui, *bufio.Reader) {
	t.Helper()
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	window := application.NewWindow("mouse test")
	window.Resize(fyne.NewSize(1280, 900))
	ui := &gui{window: window, locale: "en"}
	ui.build()
	ui.nameEntry.SetText("alice")
	clientConn, serverConn := net.Pipe()
	ui.client = newProtocolClient(clientConn)
	ui.connected = true
	if err := serverConn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ui.disconnect()
		serverConn.Close()
		window.Close()
		application.Quit()
	})
	return ui, bufio.NewReader(serverConn)
}

func expectMouseCommand(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()
	line, err := reader.ReadString('\n')
	if err != nil || line != want+"\n" {
		t.Fatalf("command = %q, %v; want %q", line, err, want)
	}
}

func TestJournalActionsUseIDsAndRevealInventory(t *testing.T) {
	ui, reader := mouseTestGUI(t)
	ui.showRoom(lookView{
		Room:  roomView{ID: "loc.test", Exits: map[string]string{"east": "loc.next"}},
		Items: []string{"item.test"}, NPCs: []string{"npc.test"}, Players: []string{"alice", "bob"},
	})
	for _, action := range []struct {
		box            fyne.CanvasObject
		label, command string
	}{
		{ui.exitBox, "Go", "MOVE east"},
		{ui.itemBox, "Take", "TAKE item.test"},
		{ui.npcBox, "Talk", "TALK npc.test"},
		{ui.npcBox, "Quest", "QUEST npc.test"},
		{ui.npcBox, "Attack", "ATTACK npc.test"},
		{ui.npcBox, "Flee", "FLEE"},
		{ui.playerBox, "Invite", "GROUP INVITE bob"},
	} {
		test.Tap(findButton(t, action.box, action.label))
		expectMouseCommand(t, reader, action.command)
	}
	ui.journal.SelectIndex(1)
	expectMouseCommand(t, reader, "INVENTORY")
	ui.handleResponse("INVENTORY", "INVENTORY", `OK ["item.test"]`)
	test.Tap(findButton(t, ui.inventoryBox, "Drop"))
	expectMouseCommand(t, reader, "DROP item.test")
	ui.journal.SelectIndex(2)
	expectMouseCommand(t, reader, "QUESTS")
}

func TestMouseGroupInviteAndJoinWithoutTextEntry(t *testing.T) {
	ui, reader := mouseTestGUI(t)
	ui.showState(stateView{Players: []string{"alice", "bob"}, Invitations: []string{"carol"}})
	ui.chooseAction("GROUP")
	test.Tap(findButton(t, ui.choiceBox, "Invite a player"))
	test.Tap(findButton(t, ui.choiceBox, "bob"))
	expectMouseCommand(t, reader, "GROUP INVITE bob")
	ui.chooseAction("GROUP")
	test.Tap(findButton(t, ui.choiceBox, "Accept an invitation"))
	test.Tap(findButton(t, ui.choiceBox, "carol"))
	expectMouseCommand(t, reader, "GROUP JOIN carol")
	if ui.choicePopup != nil || len(ui.choices) != 0 {
		t.Fatal("successful selection did not close the menu")
	}
}

func TestMouseListsIncludeTargetsAfterNineAndMenuClosesOnMove(t *testing.T) {
	ui, reader := mouseTestGUI(t)
	view := lookView{Room: roomView{ID: "loc.test"}}
	for index := range 15 {
		view.Items = append(view.Items, fmt.Sprintf("item.%02d", index))
		ui.state.Players = append(ui.state.Players, fmt.Sprintf("player%02d", index))
	}
	ui.showRoom(view)
	if len(ui.itemBox.Objects) != 15 {
		t.Fatal("journal hid targets after nine")
	}
	test.Tap(findButton(t, ui.itemBox.Objects[14], "Take"))
	expectMouseCommand(t, reader, "TAKE item.14")
	ui.chooseAction("INVITE")
	test.Tap(findButton(t, ui.choiceBox, "player14"))
	expectMouseCommand(t, reader, "GROUP INVITE player14")
	ui.chooseAction("INVITE")
	ui.showRoom(lookView{Room: roomView{ID: "loc.next"}})
	if ui.choicePopup != nil || len(ui.choices) != 0 {
		t.Fatal("menu kept targets from the previous room")
	}
}

func TestCrewZeroIsDifferentFromUnknownAndResetsOnDisconnect(t *testing.T) {
	ui, _ := mouseTestGUI(t)
	ui.handleResponse("STATE", "STATE", `OK {"crew":0,"crew_initialized":true,"players":["alice"],"group":"","invitations":[]}`)
	if ui.crewLabel.Text != "Crew: 0" {
		t.Fatalf("zero crew = %q", ui.crewLabel.Text)
	}
	ui.handleResponse("STATE", "STATE", `OK {"crew":8,"crew_initialized":true,"players":[],"group":"group.1","invitations":[]}`)
	if ui.crewLabel.Text != "Crew: 8" {
		t.Fatalf("crew = %q", ui.crewLabel.Text)
	}
	ui.disconnect()
	if ui.crewLabel.Text != "Crew: -" || ui.state.Group != "" {
		t.Fatal("disconnected player state was retained")
	}
	ui.switchLocale("ja")
	ui.showState(stateView{Crew: 12, CrewInitialized: true})
	if ui.crewLabel.Text != "仲間: 12 人" {
		t.Fatalf("Japanese crew = %q", ui.crewLabel.Text)
	}
}

func TestRFCServerFallbackKeepsMouseInvitesAndChat(t *testing.T) {
	ui, reader := mouseTestGUI(t)
	ui.showRoom(lookView{Room: roomView{ID: "loc.test"}, Players: []string{"alice", "bob"}})
	ui.handleResponse("STATE", "STATE", "ERR 400 BAD_REQUEST")
	ui.handleEvent("EVT GROUP INVITE carol")
	ui.chooseAction("INVITE")
	test.Tap(findButton(t, ui.choiceBox, "bob"))
	expectMouseCommand(t, reader, "GROUP INVITE bob")
	ui.chooseAction("JOIN")
	test.Tap(findButton(t, ui.choiceBox, "carol"))
	expectMouseCommand(t, reader, "GROUP JOIN carol")
	ui.chatEntry.SetText("hello")
	ui.sendChat()
	expectMouseCommand(t, reader, "CHAT ROOM hello")
	if !ui.stateUnavailable || ui.chatEntry.Text != "" || ui.crewLabel.Text != "Crew: -" {
		t.Fatal("RFC fallback changed gameplay or showed an invented crew value")
	}
}

func TestPollRefreshesHPAndSkipsUnsupportedState(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		t.Run(fmt.Sprintf("extension_unavailable_%t", unavailable), func(t *testing.T) {
			ui, reader := mouseTestGUI(t)
			ui.stateUnavailable = unavailable
			ui.pollOnce(ui.client)
			for _, command := range []string{"LOOK", "WHO", "STATUS"} {
				expectMouseCommand(t, reader, command)
			}
			if !unavailable {
				expectMouseCommand(t, reader, "STATE")
			}
			ui.client.mu.Lock()
			pending := append([]string(nil), ui.client.pending...)
			ui.client.mu.Unlock()
			want := 4
			if unavailable {
				want = 3
			}
			if len(pending) != want {
				t.Fatalf("poll has %v queued requests, want %d", pending, want)
			}
		})
	}
}

func TestGUIResponsiveLayoutPreservesChatAndTabs(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("responsive test")
	defer window.Close()
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ui := &gui{window: window, locale: "ja", catalog: catalog}
	ui.build()
	ui.showRoom(lookView{
		Room:    roomView{ID: "loc.argo_grove", Name: "アレスの聖なる森", Description: "古い樫の木に黄金の羊毛がかかっている。神々の声を聞きながら旅を進めよう。", Exits: map[string]string{"west": "loc.argo_bull_field"}},
		Players: []string{"alice", "bob"}, Items: []string{"item.golden_fleece", "item.medeas_draught"}, NPCs: []string{"npc.medea", "npc.colchis_dragon"},
	})
	ui.showInventory([]string{"item.golden_fleece", "item.bag_of_winds"})
	ui.showQuests([]questView{
		{QuestID: "quest.golden_fleece", Status: "active", Progress: "0/1"},
		{QuestID: "quest.phineus_harpies", Status: "completed", Progress: "1/1"},
	})
	ui.showState(stateView{Crew: 8, CrewInitialized: true})
	ui.hpLabel.SetText("HP: 84/100")
	ui.statusLabel.SetText("接続中: alice")
	ui.chatEntry.SetText("まだ送っていないメッセージ")
	ui.journal.SelectIndex(1)
	window.Show()
	var wideArtWidth float32
	for _, preview := range []struct {
		name string
		size fyne.Size
		tab  int
	}{
		{"around-wide", fyne.NewSize(1280, 900), 0},
		{"inventory-wide", fyne.NewSize(1280, 900), 1},
		{"quests-wide", fyne.NewSize(1280, 900), 2},
		{"around-narrow", fyne.NewSize(640, 900), 0},
	} {
		window.Resize(preview.size)
		ui.journal.SelectIndex(preview.tab)
		capture := window.Canvas().Capture()
		if capture == nil {
			t.Fatal("empty screenshot")
		}
		if preview.size.Width >= 960 {
			if ui.detailPanel.Position().X <= ui.scenePanel.Position().X || ui.scene.Size().Width < 850 {
				t.Fatalf("wide layout has a small scene or no sidebar: scene=%v detail=%v", ui.scene.Size(), ui.detailPanel.Position())
			}
			wideArtWidth = ui.scene.Size().Width
		} else if ui.detailPanel.Position().Y <= ui.scenePanel.Position().Y || ui.scene.Size().Width >= wideArtWidth {
			t.Fatal("narrow layout did not stack panels")
		}
		for _, object := range []fyne.CanvasObject{ui.scenePanel, ui.detailPanel} {
			if object.Position().X+object.Size().Width > ui.playArea.Size().Width+1 || object.Position().Y+object.Size().Height > ui.playArea.Size().Height+1 {
				t.Fatalf("panel extends outside play area: position=%v size=%v area=%v", object.Position(), object.Size(), ui.playArea.Size())
			}
		}
		if ui.chatEntry.Text != "まだ送っていないメッセージ" || ui.journal.SelectedIndex() != preview.tab {
			t.Fatal("resize discarded chat text or selected tab")
		}
		if directory := os.Getenv("TAP_GUI_PREVIEW_DIR"); directory != "" {
			file, err := os.Create(filepath.Join(directory, preview.name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, capture)
			closeErr := file.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("save screenshot: %v, %v", err, closeErr)
			}
		}
	}
}
