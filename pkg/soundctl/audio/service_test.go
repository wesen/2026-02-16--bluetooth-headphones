package audio

import (
	"context"
	"testing"

	sexec "soundctl/pkg/soundctl/exec"
)

func TestListSinks(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("pactl", []string{"list", "short", "sinks"}, sexec.CommandResult{Output: "47\talsa_output.pci-0000_00_1f.3.analog-stereo\tPipeWire\ts32le 2ch 48000Hz\tRUNNING"})

	svc := NewExecService(fake)
	sinks, err := svc.ListSinks(context.Background())
	if err != nil {
		t.Fatalf("ListSinks failed: %v", err)
	}
	if len(sinks) != 1 {
		t.Fatalf("expected 1 sink, got %d", len(sinks))
	}
	if sinks[0].Name != "alsa_output.pci-0000_00_1f.3.analog-stereo" {
		t.Fatalf("unexpected sink name: %s", sinks[0].Name)
	}
}

func TestSetDefaultSink(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("pactl", []string{"set-default-sink", "bluez_output.08_FF_44_2B_4C_90"}, sexec.CommandResult{})

	svc := NewExecService(fake)
	if err := svc.SetDefaultSink(context.Background(), "bluez_output.08_FF_44_2B_4C_90"); err != nil {
		t.Fatalf("SetDefaultSink failed: %v", err)
	}
}

func TestSetVolumeTargetValidation(t *testing.T) {
	svc := NewExecService(sexec.NewFakeRunner())
	if err := svc.SetVolume(context.Background(), "invalid", "x", 50); err == nil {
		t.Fatal("expected validation error for invalid target")
	}
	if err := svc.SetVolume(context.Background(), "sink", "x", 200); err == nil {
		t.Fatal("expected validation error for invalid percent")
	}
}

func TestToggleMuteUsesExpectedCommand(t *testing.T) {
	fake := sexec.NewFakeRunner()
	fake.Set("pactl", []string{"set-source-mute", "alsa_input.pci-0000_00_1f.3.analog-stereo", "toggle"}, sexec.CommandResult{})

	svc := NewExecService(fake)
	if err := svc.ToggleMute(context.Background(), "source", "alsa_input.pci-0000_00_1f.3.analog-stereo"); err != nil {
		t.Fatalf("ToggleMute failed: %v", err)
	}

	calls := fake.Calls()
	if len(calls) != 1 || calls[0] != "pactl set-source-mute alsa_input.pci-0000_00_1f.3.analog-stereo toggle" {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}
