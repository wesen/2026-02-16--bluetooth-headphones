package tui

import (
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
)

// Tab indices.
const (
	TabDevices  = 0
	TabSinks    = 1
	TabProfiles = 2
)

// TabChangedMsg requests the root model switch tabs.
type TabChangedMsg struct{ Delta int }

// StatusMsg sets a transient message in the status bar.
type StatusMsg struct{ Text string }

// ErrorMsg reports an error to the status bar.
type ErrorMsg struct{ Err error }

// --- Bluetooth domain messages ---

// DevicesLoadedMsg carries refreshed device list.
type DevicesLoadedMsg struct {
	Devices    []bluetooth.Device
	Controller bluetooth.ControllerStatus
	Err        error
}

// ConnectResultMsg reports connect outcome.
type ConnectResultMsg struct {
	Addr string
	Err  error
}

// DisconnectResultMsg reports disconnect outcome.
type DisconnectResultMsg struct {
	Addr string
	Err  error
}

// ForgetResultMsg reports remove/forget outcome.
type ForgetResultMsg struct {
	Addr string
	Err  error
}

// --- Scanner overlay messages ---

// OpenScannerMsg opens the scan overlay.
type OpenScannerMsg struct{}

// CloseScannerMsg closes the scan overlay.
type CloseScannerMsg struct{}

// DiscoveredDevicesMsg carries discovered devices from scan.
type DiscoveredDevicesMsg struct {
	Devices []bluetooth.DiscoveredDevice
	Err     error
}

// PairResultMsg reports pairing outcome.
type PairResultMsg struct {
	Addr string
	Err  error
}

// --- Audio domain messages ---

// SinksLoadedMsg carries refreshed sink list.
type SinksLoadedMsg struct {
	Sinks   []audio.ShortRecord
	Sources []audio.ShortRecord
	Err     error
}

// SetDefaultResultMsg reports set-default outcome.
type SetDefaultResultMsg struct {
	Name string
	Kind string // "sink" or "source"
	Err  error
}

// ProfilesLoadedMsg carries refreshed card/profiles list.
type ProfilesLoadedMsg struct {
	Cards []audio.ShortRecord
	Err   error
}

// SetProfileResultMsg reports set-profile outcome.
type SetProfileResultMsg struct {
	Card    string
	Profile string
	Err     error
}
