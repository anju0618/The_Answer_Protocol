package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const hubRoomID = "loc.hall_of_fates"

type gridPos struct{ x, y int }

var directionSteps = map[string]gridPos{
	"north": {0, -1}, "south": {0, 1}, "east": {1, 0}, "west": {-1, 0},
}

func arcOf(roomID string) string {
	if roomID == hubRoomID {
		return ""
	}
	name, _, _ := strings.Cut(roomID, "_")
	return name
}

func computeMapLayout(catalog *worldCatalog, arc string) map[string]gridPos {
	positions := map[string]gridPos{}
	if catalog == nil {
		return positions
	}
	taken := map[gridPos]bool{{0, 0}: true}
	positions[hubRoomID] = gridPos{0, 0}
	queue := []string{hubRoomID}
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for direction, to := range catalog.Rooms[from].Exits {
			step, ok := directionSteps[direction]
			if _, placed := positions[to]; !ok || placed {
				continue
			}
			if _, exists := catalog.Rooms[to]; !exists || (to != hubRoomID && arcOf(to) != arc) {
				continue
			}
			pos := gridPos{positions[from].x + step.x, positions[from].y + step.y}
			if taken[pos] {
				continue
			}
			taken[pos] = true
			positions[to] = pos
			queue = append(queue, to)
		}
	}
	return positions
}

func (ui *gui) markVisited(roomID string) {
	if roomID == "" {
		return
	}
	if ui.visited == nil {
		ui.visited = map[string]bool{}
	}
	if !ui.visited[roomID] {
		ui.visited[roomID] = true
	}
	ui.refreshMap()
}

type mapCell struct {
	obj fyne.CanvasObject
	pos gridPos
	to  *gridPos
}

type mapLayout struct {
	cells      []mapCell
	minX, minY int
	cols, rows int
}

func (*mapLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(200, 120) }

func (l *mapLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	if l.cols == 0 || l.rows == 0 {
		return
	}
	cell := min(size.Width/float32(l.cols), size.Height/float32(l.rows))
	originX := (size.Width - cell*float32(l.cols)) / 2
	originY := (size.Height - cell*float32(l.rows)) / 2
	center := func(p gridPos) fyne.Position {
		return fyne.NewPos(originX+(float32(p.x-l.minX)+0.5)*cell, originY+(float32(p.y-l.minY)+0.5)*cell)
	}
	for _, item := range l.cells {
		if item.to != nil {
			line := item.obj.(*canvas.Line)
			line.Position1, line.Position2 = center(item.pos), center(*item.to)
			continue
		}
		side := cell * 0.7
		item.obj.Resize(fyne.NewSize(side, side))
		item.obj.Move(center(item.pos).Subtract(fyne.NewPos(side/2, side/2)))
	}
}

var (
	mapUnknown = color.NRGBA{R: 60, G: 70, B: 96, A: 255}
	mapVisited = color.NRGBA{R: 112, G: 128, B: 170, A: 255}
	mapHazard  = color.NRGBA{R: 214, G: 140, B: 56, A: 255}
	mapLethal  = color.NRGBA{R: 190, G: 56, B: 56, A: 255}
	mapHub     = color.NRGBA{R: 70, G: 170, B: 110, A: 255}
)

func (ui *gui) roomMapColor(roomID string) color.Color {
	if roomID == hubRoomID {
		return mapHub
	}
	if entry := ui.catalog.Rooms[roomID]; entry.Hazard != nil {
		if entry.Hazard.Type == "lethal" {
			return mapLethal
		}
		return mapHazard
	}
	return mapVisited
}

func (ui *gui) refreshMap() {
	if ui.catalog == nil {
		return
	}
	arc := arcOf(ui.room.Room.ID)
	if arc == "" {
		arc = ui.lastArc
	}
	if arc == "" {
		arc = "loc.ody"
	}
	ui.lastArc = arc
	positions := computeMapLayout(ui.catalog, arc)
	for _, box := range []*fyne.Container{ui.mapBox, ui.sceneMapBox} {
		if box != nil {
			ui.redrawMap(box, positions)
		}
	}
}

func (ui *gui) redrawMap(box *fyne.Container, positions map[string]gridPos) {
	layout := &mapLayout{}
	first := true
	for _, pos := range positions {
		if first || pos.x < layout.minX {
			layout.minX = pos.x
		}
		if first || pos.y < layout.minY {
			layout.minY = pos.y
		}
		first = false
	}
	maxX, maxY := layout.minX, layout.minY
	for _, pos := range positions {
		maxX, maxY = max(maxX, pos.x), max(maxY, pos.y)
	}
	layout.cols, layout.rows = maxX-layout.minX+1, maxY-layout.minY+1

	known := map[string]bool{}
	for id := range ui.visited {
		known[id] = true
		for direction, to := range ui.catalog.Rooms[id].Exits {
			if _, ok := directionSteps[direction]; ok {
				known[to] = true
			}
		}
	}
	var lines, cells []mapCell
	for id, pos := range positions {
		if !known[id] {
			continue
		}
		if ui.visited[id] {
			for direction, to := range ui.catalog.Rooms[id].Exits {
				target, ok := positions[to]
				if _, isDir := directionSteps[direction]; ok && isDir && known[to] {
					line := canvas.NewLine(mapUnknown)
					line.StrokeWidth = 2
					lines = append(lines, mapCell{obj: line, pos: pos, to: &target})
				}
			}
		}
		rect := canvas.NewRectangle(mapUnknown)
		if ui.visited[id] {
			rect.FillColor = ui.roomMapColor(id)
		}
		if id == ui.room.Room.ID {
			rect.StrokeColor, rect.StrokeWidth = gold, 3
		} else {
			rect.StrokeColor, rect.StrokeWidth = ivory, 1
		}
		cells = append(cells, mapCell{obj: rect, pos: pos})
	}
	layout.cells = append(lines, cells...)
	box.Layout = layout
	box.Objects = nil
	for _, item := range layout.cells {
		box.Objects = append(box.Objects, item.obj)
	}
	box.Refresh()
}

func (ui *gui) buildMapTab() fyne.CanvasObject {
	ui.mapBox = container.NewWithoutLayout()
	return container.NewVScroll(ui.mapView(ui.mapBox))
}

func (ui *gui) mapView(box *fyne.Container) fyne.CanvasObject {
	legend := container.NewGridWithColumns(2,
		legendEntry(mapHub, ui.tr("Hub", "ハブ")),
		legendEntry(mapVisited, ui.tr("Visited", "訪れた")),
		legendEntry(mapHazard, ui.tr("Hazard", "危険")),
		legendEntry(mapLethal, ui.tr("Fatal", "即死")),
	)
	return container.NewBorder(nil, legend, nil, nil, box)
}

func legendEntry(c color.Color, text string) fyne.CanvasObject {
	swatch := canvas.NewRectangle(c)
	swatch.SetMinSize(fyne.NewSize(theme.Padding()*2, theme.Padding()*2))
	return container.NewBorder(nil, nil, container.NewCenter(swatch), nil, compactLabel(text))
}
