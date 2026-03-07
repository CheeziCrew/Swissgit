package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines shared keybindings across screens.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Toggle key.Binding
	All    key.Binding
	Tab    key.Binding
}

// DefaultKeyMap returns the default keybindings.
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	All: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "toggle all"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
}
