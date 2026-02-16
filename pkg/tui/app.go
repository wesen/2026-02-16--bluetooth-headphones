package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
)

var tabNames = []string{"Devices", "Sinks", "Profiles"}

// AppModel is the root Bubble Tea model. It owns the tab bar, child panes,
// scanner overlay, and status bar — rendering the outer window chrome that
// matches the spec screenshots.
type AppModel struct {
	activeTab int
	width     int
	height    int
	ready     bool

	devices  DevicesPane
	sinks    SinksPane
	profiles ProfilesPane
	scanner  ScanOverlay

	statusText string
	isError    bool
	keys       KeyMap
}

// NewAppModel creates the root app with service dependencies.
func NewAppModel(bt bluetooth.Service, au audio.Service) AppModel {
	keys := DefaultKeyMap()
	return AppModel{
		devices:  NewDevicesPane(bt, keys),
		sinks:    NewSinksPane(au, keys),
		profiles: NewProfilesPane(au, keys),
		scanner:  NewScanOverlay(bt, keys),
		keys:     keys,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.devices.Init(),
		m.sinks.Init(),
		m.profiles.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m = m.resizePanes()
		return m, nil

	case StatusMsg:
		m.statusText = msg.Text
		m.isError = false
		return m, nil

	case ErrorMsg:
		m.statusText = fmt.Sprintf("%v", msg.Err)
		m.isError = true
		return m, nil

	case OpenScannerMsg:
		var cmd tea.Cmd
		m.scanner, cmd = m.scanner.Update(msg)
		return m, cmd

	case CloseScannerMsg:
		var cmd tea.Cmd
		m.scanner, cmd = m.scanner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		// Scanner overlay captures all keys when visible.
		if m.scanner.visible {
			var cmd tea.Cmd
			m.scanner, cmd = m.scanner.Update(msg)
			return m, cmd
		}

		// Global keys.
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextTab):
			m.activeTab = (m.activeTab + 1) % len(tabNames)
			return m, nil
		case key.Matches(msg, m.keys.PrevTab):
			m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
			return m, nil
		}

		// Delegate to active pane.
		switch m.activeTab {
		case TabDevices:
			var cmd tea.Cmd
			m.devices, cmd = m.devices.Update(msg)
			cmds = append(cmds, cmd)
		case TabSinks:
			var cmd tea.Cmd
			m.sinks, cmd = m.sinks.Update(msg)
			cmds = append(cmds, cmd)
		case TabProfiles:
			var cmd tea.Cmd
			m.profiles, cmd = m.profiles.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Non-key messages: route to scanner if relevant.
	switch msg.(type) {
	case DiscoveredDevicesMsg, PairResultMsg:
		var cmd tea.Cmd
		m.scanner, cmd = m.scanner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Data messages go to their owning pane.
	switch msg.(type) {
	case DevicesLoadedMsg, ConnectResultMsg, DisconnectResultMsg, ForgetResultMsg:
		var cmd tea.Cmd
		m.devices, cmd = m.devices.Update(msg)
		cmds = append(cmds, cmd)
	case SinksLoadedMsg, SetDefaultResultMsg:
		var cmd tea.Cmd
		m.sinks, cmd = m.sinks.Update(msg)
		cmds = append(cmds, cmd)
	case ProfilesLoadedMsg, SetProfileResultMsg:
		var cmd tea.Cmd
		m.profiles, cmd = m.profiles.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Spinner ticks → scanner (when it is actively scanning).
	if m.scanner.visible && m.scanner.scanning {
		var cmd tea.Cmd
		m.scanner, cmd = m.scanner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ── Layout ──────────────────────────────────────────────────────────────────

func (m AppModel) resizePanes() AppModel {
	contentW := m.width - 4  // window border + padding
	contentH := m.height - 7 // header + tabs + footer + border
	if contentW < 30 {
		contentW = 50
	}
	if contentH < 8 {
		contentH = 16
	}
	m.devices = m.devices.SetSize(contentW, contentH)
	m.sinks = m.sinks.SetSize(contentW, contentH)
	m.profiles = m.profiles.SetSize(contentW, contentH)
	m.scanner = m.scanner.SetSize(contentW/2-2, contentH)
	return m
}

// ── View ────────────────────────────────────────────────────────────────────

func (m AppModel) View() string {
	if !m.ready {
		return "\n  Initializing SoundCtl…\n"
	}

	contentW := m.width - 6 // inside window border + padding
	if contentW < 30 {
		contentW = 50
	}

	// Title + tabs
	header := titleStyle.Render(" SoundCtl ")
	tabs := m.renderTabs()

	// Active pane content
	paneContent := m.activePaneView()

	// If scanner overlay is visible, put it side-by-side (Screen 4 style)
	if m.scanner.visible {
		leftW := contentW/2 - 1
		rightW := contentW/2 - 1
		leftStyle := lipgloss.NewStyle().Width(leftW)
		paneContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render(paneContent),
			"  ",
			lipgloss.NewStyle().Width(rightW).Render(m.scanner.View()),
		)
	}

	// Status bar
	statusLine := ""
	if m.statusText != "" {
		if m.isError {
			statusLine = errorBarStyle.Render("✗ " + m.statusText)
		} else {
			statusLine = statusBarStyle.Render("● " + m.statusText)
		}
	}

	// Help footer
	helpLine := helpStyle.Render(m.buildHelp())

	// Assemble inner content
	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		tabs,
		"",
		paneContent,
		"",
		statusLine,
		helpLine,
	)

	// Wrap in window border
	return windowStyle.Width(m.width - 2).Render(inner)
}

func (m AppModel) renderTabs() string {
	var parts []string
	for i, name := range tabNames {
		if i == m.activeTab {
			parts = append(parts, tabActiveStyle.Render("▸ "+name))
		} else {
			parts = append(parts, tabInactiveStyle.Render("  "+name))
		}
		parts = append(parts, "    ") // gap
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m AppModel) activePaneView() string {
	switch m.activeTab {
	case TabDevices:
		return m.devices.View()
	case TabSinks:
		return m.sinks.View()
	case TabProfiles:
		return m.profiles.View()
	default:
		return ""
	}
}

func (m AppModel) buildHelp() string {
	parts := []string{
		"q quit",
		"tab/shift+tab switch",
	}
	switch m.activeTab {
	case TabDevices:
		parts = append(parts, "↑↓ navigate", "enter select", "s scan", "D disconnect", "X forget")
	case TabSinks:
		parts = append(parts, "↑↓ navigate", "d set-default", "m mute")
	case TabProfiles:
		parts = append(parts, "↑↓ navigate", "enter apply")
	}
	return strings.Join(parts, "  ")
}
