package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type layoutTestTheme struct {
	fyne.Theme
	textSize float32
}

func (t layoutTestTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return t.textSize
	}
	return t.Theme.Size(name)
}

func checkJournalBounds(t *testing.T, object fyne.CanvasObject) {
	t.Helper()
	if !object.Visible() {
		return
	}
	switch child := object.(type) {
	case *widget.Label:
		if child.Size().Height+1 < child.MinSize().Height {
			t.Errorf("label %q overlaps its row: size=%v min=%v", child.Text, child.Size(), child.MinSize())
		}
		if child.Wrapping == fyne.TextWrapOff && child.Truncation == fyne.TextTruncateOff && child.Size().Width+1 < child.MinSize().Width {
			t.Errorf("label %q clips its text: size=%v min=%v", child.Text, child.Size(), child.MinSize())
		}
	case *widget.Button:
		if child.Size().Width+1 < child.MinSize().Width || child.Size().Height+1 < child.MinSize().Height {
			t.Errorf("button %q clips its text: size=%v min=%v", child.Text, child.Size(), child.MinSize())
		}
	case *fyne.Container:
		for index, item := range child.Objects {
			if item.Visible() && (item.Position().X < -1 || item.Position().Y < -1 || item.Position().X+item.Size().Width > child.Size().Width+1 || item.Position().Y+item.Size().Height > child.Size().Height+1) {
				t.Errorf("%T extends outside %T: position=%v size=%v parent=%v", item, child.Layout, item.Position(), item.Size(), child.Size())
			}
			overlays := reflect.TypeOf(child.Layout) == reflect.TypeOf(layout.NewStackLayout())
			switch child.Layout.(type) {
			case *barLayout, *mapLayout:
				overlays = true
			}
			if item.Visible() && !overlays {
				for _, other := range child.Objects[:index] {
					if other.Visible() && item.Position().X < other.Position().X+other.Size().Width-1 && other.Position().X < item.Position().X+item.Size().Width-1 && item.Position().Y < other.Position().Y+other.Size().Height-1 && other.Position().Y < item.Position().Y+item.Size().Height-1 {
						t.Errorf("%T and %T overlap in %T", item, other, child.Layout)
					}
				}
			}
			checkJournalBounds(t, item)
		}
	case *container.Scroll:
		checkJournalBounds(t, child.Content)
	}
}

func resizeLayoutWindow(window fyne.Window, size fyne.Size) {
	window.Resize(size)
	minimum := window.Content().MinSize()
	if window.Padded() {
		minimum = minimum.Add(fyne.NewSquareSize(2 * theme.Padding()))
	}
	window.Resize(fyne.NewSize(max(size.Width, minimum.Width), max(size.Height, minimum.Height)))
}

func TestGUITextAndArtworkStayWithinPanels(t *testing.T) {
	for _, locale := range []string{"ja", "en"} {
		for _, textSize := range []float32{15, 22} {
			t.Run(fmt.Sprintf("%s-text-%g", locale, textSize), func(t *testing.T) {
				application := test.NewApp()
				application.Settings().SetTheme(layoutTestTheme{Theme: retroTheme{base: theme.DarkTheme()}, textSize: textSize})
				defer application.Quit()
				window := application.NewWindow("overlap test")
				defer window.Close()
				catalog, err := loadCatalog()
				if err != nil {
					t.Fatal(err)
				}
				catalog.NPCs["npc.colchis_dragon"].Name["ja"] = "眠らぬ竜（黄金の羊毛を守るコルキスの竜）"
				ui := &gui{window: window, locale: locale, catalog: catalog}
				ui.build()
				ui.nameEntry.SetText("alice")
				ui.showRoom(lookView{
					Room: roomView{ID: "loc.argo_grove", Name: catalog.label("room", "loc.argo_grove", locale),
						Description: catalog.Rooms["loc.argo_grove"].Description.get(locale),
						Exits:       map[string]string{"west": "loc.argo_bull_field", "east": "loc.ody_laestrygonian_depths"}},
					Players: []string{"alice", "bob"}, NPCs: []string{"npc.medea", "npc.colchis_dragon"},
					Items: []string{"item.golden_fleece", "item.olive_branch_of_ithaca", "item.laurel_of_the_argo"},
				})
				ui.showInventory(append(append([]string(nil), ui.room.Items...), "item.ember_of_troy"))
				ui.showQuests([]questView{{QuestID: "quest.golden_fleece", Status: "active", Progress: "0/1"}, {QuestID: "quest.phineus_harpies", Status: "completed", Progress: "1/1"}})
				ui.showState(stateView{Crew: 8, CrewInitialized: true, Group: "group.alice"})
				ui.statusLabel.SetText(ui.tr("Connected: alice", "接続中: alice"))
				ui.chatEntry.SetText("unsent message")
				window.Show()
				for _, size := range []fyne.Size{fyne.NewSize(1280, 900), fyne.NewSize(560, 680), fyne.NewSize(1280, 680), fyne.NewSize(640, 900), fyne.NewSize(960, 680), fyne.NewSize(1280, 900)} {
					for tab := range len(ui.journal.Items) {
						t.Run(fmt.Sprintf("%gx%g-tab%d", size.Width, size.Height, tab), func(t *testing.T) {
							ui.journal.SelectIndex(tab)
							resizeLayoutWindow(window, size)
							capture := window.Canvas().Capture()
							for _, label := range []*widget.Label{ui.hpLabel, ui.crewLabel, ui.groupLabel, ui.roomCount, ui.totalCount, ui.statusLabel, ui.roomTitle} {
								checkJournalBounds(t, label)
							}
							checkJournalBounds(t, ui.journal.Items[tab].Content.(*container.Scroll).Content)
							checkJournalBounds(t, ui.commandButtons)
							checkJournalBounds(t, ui.photoStrip)
							root := window.Content().(*fyne.Container).Objects[0].(*fyne.Container)
							checkJournalBounds(t, root.Objects[0])
							for _, panel := range []fyne.CanvasObject{ui.scenePanel, ui.detailPanel} {
								if panel.Position().Y+panel.Size().Height > ui.playArea.Size().Height+1 || panel.Position().X+panel.Size().Width > ui.playArea.Size().Width+1 {
									t.Errorf("panel extends outside play area: position=%v size=%v area=%v", panel.Position(), panel.Size(), ui.playArea.Size())
								}
							}
							if ui.photoStrip.Size().Height > ui.scene.Size().Height || ui.photoStrip.Size().Width > ui.scene.Size().Width {
								t.Errorf("item artwork exceeds scene: items=%v scene=%v", ui.photoStrip.Size(), ui.scene.Size())
							}
							if ui.scene.Position().X < theme.Padding() || ui.scene.Position().Y < theme.Padding() {
								t.Errorf("room artwork touches the frame: position=%v", ui.scene.Position())
							}
							descriptionPane := ui.scenePanel.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*container.Scroll)
							if descriptionPane.Size().Height+1 < widget.NewLabel("説明文").MinSize().Height {
								t.Errorf("description viewport clips a line: size=%v", descriptionPane.Size())
							}
							if ui.chatEntry.Text != "unsent message" {
								t.Fatal("resizing discarded chat input")
							}
							saveOverlapPreview(t, fmt.Sprintf("%s-text-%g-%gx%g-tab%d", locale, textSize, size.Width, size.Height, tab), capture)
						})
					}
				}
			})
		}
	}
}

func saveOverlapPreview(t *testing.T, name string, capture image.Image) {
	t.Helper()
	if directory := os.Getenv("TAP_GUI_PREVIEW_DIR"); directory != "" {
		file, err := os.Create(filepath.Join(directory, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, capture)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("save preview: %v, %v", err, closeErr)
		}
	}
}

func TestAttackAndFleeShareRowInEnemyCard(t *testing.T) {
	for _, locale := range []string{"en", "ja"} {
		t.Run(locale, func(t *testing.T) {
			ui, reader := mouseTestGUI(t)
			ui.switchLocale(locale)
			ui.catalog = &worldCatalog{NPCs: map[string]catalogEntry{"npc.enemy": {Role: "enemy"}}}
			ui.showRoom(lookView{Room: roomView{ID: "loc.test"}, NPCs: []string{"npc.enemy", "npc.unknown"}})
			ui.window.Show()
			ui.window.Canvas().Capture()
			for _, card := range ui.npcBox.Objects {
				attack := findButton(t, card, ui.tr("Attack", "戦う"))
				flee := findButton(t, card, ui.tr("Flee", "逃げる"))
				if attack.Position().Y != flee.Position().Y || flee.Position().X < attack.Position().X+attack.Size().Width {
					t.Fatalf("combat actions are separated: attack=%v flee=%v", attack.Position(), flee.Position())
				}
				test.Tap(flee)
				expectMouseCommand(t, reader, "FLEE")
			}
		})
	}
}

func TestJapaneseGameOverContentStaysWithinDialog(t *testing.T) {
	application := test.NewApp()
	application.Settings().SetTheme(layoutTestTheme{Theme: retroTheme{base: theme.DarkTheme()}, textSize: 22})
	defer application.Quit()
	window := application.NewWindow("game over layout test")
	defer window.Close()
	window.Resize(fyne.NewSize(584, 740))
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ui := &gui{window: window, locale: "ja", catalog: catalog}
	ui.build()
	window.Show()
	ui.showGameOver("loc.ody_sealed_cave", catalog.Rooms["loc.ody_sealed_cave"])
	saveOverlapPreview(t, "ja-gameover", window.Canvas().Capture())
	overlay := window.Canvas().Overlays().Top().(fyne.Widget)
	for _, object := range test.WidgetRenderer(overlay).Objects() {
		if popup, ok := object.(*widget.PopUp); ok {
			checkJournalBounds(t, popup.Content)
			return
		}
	}
	t.Fatal("game over dialog was not displayed")
}

func TestItemArtworkScalesAndScrollsOneCard(t *testing.T) {
	ui, _ := mouseTestGUI(t)
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, Items: []string{
		"item.golden_fleece", "item.olive_branch_of_ithaca", "item.laurel_of_the_argo", "item.thread_of_fate",
		"item.bag_of_winds", "item.beeswax", "item.moly",
	}})
	ui.window.Show()
	ui.window.Canvas().Capture()
	wideSize := ui.itemPhotoBox.Objects[0].MinSize()
	resizeLayoutWindow(ui.window, fyne.NewSize(560, 680))
	ui.window.Canvas().Capture()
	if ui.itemPhotoBox.Objects[0].MinSize().Height >= wideSize.Height {
		t.Fatal("item artwork did not shrink for the smaller scene")
	}
	ui.scrollItemPhotos(1)
	want := ui.itemPhotoBox.Objects[0].Size().Width + theme.Padding()
	if ui.itemPhotoScroll.Offset.X != want {
		t.Fatalf("scroll skipped a card: offset=%g want=%g", ui.itemPhotoScroll.Offset.X, want)
	}
	ui.scrollItemPhotos(-1)
	if ui.itemPhotoScroll.Offset.X != 0 {
		t.Fatalf("scroll did not return to first card: offset=%g", ui.itemPhotoScroll.Offset.X)
	}
}

func TestActiveCombatLayoutFitsJapaneseAndEnglish(t *testing.T) {
	for _, locale := range []string{"ja", "en"} {
		t.Run(locale, func(t *testing.T) {
			application := test.NewApp()
			application.Settings().SetTheme(layoutTestTheme{Theme: retroTheme{base: theme.DarkTheme()}, textSize: 22})
			defer application.Quit()
			window := application.NewWindow("combat layout test")
			defer window.Close()
			catalog, err := loadCatalog()
			if err != nil {
				t.Fatal(err)
			}
			ui := &gui{window: window, locale: locale, catalog: catalog}
			ui.build()
			ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove", Name: catalog.label("room", "loc.argo_grove", locale)}, NPCs: []string{"npc.colchis_dragon"}, Items: []string{"item.golden_fleece"}})
			ui.showFight(&fightState{npcID: "npc.colchis_dragon", hp: 80, maxHP: 100})
			window.Show()
			for _, size := range []fyne.Size{fyne.NewSize(1280, 900), fyne.NewSize(560, 680)} {
				resizeLayoutWindow(window, size)
				saveOverlapPreview(t, fmt.Sprintf("%s-combat-%gx%g", locale, size.Width, size.Height), window.Canvas().Capture())
				checkJournalBounds(t, ui.combatPanel)
				checkJournalBounds(t, ui.detailPanel)
				if ui.journal.Size().Height+1 < ui.journal.MinSize().Height {
					t.Errorf("combat panel covers journal: journal=%v min=%v", ui.journal.Size(), ui.journal.MinSize())
				}
			}
		})
	}
}
