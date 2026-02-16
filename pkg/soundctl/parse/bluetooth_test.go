package parse

import "testing"

func TestParseBluetoothDevices(t *testing.T) {
	input := "Device 08:FF:44:2B:4C:90 Manuel's AirPods Max\nDevice C0:95:6D:A9:79:3C Big Daddy's AirPods Pro"

	devices, err := ParseBluetoothDevices(input)
	if err != nil {
		t.Fatalf("ParseBluetoothDevices returned error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Address != "08:FF:44:2B:4C:90" {
		t.Fatalf("unexpected first address: %s", devices[0].Address)
	}
	if devices[0].Name != "Manuel's AirPods Max" {
		t.Fatalf("unexpected first name: %s", devices[0].Name)
	}
}

func TestParseBluetoothInfo(t *testing.T) {
	input := `Device 08:FF:44:2B:4C:90 (public)
	Name: Manuel's AirPods Max
	Alias: Manuel's AirPods Max
	Paired: yes
	Trusted: no
	Connected: yes`

	info, err := ParseBluetoothInfo(input)
	if err != nil {
		t.Fatalf("ParseBluetoothInfo returned error: %v", err)
	}
	if info.Address != "08:FF:44:2B:4C:90" {
		t.Fatalf("unexpected address: %s", info.Address)
	}
	if info.Name != "Manuel's AirPods Max" {
		t.Fatalf("unexpected name: %s", info.Name)
	}
	if !info.Paired {
		t.Fatal("expected paired=true")
	}
	if info.Trusted {
		t.Fatal("expected trusted=false")
	}
	if !info.Connected {
		t.Fatal("expected connected=true")
	}
}
