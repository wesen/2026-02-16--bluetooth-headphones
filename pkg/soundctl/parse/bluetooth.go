package parse

import (
	"fmt"
	"strings"
)

type BluetoothDeviceRecord struct {
	Address string
	Name    string
}

type BluetoothInfoRecord struct {
	Address   string
	Name      string
	Alias     string
	Paired    bool
	Trusted   bool
	Connected bool
}

type BluetoothControllerRecord struct {
	Address     string
	Alias       string
	Powered     bool
	Pairable    bool
	Discovering bool
}

func ParseBluetoothDevices(output string) ([]BluetoothDeviceRecord, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []BluetoothDeviceRecord{}, nil
	}

	devices := make([]BluetoothDeviceRecord, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "Device ") {
			return nil, fmt.Errorf("unexpected bluetoothctl devices line: %q", line)
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("malformed bluetoothctl devices line: %q", line)
		}
		devices = append(devices, BluetoothDeviceRecord{Address: parts[1], Name: parts[2]})
	}
	return devices, nil
}

func ParseBluetoothInfo(output string) (BluetoothInfoRecord, error) {
	var info BluetoothInfoRecord
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if i == 0 {
			if !strings.HasPrefix(line, "Device ") {
				return info, fmt.Errorf("expected header line starting with Device, got %q", line)
			}
			parts := strings.SplitN(line, " ", 3)
			if len(parts) < 2 {
				return info, fmt.Errorf("malformed bluetooth info header: %q", line)
			}
			info.Address = parts[1]
			if len(parts) == 3 {
				info.Name = strings.TrimSpace(parts[2])
			}
			continue
		}

		if strings.HasPrefix(line, "Name:") {
			info.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		}
		if strings.HasPrefix(line, "Alias:") {
			info.Alias = strings.TrimSpace(strings.TrimPrefix(line, "Alias:"))
		}
		if strings.HasPrefix(line, "Paired:") {
			info.Paired = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Paired:")), "yes")
		}
		if strings.HasPrefix(line, "Trusted:") {
			info.Trusted = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Trusted:")), "yes")
		}
		if strings.HasPrefix(line, "Connected:") {
			info.Connected = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Connected:")), "yes")
		}
	}

	if info.Address == "" {
		return info, fmt.Errorf("missing bluetooth device address in info output")
	}
	return info, nil
}

func ParseBluetoothShow(output string) (BluetoothControllerRecord, error) {
	var status BluetoothControllerRecord
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if i == 0 {
			if !strings.HasPrefix(line, "Controller ") {
				return status, fmt.Errorf("expected header line starting with Controller, got %q", line)
			}
			parts := strings.SplitN(line, " ", 3)
			if len(parts) < 2 {
				return status, fmt.Errorf("malformed bluetooth show header: %q", line)
			}
			status.Address = parts[1]
			continue
		}
		if strings.HasPrefix(line, "Alias:") {
			status.Alias = strings.TrimSpace(strings.TrimPrefix(line, "Alias:"))
		}
		if strings.HasPrefix(line, "Powered:") {
			status.Powered = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Powered:")), "yes")
		}
		if strings.HasPrefix(line, "Pairable:") {
			status.Pairable = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Pairable:")), "yes")
		}
		if strings.HasPrefix(line, "Discovering:") {
			status.Discovering = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "Discovering:")), "yes")
		}
	}
	if status.Address == "" {
		return status, fmt.Errorf("missing controller address in show output")
	}
	return status, nil
}
