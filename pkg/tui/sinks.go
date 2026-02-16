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
)

// SinksPane shows output sinks, input sources, matching Screen 2.
type SinksPane struct {
	sinks   []audio.ShortRecord
	sources []audio.ShortRecord
	section int // 0=sinks, 1=sources
	cursor  int
	width   int
	height  int
	au      audio.Service
	keys    KeyMap
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m SinksPane) handleKey(msg tea.KeyMsg) (SinksPane, tea.Cmd) {
	items := m.currentItems()
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		} else if m.section > 0 {
			m.section--
			items = m.currentItems()
			m.cursor = max(0, len(items)-1)
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(items)-1 {
			m.cursor++
		} else if m.section < 1 {
			m.section++
			m.cursor = 0
		}
	case key.Matches(msg, m.keys.SetDefault):
		if rec, ok := m.selected(); ok {
			if m.section == sinksSectionOutputs {
				return m, setDefaultSinkCmd(m.au, rec.Name)
			}
			return m, setDefaultSourceCmd(m.au, rec.Name)
		}
	case key.Matches(msg, m.keys.Mute):
		// future: toggle mute on selected
	case key.Matches(msg, m.keys.Refresh):
		return m, loadSinksCmd(m.au)
	}
	return m, nil
}

func (m SinksPane) currentItems() []audio.ShortRecord {
	if m.section == sinksSectionOutputs {
		return m.sinks
	}
	return m.sources
}

func (m SinksPane) selected() (audio.ShortRecord, bool) {
	items := m.currentItems()
	if m.cursor >= 0 && m.cursor < len(items) {
		return items[m.cursor], true
	}
	return audio.ShortRecord{}, false
}

func (m *SinksPane) clampCursor() {
	items := m.currentItems()
	if m.cursor >= len(items) {
		m.cursor = max(0, len(items)-1)
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

	// ── Output Sinks section ──
	outputContent := m.renderRecordList(m.sinks, sinksSectionOutputs, innerW-4)
	outputBox := sectionBox.Width(innerW).Render(
		sectionTitle("Output Sinks") + "\n" + outputContent,
	)

	// ── Input Sources section ──
	inputContent := m.renderRecordList(m.sources, sinksSectionInputs, innerW-4)
	inputBox := sectionBox.Width(innerW).Render(
		sectionTitle("Input Sources") + "\n" + inputContent,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		outputBox,
		"",
		inputBox,
	)
}

func (m SinksPane) renderRecordList(items []audio.ShortRecord, section, rowW int) string {
	if len(items) == 0 {
		return dimStyle.Render("  (none)")
	}
	var rows []string
	for i, item := range items {
		isCursor := m.section == section && i == m.cursor
		rows = append(rows, m.renderRecordRow(i, item, section, isCursor, rowW))
	}
	return strings.Join(rows, "\n")
}

func (m SinksPane) renderRecordRow(idx int, rec audio.ShortRecord, section int, isCursor bool, rowW int) string {
	// Cursor
	cur := "  "
	if isCursor {
		cur = cursorStyle.Render("▸ ")
	}

	// Default star (first item is default for now; will improve with actual default detection)
	star := "  "
	if idx == 0 {
		star = defaultStarStyle.Render("★ ")
	}

	// Name
	nameStr := nameNormalStyle.Render(rec.Name)
	if isCursor {
		nameStr = nameHighlightStyle.Render(rec.Name)
	}

	// State badge
	state := ""
	if rec.State != "" {
		stateColor := colorDim
		if rec.State == "RUNNING" {
			stateColor = colorSuccess
		} else if rec.State == "SUSPENDED" {
			stateColor = colorWarning
		}
		state = lipgloss.NewStyle().Foreground(stateColor).Render("  " + rec.State)
	}

	// Default badge
	defaultBadge := ""
	if idx == 0 {
		defaultBadge = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true).
			Render("  [default]")
	}

	return fmt.Sprintf("%s%s%s%s%s", cur, star, nameStr, state, defaultBadge)
}

func (m SinksPane) ShortHelp() string {
	return "d set-default  m mute  ↑↓ navigate  r refresh"
}
