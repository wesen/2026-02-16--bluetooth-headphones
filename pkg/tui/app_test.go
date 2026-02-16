package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
	"soundctl/pkg/soundctl/exec"
)

func newTestApp() (AppModel, *exec.FakeRunner) {
	runner := exec.NewFakeRunner()

	// Stub controller status
	runner.Set("bluetoothctl", []string{"show"}, exec.CommandResult{
		Output: "Controller AA:BB:CC:DD:EE:FF\n\tAlias: TestController\n\tPowered: yes\n\tPairable: yes\n\tDiscovering: no",
	})

	// Stub device list (empty)
	runner.Set("bluetoothctl", []string{"devices"}, exec.CommandResult{Output: ""})

	// Stub sinks
	runner.Set("pactl", []string{"list", "short", "sinks"}, exec.CommandResult{
		Output: "1\ttest-sink\tmodule-alsa-card.c\ts16le 2ch 48000Hz\tRUNNING",
	})

	// Stub sources
	runner.Set("pactl", []string{"list", "short", "sources"}, exec.CommandResult{
		Output: "2\ttest-source\tmodule-alsa-card.c\ts16le 2ch 48000Hz\tIDLE",
	})

	// Stub cards (short form for existing service)
	runner.Set("pactl", []string{"list", "short", "cards"}, exec.CommandResult{
		Output: "0\ttest-card\tmodule-alsa-card.c",
	})

	// Stub pactl info for defaults
	runner.Set("pactl", []string{"info"}, exec.CommandResult{
		Output: "Default Sink: test-sink\nDefault Source: test-source\nServer Name: PipeWire",
	})

	// Stub sink-inputs
	runner.Set("pactl", []string{"list", "sink-inputs"}, exec.CommandResult{
		Output: "",
	})

	// Stub detailed cards for TUI profiles
	runner.Set("pactl", []string{"list", "cards"}, exec.CommandResult{
		Output: "Card #0\n\tName: test-card\n\tDriver: module-alsa-card.c\n\tProfiles:\n\t\toutput:stereo: Stereo Output (sinks: 1, sources: 0, priority: 6500, available: yes)\n\t\toff: Off (sinks: 0, sources: 0, priority: 0, available: yes)\n\tActive Profile: output:stereo",
	})

	bt := bluetooth.NewExecService(runner)
	au := audio.NewExecService(runner)

	model := NewAppModel(bt, au)
	return model, runner
}

func TestAppInit(t *testing.T) {
	model, _ := newTestApp()
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init should return batch commands")
	}
}

func TestAppTabSwitch(t *testing.T) {
	model, _ := newTestApp()

	// Simulate window size
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	if model.activeTab != TabDevices {
		t.Fatalf("expected initial tab Devices, got %d", model.activeTab)
	}

	// Press tab → Sinks
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	if model.activeTab != TabSinks {
		t.Fatalf("expected tab Sinks after tab press, got %d", model.activeTab)
	}

	// Press tab → Profiles
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	if model.activeTab != TabProfiles {
		t.Fatalf("expected tab Profiles after second tab press, got %d", model.activeTab)
	}

	// Wrap around → Devices
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	if model.activeTab != TabDevices {
		t.Fatalf("expected tab Devices after wrap, got %d", model.activeTab)
	}
}

func TestAppShiftTabReverse(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Shift-tab from Devices → Profiles (wrap backward)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = m.(AppModel)
	if model.activeTab != TabProfiles {
		t.Fatalf("expected Profiles after shift-tab, got %d", model.activeTab)
	}
}

func TestAppStatusMessage(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(StatusMsg{Text: "hello"})
	model = m.(AppModel)
	if model.statusText != "hello" {
		t.Fatalf("expected status 'hello', got %q", model.statusText)
	}
	if model.isError {
		t.Fatal("expected isError=false for StatusMsg")
	}
}

func TestAppErrorMessage(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(ErrorMsg{Err: fmt.Errorf("test error")})
	model = m.(AppModel)
	if model.statusText != "test error" {
		t.Fatalf("expected 'test error' in status, got %q", model.statusText)
	}
	if !model.isError {
		t.Fatal("expected isError=true for ErrorMsg")
	}
}

func TestAppQuit(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestDevicesLoadedMsg(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(DevicesLoadedMsg{
		Devices: []bluetooth.Device{
			{Address: "AA:BB:CC:DD:EE:01", Name: "Sony WH-1000XM5", Connected: true, Connection: "connected"},
			{Address: "AA:BB:CC:DD:EE:02", Name: "AirPods Pro", Paired: true, Connection: "paired"},
		},
		Controller: bluetooth.ControllerStatus{Alias: "TestCtrl", Powered: true},
	})
	model = m.(AppModel)

	if len(model.devices.devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(model.devices.devices))
	}
	if model.devices.devices[0].Name != "Sony WH-1000XM5" {
		t.Fatalf("expected Sony WH-1000XM5, got %s", model.devices.devices[0].Name)
	}
}

func TestSinksLoadedMsg(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(SinksLoadedMsg{
		Sinks:             []audio.ShortRecord{{ID: 1, Name: "sink1", State: "RUNNING"}},
		Sources:           []audio.ShortRecord{{ID: 2, Name: "source1", State: "IDLE"}},
		DefaultSinkName:   "sink1",
		DefaultSourceName: "source1",
	})
	model = m.(AppModel)

	if len(model.sinks.sinks) != 1 {
		t.Fatalf("expected 1 sink, got %d", len(model.sinks.sinks))
	}
	if len(model.sinks.sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(model.sinks.sources))
	}
	if model.sinks.defaultSinkName != "sink1" {
		t.Fatalf("expected default sink 'sink1', got %q", model.sinks.defaultSinkName)
	}
}

func TestProfilesLoadedMsg(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(ProfilesLoadedMsg{
		Cards: []audio.Card{{
			Index:         0,
			Name:          "card1",
			Driver:        "alsa",
			ActiveProfile: "stereo",
			Profiles: []audio.CardProfile{
				{Name: "stereo", Description: "Stereo Output", Available: true},
				{Name: "off", Description: "Off", Available: true},
			},
		}},
	})
	model = m.(AppModel)

	if len(model.profiles.cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(model.profiles.cards))
	}
	if len(model.profiles.flat) != 2 {
		t.Fatalf("expected 2 flat profiles, got %d", len(model.profiles.flat))
	}
	if !model.profiles.flat[0].isActive {
		t.Fatal("expected first profile to be active")
	}
}

func TestScannerOverlayOpenClose(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Open scanner
	m, _ = model.Update(OpenScannerMsg{})
	model = m.(AppModel)
	if !model.scanner.visible {
		t.Fatal("expected scanner to be visible after OpenScannerMsg")
	}

	// Close scanner
	m, _ = model.Update(CloseScannerMsg{})
	model = m.(AppModel)
	if model.scanner.visible {
		t.Fatal("expected scanner to be hidden after CloseScannerMsg")
	}
}

func TestViewRendersDevicesTab(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Load devices
	m, _ = model.Update(DevicesLoadedMsg{
		Devices: []bluetooth.Device{
			{Address: "AA:BB:CC:DD:EE:01", Name: "Sony WH-1000XM5", Connected: true, Connection: "connected"},
			{Address: "AA:BB:CC:DD:EE:02", Name: "AirPods Pro", Paired: true, Connection: "paired"},
		},
		Controller: bluetooth.ControllerStatus{Alias: "TestCtrl", Powered: true},
	})
	model = m.(AppModel)

	view := model.View()
	// Window should contain title, tabs, device names, volume, and help
	for _, want := range []string{"SoundCtl", "Devices", "Sinks", "Profiles", "Sony WH-1000XM5", "AirPods Pro", "Bluetooth", "Volume", "quit"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestViewRendersSinksTab(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Switch to Sinks tab
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)

	// Load data
	m, _ = model.Update(SinksLoadedMsg{
		Sinks:             []audio.ShortRecord{{ID: 1, Name: "alsa-sink-stereo", State: "RUNNING"}},
		Sources:           []audio.ShortRecord{{ID: 2, Name: "alsa-source-mono", State: "IDLE"}},
		DefaultSinkName:   "alsa-sink-stereo",
		DefaultSourceName: "alsa-source-mono",
	})
	model = m.(AppModel)

	view := model.View()
	for _, want := range []string{"Output Sinks", "Input Sources", "App Routing", "alsa-sink-stereo", "alsa-source-mono", "default"} {
		if !strings.Contains(view, want) {
			t.Errorf("sinks view missing %q", want)
		}
	}
}

func TestViewRendersProfilesTab(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Switch to Profiles tab
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)

	// Load data with full card + profiles
	m, _ = model.Update(ProfilesLoadedMsg{
		Cards: []audio.Card{{
			Index:         0,
			Name:          "alsa_card.pci_0000",
			Driver:        "module-alsa-card.c",
			ActiveProfile: "output:stereo",
			Profiles: []audio.CardProfile{
				{Name: "output:stereo", Description: "Analog Stereo Output", Available: true},
				{Name: "off", Description: "Off", Available: true},
			},
		}},
	})
	model = m.(AppModel)

	view := model.View()
	for _, want := range []string{"pci 0000", "Analog Stereo Output", "Off"} {
		if !strings.Contains(view, want) {
			t.Errorf("profiles view missing %q", want)
		}
	}
}

func TestScannerOverlayViewWithDevices(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = m.(AppModel)

	// Open scanner
	m, _ = model.Update(OpenScannerMsg{})
	model = m.(AppModel)

	// Feed discovered devices
	m, _ = model.Update(DiscoveredDevicesMsg{
		Devices: []bluetooth.DiscoveredDevice{
			{Address: "11:22:33:44:55:66", Name: "JBL Charge 5"},
			{Address: "AA:BB:CC:DD:EE:FF", Name: "Bose QC45"},
		},
	})
	model = m.(AppModel)

	view := model.View()
	for _, want := range []string{"Scanning", "JBL Charge 5", "Bose QC45", "enter pair", "esc"} {
		if !strings.Contains(view, want) {
			t.Errorf("scanner view missing %q", want)
		}
	}
}

func TestSinksWithAppRouting(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Switch to Sinks
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)

	m, _ = model.Update(SinksLoadedMsg{
		Sinks:   []audio.ShortRecord{{ID: 1, Name: "bt-sink", State: "RUNNING"}},
		Sources: []audio.ShortRecord{{ID: 2, Name: "bt-source", State: "IDLE"}},
		SinkInputs: []audio.SinkInput{
			{Index: 57, SinkIndex: 1, AppName: "Firefox", SinkName: "bt-sink"},
			{Index: 63, SinkIndex: 1, AppName: "Spotify", SinkName: "bt-sink"},
		},
		DefaultSinkName:   "bt-sink",
		DefaultSourceName: "bt-source",
	})
	model = m.(AppModel)

	view := model.View()
	for _, want := range []string{"Firefox", "Spotify", "bt-sink", "App Routing"} {
		if !strings.Contains(view, want) {
			t.Errorf("sinks view with routing missing %q", want)
		}
	}
}

func TestProfilesApplyViaEnter(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Switch to Profiles
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)

	// Load profiles
	m, _ = model.Update(ProfilesLoadedMsg{
		Cards: []audio.Card{{
			Index:         0,
			Name:          "test-card",
			ActiveProfile: "stereo",
			Profiles: []audio.CardProfile{
				{Name: "stereo", Description: "Stereo", Available: true},
				{Name: "off", Description: "Off", Available: true},
			},
		}},
	})
	model = m.(AppModel)

	// Move cursor to "off" profile
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = m.(AppModel)

	if model.profiles.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", model.profiles.cursor)
	}

	// Press enter — should produce a setProfile command
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected setProfile command on enter for inactive profile")
	}
}

func TestDevicesCursorNavigation(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	m, _ = model.Update(DevicesLoadedMsg{
		Devices: []bluetooth.Device{
			{Address: "01", Name: "Dev1", Connection: "paired"},
			{Address: "02", Name: "Dev2", Connection: "saved"},
			{Address: "03", Name: "Dev3", Connection: "connected", Connected: true},
		},
		Controller: bluetooth.ControllerStatus{Alias: "Ctrl"},
	})
	model = m.(AppModel)

	if model.devices.cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", model.devices.cursor)
	}

	// Move down twice
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = m.(AppModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = m.(AppModel)
	if model.devices.cursor != 2 {
		t.Fatalf("expected cursor at 2, got %d", model.devices.cursor)
	}

	// Can't go past end
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = m.(AppModel)
	if model.devices.cursor != 2 {
		t.Fatalf("expected cursor clamped at 2, got %d", model.devices.cursor)
	}

	// Move up
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = m.(AppModel)
	if model.devices.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", model.devices.cursor)
	}
}
