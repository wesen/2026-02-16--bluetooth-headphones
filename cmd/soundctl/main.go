package main

import (
	"fmt"
	"os"

	"soundctl/pkg/cmd"
	"soundctl/pkg/soundctl/audio"
	"soundctl/pkg/soundctl/bluetooth"
	sexec "soundctl/pkg/soundctl/exec"
	"soundctl/pkg/soundctl/preset"
)

func main() {
	runner := sexec.NewOSRunner()
	rootCmd, err := cmd.NewRootCommand(cmd.Dependencies{
		Bluetooth:   bluetooth.NewExecService(runner),
		Audio:       audio.NewExecService(runner),
		PresetStore: preset.NewStore(""),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize root command: %v\n", err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
