package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
)

// --- Bluetooth commands ---

func loadDevicesCmd(bt bluetooth.Service) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		controller, err := bt.ControllerStatus(ctx)
		if err != nil {
			return DevicesLoadedMsg{Err: err}
		}
		devices, err := bt.ListDevices(ctx)
		return DevicesLoadedMsg{Devices: devices, Controller: controller, Err: err}
	}
}

func connectCmd(bt bluetooth.Service, addr string) tea.Cmd {
	return func() tea.Msg {
		err := bt.Connect(context.Background(), addr)
		return ConnectResultMsg{Addr: addr, Err: err}
	}
}

func disconnectCmd(bt bluetooth.Service, addr string) tea.Cmd {
	return func() tea.Msg {
		err := bt.Disconnect(context.Background(), addr)
		return DisconnectResultMsg{Addr: addr, Err: err}
	}
}

func forgetCmd(bt bluetooth.Service, addr string) tea.Cmd {
	return func() tea.Msg {
		err := bt.Remove(context.Background(), addr)
		return ForgetResultMsg{Addr: addr, Err: err}
	}
}

// --- Scanner commands ---

func discoverCmd(bt bluetooth.Service, seconds int) tea.Cmd {
	return func() tea.Msg {
		found, err := bt.Discover(context.Background(), seconds)
		return DiscoveredDevicesMsg{Devices: found, Err: err}
	}
}

func pairCmd(bt bluetooth.Service, addr string) tea.Cmd {
	return func() tea.Msg {
		if err := bt.Pair(context.Background(), addr); err != nil {
			return PairResultMsg{Addr: addr, Err: err}
		}
		if err := bt.Trust(context.Background(), addr); err != nil {
			return PairResultMsg{Addr: addr, Err: err}
		}
		if err := bt.Connect(context.Background(), addr); err != nil {
			return PairResultMsg{Addr: addr, Err: err}
		}
		return PairResultMsg{Addr: addr}
	}
}

// --- Audio commands ---

func loadSinksCmd(au audio.Service) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		sinks, err := au.ListSinks(ctx)
		if err != nil {
			return SinksLoadedMsg{Err: err}
		}
		sources, err := au.ListSources(ctx)
		if err != nil {
			return SinksLoadedMsg{Err: err}
		}
		inputs, err := au.ListSinkInputs(ctx)
		if err != nil {
			return SinksLoadedMsg{Err: err}
		}
		defaults, err := au.GetDefaults(ctx)
		if err != nil {
			return SinksLoadedMsg{Err: err}
		}
		return SinksLoadedMsg{
			Sinks:             sinks,
			Sources:           sources,
			SinkInputs:        inputs,
			DefaultSinkName:   defaults.DefaultSinkName,
			DefaultSourceName: defaults.DefaultSourceName,
		}
	}
}

func moveSinkInputCmd(au audio.Service, streamID int, sink string) tea.Cmd {
	return func() tea.Msg {
		err := au.MoveSinkInput(context.Background(), streamID, sink)
		return MoveStreamResultMsg{StreamID: streamID, Sink: sink, Err: err}
	}
}

func setDefaultSinkCmd(au audio.Service, name string) tea.Cmd {
	return func() tea.Msg {
		err := au.SetDefaultSink(context.Background(), name)
		return SetDefaultResultMsg{Name: name, Kind: "sink", Err: err}
	}
}

func setDefaultSourceCmd(au audio.Service, name string) tea.Cmd {
	return func() tea.Msg {
		err := au.SetDefaultSource(context.Background(), name)
		return SetDefaultResultMsg{Name: name, Kind: "source", Err: err}
	}
}

func loadProfilesCmd(au audio.Service) tea.Cmd {
	return func() tea.Msg {
		cards, err := au.ListCardsDetailed(context.Background())
		return ProfilesLoadedMsg{Cards: cards, Err: err}
	}
}

func setProfileCmd(au audio.Service, card, profile string) tea.Cmd {
	return func() tea.Msg {
		err := au.SetCardProfile(context.Background(), card, profile)
		return SetProfileResultMsg{Card: card, Profile: profile, Err: err}
	}
}
