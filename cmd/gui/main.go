package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	prompt  string
}

type gui struct {
	window          fyne.Window
	catalog         *worldCatalog
	client          *protocolClient
	pollStop        chan struct{}
	dialing         bool
	connected       bool
	locale          string
	room            lookView
	inventory       []string
	choices         []menuChoice
	hostEntry       *widget.Entry
	nameEntry       *widget.Entry
	languageSelect  *widget.Select
	connectButton   *widget.Button
	quitButton      *widget.Button
	statusLabel     *widget.Label
	roomTitle       *widget.Label
	roomID          *widget.Label
	roomDesc        *widget.Label
	scene           *canvas.Image
	roomCount       *widget.Label
	totalCount      *widget.Label
	hpLabel         *widget.Label
	groupLabel      *widget.Label
	exitBox         *fyne.Container
	playerBox       *fyne.Container
	itemBox         *fyne.Container
	itemPhotoBox    *fyne.Container
	itemPhotoScroll *container.Scroll
	npcBox          *fyne.Container
	inventoryBox    *fyne.Container
	questBox        *fyne.Container
	choiceTitle     *widget.Label
	choiceBox       *fyne.Container
	commandButtons  *fyne.Container
	chatScope       *widget.Select
	chatEntry       *widget.Entry
	rawEntry        *widget.Entry
	chatLabel       *widget.Label
	storyLabel      *widget.Label
	logLabel        *widget.Label
	chatScroll      *container.Scroll
	storyScroll     *container.Scroll
	logScroll       *container.Scroll
	messages        *container.AppTabs
	chatLines       []string
	storyLines      []string
	logLines        []string
}

func main() {
	application := app.New()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	window := application.NewWindow("The Answer Protocol")
	window.Resize(fyne.NewSize(1280, 900))
	catalog, catalogErr := loadCatalog()
	ui := &gui{window: window, catalog: catalog, locale: "en"}
	ui.build()
	window.Canvas().Focus(ui.nameEntry)
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

func (ui *gui) build() {
	ui.hostEntry = widget.NewEntry()
	ui.hostEntry.SetText("127.0.0.1:4242")
	ui.hostEntry.SetPlaceHolder("host:port")
	ui.hostEntry.OnSubmitted = func(string) { ui.window.Canvas().Focus(ui.nameEntry) }
	ui.nameEntry = widget.NewEntry()
	ui.nameEntry.SetPlaceHolder(ui.tr("player name", "プレイヤー名"))
	ui.nameEntry.OnSubmitted = func(string) { ui.connect() }
	ui.languageSelect = widget.NewSelect([]string{"English", japaneseLanguageOption}, nil)
	ui.languageSelect.SetSelected(ui.tr("English", japaneseLanguageOption))
	ui.connectButton = widget.NewButton(ui.tr("Connect", "接続"), ui.connect)
	ui.quitButton = widget.NewButton("QUIT", func() { ui.send("QUIT") })
	ui.quitButton.Disable()
	ui.statusLabel = widget.NewLabel(ui.tr("Not connected", "未接続"))
	ui.roomCount = widget.NewLabel(ui.tr("Room players: -", "部屋: - 人"))
	ui.totalCount = widget.NewLabel(ui.tr("Online players: -", "全体: - 人"))
	ui.hpLabel = widget.NewLabel("HP: -")
	ui.groupLabel = widget.NewLabel("Group: -")

	ui.roomTitle = widget.NewLabelWithStyle(ui.tr("Room", "部屋"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ui.roomID = widget.NewLabel("")
	ui.roomDesc = widget.NewLabel(ui.tr("Connect to see the room.", "接続後にLOOKで表示します。"))
	ui.roomDesc.Wrapping = fyne.TextWrapWord
	ui.exitBox = container.NewVBox(widget.NewLabel("-"))
	ui.playerBox = container.NewVBox(widget.NewLabel("-"))
	ui.itemBox = container.NewVBox(widget.NewLabel("-"))
	ui.itemPhotoBox = container.NewHBox()
	ui.itemPhotoScroll = container.NewHScroll(ui.itemPhotoBox)
	ui.itemPhotoScroll.Hide()
	ui.npcBox = container.NewVBox(widget.NewLabel("-"))
	ui.inventoryBox = container.NewVBox(widget.NewLabel("-"))
	ui.questBox = container.NewVBox(widget.NewLabel("-"))
	ui.choiceTitle = widget.NewLabel(ui.tr("Choose a target by number", "番号で対象を選ぶ"))
	ui.choiceBox = container.NewVBox(widget.NewLabel("-"))
	ui.rawEntry = widget.NewEntry()
	ui.rawEntry.SetPlaceHolder(ui.tr("/ : RFC command", "/ : RFCコマンド"))
	ui.rawEntry.OnSubmitted = func(line string) {
		ui.send(strings.TrimSpace(line))
		ui.rawEntry.SetText("")
		ui.window.Canvas().Unfocus()
	}
	ui.chatScope = widget.NewSelect([]string{"GLOBAL", "ROOM", "GROUP"}, nil)
	ui.chatScope.SetSelected("ROOM")
	ui.chatEntry = widget.NewEntry()
	ui.chatEntry.SetPlaceHolder(ui.tr("C : Chat / F1-F3 : Scope", "C : 会話 / F1-F3 : 範囲"))
	ui.chatEntry.OnSubmitted = func(string) {
		ui.sendChat()
		ui.window.Canvas().Unfocus()
	}

	ui.scene = sceneImage()
	photoViewport := container.NewGridWrap(fyne.NewSize(224, 112), ui.itemPhotoScroll)
	titlePlate := canvas.NewRectangle(ink)
	titlePlate.StrokeColor = gold
	titlePlate.StrokeWidth = 1
	titleBadge := container.NewGridWrap(fyne.NewSize(280, 40), container.NewStack(titlePlate, container.NewPadded(ui.roomTitle)))
	sceneVisual := container.NewStack(ui.scene, container.NewBorder(
		titleBadge, container.NewPadded(photoViewport), nil, nil,
	))
	sceneFrame := canvas.NewRectangle(ink)
	sceneFrame.StrokeColor = ivory
	sceneFrame.StrokeWidth = 2
	scene := container.NewStack(sceneFrame, container.NewPadded(sceneVisual))
	surroundings := container.NewAppTabs(
		container.NewTabItem(ui.tr("Exits", "出口"), container.NewVScroll(ui.exitBox)),
		container.NewTabItem(ui.tr("Players", "人"), container.NewVScroll(ui.playerBox)),
		container.NewTabItem(ui.tr("Items", "道具"), container.NewVScroll(ui.itemBox)),
		container.NewTabItem("NPC", container.NewVScroll(ui.npcBox)),
	)
	ui.commandButtons = container.NewGridWithColumns(4,
		ui.commandButton(ui.tr("L Look", "L 調べる"), func() { ui.send("LOOK") }),
		ui.commandButton(ui.tr("M Move", "M 移動"), func() { ui.chooseAction("MOVE") }),
		ui.commandButton(ui.tr("T Take", "T 取る"), func() { ui.chooseAction("TAKE") }),
		ui.commandButton(ui.tr("D Drop", "D 捨てる"), func() { ui.chooseAction("DROP") }),
		ui.commandButton(ui.tr("N Talk", "N 話す"), func() { ui.chooseAction("TALK") }),
		ui.commandButton(ui.tr("A Attack", "A 戦う"), func() { ui.chooseAction("ATTACK") }),
		ui.commandButton(ui.tr("E Quest", "E 依頼"), func() { ui.chooseAction("QUEST") }),
		ui.commandButton(ui.tr("I Inventory", "I 持ち物"), func() { ui.send("INVENTORY") }),
		ui.commandButton(ui.tr("S Status", "S 状態"), func() { ui.send("STATUS") }),
		ui.commandButton(ui.tr("Q Quests", "Q 依頼一覧"), func() { ui.send("QUESTS") }),
		ui.commandButton(ui.tr("W Who", "W 人数"), func() { ui.send("WHO") }),
		ui.commandButton(ui.tr("G Group", "G 仲間"), func() { ui.chooseAction("GROUP") }),
		ui.commandButton(ui.tr("F Flee", "F 逃げる"), func() { ui.send("FLEE") }),
		ui.commandButton(ui.tr("C Chat", "C 会話"), ui.focusChat),
		ui.commandButton(ui.tr("/ Command", "/ 入力"), func() { ui.window.Canvas().Focus(ui.rawEntry) }),
		ui.commandButton(ui.tr("X Quit", "X 終了"), func() { ui.send("QUIT") }),
	)
	choiceArea := container.NewBorder(ui.choiceTitle, nil, nil, nil, container.NewVScroll(ui.choiceBox))
	commandTop := container.NewVBox(
		widget.NewLabel(ui.tr("Arrows: move / Numbers: select / [ ]: item art", "矢印: 移動  /  数字: 対象を選択  /  [ ]: イラスト")),
		ui.commandButtons, widget.NewSeparator(),
	)
	commandPane := framed(ui.tr("Commands", "コマンド"), container.NewBorder(commandTop, ui.rawEntry, nil, nil, choiceArea))
	journal := container.NewAppTabs(
		container.NewTabItem(ui.tr("Around", "まわり"), surroundings),
		container.NewTabItem(ui.tr("Inventory", "もちもの"), container.NewVScroll(ui.inventoryBox)),
		container.NewTabItem(ui.tr("Quests", "クエスト"), container.NewVScroll(ui.questBox)),
	)
	right := container.NewVSplit(commandPane, framed(ui.tr("Journal", "記録"), journal))
	right.Offset = 0.62
	worldSplit := container.NewHSplit(scene, right)
	worldSplit.Offset = 0.61

	ui.chatLabel = widget.NewLabel("")
	ui.chatLabel.Wrapping = fyne.TextWrapWord
	ui.chatScroll = container.NewVScroll(ui.chatLabel)
	chatInput := container.NewBorder(nil, nil, ui.chatScope, widget.NewButton(ui.tr("Send", "送信"), ui.sendChat), ui.chatEntry)
	chatPane := container.NewBorder(nil, chatInput, nil, nil, ui.chatScroll)

	ui.logLabel = widget.NewLabel("")
	ui.logLabel.Wrapping = fyne.TextWrapWord
	ui.logScroll = container.NewVScroll(ui.logLabel)
	ui.storyLabel = widget.NewLabel(ui.tr("The gods of Greece await you.", "ギリシアの神々があなたを待っている。"))
	ui.storyLabel.Wrapping = fyne.TextWrapWord
	ui.storyScroll = container.NewVScroll(ui.storyLabel)
	adventure := container.NewBorder(ui.roomDesc, nil, nil, nil, ui.storyScroll)
	ui.messages = container.NewAppTabs(
		container.NewTabItem(ui.tr("Adventure", "ぼうけん"), adventure),
		container.NewTabItem(ui.tr("Chat", "チャット"), chatPane),
		container.NewTabItem(ui.tr("Log", "ログ"), ui.logScroll),
	)
	mainSplit := container.NewVSplit(worldSplit, framed(ui.tr("Messages", "ことば"), ui.messages))
	mainSplit.Offset = 0.64
	connectionRow := container.NewBorder(nil, nil, widget.NewLabel(ui.tr("Server", "サーバー")), nil, ui.hostEntry)
	nameRow := container.NewBorder(nil, nil, widget.NewLabel(ui.tr("Name", "名前")),
		container.NewHBox(ui.languageSelect, ui.connectButton, ui.quitButton), ui.nameEntry)
	header := container.NewVBox(
		connectionRow, nameRow,
		container.NewGridWithColumns(5, ui.statusLabel, ui.roomCount, ui.totalCount, ui.hpLabel, ui.groupLabel),
	)
	ui.window.SetContent(container.NewBorder(framed("THE ANSWER PROTOCOL", header), nil, nil, nil, mainSplit))
	ui.bindKeyboard()
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
	if len(ui.chatLines) > 0 {
		ui.chatLabel.SetText(strings.Join(ui.chatLines, "\n"))
	}
	if len(ui.storyLines) > 0 {
		ui.storyLabel.SetText(strings.Join(ui.storyLines, "\n"))
	}
	if len(ui.logLines) > 0 {
		ui.logLabel.SetText(strings.Join(ui.logLines, "\n"))
	}
}

func (ui *gui) commandButton(text string, action func()) *widget.Button {
	return widget.NewButton(text, func() {
		action()
		if ui.window.Canvas().Focused() != ui.chatEntry && ui.window.Canvas().Focused() != ui.rawEntry {
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
		ui.addLog(ui.tr("Connect before sending commands", "接続後にコマンドを送信してください"))
		return
	}
	var choices []menuChoice
	switch action {
	case "MOVE":
		var directions []string
		for direction := range ui.room.Room.Exits {
			directions = append(directions, direction)
		}
		sort.Strings(directions)
		for _, direction := range directions {
			choices = append(choices, menuChoice{label: ui.exitLabel(direction, ui.room.Room.Exits[direction]), command: "MOVE " + direction})
		}
	case "TAKE":
		for _, id := range ui.room.Items {
			choices = append(choices, menuChoice{label: ui.catalog.label("item", id, ui.locale), command: "TAKE " + id})
		}
	case "DROP":
		for _, id := range ui.inventory {
			choices = append(choices, menuChoice{label: ui.catalog.label("item", id, ui.locale), command: "DROP " + id})
		}
	case "TALK", "ATTACK", "QUEST":
		for _, id := range ui.room.NPCs {
			choices = append(choices, menuChoice{label: ui.catalog.label("npc", id, ui.locale), command: action + " " + id})
		}
	case "GROUP":
		choices = []menuChoice{
			{label: "CREATE", command: "GROUP CREATE"},
			{label: ui.tr("INVITE player", "INVITE 名前"), prompt: "GROUP INVITE "},
			{label: ui.tr("JOIN leader", "JOIN リーダー"), prompt: "GROUP JOIN "},
			{label: "LEAVE", command: "GROUP LEAVE"},
		}
	}
	ui.choices = choices
	if len(choices) == 0 {
		ui.choiceTitle.SetText(action + ui.tr(": no targets", ": 対象なし"))
		setRows(ui.choiceBox, nil)
		return
	}
	ui.choiceTitle.SetText(action + ui.tr(": select a number (Esc to cancel)", ": 数字を押す (Escで戻る)"))
	var rows []fyne.CanvasObject
	for index, choice := range choices {
		if index >= 9 {
			break
		}
		index := index
		rows = append(rows, ui.commandButton(fmt.Sprintf("%d  %s", index+1, choice.label), func() { ui.runChoice(index) }))
	}
	if len(choices) > 9 {
		rows = append(rows, widget.NewLabel(ui.tr("For targets 10+, type the command after /", "10件目以降は / でコマンドを直接入力")))
	}
	setRows(ui.choiceBox, rows)
}

func (ui *gui) runChoice(index int) {
	if index < 0 || index >= len(ui.choices) {
		return
	}
	choice := ui.choices[index]
	if choice.prompt != "" {
		ui.rawEntry.SetText(choice.prompt)
		ui.window.Canvas().Focus(ui.rawEntry)
		return
	}
	ui.send(choice.command)
	ui.choices = nil
	ui.choiceTitle.SetText(ui.tr("Choose a target by number", "番号で対象を選ぶ"))
	setRows(ui.choiceBox, nil)
}

func (ui *gui) bindKeyboard() {
	ui.window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyF1:
			ui.chatScope.SetSelected("GLOBAL")
		case fyne.KeyF2:
			ui.chatScope.SetSelected("ROOM")
		case fyne.KeyF3:
			ui.chatScope.SetSelected("GROUP")
		}
		if ui.window.Canvas().Focused() != nil {
			return
		}
		if event.Name == fyne.KeyEscape {
			ui.choices = nil
			ui.choiceTitle.SetText(ui.tr("Choose a target by number", "番号で対象を選ぶ"))
			setRows(ui.choiceBox, nil)
			return
		}
		directions := map[fyne.KeyName]string{fyne.KeyUp: "north", fyne.KeyDown: "south", fyne.KeyLeft: "west", fyne.KeyRight: "east"}
		if direction, ok := directions[event.Name]; ok {
			if _, exists := ui.room.Room.Exits[direction]; exists {
				ui.send("MOVE " + direction)
			}
		}
	})
	ui.window.Canvas().SetOnTypedRune(func(value rune) {
		if ui.window.Canvas().Focused() != nil {
			return
		}
		if value >= '1' && value <= '9' {
			ui.runChoice(int(value - '1'))
			return
		}
		switch unicode.ToUpper(value) {
		case '[':
			ui.scrollItemPhotos(-110)
		case ']':
			ui.scrollItemPhotos(110)
		case 'L':
			ui.send("LOOK")
		case 'M':
			ui.chooseAction("MOVE")
		case 'T':
			ui.chooseAction("TAKE")
		case 'D':
			ui.chooseAction("DROP")
		case 'N':
			ui.chooseAction("TALK")
		case 'A':
			ui.chooseAction("ATTACK")
		case 'E':
			ui.chooseAction("QUEST")
		case 'I':
			ui.send("INVENTORY")
		case 'S':
			ui.send("STATUS")
		case 'Q':
			ui.send("QUESTS")
		case 'W':
			ui.send("WHO")
		case 'G':
			ui.chooseAction("GROUP")
		case 'F':
			ui.send("FLEE")
		case 'C':
			ui.focusChat()
		case '/':
			ui.window.Canvas().Focus(ui.rawEntry)
		case 'X':
			ui.send("QUIT")
		}
	})
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
			if client.canPoll() {
				client.Send("LOOK")
				client.Send("WHO")
			}
		case <-stop:
			return
		case <-client.done:
			return
		}
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
	ui.room = lookView{}
	ui.inventory = nil
	ui.choices = nil
	ui.choiceTitle.SetText(ui.tr("Choose a target by number", "番号で対象を選ぶ"))
	setRows(ui.choiceBox, nil)
	ui.connectButton.Enable()
	ui.quitButton.Disable()
	ui.languageSelect.Enable()
	ui.statusLabel.SetText(ui.tr("Not connected", "未接続"))
	ui.roomCount.SetText(ui.tr("Room players: -", "部屋: - 人"))
	ui.totalCount.SetText(ui.tr("Online players: -", "全体: - 人"))
	ui.hpLabel.SetText("HP: -")
	ui.groupLabel.SetText("Group: -")
	ui.roomTitle.SetText(ui.tr("Room", "部屋"))
	ui.scene.Resource = nil
	ui.scene.Image = loadArt("rooms", "unknown")
	ui.scene.Refresh()
	ui.showItemPhotos(nil)
	ui.roomID.SetText("")
	ui.roomDesc.SetText(ui.tr("Connect to see the room.", "接続後にLOOKで表示します。"))
	setRows(ui.exitBox, nil)
	setRows(ui.playerBox, nil)
	setRows(ui.itemBox, nil)
	setRows(ui.npcBox, nil)
	setRows(ui.inventoryBox, nil)
	setRows(ui.questBox, nil)
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
		ui.addStory(text)
		switch kind {
		case "DEATH":
			ui.refresh("LOOK", "INVENTORY", "STATUS")
		case "QUEST":
			ui.refresh("STATUS", "QUESTS")
		case "ENDING":
			ui.refresh("INVENTORY")
		case "TEAM":
			ui.refresh("STATUS", "QUESTS")
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
		ui.addStory(strings.TrimPrefix(line, "EVT ROOM COMBAT "))
		ui.send("LOOK")
		ui.send("STATUS")
	}
	if strings.HasPrefix(line, "EVT GROUP INVITE ") {
		leader := strings.TrimPrefix(line, "EVT GROUP INVITE ")
		ui.showMessage("GROUP INVITE", leader+ui.tr(" invited you. Choose G: JOIN.", " から招待されました。G : JOIN を選んでください。"))
	}
}

func (ui *gui) handleResponse(command, request, line string) {
	if strings.HasPrefix(line, "ERR ") {
		ui.addStory(line)
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
		for _, cmd := range []string{"LOOK", "INVENTORY", "STATUS", "QUESTS", "WHO"} {
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
		ui.hpLabel.SetText(fmt.Sprintf("HP: %d/%d (%s)", status.HP, status.MaxHP, ui.statusWord(status.Status)))
	case "WHO":
		if strings.HasPrefix(line, "OK players=") {
			ui.setTotal(strings.TrimPrefix(line, "OK players="))
		}
	case "MOVE":
		ui.addStory(ui.tr("Moved to: ", "移動: ") + strings.TrimPrefix(line, "OK room="))
		ui.refresh("LOOK", "STATUS", "QUESTS")
	case "TAKE", "DROP":
		id := strings.TrimPrefix(strings.TrimPrefix(line, "OK taken="), "OK dropped=")
		if command == "TAKE" {
			ui.addStory(ui.tr("Taken: ", "手に入れた: ") + ui.catalog.label("item", id, ui.locale))
		} else {
			ui.addStory(ui.tr("Dropped: ", "置いた: ") + ui.catalog.label("item", id, ui.locale))
		}
		ui.refresh("LOOK", "INVENTORY", "STATUS", "QUESTS")
	case "ATTACK", "FLEE":
		ui.addStory(strings.TrimPrefix(line, "OK "))
		ui.refresh("LOOK", "STATUS", "QUESTS")
	case "TALK":
		words := strings.TrimPrefix(line, "OK ")
		ui.addStory(words)
		ui.showMessage("TALK", words)
		if line == "OK dead" {
			ui.refresh("LOOK", "STATUS")
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
			ui.groupLabel.SetText("Group: " + strings.TrimPrefix(line, "OK group="))
		} else if strings.EqualFold(request, "GROUP LEAVE") {
			ui.groupLabel.SetText("Group: -")
		}
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
		ui.send(command)
	}
}

func (ui *gui) setTotal(value string) {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil && number >= 0 {
		ui.totalCount.SetText(fmt.Sprintf(ui.tr("Online players: %d", "全体: %d 人"), number))
	}
}

func (ui *gui) showRoom(view lookView) {
	ui.room = view
	ui.scene.Resource = nil
	ui.scene.Image = composeScene(view.Room.ID, view.NPCs)
	ui.scene.Refresh()
	ui.showItemPhotos(view.Items)
	ui.roomTitle.SetText(view.Room.Name)
	ui.roomID.SetText(view.Room.ID)
	ui.roomDesc.SetText(view.Room.Description)
	ui.roomCount.SetText(fmt.Sprintf(ui.tr("Room players: %d", "部屋: %d 人"), len(view.Players)))
	var exits []fyne.CanvasObject
	directions := make([]string, 0, len(view.Room.Exits))
	for direction := range view.Room.Exits {
		directions = append(directions, direction)
	}
	sort.Strings(directions)
	for _, direction := range directions {
		destination := view.Room.Exits[direction]
		exits = append(exits, widget.NewLabel(ui.exitLabel(direction, destination)))
	}
	setRows(ui.exitBox, exits)
	var players []fyne.CanvasObject
	for _, name := range view.Players {
		players = append(players, widget.NewLabel(name))
	}
	setRows(ui.playerBox, players)
	var items []fyne.CanvasObject
	for _, id := range view.Items {
		label := widget.NewLabel(ui.catalog.label("item", id, ui.locale))
		label.Wrapping = fyne.TextWrapWord
		items = append(items, label)
	}
	setRows(ui.itemBox, items)
	var npcs []fyne.CanvasObject
	for _, id := range view.NPCs {
		label := widget.NewLabel(ui.catalog.label("npc", id, ui.locale))
		label.Wrapping = fyne.TextWrapWord
		npcs = append(npcs, label)
	}
	setRows(ui.npcBox, npcs)
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
	} else {
		ui.itemPhotoScroll.Show()
	}
	ui.itemPhotoScroll.ScrollToOffset(fyne.Position{})
}

func (ui *gui) scrollItemPhotos(delta float32) {
	offset := ui.itemPhotoScroll.Offset
	offset.X += delta
	if offset.X < 0 {
		offset.X = 0
	}
	ui.itemPhotoScroll.ScrollToOffset(offset)
}

func (ui *gui) showInventory(ids []string) {
	ui.inventory = append([]string(nil), ids...)
	var rows []fyne.CanvasObject
	for _, id := range ids {
		label := widget.NewLabel(ui.catalog.label("item", id, ui.locale))
		label.Wrapping = fyne.TextWrapWord
		rows = append(rows, label)
	}
	setRows(ui.inventoryBox, rows)
}

func (ui *gui) showQuests(quests []questView) {
	var rows []fyne.CanvasObject
	for _, quest := range quests {
		label := widget.NewLabel(ui.catalog.label("quest", quest.QuestID, ui.locale) + " - " + ui.statusWord(quest.Status) + " " + quest.Progress)
		label.Wrapping = fyne.TextWrapWord
		rows = append(rows, label)
	}
	setRows(ui.questBox, rows)
}

func setRows(box *fyne.Container, rows []fyne.CanvasObject) {
	if len(rows) == 0 {
		rows = []fyne.CanvasObject{widget.NewLabel("- ")}
	}
	box.Objects = rows
	box.Refresh()
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

func (ui *gui) addChat(line string) {
	ui.chatLines = appendLine(ui.chatLines, line)
	ui.chatLabel.SetText(strings.Join(ui.chatLines, "\n"))
	ui.chatScroll.ScrollToBottom()
}

func (ui *gui) addStory(line string) {
	ui.storyLines = appendLine(ui.storyLines, line)
	ui.storyLabel.SetText(strings.Join(ui.storyLines, "\n"))
	ui.storyScroll.ScrollToBottom()
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
