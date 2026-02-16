package preset

import (
	"context"
	"testing"

	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/exec"
)

func TestSnapshotCurrent(t *testing.T) {
	runner := exec.NewFakeRunner()

	// pactl info
	runner.Set("pactl", []string{"info"}, exec.CommandResult{
		Output: "Default Sink: bt-sink\nDefault Source: bt-source\nServer Name: PipeWire",
	})

	// pactl list cards (detailed)
	runner.Set("pactl", []string{"list", "cards"}, exec.CommandResult{
		Output: "Card #0\n\tName: bluez_card.sony\n\tDriver: bluez5\n\tProfiles:\n\t\ta2dp-sink: A2DP (sinks: 1, sources: 0, priority: 40, available: yes)\n\t\toff: Off (sinks: 0, sources: 0, priority: 0, available: yes)\n\tActive Profile: a2dp-sink",
	})

	// pactl list sink-inputs
	runner.Set("pactl", []string{"list", "sink-inputs"}, exec.CommandResult{
		Output: "Sink Input #57\n\tSink: 1\n\tProperties:\n\t\tapplication.name = \"Firefox\"\n",
	})

	// pactl list short sinks (for sink name resolution + volume snapshot)
	runner.Set("pactl", []string{"list", "short", "sinks"}, exec.CommandResult{
		Output: "1\tbt-sink\tbluez5\tspec\tRUNNING",
	})

	au := audio.NewExecService(runner)

	p, err := SnapshotCurrent(context.Background(), au)
	if err != nil {
		t.Fatalf("SnapshotCurrent: %v", err)
	}

	if p.DefaultSink != "bt-sink" {
		t.Fatalf("expected default sink bt-sink, got %q", p.DefaultSink)
	}

	if p.CardProfiles["bluez_card.sony"] != "a2dp-sink" {
		t.Fatalf("expected a2dp-sink profile, got %q", p.CardProfiles["bluez_card.sony"])
	}

	if p.AppRoutes["Firefox"] != "follow_default" {
		t.Fatalf("expected Firefox→follow_default, got %q", p.AppRoutes["Firefox"])
	}

	if _, ok := p.Volumes["bt-sink"]; !ok {
		t.Fatal("expected bt-sink volume entry")
	}
}
