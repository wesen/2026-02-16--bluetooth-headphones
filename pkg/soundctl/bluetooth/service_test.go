package bluetooth

import (
	"context"
	"errors"
	"testing"

	sexec "soundctl/pkg/soundctl/exec"
)

func TestListDevices(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"devices"}, sexec.CommandResult{Output: "Device 08:FF:44:2B:4C:90 AirPods Max"})
	fake.Set("bluetoothctl", []string{"info", "08:FF:44:2B:4C:90"}, sexec.CommandResult{Output: `Device 08:FF:44:2B:4C:90 (public)
	Name: AirPods Max
	Alias: AirPods Max
	Paired: yes
	Trusted: yes
	Connected: no`})

	svc := NewExecService(fake)
	devices, err := svc.ListDevices(context.Background())
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Address != "08:FF:44:2B:4C:90" {
		t.Fatalf("unexpected address: %s", devices[0].Address)
	}
	if devices[0].Connection != "paired" {
		t.Fatalf("unexpected connection mode: %s", devices[0].Connection)
	}
	if !devices[0].Trusted {
		t.Fatal("expected trusted=true")
	}
}

func TestInfo(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"info", "08:FF:44:2B:4C:90"}, sexec.CommandResult{Output: `Device 08:FF:44:2B:4C:90 (public)
	Name: AirPods Max
	Alias: AirPods Max
	Paired: yes
	Trusted: no
	Connected: yes`})

	svc := NewExecService(fake)
	info, err := svc.Info(context.Background(), "08:FF:44:2B:4C:90")
	if err != nil {
		t.Fatalf("Info failed: %v", err)
	}
	if !info.Paired || !info.Connected {
		t.Fatalf("unexpected info state: %#v", info)
	}
	if info.Trusted {
		t.Fatalf("expected trusted=false, got %#v", info)
	}
}

func TestConnectRunsExpectedCommand(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"connect", "AA:BB:CC:DD:EE:FF"}, sexec.CommandResult{})

	svc := NewExecService(fake)
	if err := svc.Connect(context.Background(), "AA:BB:CC:DD:EE:FF"); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	calls := fake.Calls()
	if len(calls) != 1 || calls[0] != "bluetoothctl connect AA:BB:CC:DD:EE:FF" {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func TestControllerStatus(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"show"}, sexec.CommandResult{Output: `Controller 10:A5:1D:00:C6:6F (public)
	Alias: f
	Powered: yes
	Pairable: no
	Discovering: yes`})

	svc := NewExecService(fake)
	status, err := svc.ControllerStatus(context.Background())
	if err != nil {
		t.Fatalf("ControllerStatus failed: %v", err)
	}
	if !status.Discovering {
		t.Fatal("expected discovering=true")
	}
	if status.Pairable {
		t.Fatal("expected pairable=false")
	}
}

func TestPairAlreadyExistsIsNonFatal(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"pair", "08:FF:44:2B:4C:90"}, sexec.CommandResult{
		Output: "Failed to pair: org.bluez.Error.AlreadyExists",
		Err:    errors.New("exit status 1"),
	})

	svc := NewExecService(fake)
	if err := svc.Pair(context.Background(), "08:FF:44:2B:4C:90"); err != nil {
		t.Fatalf("Pair should ignore AlreadyExists, got error: %v", err)
	}
}

func TestDiscoverParsesDevices(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"--timeout", "8", "scan", "on"}, sexec.CommandResult{
		Output: "Discovery started\n[NEW] Device 90:62:3F:92:B1:A7 AirPods Max #3 - Find My",
	})

	svc := NewExecService(fake)
	found, err := svc.Discover(context.Background(), 8)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 discovered device, got %d", len(found))
	}
	if found[0].Address != "90:62:3F:92:B1:A7" {
		t.Fatalf("unexpected discovered address: %s", found[0].Address)
	}
}
