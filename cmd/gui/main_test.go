package main

import (
	"bufio"
	"crypto/sha256"
	"image"
	"image/png"
	"net"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestGUILanguageSelectionUpdatesControls(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()

	checkInventoryButton := func(want string) {
		t.Helper()
		button := ui.commandButtons.Objects[7].(*widget.Button)
		if button.Text != want {
			t.Errorf("inventory button = %q, want %q", button.Text, want)
		}
	}
	checkInventoryButton("I Inventory")
	if ui.connectButton.Text != "Connect" || ui.messages.Items[0].Text != "Adventure" {
		t.Fatal("English controls were not shown by default")
	}

	ui.hostEntry.SetText("example.org:4242")
	ui.nameEntry.SetText("alice")
	ui.rawEntry.SetText("LOOK")
	ui.chatEntry.SetText("hello")
	ui.messages.SelectIndex(1)
	ui.languageSelect.SetSelected(japaneseLanguageOption)
	checkInventoryButton("I 持ち物")
	if ui.connectButton.Text != "接続" || ui.messages.Items[0].Text != "ぼうけん" {
		t.Fatal("Japanese controls were not shown after selection")
	}
	if ui.hostEntry.Text != "example.org:4242" || ui.nameEntry.Text != "alice" {
		t.Fatal("switching languages discarded connection inputs")
	}
	if ui.rawEntry.Text != "LOOK" || ui.chatEntry.Text != "hello" || ui.messages.SelectedIndex() != 1 {
		t.Fatal("switching languages discarded unsent text or the selected tab")
	}

	ui.languageSelect.SetSelected("English")
	checkInventoryButton("I Inventory")
	if ui.connectButton.Text != "Connect" || ui.messages.Items[0].Text != "Adventure" {
		t.Fatal("English controls were not restored")
	}
	ui.setTotal("2")
	if ui.totalCount.Text != "Online players: 2" {
		t.Errorf("online count = %q", ui.totalCount.Text)
	}
	ui.connected = true
	ui.chooseAction("TAKE")
	if ui.choiceTitle.Text != "TAKE: no targets" {
		t.Errorf("action prompt = %q", ui.choiceTitle.Text)
	}
	ui.connected = false
	ui.languageSelect.SetSelected(japaneseLanguageOption)
	ui.handleResponse("LANG", "LANG ja", "ERR 400 BAD_REQUEST")
	if ui.locale != "en" || ui.languageSelect.Selected != "English" {
		t.Fatal("rejected Japanese extension did not restore English controls")
	}
}

func TestKeyboardTakeSendsItemID(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en", catalog: &worldCatalog{
		Items: map[string]catalogEntry{"item.sword": {Name: localizedName{"en": "Bronze Sword"}}},
	}}
	ui.build()
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	ui.client = newProtocolClient(clientConn)
	defer ui.client.Close()
	ui.connected = true
	ui.showRoom(lookView{Room: roomView{ID: "room.hall", Name: "Hall"}, Items: []string{"item.sword"}})
	window.Canvas().Unfocus()
	window.Canvas().OnTypedRune()('t')
	window.Canvas().OnTypedRune()('1')
	if err := serverConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewScanner(serverConn)
	if !reader.Scan() {
		t.Fatalf("no command received: %v", reader.Err())
	}
	if got := reader.Text(); got != "TAKE item.sword" {
		t.Fatalf("command = %q", got)
	}
	window.Canvas().OnTypedRune()('c')
	if ui.messages.SelectedIndex() != 1 || window.Canvas().Focused() != ui.chatEntry {
		t.Fatal("chat shortcut did not reveal and focus chat input")
	}
}

func TestGUIRenders(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	window.Resize(fyne.NewSize(1180, 840))
	ui := &gui{window: window, locale: "en"}
	ui.build()
	ui.showRoom(lookView{
		Room: roomView{
			ID:          "loc.argo_grove",
			Name:        "Grove of Ares",
			Description: "The Golden Fleece hangs from an ancient oak.",
			Exits:       map[string]string{"west": "loc.argo_bull_field"},
		},
		Players: []string{"alice"},
		Items:   []string{"item.golden_fleece", "item.medeas_draught"},
		NPCs:    []string{"npc.medea", "npc.colchis_dragon"},
	})
	ui.statusLabel.SetText("Playing as: alice")
	ui.hpLabel.SetText("HP: 100/100")
	ui.totalCount.SetText("Online players: 1")
	window.Show()
	image := window.Canvas().Capture()
	if image == nil || image.Bounds().Dx() < 500 || image.Bounds().Dy() < 300 {
		t.Fatalf("invalid GUI image: %v", image)
	}
	if path := os.Getenv("TAP_GUI_PREVIEW"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err := png.Encode(file, image); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGUIRoomArtUpdates(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, NPCs: []string{"npc.medea"}})
	first, ok := ui.scene.Image.(*image.RGBA)
	if !ok {
		t.Fatal("LOOK did not install room artwork")
	}
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, NPCs: []string{"npc.colchis_dragon"}})
	second := ui.scene.Image.(*image.RGBA)
	if sha256.Sum256(first.Pix) == sha256.Sum256(second.Pix) {
		t.Fatal("NPC change did not update artwork")
	}
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_bull_field"}, NPCs: []string{"npc.khalkotauroi"}})
	third := ui.scene.Image.(*image.RGBA)
	if sha256.Sum256(second.Pix) == sha256.Sum256(third.Pix) {
		t.Fatal("room change did not update artwork")
	}
}

func TestMoveRefreshesRoomArt(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, NPCs: []string{"npc.medea"}})
	before := sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix)
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	ui.client = newProtocolClient(clientConn)
	defer ui.client.Close()
	ui.connected = true
	ui.handleResponse("MOVE", "MOVE east", "OK room=loc.argo_bull_field")
	if err := serverConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewScanner(serverConn)
	if !reader.Scan() || reader.Text() != "LOOK" {
		t.Fatalf("MOVE did not request LOOK: %q, %v", reader.Text(), reader.Err())
	}
	ui.handleResponse("LOOK", "LOOK", `OK {"room":{"id":"loc.argo_bull_field","name":"Bull Field","exits":{}},"players":[],"items":[],"npcs":["npc.khalkotauroi"]}`)
	after := sha256.Sum256(ui.scene.Image.(*image.RGBA).Pix)
	if before == after {
		t.Fatal("art did not change after MOVE and LOOK")
	}
}

func TestQuitErrorKeepsGUIConnected(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	ui.client = newProtocolClient(clientConn)
	defer ui.client.Close()
	ui.connected = true
	ui.quitButton.Enable()
	if !ui.send("QUIT") {
		t.Fatal("QUIT was not queued")
	}
	ui.handleResponse("QUIT", "QUIT", "ERR 500 STATE_ERROR")
	if !ui.connected || ui.quitButton.Disabled() || ui.pollStop == nil {
		t.Fatal("QUIT error left the GUI in a disconnected state")
	}
	ui.stopPolling()
}
