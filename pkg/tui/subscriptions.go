package tui

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ── PulseAudio subscription ────────────────────────────────────────────────

// PulseAudioEventMsg signals that something changed in PulseAudio.
type PulseAudioEventMsg struct {
	EventType string // "change", "new", "remove"
	Facility  string // "sink", "source", "card", "sink-input", etc.
}

// PulseAudioSubscription manages a `pactl subscribe` child process.
type PulseAudioSubscription struct {
	ctx    context.Context
	cancel context.CancelFunc
	events chan PulseAudioEventMsg
}

// NewPulseAudioSubscription starts listening for PulseAudio events.
func NewPulseAudioSubscription(parentCtx context.Context) *PulseAudioSubscription {
	ctx, cancel := context.WithCancel(parentCtx)
	sub := &PulseAudioSubscription{
		ctx:    ctx,
		cancel: cancel,
		events: make(chan PulseAudioEventMsg, 32),
	}
	go sub.run()
	return sub
}

func (s *PulseAudioSubscription) run() {
	defer close(s.events)

	cmd := exec.CommandContext(s.ctx, "pactl", "subscribe")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		// Lines look like: Event 'change' on sink #47
		//                   Event 'new' on sink-input #63
		ev := parsePactlSubscribeLine(line)
		if ev.EventType != "" {
			select {
			case s.events <- ev:
			case <-s.ctx.Done():
				return
			}
		}
	}
	// Wait for process to finish
	_ = cmd.Wait()
}

func parsePactlSubscribeLine(line string) PulseAudioEventMsg {
	// Format: Event 'change' on sink #47
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "Event '") {
		return PulseAudioEventMsg{}
	}
	// Extract event type
	rest := strings.TrimPrefix(line, "Event '")
	quoteEnd := strings.Index(rest, "'")
	if quoteEnd < 0 {
		return PulseAudioEventMsg{}
	}
	eventType := rest[:quoteEnd]
	rest = rest[quoteEnd+1:]

	// Extract facility: " on <facility> #<id>"
	rest = strings.TrimPrefix(rest, " on ")
	parts := strings.Fields(rest)
	facility := ""
	if len(parts) > 0 {
		facility = parts[0]
	}

	return PulseAudioEventMsg{
		EventType: eventType,
		Facility:  facility,
	}
}

// Stop terminates the subscription.
func (s *PulseAudioSubscription) Stop() {
	s.cancel()
}

// WaitCmd returns a tea.Cmd that waits for the next PulseAudio event.
func (s *PulseAudioSubscription) WaitCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case ev, ok := <-s.events:
			if !ok {
				return nil
			}
			return ev
		case <-s.ctx.Done():
			return nil
		}
	}
}

// ── Bluetooth subscription ─────────────────────────────────────────────────

// BluetoothEventMsg signals that something changed in bluetooth state.
type BluetoothEventMsg struct {
	EventType string // "property-changed", "device-added", "device-removed"
	Detail    string
}

// BluetoothSubscription monitors bluetooth events via dbus-monitor.
type BluetoothSubscription struct {
	ctx    context.Context
	cancel context.CancelFunc
	events chan BluetoothEventMsg
}

// NewBluetoothSubscription starts listening for BlueZ DBus events.
func NewBluetoothSubscription(parentCtx context.Context) *BluetoothSubscription {
	ctx, cancel := context.WithCancel(parentCtx)
	sub := &BluetoothSubscription{
		ctx:    ctx,
		cancel: cancel,
		events: make(chan BluetoothEventMsg, 32),
	}
	go sub.run()
	return sub
}

func (s *BluetoothSubscription) run() {
	defer close(s.events)

	cmd := exec.CommandContext(s.ctx, "dbus-monitor", "--system",
		"type='signal',sender='org.bluez'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		ev := parseDbusMonitorLine(line)
		if ev.EventType != "" {
			select {
			case s.events <- ev:
			case <-s.ctx.Done():
				return
			}
		}
	}
	_ = cmd.Wait()
}

func parseDbusMonitorLine(line string) BluetoothEventMsg {
	line = strings.TrimSpace(line)
	// Signal lines look like:
	//   signal time=... sender=:1.4 -> destination=(null) serial=1234
	//     path=/org/bluez/hci0/dev_AA_BB_CC_DD_EE_FF; interface=org.freedesktop.DBus.Properties; member=PropertiesChanged
	if strings.Contains(line, "member=PropertiesChanged") {
		return BluetoothEventMsg{EventType: "property-changed", Detail: line}
	}
	if strings.Contains(line, "member=InterfacesAdded") {
		return BluetoothEventMsg{EventType: "device-added", Detail: line}
	}
	if strings.Contains(line, "member=InterfacesRemoved") {
		return BluetoothEventMsg{EventType: "device-removed", Detail: line}
	}
	return BluetoothEventMsg{}
}

// Stop terminates the subscription.
func (s *BluetoothSubscription) Stop() {
	s.cancel()
}

// WaitCmd returns a tea.Cmd that waits for the next bluetooth event.
func (s *BluetoothSubscription) WaitCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case ev, ok := <-s.events:
			if !ok {
				return nil
			}
			return ev
		case <-s.ctx.Done():
			return nil
		}
	}
}

// ── Debounced refresh ──────────────────────────────────────────────────────

// RefreshTickMsg is sent after a debounce period to trigger data reload.
type RefreshTickMsg struct{}

// debounceRefreshCmd waits a short period then triggers a refresh.
// This prevents rapid-fire reloads when many events arrive at once.
func debounceRefreshCmd() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}
