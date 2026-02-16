package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
	"soundctl/pkg/soundctl/exec"
	"soundctl/pkg/soundctl/preset"
)

func newTestApp() (AppModel, *exec.FakeRunner) {
	dir, _ := os.MkdirTemp("", "soundctl-test-*")
	return newTestAppWithDir(dir)
}

func newTestAppWithDir(tmpDir string) (AppModel, *exec.FakeRunner) {
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
	store := preset.NewStore(filepath.Join(tmpDir, "presets.yaml"))

	model := NewAppModel(bt, au, store)
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

	// Press tab → Presets
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	if model.activeTab != TabPresets {
		t.Fatalf("expected tab Presets after third tab press, got %d", model.activeTab)
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

	// Shift-tab from Devices → Presets (wrap backward to last tab)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = m.(AppModel)
	if model.activeTab != TabPresets {
		t.Fatalf("expected Presets after shift-tab, got %d", model.activeTab)
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

func TestPulseAudioEventTriggersRefresh(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Simulate a PulseAudio event
	m, cmd := model.Update(PulseAudioEventMsg{EventType: "change", Facility: "sink"})
	model = m.(AppModel)

	if !model.refreshPending {
		t.Fatal("expected refreshPending=true after PA event")
	}
	if cmd == nil {
		t.Fatal("expected debounce command from PA event")
	}
}

func TestBluetoothEventTriggersRefresh(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Simulate a Bluetooth event
	m, cmd := model.Update(BluetoothEventMsg{EventType: "property-changed", Detail: "test"})
	model = m.(AppModel)

	if !model.refreshPending {
		t.Fatal("expected refreshPending=true after BT event")
	}
	if cmd == nil {
		t.Fatal("expected debounce command from BT event")
	}
}

func TestRefreshTickResetsAndReloads(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// Set pending
	model.refreshPending = true

	// Fire refresh tick
	m, cmd := model.Update(RefreshTickMsg{})
	model = m.(AppModel)

	if model.refreshPending {
		t.Fatal("expected refreshPending=false after RefreshTickMsg")
	}
	if cmd == nil {
		t.Fatal("expected reload commands from RefreshTickMsg")
	}
}

func TestDebouncePreventsDoubleRefresh(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(AppModel)

	// First event sets pending
	m, _ = model.Update(PulseAudioEventMsg{EventType: "change", Facility: "sink"})
	model = m.(AppModel)
	if !model.refreshPending {
		t.Fatal("expected pending after first event")
	}

	// Second event should not schedule another debounce
	m, _ = model.Update(PulseAudioEventMsg{EventType: "change", Facility: "source"})
	model = m.(AppModel)
	// Still pending — only one debounce timer should be active
	if !model.refreshPending {
		t.Fatal("expected pending still true")
	}
}

func TestIntegrationFullDataFlow(t *testing.T) {
	// Integration test: load data for all panes, navigate, verify views
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = m.(AppModel)

	// Load all data
	m, _ = model.Update(DevicesLoadedMsg{
		Devices: []bluetooth.Device{
			{Address: "01:02:03:04:05:06", Name: "Sony WH-1000XM5", Connected: true, Connection: "connected"},
			{Address: "07:08:09:0A:0B:0C", Name: "AirPods Pro", Paired: true, Connection: "paired"},
			{Address: "0D:0E:0F:10:11:12", Name: "JBL Flip 6", Connection: "saved"},
		},
		Controller: bluetooth.ControllerStatus{Alias: "hci0", Powered: true, Discovering: false},
	})
	model = m.(AppModel)

	m, _ = model.Update(SinksLoadedMsg{
		Sinks: []audio.ShortRecord{
			{ID: 47, Name: "bluez_sink.sony", State: "RUNNING"},
			{ID: 48, Name: "alsa_output.stereo", State: "SUSPENDED"},
		},
		Sources: []audio.ShortRecord{
			{ID: 49, Name: "bluez_source.sony", State: "RUNNING"},
		},
		SinkInputs: []audio.SinkInput{
			{Index: 57, SinkIndex: 47, AppName: "Firefox", SinkName: "bluez_sink.sony"},
			{Index: 63, SinkIndex: 47, AppName: "Spotify", SinkName: "bluez_sink.sony"},
		},
		DefaultSinkName:   "bluez_sink.sony",
		DefaultSourceName: "bluez_source.sony",
	})
	model = m.(AppModel)

	m, _ = model.Update(ProfilesLoadedMsg{
		Cards: []audio.Card{
			{
				Index: 62, Name: "bluez_card.sony", Driver: "bluez5",
				ActiveProfile: "a2dp-sink",
				Profiles: []audio.CardProfile{
					{Name: "a2dp-sink", Description: "High Fidelity Playback (A2DP Sink)", Available: true},
					{Name: "headset-head-unit", Description: "Headset Head Unit (HSP/HFP)", Available: true},
					{Name: "off", Description: "Off", Available: true},
				},
			},
			{
				Index: 47, Name: "alsa_card.pci-0000_00_1f", Driver: "alsa",
				ActiveProfile: "output:analog-stereo+input:analog-stereo",
				Profiles: []audio.CardProfile{
					{Name: "output:analog-stereo+input:analog-stereo", Description: "Analog Stereo Duplex", Available: true},
					{Name: "output:analog-stereo", Description: "Analog Stereo Output", Available: true},
					{Name: "off", Description: "Off", Available: true},
				},
			},
		},
	})
	model = m.(AppModel)

	// ── Verify Devices tab (Screen 1) ──
	view := model.View()
	for _, want := range []string{
		"SoundCtl", "Devices", "Sinks", "Profiles",
		"Sony WH-1000XM5", "AirPods Pro", "JBL Flip 6",
		"Bluetooth", "Volume",
		"●", "○",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("Devices tab missing %q", want)
		}
	}

	// ── Switch to Sinks tab (Screen 2) ──
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	view = model.View()
	for _, want := range []string{
		"Output Sinks", "Input Sources", "App Routing",
		"bluez_sink.sony", "[default]", "★",
		"Firefox", "Spotify", "→",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("Sinks tab missing %q", want)
		}
	}

	// ── Switch to Profiles tab (Screen 3) ──
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = m.(AppModel)
	view = model.View()
	for _, want := range []string{
		"High Fidelity Playback",
		"Headset Head Unit",
		"Analog Stereo Duplex",
		"Analog Stereo Output",
		"●", "○",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("Profiles tab missing %q", want)
		}
	}

	// ── Open scanner overlay (Screen 4) ──
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // back to Devices
	model = m.(AppModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = m.(AppModel)

	m, _ = model.Update(OpenScannerMsg{})
	model = m.(AppModel)
	if !model.scanner.visible {
		t.Fatal("expected scanner visible")
	}

	// Feed discovered devices
	m, _ = model.Update(DiscoveredDevicesMsg{
		Devices: []bluetooth.DiscoveredDevice{
			{Address: "AA:BB:CC:DD:EE:FF", Name: "Bose QC45"},
		},
	})
	model = m.(AppModel)
	view = model.View()
	for _, want := range []string{"Scanning", "Bose QC45", "enter pair", "esc"} {
		if !strings.Contains(view, want) {
			t.Errorf("Scanner overlay missing %q", want)
		}
	}

	// Close scanner
	m, _ = model.Update(CloseScannerMsg{})
	model = m.(AppModel)
	if model.scanner.visible {
		t.Fatal("expected scanner hidden after close")
	}
}

func TestPresetsTabView(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model = m.(AppModel)

	// Switch to Presets tab
	for i := 0; i < 3; i++ {
		m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = m.(AppModel)
	}
	if model.activeTab != TabPresets {
		t.Fatalf("expected Presets tab, got %d", model.activeTab)
	}

	// Load presets
	m, _ = model.Update(PresetsLoadedMsg{
		Presets: []preset.Preset{
			{Name: "Music Mode", DefaultSink: "bt-sink", CardProfiles: map[string]string{"sony": "a2dp"}},
			{Name: "Video Call", DefaultSink: "bt-sink"},
		},
	})
	model = m.(AppModel)

	view := model.View()
	for _, want := range []string{"Saved Presets", "Music Mode", "Video Call"} {
		if !strings.Contains(view, want) {
			t.Errorf("presets view missing %q", want)
		}
	}
}

func TestPresetsApplyConfirmFlow(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model = m.(AppModel)

	// Switch to Presets tab
	for i := 0; i < 3; i++ {
		m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = m.(AppModel)
	}

	// Load presets
	m, _ = model.Update(PresetsLoadedMsg{
		Presets: []preset.Preset{
			{Name: "Test Preset", DefaultSink: "sink"},
		},
	})
	model = m.(AppModel)

	// Press enter to open confirm
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(AppModel)

	// The confirm msg should be routed
	m, _ = model.Update(OpenConfirmMsg{
		Preset: preset.Preset{Name: "Test Preset", DefaultSink: "sink"},
		Diffs:  []preset.DiffLine{{Field: "Default sink", From: "", To: "sink"}},
	})
	model = m.(AppModel)

	if !model.presets.confirmVisible {
		t.Fatal("expected confirm overlay visible")
	}

	view := model.View()
	if !strings.Contains(view, "Apply") {
		t.Error("confirm view missing 'Apply'")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("confirm view missing 'Cancel'")
	}

	// Press esc to close
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(AppModel)
	if model.presets.confirmVisible {
		t.Fatal("expected confirm overlay closed after esc")
	}
}

func TestPresetsActiveMarker(t *testing.T) {
	model, _ := newTestApp()
	m, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model = m.(AppModel)

	// Switch to Presets tab
	for i := 0; i < 3; i++ {
		m, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = m.(AppModel)
	}

	// Load + mark active
	m, _ = model.Update(PresetsLoadedMsg{
		Presets: []preset.Preset{
			{Name: "Active One"},
			{Name: "Other"},
		},
	})
	model = m.(AppModel)

	m, _ = model.Update(ApplyPresetResultMsg{
		Name:   "Active One",
		Result: preset.ApplyResult{Applied: []string{"test"}},
	})
	model = m.(AppModel)

	if model.presets.activePreset != "Active One" {
		t.Fatalf("expected activePreset='Active One', got %q", model.presets.activePreset)
	}

	view := model.View()
	if !strings.Contains(view, "[active]") {
		t.Error("view missing [active] badge")
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
