package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/preset"
)

// ── Messages ────────────────────────────────────────────────────────────────

// PresetsLoadedMsg carries refreshed preset list.
type PresetsLoadedMsg struct {
	Presets []preset.Preset
	Err     error
}

// ApplyPresetResultMsg reports apply outcome.
type ApplyPresetResultMsg struct {
	Name   string
	Result preset.ApplyResult
}

// DeletePresetResultMsg reports delete outcome.
type DeletePresetResultMsg struct {
	Name string
	Err  error
}

// SavePresetResultMsg reports save/snapshot outcome.
type SavePresetResultMsg struct {
	Name string
	Err  error
}

// OpenConfirmMsg opens the apply confirmation overlay.
type OpenConfirmMsg struct {
	Preset preset.Preset
	Diffs  []preset.DiffLine
}

// CloseConfirmMsg closes the confirmation overlay.
type CloseConfirmMsg struct{}

// ── Commands ────────────────────────────────────────────────────────────────

func loadPresetsCmd(store *preset.Store) tea.Cmd {
	return func() tea.Msg {
		presets, err := store.List()
		return PresetsLoadedMsg{Presets: presets, Err: err}
	}
}

func applyPresetCmd(au audio.Service, p preset.Preset) tea.Cmd {
	return func() tea.Msg {
		result := preset.Apply(context.Background(), au, p)
		return ApplyPresetResultMsg{Name: p.Name, Result: result}
	}
}

func deletePresetCmd(store *preset.Store, name string) tea.Cmd {
	return func() tea.Msg {
		err := store.Delete(name)
		return DeletePresetResultMsg{Name: name, Err: err}
	}
}

func snapshotPresetCmd(store *preset.Store, au audio.Service, name string) tea.Cmd {
	return func() tea.Msg {
		p, err := preset.SnapshotCurrent(context.Background(), au)
		if err != nil {
			return SavePresetResultMsg{Name: name, Err: err}
		}
		p.Name = name
		err = store.Save(p)
		return SavePresetResultMsg{Name: name, Err: err}
	}
}

// ── PresetsPane ─────────────────────────────────────────────────────────────

// PresetsPane shows saved presets with apply/delete actions (Screen 5).
type PresetsPane struct {
	presets      []preset.Preset
	activePreset string // name of currently applied preset
	cursor       int
	width        int
	height       int
	store        *preset.Store
	au           audio.Service
	keys         KeyMap

	// Confirmation overlay (Screen 7)
	confirmVisible bool
	confirmPreset  preset.Preset
	confirmDiffs   []preset.DiffLine
	confirmCursor  int // 0=apply, 1=cancel
}

func NewPresetsPane(store *preset.Store, au audio.Service, keys KeyMap) PresetsPane {
	return PresetsPane{store: store, au: au, keys: keys}
}

func (m PresetsPane) Init() tea.Cmd {
	return loadPresetsCmd(m.store)
}

func (m PresetsPane) Update(msg tea.Msg) (PresetsPane, tea.Cmd) {
	switch msg := msg.(type) {
	case PresetsLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		m.presets = msg.Presets
		if m.cursor >= len(m.presets) {
			m.cursor = max(0, len(m.presets)-1)
		}

	case ApplyPresetResultMsg:
		m.activePreset = msg.Name
		if len(msg.Result.Errors) > 0 {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("preset %q applied with %d error(s)", msg.Name, len(msg.Result.Errors))}
			}
		}
		m.confirmVisible = false
		return m, tea.Batch(
			loadPresetsCmd(m.store),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("✓ Preset %q applied", msg.Name)} },
		)

	case DeletePresetResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		return m, tea.Batch(
			loadPresetsCmd(m.store),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Deleted preset %q", msg.Name)} },
		)

	case SavePresetResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		return m, tea.Batch(
			loadPresetsCmd(m.store),
			func() tea.Msg { return StatusMsg{Text: fmt.Sprintf("Saved preset %q", msg.Name)} },
		)

	case OpenConfirmMsg:
		m.confirmVisible = true
		m.confirmPreset = msg.Preset
		m.confirmDiffs = msg.Diffs
		m.confirmCursor = 0

	case CloseConfirmMsg:
		m.confirmVisible = false

	case tea.KeyMsg:
		if m.confirmVisible {
			return m.handleConfirmKey(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m PresetsPane) handleKey(msg tea.KeyMsg) (PresetsPane, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.presets)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if p, ok := m.selected(); ok {
			// Build diff against a "current state" pseudo-preset
			currentPseudo := preset.Preset{} // simplified — full diff requires snapshot
			diffs := preset.Diff(currentPseudo, p)
			return m, func() tea.Msg {
				return OpenConfirmMsg{Preset: p, Diffs: diffs}
			}
		}
	case key.Matches(msg, m.keys.Forget): // X = delete
		if p, ok := m.selected(); ok {
			return m, deletePresetCmd(m.store, p.Name)
		}
	case key.Matches(msg, m.keys.Scan): // s = snapshot
		return m, snapshotPresetCmd(m.store, m.au, fmt.Sprintf("Snapshot %s", timeLabel()))
	case key.Matches(msg, m.keys.Refresh):
		return m, loadPresetsCmd(m.store)
	}
	return m, nil
}

func (m PresetsPane) handleConfirmKey(msg tea.KeyMsg) (PresetsPane, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.confirmVisible = false
	case key.Matches(msg, m.keys.Enter):
		if m.confirmCursor == 0 {
			// Apply
			return m, applyPresetCmd(m.au, m.confirmPreset)
		}
		m.confirmVisible = false
	case msg.String() == "left", msg.String() == "h":
		m.confirmCursor = 0
	case msg.String() == "right", msg.String() == "l":
		m.confirmCursor = 1
	}
	return m, nil
}

func (m PresetsPane) selected() (preset.Preset, bool) {
	if m.cursor >= 0 && m.cursor < len(m.presets) {
		return m.presets[m.cursor], true
	}
	return preset.Preset{}, false
}

func (m PresetsPane) SetSize(w, h int) PresetsPane {
	m.width = w
	m.height = h
	return m
}

// ── View ────────────────────────────────────────────────────────────────────

func (m PresetsPane) View() string {
	innerW := m.width - 6
	if innerW < 30 {
		innerW = 50
	}

	// Preset list
	var rows []string
	if len(m.presets) == 0 {
		rows = append(rows, dimStyle.Render("  No presets saved. Press S to snapshot current state."))
	}
	for i, p := range m.presets {
		rows = append(rows, m.renderPresetRow(i, p))
	}

	listContent := strings.Join(rows, "\n")
	listBox := sectionBox.Width(innerW).Render(
		sectionTitle("Saved Presets") + "\n" + listContent,
	)

	// Overlay
	if m.confirmVisible {
		overlay := m.renderConfirm(innerW - 8)
		return listBox + "\n\n" + overlay
	}

	return listBox
}

func (m PresetsPane) renderPresetRow(idx int, p preset.Preset) string {
	isCursor := idx == m.cursor
	isActive := p.Name == m.activePreset

	cur := "  "
	if isCursor {
		cur = cursorStyle.Render("▸ ")
	}

	star := "  "
	if isActive {
		star = defaultStarStyle.Render("★ ")
	}

	nameStr := nameNormalStyle.Render(p.Name)
	if isCursor {
		nameStr = nameHighlightStyle.Render(p.Name)
	}

	badge := ""
	if isActive {
		badge = "  " + lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("[active]")
	}

	// Summary line
	summary := m.presetSummary(p)
	summaryStr := dimStyle.Render("    " + summary)

	return fmt.Sprintf("%s%s%s%s\n%s", cur, star, nameStr, badge, summaryStr)
}

func (m PresetsPane) presetSummary(p preset.Preset) string {
	var parts []string
	for card, prof := range p.CardProfiles {
		short := card
		if idx := strings.LastIndex(card, "."); idx >= 0 {
			short = card[idx+1:]
		}
		parts = append(parts, short+"→"+prof)
	}
	if p.DefaultSink != "" {
		parts = append(parts, "Sink:"+friendlyName(p.DefaultSink))
	}
	for ch, vol := range p.Volumes {
		label := ch
		if idx := strings.LastIndex(ch, "."); idx >= 0 {
			label = ch[idx+1:]
		}
		s := fmt.Sprintf("%s:%d%%", label, vol.Level)
		if vol.Muted {
			s += "🔇"
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return strings.Join(parts, "  ")
}

func (m PresetsPane) renderConfirm(w int) string {
	title := scannerTitleStyle.Render(fmt.Sprintf("Apply %q?", m.confirmPreset.Name))

	var rows []string
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(colorBright).Render("  Changes:"))

	if len(m.confirmDiffs) == 0 {
		rows = append(rows, dimStyle.Render("    (no detected changes)"))
	}
	for _, d := range m.confirmDiffs {
		rows = append(rows, dimStyle.Render(fmt.Sprintf("    %s  %s → %s", d.Field, d.From, d.To)))
	}
	rows = append(rows, "")

	// Buttons
	applyBtn := buttonStyle.Render("Apply")
	cancelBtn := buttonStyle.Render("Cancel")
	if m.confirmCursor == 0 {
		applyBtn = buttonActiveStyle.Render("Apply")
	} else {
		cancelBtn = buttonActiveStyle.Render("Cancel")
	}
	gap := "    "
	rows = append(rows, fmt.Sprintf("      %s%s%s", applyBtn, gap, cancelBtn))
	rows = append(rows, "")

	content := strings.Join(rows, "\n")
	box := scannerBox.Width(w).Render(title + "\n" + content)
	return box
}

func (m PresetsPane) ShortHelp() string {
	return "enter apply  S snapshot  X delete  ↑↓ navigate  r refresh"
}

func timeLabel() string {
	return time.Now().Format("2006-01-02 15:04")
}
