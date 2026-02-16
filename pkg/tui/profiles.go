package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/audio"
)

// ProfilesPane shows audio cards and allows profile switching,
// matching the spec Screen 3 layout with card-grouped profiles.
type ProfilesPane struct {
	cards  []audio.ShortRecord
	cursor int
	width  int
	height int
	au     audio.Service
	keys   KeyMap
}

func NewProfilesPane(au audio.Service, keys KeyMap) ProfilesPane {
	return ProfilesPane{au: au, keys: keys}
}

func (m ProfilesPane) Init() tea.Cmd {
	return loadProfilesCmd(m.au)
}

func (m ProfilesPane) Update(msg tea.Msg) (ProfilesPane, tea.Cmd) {
	switch msg := msg.(type) {
	case ProfilesLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: msg.Err} }
		}
		m.cards = msg.Cards
		if m.cursor >= len(m.cards) {
			m.cursor = max(0, len(m.cards)-1)
		}

	case SetProfileResultMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("set profile %s/%s: %w", msg.Card, msg.Profile, msg.Err)}
			}
		}
		return m, tea.Batch(
			loadProfilesCmd(m.au),
			func() tea.Msg {
				return StatusMsg{Text: fmt.Sprintf("Profile set: %s → %s", msg.Card, msg.Profile)}
			},
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m ProfilesPane) handleKey(msg tea.KeyMsg) (ProfilesPane, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.cards)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Refresh):
		return m, loadProfilesCmd(m.au)
	}
	return m, nil
}

func (m ProfilesPane) selected() (audio.ShortRecord, bool) {
	if m.cursor >= 0 && m.cursor < len(m.cards) {
		return m.cards[m.cursor], true
	}
	return audio.ShortRecord{}, false
}

func (m ProfilesPane) SetSize(w, h int) ProfilesPane {
	m.width = w
	m.height = h
	return m
}

func (m ProfilesPane) View() string {
	innerW := m.width - 6
	if innerW < 30 {
		innerW = 50
	}

	if len(m.cards) == 0 {
		emptyBox := sectionBox.Width(innerW).Render(
			sectionTitle("Audio Cards") + "\n" +
				dimStyle.Render("  No cards found."),
		)
		return emptyBox
	}

	// Render each card as its own box section.
	var boxes []string
	for i, card := range m.cards {
		isCursor := i == m.cursor

		// Card name becomes the section title
		cardTitle := sectionTitle(friendlyCardName(card.Name))

		// Cursor + bullet
		cur := "  "
		if isCursor {
			cur = cursorStyle.Render("▸ ")
		}
		bullet := disconnectedStyle.Render("○")
		if isCursor {
			bullet = connectedStyle.Render("●")
		}

		nameStr := nameNormalStyle.Render(card.Name)
		if isCursor {
			nameStr = nameHighlightStyle.Render(card.Name)
		}

		driverStr := ""
		if card.Driver != "" {
			driverStr = dimStyle.Render("  " + card.Driver)
		}

		content := fmt.Sprintf("%s%s %s%s", cur, bullet, nameStr, driverStr)

		box := sectionBox.Width(innerW).Render(
			cardTitle + "\n" + content,
		)
		boxes = append(boxes, box)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(boxes, "\n"),
	)
}

// friendlyCardName derives a human-friendly name from a pactl card name.
func friendlyCardName(name string) string {
	// Strip common prefixes like "alsa_card."
	name = strings.TrimPrefix(name, "alsa_card.")
	// Replace underscores/dots with spaces for readability
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, ".", " ")
	if len(name) > 40 {
		name = name[:37] + "..."
	}
	return name
}

func (m ProfilesPane) ShortHelp() string {
	return "enter apply  ↑↓ navigate  r refresh"
}
