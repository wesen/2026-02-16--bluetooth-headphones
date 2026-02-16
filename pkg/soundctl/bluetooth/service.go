package bluetooth

import (
	"context"
	"fmt"

	sexec "soundctl/pkg/soundctl/exec"
	"soundctl/pkg/soundctl/parse"
)

type Device struct {
	Address string
	Name    string
}

type DeviceInfo struct {
	Address   string
	Name      string
	Alias     string
	Paired    bool
	Trusted   bool
	Connected bool
}

type Service interface {
	ListDevices(ctx context.Context) ([]Device, error)
	Info(ctx context.Context, address string) (DeviceInfo, error)
	Connect(ctx context.Context, address string) error
	Disconnect(ctx context.Context, address string) error
	Trust(ctx context.Context, address string) error
	Remove(ctx context.Context, address string) error
	Pair(ctx context.Context, address string) error
	StartScan(ctx context.Context) error
	StopScan(ctx context.Context) error
}

type ExecService struct {
	runner sexec.Runner
}

func NewExecService(runner sexec.Runner) *ExecService {
	return &ExecService{runner: runner}
}

func (s *ExecService) ListDevices(ctx context.Context) ([]Device, error) {
	out, err := s.runner.Run(ctx, "bluetoothctl", "devices")
	if err != nil {
		return nil, err
	}
	recs, err := parse.ParseBluetoothDevices(out)
	if err != nil {
		return nil, err
	}
	devices := make([]Device, 0, len(recs))
	for _, rec := range recs {
		devices = append(devices, Device{Address: rec.Address, Name: rec.Name})
	}
	return devices, nil
}

func (s *ExecService) Info(ctx context.Context, address string) (DeviceInfo, error) {
	if address == "" {
		return DeviceInfo{}, fmt.Errorf("address is required")
	}
	out, err := s.runner.Run(ctx, "bluetoothctl", "info", address)
	if err != nil {
		return DeviceInfo{}, err
	}
	rec, err := parse.ParseBluetoothInfo(out)
	if err != nil {
		return DeviceInfo{}, err
	}
	return DeviceInfo{
		Address:   rec.Address,
		Name:      rec.Name,
		Alias:     rec.Alias,
		Paired:    rec.Paired,
		Trusted:   rec.Trusted,
		Connected: rec.Connected,
	}, nil
}

func (s *ExecService) Connect(ctx context.Context, address string) error {
	return s.runOnAddress(ctx, "connect", address)
}

func (s *ExecService) Disconnect(ctx context.Context, address string) error {
	return s.runOnAddress(ctx, "disconnect", address)
}

func (s *ExecService) Trust(ctx context.Context, address string) error {
	return s.runOnAddress(ctx, "trust", address)
}

func (s *ExecService) Remove(ctx context.Context, address string) error {
	return s.runOnAddress(ctx, "remove", address)
}

func (s *ExecService) Pair(ctx context.Context, address string) error {
	return s.runOnAddress(ctx, "pair", address)
}

func (s *ExecService) StartScan(ctx context.Context) error {
	_, err := s.runner.Run(ctx, "bluetoothctl", "scan", "on")
	return err
}

func (s *ExecService) StopScan(ctx context.Context) error {
	_, err := s.runner.Run(ctx, "bluetoothctl", "scan", "off")
	return err
}

func (s *ExecService) runOnAddress(ctx context.Context, operation string, address string) error {
	if address == "" {
		return fmt.Errorf("address is required")
	}
	_, err := s.runner.Run(ctx, "bluetoothctl", operation, address)
	return err
}
