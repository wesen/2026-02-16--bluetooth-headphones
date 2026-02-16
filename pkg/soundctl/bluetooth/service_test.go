package bluetooth

import (
	"context"
	"testing"

	sexec "soundctl/pkg/soundctl/exec"
)

func TestListDevices(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("bluetoothctl", []string{"devices"}, sexec.CommandResult{Output: "Device 08:FF:44:2B:4C:90 AirPods Max"})

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
