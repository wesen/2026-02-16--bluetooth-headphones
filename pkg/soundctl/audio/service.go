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

// DefaultsInfo reports the current default sink and source names.
type DefaultsInfo struct {
	DefaultSinkName   string
	DefaultSourceName string
	ServerName        string
}

// SinkInput represents an active stream routed to a sink.
type SinkInput struct {
	Index     int
	SinkIndex int
	AppName   string
	MediaName string
	SinkName  string // resolved from sink index
}

// Card represents an audio card with its available profiles.
type Card struct {
	Index         int
	Name          string
	Driver        string
	Profiles      []CardProfile
	ActiveProfile string
}

// CardProfile represents a profile available on a card.
type CardProfile struct {
	Name        string
	Description string
	Available   bool
}

type Service interface {
	ListSinks(ctx context.Context) ([]ShortRecord, error)
	ListSources(ctx context.Context) ([]ShortRecord, error)
	ListCards(ctx context.Context) ([]ShortRecord, error)
	GetDefaults(ctx context.Context) (DefaultsInfo, error)
	ListSinkInputs(ctx context.Context) ([]SinkInput, error)
	ListCardsDetailed(ctx context.Context) ([]Card, error)
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

func (s *ExecService) GetDefaults(ctx context.Context) (DefaultsInfo, error) {
	out, err := s.runner.Run(ctx, "pactl", "info")
	if err != nil {
		return DefaultsInfo{}, err
	}
	rec, err := parse.ParsePactlInfo(out)
	if err != nil {
		return DefaultsInfo{}, err
	}
	return DefaultsInfo{
		DefaultSinkName:   rec.DefaultSinkName,
		DefaultSourceName: rec.DefaultSourceName,
		ServerName:        rec.ServerName,
	}, nil
}

func (s *ExecService) ListSinkInputs(ctx context.Context) ([]SinkInput, error) {
	out, err := s.runner.Run(ctx, "pactl", "list", "sink-inputs")
	if err != nil {
		return nil, err
	}
	recs, err := parse.ParsePactlSinkInputs(out)
	if err != nil {
		return nil, err
	}

	// Resolve sink names by listing sinks
	sinks, err := s.ListSinks(ctx)
	if err != nil {
		return nil, err
	}
	sinkMap := make(map[int]string)
	for _, sink := range sinks {
		sinkMap[sink.ID] = sink.Name
	}

	inputs := make([]SinkInput, 0, len(recs))
	for _, rec := range recs {
		inputs = append(inputs, SinkInput{
			Index:     rec.Index,
			SinkIndex: rec.SinkIndex,
			AppName:   rec.AppName,
			MediaName: rec.MediaName,
			SinkName:  sinkMap[rec.SinkIndex],
		})
	}
	return inputs, nil
}

func (s *ExecService) ListCardsDetailed(ctx context.Context) ([]Card, error) {
	out, err := s.runner.Run(ctx, "pactl", "list", "cards")
	if err != nil {
		return nil, err
	}
	recs, err := parse.ParsePactlCards(out)
	if err != nil {
		return nil, err
	}
	cards := make([]Card, 0, len(recs))
	for _, rec := range recs {
		profiles := make([]CardProfile, 0, len(rec.Profiles))
		for _, p := range rec.Profiles {
			profiles = append(profiles, CardProfile{
				Name:        p.Name,
				Description: p.Description,
				Available:   p.Available,
			})
		}
		cards = append(cards, Card{
			Index:         rec.Index,
			Name:          rec.Name,
			Driver:        rec.Driver,
			Profiles:      profiles,
			ActiveProfile: rec.ActiveProfile,
		})
	}
	return cards, nil
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
