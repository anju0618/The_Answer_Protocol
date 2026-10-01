package main

import (
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// storyKind says what a line of the Adventure tab is about, so we can colour it.
type storyKind int

const (
	storyPlain storyKind = iota
	storyDeath
	storyQuest
	storyCombat
	storyEnding
	storyTeam
	storyError
)

type storyEntry struct {
	kind storyKind
	time string
	text string
}

// Custom theme colour names. RichText styles pick colours by name,
// and retroTheme.Color turns each name into a real colour.
const (
	colorStoryDeath  fyne.ThemeColorName = "storyDeath"
	colorStoryQuest  fyne.ThemeColorName = "storyQuest"
	colorStoryCombat fyne.ThemeColorName = "storyCombat"
	colorStoryEnding fyne.ThemeColorName = "storyEnding"
	colorStoryTeam   fyne.ThemeColorName = "storyTeam"
)

func storyColorName(kind storyKind) fyne.ThemeColorName {
	switch kind {
	case storyDeath, storyError:
		return colorStoryDeath
	case storyQuest:
		return colorStoryQuest
	case storyCombat:
		return colorStoryCombat
	case storyEnding:
		return colorStoryEnding
	case storyTeam:
		return colorStoryTeam
	}
	return "foreground"
}

// storyColor returns the colour for a custom story colour name (and false for any other name).
func storyColor(name fyne.ThemeColorName) (color.Color, bool) {
	switch name {
	case colorStoryDeath:
		return color.NRGBA{R: 232, G: 96, B: 96, A: 255}, true
	case colorStoryQuest:
		return gold, true
	case colorStoryCombat:
		return color.NRGBA{R: 236, G: 154, B: 84, A: 255}, true
	case colorStoryEnding:
		return color.NRGBA{R: 190, G: 140, B: 240, A: 255}, true
	case colorStoryTeam:
		return color.NRGBA{R: 110, G: 200, B: 140, A: 255}, true
	}
	return nil, false
}

// storyKindOf maps the kind word of "EVT PLAYER <kind> <text>" to a storyKind.
func storyKindOf(word string) storyKind {
	switch word {
	case "DEATH":
		return storyDeath
	case "QUEST":
		return storyQuest
	case "ENDING":
		return storyEnding
	case "TEAM":
		return storyTeam
	}
	return storyPlain
}

func newStoryText() *widget.RichText {
	text := widget.NewRichText()
	text.Wrapping = fyne.TextWrapWord
	return text
}

// storySegments turns entries into coloured RichText paragraphs.
func storySegments(entries []storyEntry) []widget.RichTextSegment {
	segments := make([]widget.RichTextSegment, 0, len(entries))
	for _, entry := range entries {
		segments = append(segments, &widget.TextSegment{
			Text: entry.time + "  " + entry.text,
			Style: widget.RichTextStyle{
				ColorName: storyColorName(entry.kind),
				Inline:    false,
			},
		})
	}
	return segments
}

func (ui *gui) addStory(line string) { ui.addStoryKind(storyPlain, line) }

func (ui *gui) addStoryKind(kind storyKind, line string) {
	ui.storyLines = append(ui.storyLines, storyEntry{kind: kind, time: time.Now().Format("15:04:05"), text: strings.TrimSpace(line)})
	if len(ui.storyLines) > 300 {
		ui.storyLines = ui.storyLines[len(ui.storyLines)-300:]
	}
	ui.storyText.Segments = storySegments(ui.storyLines)
	ui.storyText.Refresh()
	ui.storyScroll.ScrollToBottom()
}
