package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/bluetooth"
)

// DevicesPane shows bluetooth devices with connect/disconnect/forget actions
// and a volume section, matching the spec Screen 1 layout.
type DevicesPane struct {
	devices    []bluetooth.Device
	controller bluetooth.ControllerStatus
	cursor     int
	width      int
	height     int
	bt         bluetooth.Service
	keys       KeyMap
}

func NewDevicesPane(bt bluetooth.Service, keys KeyMap) DevicesPane {
	return DevicesPane{bt: bt, keys: keys}
}

func (m DevicesPane) Init() tea.Cmd {
	return loadDevicesCmd(m.bt)
}

func (m DevicesPane) Update(msg tea.Msg) (DevicesPane, tea.Cmd) {
	switch msg := msg.(type) {
	case DevicesLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		m.devices = msg.Devices
		m.controller = msg.Controller
		if m.cursor >= len(m.devices) {
			m.cursor = max(0, len(m.devices)-1)
		}

	case ConnectResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("connect %s: %w", msg.Addr, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadDevicesCmd(m.bt),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Connected %s", msg.Addr)} },
		)

	case DisconnectResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("disconnect %s: %w", msg.Addr, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadDevicesCmd(m.bt),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Disconnected %s", msg.Addr)} },
		)

	case ForgetResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("forget %s: %w", msg.Addr, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadDevicesCmd(m.bt),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Removed %s", msg.Addr)} },
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m DevicesPane) handleKey(msg tea.KeyMsg) (DevicesPane, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.devices)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if d, ok := m.selected(); ok {
			if d.Connected {
				return m, disconnectCmd(m.bt, d.Address)
			}
			return m, connectCmd(m.bt, d.Address)
		}
	case key.Matches(msg, m.keys.Disconnect):
		if d, ok := m.selected(); ok && d.Connected {
			return m, disconnectCmd(m.bt, d.Address)
		}
	case key.Matches(msg, m.keys.Forget):
		if d, ok := m.selected(); ok {
			return m, forgetCmd(m.bt, d.Address)
		}
	case key.Matches(msg, m.keys.Scan):
		return m, func() tea.Msg { return OpenScannerMsg{} }
	case key.Matches(msg, m.keys.Refresh):
		return m, loadDevicesCmd(m.bt)
	}
	return m, nil
}

func (m DevicesPane) selected() (bluetooth.Device, bool) {
	if m.cursor >= 0 && m.cursor < len(m.devices) {
		return m.devices[m.cursor], true
	}
	return bluetooth.Device{}, false
}

func (m DevicesPane) SetSize(w, h int) DevicesPane {
	m.width = w
	m.height = h
	return m
}

func (m DevicesPane) View() string {
	innerW := m.width - 6 // account for box padding/border
	if innerW < 30 {
		innerW = 50
	}

	// ── Bluetooth section ──
	var btRows []string
	if len(m.devices) == 0 {
		btRows = append(btRows, dimStyle.Render("  No devices found. Press s to scan."))
	}
	for i, d := range m.devices {
		btRows = append(btRows, m.renderDeviceRow(i, d, innerW-4))
	}
	btRows = append(btRows, "") // blank line before buttons
	btRows = append(btRows, m.renderButtons())

	btContent := strings.Join(btRows, "\n")
	btBox := sectionBox.Width(innerW).Render(
		sectionTitle("Bluetooth") + "\n" + btContent,
	)

	// ── Volume section ──
	volBarW := innerW - 18 // room for label + pct
	if volBarW < 10 {
		volBarW = 10
	}
	var volRows []string
	// Use a dummy volume for now – Phase 2.3 will add live subscriptions
	volRows = append(volRows, renderVolumeLine("Master", 72, volBarW))
	volRows = append(volRows, renderVolumeLine("Media", 90, volBarW))
	volRows = append(volRows, renderVolumeLine("Alerts", 35, volBarW))

	volContent := strings.Join(volRows, "\n")
	volBox := sectionBox.Width(innerW).Render(
		sectionTitle("Volume") + "\n" + volContent,
	)

	// ── Controller info line ──
	scanLabel := "off"
	if m.controller.Discovering {
		scanLabel = connectedStyle.Render("scanning")
	}
	controllerLine := dimStyle.Render(
		fmt.Sprintf("  Controller: %s  Scan: %s", m.controller.Alias, scanLabel),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		btBox,
		"",
		volBox,
		"",
		controllerLine,
	)
}

func (m DevicesPane) renderDeviceRow(idx int, d bluetooth.Device, rowW int) string {
	// Cursor
	cur := "  "
	if idx == m.cursor {
		cur = cursorStyle.Render("▸ ")
	}

	// Bullet icon
	icon := disconnectedStyle.Render("○")
	if d.Connected {
		icon = connectedStyle.Render("●")
	}

	// Name
	nameW := 28
	name := d.Name
	if len(name) > nameW {
		name = name[:nameW-1] + "…"
	}
	nameStr := nameNormalStyle.Render(fmt.Sprintf("%-*s", nameW, name))
	if idx == m.cursor {
		nameStr = nameHighlightStyle.Render(fmt.Sprintf("%-*s", nameW, name))
	}

	// Status
	status := statusLabelStyle.Render(capitalize(d.Connection))

	return fmt.Sprintf("%s%s %s %s", cur, icon, nameStr, status)
}

func (m DevicesPane) renderButtons() string {
	scanBtn := buttonStyle.Render("Scan")
	disconnBtn := buttonStyle.Render("Disconnect")
	forgetBtn := buttonStyle.Render("Forget")

	// Highlight the contextual button
	if d, ok := m.selected(); ok && d.Connected {
		disconnBtn = buttonActiveStyle.Render("Disconnect")
	}

	gap := lipgloss.NewStyle().Width(2).Render("")
	return "  " + scanBtn + gap + disconnBtn + gap + forgetBtn
}

func renderVolumeLine(label string, pct, barW int) string {
	labelStr := lipgloss.NewStyle().
		Width(10).
		Foreground(colorDim).
		Render(label)

	bar := volumeBar(pct, barW)
	pctStr := lipgloss.NewStyle().
		Width(5).
		Align(lipgloss.Right).
		Foreground(colorBright).
		Render(fmt.Sprintf("%d%%", pct))

	return fmt.Sprintf("  %s %s %s", labelStr, bar, pctStr)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (m DevicesPane) ShortHelp() string {
	return "enter connect  s scan  D disconnect  X forget  r refresh"
}
