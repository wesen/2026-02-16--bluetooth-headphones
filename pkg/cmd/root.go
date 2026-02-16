package cmd

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
	"soundctl/pkg/cmd/devices"
	"soundctl/pkg/cmd/mute"
	"soundctl/pkg/cmd/profiles"
	"soundctl/pkg/cmd/scan"
	"soundctl/pkg/cmd/sinks"
	"soundctl/pkg/cmd/sources"
	"soundctl/pkg/cmd/volume"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
)

type Dependencies struct {
	Bluetooth bluetooth.Service
	Audio     audio.Service
}

func NewRootCommand(deps Dependencies) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "soundctl",
		Short: "SoundCtl CLI for bluetooth/audio operations",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "soundctl"); err != nil {
		return nil, err
	}

	helpSystem := help.NewHelpSystem()
	if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
		return nil, err
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	groups := []*cobra.Command{
		{Use: "devices", Short: "Bluetooth device operations"},
		{Use: "scan", Short: "Bluetooth scanning operations"},
		{Use: "sinks", Short: "Audio sink operations"},
		{Use: "sources", Short: "Audio source operations"},
		{Use: "profiles", Short: "Audio profile/card operations"},
		{Use: "volume", Short: "Volume operations"},
		{Use: "mute", Short: "Mute operations"},
	}
	for _, g := range groups {
		rootCmd.AddCommand(g)
	}

	if err := devices.Register(groups[0], deps.Bluetooth); err != nil {
		return nil, fmt.Errorf("register devices commands: %w", err)
	}
	if err := scan.Register(groups[1], deps.Bluetooth); err != nil {
		return nil, fmt.Errorf("register scan commands: %w", err)
	}
	if err := sinks.Register(groups[2], deps.Audio); err != nil {
		return nil, fmt.Errorf("register sinks commands: %w", err)
	}
	if err := sources.Register(groups[3], deps.Audio); err != nil {
		return nil, fmt.Errorf("register sources commands: %w", err)
	}
	if err := profiles.Register(groups[4], deps.Audio); err != nil {
		return nil, fmt.Errorf("register profiles commands: %w", err)
	}
	if err := volume.Register(groups[5], deps.Audio); err != nil {
		return nil, fmt.Errorf("register volume commands: %w", err)
	}
	if err := mute.Register(groups[6], deps.Audio); err != nil {
		return nil, fmt.Errorf("register mute commands: %w", err)
	}

	return rootCmd, nil
}
