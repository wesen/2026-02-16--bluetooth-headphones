package preset

import (
	"context"
	"testing"

	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/exec"
)

func fakeAudioService(runner *exec.FakeRunner) audio.Service {
	return audio.NewExecService(runner)
}

func TestApplyCardProfiles(t *testing.T) {
	runner := exec.NewFakeRunner()
	runner.Set("pactl", []string{"set-card-profile", "bluez_card.sony", "a2dp-sink"}, exec.CommandResult{})
	runner.Set("pactl", []string{"set-default-sink", "bt-sink"}, exec.CommandResult{})

	au := fakeAudioService(runner)
	p := Preset{
		Name:         "Test",
		CardProfiles: map[string]string{"bluez_card.sony": "a2dp-sink"},
		DefaultSink:  "bt-sink",
	}

	result := Apply(context.Background(), au, p)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Applied) != 2 {
		t.Fatalf("expected 2 applied changes, got %d: %v", len(result.Applied), result.Applied)
	}
}

func TestApplyVolumes(t *testing.T) {
	runner := exec.NewFakeRunner()
	runner.Set("pactl", []string{"set-sink-volume", "master", "80%"}, exec.CommandResult{})

	au := fakeAudioService(runner)
	p := Preset{
		Name:    "Vol",
		Volumes: map[string]VolumeSpec{"master": {Level: 80}},
	}

	result := Apply(context.Background(), au, p)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestApplyAppRoutes(t *testing.T) {
	runner := exec.NewFakeRunner()

	// Stub sink-inputs listing
	runner.Set("pactl", []string{"list", "sink-inputs"}, exec.CommandResult{
		Output: "Sink Input #57\n\tSink: 1\n\tProperties:\n\t\tapplication.name = \"Firefox\"\n",
	})
	// Stub sinks for name resolution
	runner.Set("pactl", []string{"list", "short", "sinks"}, exec.CommandResult{
		Output: "1\told-sink\tdriver\tspec\tRUNNING",
	})
	// Stub set-default-sink (called because DefaultSink is set)
	runner.Set("pactl", []string{"set-default-sink", "bt-sink"}, exec.CommandResult{})
	// Stub move command
	runner.Set("pactl", []string{"move-sink-input", "57", "bt-sink"}, exec.CommandResult{})

	au := fakeAudioService(runner)
	p := Preset{
		Name:        "Route",
		DefaultSink: "bt-sink",
		AppRoutes:   map[string]string{"Firefox": "bt-sink"},
	}

	result := Apply(context.Background(), au, p)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	calls := runner.Calls()
	found := false
	for _, c := range calls {
		if c == "pactl move-sink-input 57 bt-sink" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected move-sink-input call, got %v", calls)
	}
}

func TestApplyFollowDefault(t *testing.T) {
	runner := exec.NewFakeRunner()

	runner.Set("pactl", []string{"list", "sink-inputs"}, exec.CommandResult{
		Output: "Sink Input #10\n\tSink: 1\n\tProperties:\n\t\tapplication.name = \"Spotify\"\n",
	})
	runner.Set("pactl", []string{"list", "short", "sinks"}, exec.CommandResult{
		Output: "1\told-sink\tdriver\tspec\tRUNNING",
	})
	runner.Set("pactl", []string{"set-default-sink", "my-default"}, exec.CommandResult{})
	runner.Set("pactl", []string{"move-sink-input", "10", "my-default"}, exec.CommandResult{})

	au := fakeAudioService(runner)
	p := Preset{
		Name:        "Follow",
		DefaultSink: "my-default",
		AppRoutes:   map[string]string{"Spotify": "follow_default"},
	}

	result := Apply(context.Background(), au, p)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestDiff(t *testing.T) {
	current := Preset{
		CardProfiles: map[string]string{"card1": "a2dp"},
		Volumes:      map[string]VolumeSpec{"Master": {Level: 80, Muted: false}},
		DefaultSink:  "sink-a",
		AppRoutes:    map[string]string{"Firefox": "sink-a"},
	}
	target := Preset{
		CardProfiles: map[string]string{"card1": "hsp"},
		Volumes:      map[string]VolumeSpec{"Master": {Level: 60, Muted: false}},
		DefaultSink:  "sink-b",
		AppRoutes:    map[string]string{"Firefox": "sink-b"},
	}

	diffs := Diff(current, target)
	if len(diffs) != 4 {
		t.Fatalf("expected 4 diffs, got %d: %+v", len(diffs), diffs)
	}

	// Check for expected fields
	fields := map[string]bool{}
	for _, d := range diffs {
		fields[d.Field] = true
	}
	for _, want := range []string{"card1 profile", "Default sink", "Master volume", "Firefox route"} {
		if !fields[want] {
			t.Errorf("missing diff field %q", want)
		}
	}
}

func TestDiffNoChanges(t *testing.T) {
	p := Preset{
		CardProfiles: map[string]string{"card1": "a2dp"},
		DefaultSink:  "sink-a",
	}
	diffs := Diff(p, p)
	if len(diffs) != 0 {
		t.Fatalf("expected 0 diffs for identical presets, got %d", len(diffs))
	}
}
