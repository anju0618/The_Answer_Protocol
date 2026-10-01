package main

import (
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type menuChoice struct {
	label   string
	command string
	action  string
}

type gui struct {
	window           fyne.Window
	catalog          *worldCatalog
	client           *protocolClient
	pollStop         chan struct{}
	dialing          bool
	connected        bool
	locale           string
	room             lookView
	inventory        []string
	choices          []menuChoice
	state            stateView
	stateUnavailable bool
	quests           []questView
	hostEntry        *widget.Entry
	nameEntry        *widget.Entry
	languageSelect   *widget.Select
	connectButton    *widget.Button
	quitButton       *widget.Button
	settingsButton   *widget.Button
	statusLabel      *widget.Label
	roomTitle        *widget.Label
	roomDesc         *widget.Label
	scene            *canvas.Image
	roomCount        *widget.Label
	totalCount       *widget.Label
	hpLabel          *widget.Label
	hpBar            *statBar
	combatPanel      *fyne.Container
	combatName       *widget.Label
	combatHP         *widget.Label
	combatBar        *statBar
	fight            *fightState
	visited          map[string]bool
	endingBox        *fyne.Container
	flash            *canvas.Rectangle
	flashAnim        *fyne.Animation
	lastArc          string
	mapBox           *fyne.Container
	crewBar          *statBar
	crewLabel        *widget.Label
	groupLabel       *widget.Label
	exitBox          *fyne.Container
	playerBox        *fyne.Container
	itemBox          *fyne.Container
	itemPhotoBox     *fyne.Container
	itemPhotoScroll  *container.Scroll
	photoStrip       *fyne.Container
	npcBox           *fyne.Container
	inventoryBox     *fyne.Container
	questBox         *fyne.Container
	choiceTitle      *widget.Label
	choiceBox        *fyne.Container
	choicePopup      *widget.PopUp
	commandButtons   *fyne.Container
	journal          *container.AppTabs
	playArea         *fyne.Container
	scenePanel       fyne.CanvasObject
	detailPanel      fyne.CanvasObject
	chatScope        *widget.Select
	chatEntry        *widget.Entry
	chatLabel        *widget.Label
	storyText        *widget.RichText
	logLabel         *widget.Label
	chatScroll       *container.Scroll
	storyScroll      *container.Scroll
	logScroll        *container.Scroll
	messages         *container.AppTabs
	chatLines        []string
	storyLines       []storyEntry
	logLines         []string
}

func main() {
	application := app.New()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	window := application.NewWindow("The Answer Protocol")
	window.Resize(fyne.NewSize(1280, 900))
	catalog, catalogErr := loadCatalog()
	ui := &gui{window: window, catalog: catalog, locale: "en"}
	ui.build()
	if catalogErr != nil {
		ui.addLog(ui.tr("Could not load display names: ", "表示名データを読み込めません: ") + catalogErr.Error())
	}
	window.SetOnClosed(func() {
		if ui.client != nil {
			ui.client.Close()
		}
	})
	window.Show()
	time.AfterFunc(300*time.Millisecond, func() { fyne.Do(window.RequestFocus) })
	application.Run()
}

func (ui *gui) sendChat() {
	message := strings.TrimSpace(ui.chatEntry.Text)
	if message == "" {
		return
	}
	if ui.send("CHAT " + ui.chatScope.Selected + " " + message) {
		ui.chatEntry.SetText("")
	}
}

func (ui *gui) connect() {
	if ui.dialing || ui.client != nil {
		return
	}
	address := strings.TrimSpace(ui.hostEntry.Text)
	name := strings.TrimSpace(ui.nameEntry.Text)
	if address == "" || len(strings.Fields(name)) != 1 || strings.ContainsAny(name, "\r\n") {
		ui.showMessage(ui.tr("Connect", "接続"), ui.tr("Enter the server host:port and a name without spaces.", "サーバーのhost:portと空白のない名前を入力してください。"))
		return
	}
	ui.connectButton.Disable()
	ui.settingsButton.Disable()
	ui.languageSelect.Disable()
	ui.dialing = true
	ui.statusLabel.SetText(ui.tr("Connecting...", "接続中..."))
	ui.locale = "en"
	if ui.languageSelect.Selected == japaneseLanguageOption {
		ui.locale = "ja"
	}
	go func() {
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		fyne.Do(func() {
			ui.dialing = false
			if err != nil {
				ui.statusLabel.SetText(ui.tr("Connection failed", "接続失敗"))
				ui.connectButton.Enable()
				ui.settingsButton.Enable()
				ui.languageSelect.Enable()
				ui.addLog(ui.tr("Connection failed: ", "接続失敗: ") + err.Error())
				return
			}
			ui.client = newProtocolClient(conn)
			ui.window.Canvas().Unfocus()
			ui.addLog(ui.tr("Connected to server: ", "接続: ") + address)
			client := ui.client
			go ui.receive(client)
			if ui.locale == "ja" && !ui.sendHandshake("LANG ja") {
				return
			}
			ui.sendHandshake("CONNECT " + name)
		})
	}()
}

func (ui *gui) sendHandshake(line string) bool {
	if err := ui.client.Send(line); err != nil {
		ui.addLog(ui.tr("Send failed: ", "送信失敗: ") + err.Error())
		ui.disconnect()
		return false
	}
	ui.addLog("> " + line)
	return true
}

func (ui *gui) receive(client *protocolClient) {
	for message := range client.incoming {
		message := message
		fyne.Do(func() {
			if ui.client == client {
				ui.handleMessage(message)
			}
		})
	}
}

func (ui *gui) poll(client *protocolClient, stop <-chan struct{}) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fyne.Do(func() {
				select {
				case <-stop:
					return
				default:
				}
				ui.pollOnce(client)
			})
		case <-stop:
			return
		case <-client.done:
			return
		}
	}
}

func (ui *gui) pollOnce(client *protocolClient) {
	if ui.client == client && ui.connected && client.canPoll() {
		ui.refresh("LOOK", "WHO", "STATUS", "STATE")
	}
}

func (ui *gui) stopPolling() {
	if ui.pollStop != nil {
		close(ui.pollStop)
		ui.pollStop = nil
	}
}

func (ui *gui) send(line string) bool {
	if !ui.connected || ui.client == nil {
		ui.addLog(ui.tr("Connect before sending commands", "接続後にコマンドを送信してください"))
		return false
	}
	if err := ui.client.Send(line); err != nil {
		ui.addLog(ui.tr("Send failed: ", "送信失敗: ") + err.Error())
		return false
	}
	ui.addLog("> " + line)
	if strings.EqualFold(strings.TrimSpace(line), "QUIT") {
		ui.stopPolling()
		ui.quitButton.Disable()
	}
	return true
}

func (ui *gui) disconnect() {
	ui.stopPolling()
	if ui.client != nil {
		ui.client.Close()
		ui.client = nil
	}
	ui.connected = false
	ui.stateUnavailable = false
	ui.state = stateView{}
	ui.room = lookView{}
	ui.inventory = nil
	ui.quests = nil
	ui.clearChoices()
	ui.connectButton.Enable()
	ui.settingsButton.Enable()
	ui.quitButton.Disable()
	ui.languageSelect.Enable()
	ui.statusLabel.SetText(ui.tr("Not connected", "未接続"))
	ui.roomCount.SetText(ui.tr("Here: -", "部屋: - 人"))
	ui.totalCount.SetText(ui.tr("Online: -", "全体: - 人"))
	ui.hpLabel.SetText("HP: -")
	ui.hpBar.Set(0, 1)
	ui.showFight(nil)
	ui.visited, ui.lastArc = nil, ""
	ui.showState(stateView{})
	ui.scene.Resource = nil
	ui.scene.Image = loadArt("rooms", "unknown")
	ui.scene.Refresh()
	ui.showItemPhotos(nil)
	ui.showRoom(lookView{})
	ui.showInventory(nil)
	ui.showQuests(nil)
}

func (ui *gui) handleMessage(message serverMessage) {
	switch message.kind {
	case messageDisconnect:
		ui.addLog(ui.tr("Disconnected: ", "切断: ") + message.err.Error())
		ui.disconnect()
	case messageEvent:
		ui.addLog(message.line)
		ui.handleEvent(message.line)
	case messageGreeting, messageNotice:
		ui.addLog(message.line)
	case messageResponse:
		ui.addLog(message.line)
		ui.handleResponse(message.command, message.request, message.line)
	}
}

func (ui *gui) handleEvent(line string) {
	parts := strings.SplitN(line, " ", 5)
	if len(parts) >= 5 && parts[0] == "EVT" && parts[2] == "CHAT" {
		ui.addChat("[" + parts[1] + "] " + parts[3] + ": " + parts[4])
		return
	}
	if strings.HasPrefix(line, "EVT PLAYER ") {

		kind, text, _ := strings.Cut(strings.TrimPrefix(line, "EVT PLAYER "), " ")
		ui.addStoryKind(storyKindOf(kind), text)
		switch kind {
		case "DEATH":
			ui.flashScene(flashDeath, 900*time.Millisecond)
			ui.showFight(nil)
			ui.refresh("LOOK", "INVENTORY", "STATUS", "STATE")
		case "QUEST":
			ui.refresh("STATUS", "QUESTS")
		case "ENDING":
			ui.refresh("INVENTORY")
		case "TEAM":
			ui.refresh("STATUS", "QUESTS", "STATE")
		}
		return
	}
	if strings.HasPrefix(line, "EVT STATS players=") {
		ui.setTotal(strings.TrimPrefix(line, "EVT STATS players="))
	}
	if strings.HasPrefix(line, "EVT ROOM PRESENCE ") {
		ui.send("LOOK")
	}
	if strings.HasPrefix(line, "EVT ROOM COMBAT ") {
		ui.addStoryKind(storyCombat, strings.TrimPrefix(line, "EVT ROOM COMBAT "))
		ui.send("LOOK")
		ui.send("STATUS")
		ui.refresh("STATE")
	}
	if strings.HasPrefix(line, "EVT GROUP INVITE ") {
		leader := strings.TrimPrefix(line, "EVT GROUP INVITE ")
		if !slices.Contains(ui.state.Invitations, leader) {
			ui.state.Invitations = append(ui.state.Invitations, leader)
		}
		ui.addStory(leader + ui.tr(" invited you. Open Group to accept.", " から招待されました。「グループ」から参加できます。"))
	}
	if strings.HasPrefix(line, "EVT GROUP JOIN ") || strings.HasPrefix(line, "EVT GROUP LEAVE ") {
		ui.refresh("STATE")
	}
}

func (ui *gui) handleResponse(command, request, line string) {
	if strings.HasPrefix(line, "ERR ") {
		if command == "STATE" && strings.HasPrefix(line, "ERR 400 ") {
			ui.stateUnavailable = true
			ui.addLog(ui.tr("Crew and online names are unavailable on this server.", "このサーバーでは仲間の人数と全体の名前一覧を取得できません。"))
			return
		}
		ui.addStoryKind(storyError, line)
		if command == "QUIT" {
			ui.quitButton.Enable()
			if ui.pollStop == nil && ui.client != nil {
				ui.pollStop = make(chan struct{})
				go ui.poll(ui.client, ui.pollStop)
			}
		}
		if command == "LANG" {
			ui.switchLocale("en")
			ui.addLog("Japanese extension unavailable; continuing in English")
		}
		if command == "CONNECT" {
			ui.disconnect()
		}
		return
	}
	switch command {
	case "CONNECT":
		ui.connected = true
		ui.statusLabel.SetText(ui.tr("Playing as: ", "接続中: ") + ui.nameEntry.Text)
		ui.quitButton.Enable()
		ui.pollStop = make(chan struct{})
		go ui.poll(ui.client, ui.pollStop)
		for _, cmd := range []string{"LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO", "STATE"} {
			ui.send(cmd)
		}
	case "LOOK":
		var view lookView
		if err := decodeOK(line, &view); err != nil || view.Room.ID == "" {
			ui.addLog(ui.tr("Could not parse LOOK: ", "LOOKを解析できません: ") + parseError(err))
			return
		}
		ui.showRoom(view)
	case "INVENTORY":
		var items []string
		if err := decodeOK(line, &items); err != nil {
			ui.addLog(ui.tr("Could not parse INVENTORY: ", "INVENTORYを解析できません: ") + err.Error())
			return
		}
		ui.showInventory(items)
	case "STATUS":
		var status statusView
		if err := decodeOK(line, &status); err != nil {
			ui.addLog(ui.tr("Could not parse STATUS: ", "STATUSを解析できません: ") + err.Error())
			return
		}
		ui.hpLabel.SetText(fmt.Sprintf("HP: %d/%d", status.HP, status.MaxHP))
		ui.hpBar.Set(status.HP, status.MaxHP)
		if status.Status != "healthy" {
			ui.hpLabel.SetText(ui.hpLabel.Text + " (" + ui.statusWord(status.Status) + ")")
		}
	case "STATE":
		var state stateView
		if err := decodeOK(line, &state); err != nil {
			ui.addLog(err.Error())
			return
		}
		ui.showState(state)
	case "WHO":
		if strings.HasPrefix(line, "OK players=") {
			ui.setTotal(strings.TrimPrefix(line, "OK players="))
		}
	case "MOVE":
		destination := strings.TrimPrefix(line, "OK room=")
		ui.addStory(ui.tr("Moved to: ", "移動: ") + destination)
		ui.flashScene(flashTravel, 450*time.Millisecond)
		ui.showFight(nil)
		ui.markVisited(destination)
		ui.refresh("LOOK", "STATUS", "QUESTS", "STATE")
		if room, ok := ui.catalog.gameOverRoom(destination); ok {
			ui.recordFatalRoom(destination)
			ui.showGameOver(destination, room)
		}
	case "TAKE", "DROP":
		id := strings.TrimPrefix(strings.TrimPrefix(line, "OK taken="), "OK dropped=")
		if command == "TAKE" {
			ui.addStory(ui.tr("Taken: ", "手に入れた: ") + ui.catalog.label("item", id, ui.locale))
		} else {
			ui.addStory(ui.tr("Dropped: ", "置いた: ") + ui.catalog.label("item", id, ui.locale))
		}
		ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS", "STATE")
	case "DEFEND":
		if !ui.handleDefend(line) {
			ui.addStory(strings.TrimPrefix(line, "OK "))
		}
		ui.refresh("STATUS")
	case "ATTACK", "FLEE":
		handled := ui.handleAttack(request, line)
		if command == "FLEE" {
			handled = ui.handleFlee(line)
		}
		if !handled {
			ui.addStory(strings.TrimPrefix(line, "OK "))
		}
		ui.refresh("LOOK", "STATUS", "QUESTS", "STATE")
	case "TALK":
		words := strings.TrimPrefix(line, "OK ")
		ui.addStory(words)
		ui.showMessage("TALK", words)
		if line == "OK dead" {
			ui.refresh("LOOK", "STATUS", "STATE")
		}
	case "QUEST":
		words := strings.TrimPrefix(line, "OK ")
		var quest struct {
			QuestID     string `json:"quest_id"`
			Description string `json:"description"`
			Reward      int    `json:"reward"`
		}
		if json.Unmarshal([]byte(words), &quest) == nil && quest.Description != "" {
			words = ui.catalog.label("quest", quest.QuestID, ui.locale) + "\n" + quest.Description + fmt.Sprintf("\nHP +%d", quest.Reward)
		}
		ui.addStory(words)
		ui.showMessage("QUEST", words)
		ui.refresh("QUESTS")
	case "QUESTS":
		var quests []questView
		if err := decodeOK(line, &quests); err != nil {
			ui.addLog(ui.tr("Could not parse QUESTS: ", "QUESTSを解析できません: ") + err.Error())
			return
		}
		ui.showQuests(quests)
	case "GROUP":
		ui.addStory(line)
		if strings.HasPrefix(line, "OK group=") {
			ui.state.Group = strings.TrimPrefix(line, "OK group=")
			ui.state.Invitations = nil
		} else if strings.EqualFold(request, "GROUP LEAVE") {
			ui.state.Group = ""
		}
		ui.showState(ui.state)
		ui.refresh("STATE")
	case "QUIT":
		ui.addStory(ui.tr("Until our next journey.", "また旅を続けよう。"))
		ui.disconnect()
	}
}

func parseError(err error) string {
	if err == nil {
		return "room.id is empty"
	}
	return err.Error()
}

func (ui *gui) refresh(commands ...string) {
	for _, command := range commands {
		if command == "STATE" && ui.stateUnavailable {
			continue
		}
		ui.send(command)
	}
}

func (ui *gui) setTotal(value string) {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil && number >= 0 {
		ui.totalCount.SetText(fmt.Sprintf(ui.tr("Online: %d", "全体: %d 人"), number))
	}
}

func (ui *gui) showMessage(title, message string) {
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord
	var popup *widget.PopUp
	closeButton := widget.NewButton(ui.tr("Close", "とじる"), func() {
		popup.Hide()
		ui.window.Canvas().Unfocus()
	})
	content := framed(title, container.NewBorder(nil, closeButton, nil, nil, container.NewVScroll(label)))
	popup = widget.NewModalPopUp(container.NewGridWrap(fyne.NewSize(520, 220), content), ui.window.Canvas())
	popup.Show()
	ui.window.Canvas().Focus(closeButton)
}

func (ui *gui) showGameOver(roomID string, room catalogEntry) {
	scene := canvas.NewImageFromImage(loadArt("rooms", roomID))
	scene.FillMode = canvas.ImageFillContain
	scene.ScaleMode = canvas.ImageScaleSmooth
	scene.SetMinSize(fyne.NewSize(480, 288))
	title := widget.NewLabelWithStyle(room.Name.get(ui.locale), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapWord
	story := widget.NewLabel(room.Description.get(ui.locale))
	story.Wrapping = fyne.TextWrapWord
	var popup *widget.PopUp
	closeButton := widget.NewButton(ui.tr("Return to the Hall of the Fates", "運命の間へ戻る"), func() {
		popup.Hide()
		ui.window.Canvas().Unfocus()
	})
	content := framed(ui.tr("GAME OVER", "ゲームオーバー"),
		container.NewBorder(nil, closeButton, nil, nil, container.NewVScroll(textVBox(title, scene, story))))
	popup = widget.NewModalPopUp(container.NewGridWrap(fyne.NewSize(540, 520), content), ui.window.Canvas())
	popup.Show()
	ui.window.Canvas().Focus(closeButton)
}

func (ui *gui) addChat(line string) {
	ui.chatLines = appendLine(ui.chatLines, line)
	ui.chatLabel.SetText(strings.Join(ui.chatLines, "\n"))
	ui.chatScroll.ScrollToBottom()
}

func (ui *gui) addLog(line string) {
	ui.logLines = appendLine(ui.logLines, line)
	ui.logLabel.SetText(strings.Join(ui.logLines, "\n"))
	ui.logScroll.ScrollToBottom()
}

func appendLine(lines []string, line string) []string {
	lines = append(lines, time.Now().Format("15:04:05")+"  "+line)
	if len(lines) > 300 {
		return lines[len(lines)-300:]
	}
	return lines
}
