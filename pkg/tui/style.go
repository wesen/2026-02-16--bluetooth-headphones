package tui

import "github.com/charmbracelet/lipgloss"

// ── Colour palette ──────────────────────────────────────────────────────────

var (
	colorPrimary   = lipgloss.Color("#7C3AED") // violet
	colorAccent    = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#22C55E") // green
	colorWarning   = lipgloss.Color("#EAB308") // yellow
	colorDanger    = lipgloss.Color("#EF4444") // red
	colorMuted     = lipgloss.Color("#6B7280") // gray-500
	colorDim       = lipgloss.Color("#9CA3AF") // gray-400
	colorBright    = lipgloss.Color("#F9FAFB") // gray-50
	colorSurface   = lipgloss.Color("#1F2937") // gray-800
	colorHighlight = lipgloss.Color("#374151") // gray-700
	colorScanner   = lipgloss.Color("#EC4899") // pink
)

// ── Shared styles ───────────────────────────────────────────────────────────

var (
	// Outer window frame
	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	// Section boxes inside panes
	sectionBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	// Section box for scanner overlay
	scannerBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorScanner).
			Padding(0, 1)

	// Title shown in the window border area
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			Background(colorPrimary).
			Padding(0, 1)

	// Section heading inside a box ("─ Bluetooth ─")
	sectionHeadingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	// Tab: active
	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright)

	// Tab: inactive
	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	// Cursor indicator
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	// Connected / active indicator
	connectedStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Disconnected / inactive indicator
	disconnectedStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	// Status label (e.g. "Connected", "Paired")
	statusLabelStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Width(12)

	// Volume bar: filled
	barFilledStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Volume bar: empty
	barEmptyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight)

	// Bottom help line
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Status bar message
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Error message in status bar
	errorBarStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	// Default indicator star
	defaultStarStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	// Dim info text
	dimStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Bold name in cursor row
	nameHighlightStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBright)

	// Normal name
	nameNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")) // gray-300

	// Action button
	buttonStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			Background(colorHighlight).
			Padding(0, 1)

	// Active action button (highlighted)
	buttonActiveStyle = lipgloss.NewStyle().
				Foreground(colorBright).
				Background(colorPrimary).
				Padding(0, 1).
				Bold(true)

	// Scanner title
	scannerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorScanner)
)

// ── Helpers ─────────────────────────────────────────────────────────────────

// volumeBar renders a ▓░ bar for a percentage at a given total width.
func volumeBar(percent, width int) string {
	if width < 4 {
		width = 20
	}
	filled := width * percent / 100
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := barFilledStyle.Render(repeatRune('▓', filled)) +
		barEmptyStyle.Render(repeatRune('░', empty))
	return bar
}

func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, n*3) // up to 3 bytes per rune
	s := string(r)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

// sectionTitle renders a styled heading like "─ Bluetooth ─".
func sectionTitle(label string) string {
	return sectionHeadingStyle.Render("─ " + label + " ─")
}
