package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Search        key.Binding
	Sort          key.Binding
	Filter        key.Binding
	Layout        key.Binding
	Add           key.Binding
	Edit          key.Binding
	Delete        key.Binding
	First         key.Binding
	Last          key.Binding
	Quit          key.Binding
	Save          key.Binding
	Cancel        key.Binding
	NextField     key.Binding
	PrevField     key.Binding
	Confirm       key.Binding
	Deny          key.Binding
	Tasksets      key.Binding
	TagFilter     key.Binding
	RevealPath    key.Binding
	Archive       key.Binding
	ToggleArchive key.Binding
	ThemeCycle    key.Binding
	Select        key.Binding
	BulkDelete    key.Binding
	BulkArchive   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open project"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Layout: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "layout"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "x"),
			key.WithHelp("d/x", "delete"),
		),
		First: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "first"),
		),
		Last: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "last"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("tab", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y", "confirm"),
		),
		Deny: key.NewBinding(
			key.WithKeys("n", "esc", "q"),
			key.WithHelp("n", "cancel"),
		),
		Tasksets: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tasksets"),
		),
		TagFilter: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "tag filter"),
		),
		RevealPath: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "reveal"),
		),
		Archive: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "archive"),
		),
		ToggleArchive: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "show archived"),
		),
		ThemeCycle: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "theme"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		BulkDelete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "bulk delete"),
		),
		BulkArchive: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "bulk archive"),
		),
	}
}
