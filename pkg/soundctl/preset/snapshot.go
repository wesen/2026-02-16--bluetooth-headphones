package preset

import (
	"context"
	"time"

	"soundctl/pkg/soundctl/audio"
)

// SnapshotCurrent captures the current live audio state as a Preset.
// The returned preset has no name (the caller should set one).
func SnapshotCurrent(ctx context.Context, au audio.Service) (Preset, error) {
	now := time.Now()
	p := Preset{
		CreatedAt:    now,
		UpdatedAt:    now,
		CardProfiles: make(map[string]string),
		Volumes:      make(map[string]VolumeSpec),
		AppRoutes:    make(map[string]string),
	}

	// Get defaults
	defaults, err := au.GetDefaults(ctx)
	if err != nil {
		return p, err
	}
	p.DefaultSink = defaults.DefaultSinkName

	// Get detailed cards with active profiles
	cards, err := au.ListCardsDetailed(ctx)
	if err != nil {
		return p, err
	}
	for _, card := range cards {
		if card.ActiveProfile != "" {
			p.CardProfiles[card.Name] = card.ActiveProfile
		}
	}

	// Get current sink inputs for app routing
	inputs, err := au.ListSinkInputs(ctx)
	if err != nil {
		return p, err
	}
	for _, si := range inputs {
		if si.AppName != "" {
			if si.SinkName == p.DefaultSink {
				p.AppRoutes[si.AppName] = "follow_default"
			} else {
				p.AppRoutes[si.AppName] = si.SinkName
			}
		}
	}

	// Note: actual per-sink volume levels would require `pactl list sinks`
	// with volume parsing. For now, we capture sinks with a placeholder.
	// This is a known limitation documented in the diary.
	sinks, err := au.ListSinks(ctx)
	if err != nil {
		return p, err
	}
	for _, sink := range sinks {
		// Use sink name as channel name for now
		p.Volumes[sink.Name] = VolumeSpec{Level: 100, Muted: false}
	}

	return p, nil
}
