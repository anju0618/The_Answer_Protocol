package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type commandEntry struct {
	widget.Entry
	onEscape func()
}

func newCommandEntry(onEscape func()) *commandEntry {
	entry := &commandEntry{onEscape: onEscape}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (entry *commandEntry) TypedKey(event *fyne.KeyEvent) {
	if event.Name == fyne.KeyEscape {
		entry.onEscape()
		return
	}
	entry.Entry.TypedKey(event)
}
