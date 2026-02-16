package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/audio"
)

const (
	sinksSectionOutputs = 0
	sinksSectionInputs  = 1
	sinksSectionRoutes  = 2
)

// SinksPane shows output sinks, input sources, and app routing (Screen 2).
type SinksPane struct {
	sinks             []audio.ShortRecord
	sources           []audio.ShortRecord
	sinkInputs        []audio.SinkInput
	defaultSinkName   string
	defaultSourceName string
	section           int // 0=outputs, 1=inputs, 2=routes
	cursor            int
	width             int
	height            int
	au                audio.Service
	keys              KeyMap
}

func NewSinksPane(au audio.Service, keys KeyMap) SinksPane {
	return SinksPane{au: au, keys: keys}
}

func (m SinksPane) Init() tea.Cmd {
	return loadSinksCmd(m.au)
}

func (m SinksPane) Update(msg tea.Msg) (SinksPane, tea.Cmd) {
	switch msg := msg.(type) {
	case SinksLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		m.sinks = msg.Sinks
		m.sources = msg.Sources
		m.sinkInputs = msg.SinkInputs
		m.defaultSinkName = msg.DefaultSinkName
		m.defaultSourceName = msg.DefaultSourceName
		m.clampCursor()

	case SetDefaultResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("set default %s %s: %w", msg.Kind, msg.Name, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadSinksCmd(m.au),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Set default %s: %s", msg.Kind, msg.Name)} },
		)

	case MoveStreamResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("move stream %d→%s: %w", msg.StreamID, msg.Sink, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadSinksCmd(m.au),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Moved stream %d → %s", msg.StreamID, msg.Sink)} },
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m SinksPane) handleKey(msg tea.KeyMsg) (SinksPane, tea.Cmd) {
	items := m.currentItemCount()
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		} else if m.section > 0 {
			m.section--
			m.cursor = max(0, m.currentItemCount()-1)
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < items-1 {
			m.cursor++
		} else if m.section < 2 {
			m.section++
			m.cursor = 0
		}
	case key.Matches(msg, m.keys.SetDefault):
		if m.section == sinksSectionOutputs {
			if rec, ok := m.selectedSink(); ok {
				return m, setDefaultSinkCmd(m.au, rec.Name)
			}
		} else if m.section == sinksSectionInputs {
			if rec, ok := m.selectedSource(); ok {
				return m, setDefaultSourceCmd(m.au, rec.Name)
			}
		}
	case key.Matches(msg, m.keys.Refresh):
		return m, loadSinksCmd(m.au)
	}
	return m, nil
}

func (m SinksPane) currentItemCount() int {
	switch m.section {
	case sinksSectionOutputs:
		return len(m.sinks)
	case sinksSectionInputs:
		return len(m.sources)
	case sinksSectionRoutes:
		return len(m.sinkInputs)
	}
	return 0
}

func (m SinksPane) selectedSink() (audio.ShortRecord, bool) {
	if m.section == sinksSectionOutputs && m.cursor >= 0 && m.cursor < len(m.sinks) {
		return m.sinks[m.cursor], true
	}
	return audio.ShortRecord{}, false
}

func (m SinksPane) selectedSource() (audio.ShortRecord, bool) {
	if m.section == sinksSectionInputs && m.cursor >= 0 && m.cursor < len(m.sources) {
		return m.sources[m.cursor], true
	}
	return audio.ShortRecord{}, false
}

func (m *SinksPane) clampCursor() {
	count := m.currentItemCount()
	if m.cursor >= count {
		m.cursor = max(0, count-1)
	}
}

func (m SinksPane) SetSize(w, h int) SinksPane {
	m.width = w
	m.height = h
	return m
}

func (m SinksPane) View() string {
	innerW := m.width - 6
	if innerW < 30 {
		innerW = 50
	}

	// ── Output Sinks ──
	outputBox := sectionBox.Width(innerW).Render(
		sectionTitle("Output Sinks") + "\n" + m.renderSinkList(m.sinks, sinksSectionOutputs, m.defaultSinkName),
	)

	// ── Input Sources ──
	inputBox := sectionBox.Width(innerW).Render(
		sectionTitle("Input Sources") + "\n" + m.renderSinkList(m.sources, sinksSectionInputs, m.defaultSourceName),
	)

	// ── App Routing ──
	routeBox := sectionBox.Width(innerW).Render(
		sectionTitle("App Routing") + "\n" + m.renderRoutes(),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		outputBox,
		"",
		inputBox,
		"",
		routeBox,
	)
}

func (m SinksPane) renderSinkList(items []audio.ShortRecord, section int, defaultName string) string {
	if len(items) == 0 {
		return dimStyle.Render("  (none)")
	}
	var rows []string
	for i, item := range items {
		isCursor := m.section == section && i == m.cursor
		isDefault := item.Name == defaultName

		cur := "  "
		if isCursor {
			cur = cursorStyle.Render("▸ ")
		}

		star := "  "
		if isDefault {
			star = defaultStarStyle.Render("★ ")
		}

		nameStr := nameNormalStyle.Render(item.Name)
		if isCursor {
			nameStr = nameHighlightStyle.Render(item.Name)
		}

		// Default badge
		badge := ""
		if isDefault {
			badge = lipgloss.NewStyle().
				Foreground(colorWarning).Bold(true).
				Render("  [default]")
		}

		// State
		state := ""
		if item.State != "" {
			stateColor := colorDim
			if item.State == "RUNNING" {
				stateColor = colorSuccess
			} else if item.State == "SUSPENDED" {
				stateColor = colorWarning
			}
			state = "  " + lipgloss.NewStyle().Foreground(stateColor).Render(item.State)
		}

		rows = append(rows, fmt.Sprintf("%s%s%s%s%s", cur, star, nameStr, badge, state))
	}
	return strings.Join(rows, "\n")
}

func (m SinksPane) renderRoutes() string {
	if len(m.sinkInputs) == 0 {
		return dimStyle.Render("  No active streams")
	}
	var rows []string
	for i, si := range m.sinkInputs {
		isCursor := m.section == sinksSectionRoutes && i == m.cursor

		cur := "  "
		if isCursor {
			cur = cursorStyle.Render("▸ ")
		}

		appName := si.AppName
		if appName == "" {
			appName = fmt.Sprintf("Stream #%d", si.Index)
		}

		sinkName := si.SinkName
		if sinkName == "" {
			sinkName = fmt.Sprintf("sink:%d", si.SinkIndex)
		}

		appStr := nameNormalStyle.Render(fmt.Sprintf("%-16s", appName))
		if isCursor {
			appStr = nameHighlightStyle.Render(fmt.Sprintf("%-16s", appName))
		}

		arrow := dimStyle.Render(" → ")
		sinkStr := lipgloss.NewStyle().Foreground(colorAccent).Render(friendlyName(sinkName))

		reroute := ""
		if isCursor {
			reroute = "  " + lipgloss.NewStyle().Foreground(colorScanner).Render("🔀 reroute")
		}

		rows = append(rows, fmt.Sprintf("%s%s%s%s%s", cur, appStr, arrow, sinkStr, reroute))
	}
	return strings.Join(rows, "\n")
}

// friendlyName strips common pactl name prefixes for display.
func friendlyName(name string) string {
	for _, prefix := range []string{"alsa_output.", "alsa_input.", "bluez_sink.", "bluez_source."} {
		if strings.HasPrefix(name, prefix) {
			return strings.TrimPrefix(name, prefix)
		}
	}
	return name
}

func (m SinksPane) ShortHelp() string {
	return "d set-default  r reroute  m mute  ↑↓ navigate"
}
