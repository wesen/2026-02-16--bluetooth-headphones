package audio

import (
	"context"
	"fmt"
	"strconv"

	sexec "soundctl/pkg/soundctl/exec"
	"soundctl/pkg/soundctl/parse"
)

type ShortRecord struct {
	ID         int
	Name       string
	Driver     string
	SampleSpec string
	State      string
}

type Service interface {
	ListSinks(ctx context.Context) ([]ShortRecord, error)
	ListSources(ctx context.Context) ([]ShortRecord, error)
	ListCards(ctx context.Context) ([]ShortRecord, error)
	SetDefaultSink(ctx context.Context, sink string) error
	SetDefaultSource(ctx context.Context, source string) error
	MoveSinkInput(ctx context.Context, streamID int, sink string) error
	SetCardProfile(ctx context.Context, card string, profile string) error
	SetVolume(ctx context.Context, target string, name string, percent int) error
	ToggleMute(ctx context.Context, target string, name string) error
}

type ExecService struct {
	runner sexec.Runner
}

func NewExecService(runner sexec.Runner) *ExecService {
	return &ExecService{runner: runner}
}

func (s *ExecService) ListSinks(ctx context.Context) ([]ShortRecord, error) {
	return s.listShort(ctx, "sinks")
}

func (s *ExecService) ListSources(ctx context.Context) ([]ShortRecord, error) {
	return s.listShort(ctx, "sources")
}

func (s *ExecService) ListCards(ctx context.Context) ([]ShortRecord, error) {
	return s.listShort(ctx, "cards")
}

func (s *ExecService) listShort(ctx context.Context, noun string) ([]ShortRecord, error) {
	out, err := s.runner.Run(ctx, "pactl", "list", "short", noun)
	if err != nil {
		return nil, err
	}
	recs, err := parse.ParsePactlShort(out)
	if err != nil {
		return nil, err
	}
	rows := make([]ShortRecord, 0, len(recs))
	for _, rec := range recs {
		rows = append(rows, ShortRecord{
			ID:         rec.ID,
			Name:       rec.Name,
			Driver:     rec.Driver,
			SampleSpec: rec.SampleSpec,
			State:      rec.State,
		})
	}
	return rows, nil
}

func (s *ExecService) SetDefaultSink(ctx context.Context, sink string) error {
	if sink == "" {
		return fmt.Errorf("sink is required")
	}
	_, err := s.runner.Run(ctx, "pactl", "set-default-sink", sink)
	return err
}

func (s *ExecService) SetDefaultSource(ctx context.Context, source string) error {
	if source == "" {
		return fmt.Errorf("source is required")
	}
	_, err := s.runner.Run(ctx, "pactl", "set-default-source", source)
	return err
}

func (s *ExecService) MoveSinkInput(ctx context.Context, streamID int, sink string) error {
	if sink == "" {
		return fmt.Errorf("sink is required")
	}
	_, err := s.runner.Run(ctx, "pactl", "move-sink-input", strconv.Itoa(streamID), sink)
	return err
}

func (s *ExecService) SetCardProfile(ctx context.Context, card string, profile string) error {
	if card == "" {
		return fmt.Errorf("card is required")
	}
	if profile == "" {
		return fmt.Errorf("profile is required")
	}
	_, err := s.runner.Run(ctx, "pactl", "set-card-profile", card, profile)
	return err
}

func (s *ExecService) SetVolume(ctx context.Context, target string, name string, percent int) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if percent < 0 || percent > 150 {
		return fmt.Errorf("percent must be between 0 and 150")
	}
	cmd, err := volumeCommand(target)
	if err != nil {
		return err
	}
	_, err = s.runner.Run(ctx, "pactl", cmd, name, fmt.Sprintf("%d%%", percent))
	return err
}

func (s *ExecService) ToggleMute(ctx context.Context, target string, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	cmd, err := muteCommand(target)
	if err != nil {
		return err
	}
	_, err = s.runner.Run(ctx, "pactl", cmd, name, "toggle")
	return err
}

func volumeCommand(target string) (string, error) {
	switch target {
	case "sink":
		return "set-sink-volume", nil
	case "source":
		return "set-source-volume", nil
	default:
		return "", fmt.Errorf("invalid target %q: expected sink or source", target)
	}
}

func muteCommand(target string) (string, error) {
	switch target {
	case "sink":
		return "set-sink-mute", nil
	case "source":
		return "set-source-mute", nil
	default:
		return "", fmt.Errorf("invalid target %q: expected sink or source", target)
	}
}
