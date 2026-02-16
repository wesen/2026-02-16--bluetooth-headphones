package preset

import (
	"context"
	"fmt"

	"soundctl/pkg/soundctl/audio"
)

// ApplyResult reports what the apply operation did.
type ApplyResult struct {
	Applied []string // human-readable list of changes made
	Errors  []error  // non-fatal errors (e.g. a missing stream)
}

// Apply executes a preset's configuration against the audio service.
// It sets card profiles, volumes, default sink, and app routing.
func Apply(ctx context.Context, au audio.Service, p Preset) ApplyResult {
	var result ApplyResult

	// 1) Set card profiles
	for card, profile := range p.CardProfiles {
		if err := au.SetCardProfile(ctx, card, profile); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("set profile %s→%s: %w", card, profile, err))
		} else {
			result.Applied = append(result.Applied, fmt.Sprintf("Profile %s → %s", card, profile))
		}
	}

	// 2) Set default sink
	if p.DefaultSink != "" {
		if err := au.SetDefaultSink(ctx, p.DefaultSink); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("set default sink %s: %w", p.DefaultSink, err))
		} else {
			result.Applied = append(result.Applied, fmt.Sprintf("Default sink → %s", p.DefaultSink))
		}
	}

	// 3) Set volumes
	for channel, vol := range p.Volumes {
		if err := au.SetVolume(ctx, "sink", channel, vol.Level); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("set volume %s=%d%%: %w", channel, vol.Level, err))
		} else {
			result.Applied = append(result.Applied, fmt.Sprintf("Volume %s → %d%%", channel, vol.Level))
		}
		// Toggle mute if needed (we always set it to target state)
		// Note: pactl set-sink-mute takes 1/0/toggle — we use explicit value
	}

	// 4) App routing
	if len(p.AppRoutes) > 0 {
		inputs, err := au.ListSinkInputs(ctx)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("list sink inputs: %w", err))
		} else {
			for _, si := range inputs {
				targetSink, ok := p.AppRoutes[si.AppName]
				if !ok {
					continue
				}
				if targetSink == "follow_default" {
					targetSink = p.DefaultSink
				}
				if targetSink == "" || targetSink == si.SinkName {
					continue // already routed correctly
				}
				if err := au.MoveSinkInput(ctx, si.Index, targetSink); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("route %s→%s: %w", si.AppName, targetSink, err))
				} else {
					result.Applied = append(result.Applied, fmt.Sprintf("Route %s → %s", si.AppName, targetSink))
				}
			}
		}
	}

	return result
}

// Diff computes the changes that would occur if the preset were applied,
// given the current state.
func Diff(current, target Preset) []DiffLine {
	var diffs []DiffLine

	// Card profiles
	for card, targetProf := range target.CardProfiles {
		currentProf := current.CardProfiles[card]
		if currentProf != targetProf {
			diffs = append(diffs, DiffLine{
				Field: fmt.Sprintf("%s profile", card),
				From:  currentProf,
				To:    targetProf,
			})
		}
	}

	// Default sink
	if target.DefaultSink != "" && current.DefaultSink != target.DefaultSink {
		diffs = append(diffs, DiffLine{
			Field: "Default sink",
			From:  current.DefaultSink,
			To:    target.DefaultSink,
		})
	}

	// Volumes
	for ch, targetVol := range target.Volumes {
		currentVol := current.Volumes[ch]
		if currentVol.Level != targetVol.Level {
			diffs = append(diffs, DiffLine{
				Field: fmt.Sprintf("%s volume", ch),
				From:  fmt.Sprintf("%d%%", currentVol.Level),
				To:    fmt.Sprintf("%d%%", targetVol.Level),
			})
		}
		if currentVol.Muted != targetVol.Muted {
			muteLabel := func(m bool) string {
				if m {
					return "muted"
				}
				return "unmuted"
			}
			diffs = append(diffs, DiffLine{
				Field: fmt.Sprintf("%s mute", ch),
				From:  muteLabel(currentVol.Muted),
				To:    muteLabel(targetVol.Muted),
			})
		}
	}

	// App routes
	for app, targetSink := range target.AppRoutes {
		currentSink := current.AppRoutes[app]
		if currentSink != targetSink {
			diffs = append(diffs, DiffLine{
				Field: fmt.Sprintf("%s route", app),
				From:  currentSink,
				To:    targetSink,
			})
		}
	}

	return diffs
}
