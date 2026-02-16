package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"soundctl/pkg/soundctl/audio"
)

// flatProfile is a flattened view: one row per profile across all cards.
type flatProfile struct {
	cardIndex     int
	cardName      string
	profName      string
	profDesc      string
	isActive      bool
	isAvailable   bool
	isFirstInCard bool // marks the start of a new card group
}

// ProfilesPane shows audio cards with their profiles in radio-button style
// (spec Screen 3). Cursor navigates across all profiles; enter applies.
type ProfilesPane struct {
	cards  []audio.Card
	flat   []flatProfile
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
		m.flat = m.flattenProfiles()
		if m.cursor >= len(m.flat) {
			m.cursor = max(0, len(m.flat)-1)
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
				return StatusMsg{Text: fmt.Sprintf("Profile applied: %s → %s", msg.Card, msg.Profile)}
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
		if m.cursor < len(m.flat)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.cursor >= 0 && m.cursor < len(m.flat) {
			fp := m.flat[m.cursor]
			if !fp.isActive {
				return m, setProfileCmd(m.au, fp.cardName, fp.profName)
			}
		}
	case key.Matches(msg, m.keys.Refresh):
		return m, loadProfilesCmd(m.au)
	}
	return m, nil
}

func (m ProfilesPane) flattenProfiles() []flatProfile {
	var out []flatProfile
	for _, card := range m.cards {
		first := true
		for _, prof := range card.Profiles {
			out = append(out, flatProfile{
				cardIndex:     card.Index,
				cardName:      card.Name,
				profName:      prof.Name,
				profDesc:      prof.Description,
				isActive:      prof.Name == card.ActiveProfile,
				isAvailable:   prof.Available,
				isFirstInCard: first,
			})
			first = false
		}
	}
	return out
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

	if len(m.flat) == 0 {
		emptyBox := sectionBox.Width(innerW).Render(
			sectionTitle("Audio Cards") + "\n" +
				dimStyle.Render("  No cards found."),
		)
		return emptyBox
	}

	// Group profiles by card, render each card as a bordered section.
	type cardGroup struct {
		name string
		rows []string
	}
	groups := make(map[int]*cardGroup)
	var order []int

	for i, fp := range m.flat {
		if fp.isFirstInCard {
			groups[fp.cardIndex] = &cardGroup{name: fp.cardName}
			order = append(order, fp.cardIndex)
		}
		g := groups[fp.cardIndex]
		g.rows = append(g.rows, m.renderProfileRow(i, fp))
	}

	var boxes []string
	for _, idx := range order {
		g := groups[idx]
		title := sectionTitle(friendlyCardName(g.name))
		content := strings.Join(g.rows, "\n")
		box := sectionBox.Width(innerW).Render(title + "\n" + content)
		boxes = append(boxes, box)
	}

	return lipgloss.JoinVertical(lipgloss.Left, boxes...)
}

// friendlyCardName derives a human-friendly name from a pactl card name.
func friendlyCardName(name string) string {
	name = strings.TrimPrefix(name, "alsa_card.")
	name = strings.TrimPrefix(name, "bluez_card.")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, ".", " ")
	if len(name) > 40 {
		name = name[:37] + "..."
	}
	return name
}

func (m ProfilesPane) renderProfileRow(flatIdx int, fp flatProfile) string {
	isCursor := flatIdx == m.cursor

	cur := "  "
	if isCursor {
		cur = cursorStyle.Render("▸ ")
	}

	// Radio bullet: ● active, ○ inactive
	bullet := disconnectedStyle.Render("○")
	if fp.isActive {
		bullet = connectedStyle.Render("●")
	}

	// Description (or fallback to name)
	desc := fp.profDesc
	if desc == "" {
		desc = fp.profName
	}

	descStr := nameNormalStyle.Render(desc)
	if isCursor {
		descStr = nameHighlightStyle.Render(desc)
	}

	// Availability badge
	avail := ""
	if !fp.isAvailable {
		avail = "  " + dimStyle.Render("(unavailable)")
	}

	return fmt.Sprintf("%s%s %s%s", cur, bullet, descStr, avail)
}

func (m ProfilesPane) ShortHelp() string {
	return "enter apply  ↑↓ navigate  r refresh"
}
