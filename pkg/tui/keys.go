package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI.
type KeyMap struct {
	Quit       key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Scan       key.Binding
	Disconnect key.Binding
	Forget     key.Binding
	SetDefault key.Binding
	Mute       key.Binding
	Escape     key.Binding
	Help       key.Binding
	Refresh    key.Binding
}

// DefaultKeyMap returns the standard keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		NextTab:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
		PrevTab:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev tab")),
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Scan:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scan")),
		Disconnect: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "disconnect")),
		Forget:     key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "forget")),
		SetDefault: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "set default")),
		Mute:       key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mute")),
		Escape:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close/back")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}
}
