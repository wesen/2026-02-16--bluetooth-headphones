package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/bluetooth"
)

// ScanOverlay is the scanning overlay that appears alongside the devices pane,
// matching the spec Screen 4 layout.
type ScanOverlay struct {
	visible    bool
	scanning   bool
	discovered []bluetooth.DiscoveredDevice
	cursor     int
	spinner    spinner.Model
	width      int
	height     int
	bt         bluetooth.Service
	keys       KeyMap
}

func NewScanOverlay(bt bluetooth.Service, keys KeyMap) ScanOverlay {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorScanner)
	return ScanOverlay{bt: bt, keys: keys, spinner: sp}
}

func (m ScanOverlay) Init() tea.Cmd {
	return nil
}

func (m ScanOverlay) Update(msg tea.Msg) (ScanOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case OpenScannerMsg:
		m.visible = true
		m.scanning = true
		m.discovered = nil
		m.cursor = 0
		return m, tea.Batch(m.spinner.Tick, discoverCmd(m.bt, 8))

	case CloseScannerMsg:
		m.visible = false
		m.scanning = false
		m.discovered = nil
		m.cursor = 0

	case DiscoveredDevicesMsg:
		m.scanning = false
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		m.discovered = msg.Devices
		m.cursor = 0

	case PairResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("pair %s: %w", msg.Addr, msg.Err)}
			}
		}
		m.visible = false
		m.scanning = false
		return m, tea.Batch(
			loadDevicesCmd(m.bt),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Paired + connected %s", msg.Addr)} },
		)

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.visible {
			return m.handleKey(msg)
		}
	}
	return m, nil
}

func (m ScanOverlay) handleKey(msg tea.KeyMsg) (ScanOverlay, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		return m, func() tea.Msg { return CloseScannerMsg{} }
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.discovered)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if !m.scanning && m.cursor >= 0 && m.cursor < len(m.discovered) {
			d := m.discovered[m.cursor]
			return m, pairCmd(m.bt, d.Address)
		}
	case key.Matches(msg, m.keys.Scan):
		if !m.scanning {
			m.scanning = true
			m.discovered = nil
			return m, tea.Batch(m.spinner.Tick, discoverCmd(m.bt, 8))
		}
	}
	return m, nil
}

func (m ScanOverlay) SetSize(w, h int) ScanOverlay {
	m.width = w
	m.height = h
	return m
}

func (m ScanOverlay) View() string {
	if !m.visible {
		return ""
	}

	var rows []string

	// Title with spinner
	if m.scanning {
		rows = append(rows, scannerTitleStyle.Render(
			fmt.Sprintf("Scanning... %s", m.spinner.View()),
		))
	} else {
		rows = append(rows, scannerTitleStyle.Render("Scan Complete"))
	}
	rows = append(rows, "")

	if len(m.discovered) == 0 && !m.scanning {
		rows = append(rows, dimStyle.Render("  No new devices found."))
	}

	for i, d := range m.discovered {
		cur := "  "
		if i == m.cursor {
			cur = cursorStyle.Render("▸ ")
		}

		name := d.Name
		if name == "" || name == d.Address {
			// Show truncated address for unknown devices
			name = fmt.Sprintf("Unknown (%s…)", d.Address[:8])
		}

		nameStr := nameNormalStyle.Render(name)
		if i == m.cursor {
			nameStr = nameHighlightStyle.Render(name)
		}
		rows = append(rows, fmt.Sprintf("%s%s", cur, nameStr))
	}

	rows = append(rows, "")
	rows = append(rows, helpStyle.Render("  enter pair"))
	rows = append(rows, helpStyle.Render("  s     rescan"))
	rows = append(rows, helpStyle.Render("  esc   cancel"))

	content := strings.Join(rows, "\n")

	overlayW := m.width
	if overlayW < 24 {
		overlayW = 28
	}

	return scannerBox.Width(overlayW).Render(
		sectionTitle("Scanning") + "\n" + content,
	)
}
