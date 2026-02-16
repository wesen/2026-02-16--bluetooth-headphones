package parse

import (
	"testing"
)

func TestParsePactlInfo(t *testing.T) {
	input := `Server String: /run/user/1000/pulse/native
Library Protocol Version: 35
Server Protocol Version: 35
Is Local: yes
Client Index: 42
Tile Size: 65536
User Name: user
Host Name: box
Server Name: PulseAudio (on PipeWire 0.3.77)
Server Version: 15.0.0
Default Sample Specification: float32le 2ch 48000Hz
Default Channel Map: front-left,front-right
Default Sink: alsa_output.pci-0000_00_1f.3.analog-stereo
Default Source: alsa_input.pci-0000_00_1f.3.analog-stereo
Cookie: 1234:abcd`

	info, err := ParsePactlInfo(input)
	if err != nil {
		t.Fatalf("ParsePactlInfo: %v", err)
	}
	if info.DefaultSinkName != "alsa_output.pci-0000_00_1f.3.analog-stereo" {
		t.Fatalf("unexpected default sink: %s", info.DefaultSinkName)
	}
	if info.DefaultSourceName != "alsa_input.pci-0000_00_1f.3.analog-stereo" {
		t.Fatalf("unexpected default source: %s", info.DefaultSourceName)
	}
	if info.ServerName != "PulseAudio (on PipeWire 0.3.77)" {
		t.Fatalf("unexpected server name: %s", info.ServerName)
	}
}

func TestParsePactlSinkInputs(t *testing.T) {
	input := `Sink Input #57
	Driver: PipeWire
	Owner Module: n/a
	Client: 38
	Sink: 47
	Sample Specification: float32le 2ch 48000Hz
	Channel Map: front-left,front-right
	Format: pcm, format.sample_format = "\"float32le\""
	Corked: no
	Mute: no
	Volume: front-left: 65536 / 100% / 0.00 dB,   front-right: 65536 / 100% / 0.00 dB
	        balance 0.00
	Buffer Latency: 0 usec
	Sink Latency: 0 usec
	Resample method: PipeWire
	Properties:
		media.name = "Playback"
		application.name = "Firefox"
		node.name = "Firefox"
Sink Input #63
	Driver: PipeWire
	Sink: 47
	Properties:
		media.name = "Music"
		application.name = "Spotify"
		node.name = "spotify"`

	records, err := ParsePactlSinkInputs(input)
	if err != nil {
		t.Fatalf("ParsePactlSinkInputs: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 sink inputs, got %d", len(records))
	}
	if records[0].Index != 57 {
		t.Fatalf("expected index 57, got %d", records[0].Index)
	}
	if records[0].AppName != "Firefox" {
		t.Fatalf("expected Firefox, got %s", records[0].AppName)
	}
	if records[0].SinkIndex != 47 {
		t.Fatalf("expected sink 47, got %d", records[0].SinkIndex)
	}
	if records[1].AppName != "Spotify" {
		t.Fatalf("expected Spotify, got %s", records[1].AppName)
	}
}

func TestParsePactlCards(t *testing.T) {
	input := `Card #47
	Name: alsa_card.pci-0000_00_1f.3
	Driver: module-alsa-card.c
	Owner Module: 6
	Properties:
		alsa.card = "0"
	Profiles:
		output:analog-stereo: Analog Stereo Output (sinks: 1, sources: 0, priority: 6500, available: yes)
		output:analog-stereo+input:analog-stereo: Analog Stereo Duplex (sinks: 1, sources: 1, priority: 6565, available: yes)
		off: Off (sinks: 0, sources: 0, priority: 0, available: yes)
	Active Profile: output:analog-stereo+input:analog-stereo
	Ports:
		analog-output: Analog Output (type: Unknown, priority: 9900, latency offset: 0 usec, availability unknown)
Card #62
	Name: bluez_card.AA_BB_CC_DD_EE_FF
	Driver: module-bluez5-device.c
	Profiles:
		a2dp-sink: High Fidelity Playback (A2DP Sink, codec SBC) (sinks: 1, sources: 0, priority: 40, available: yes)
		headset-head-unit: Headset Head Unit (HSP/HFP) (sinks: 1, sources: 1, priority: 30, available: yes)
		off: Off (sinks: 0, sources: 0, priority: 0, available: yes)
	Active Profile: a2dp-sink`

	cards, err := ParsePactlCards(input)
	if err != nil {
		t.Fatalf("ParsePactlCards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}

	// Card 1
	if cards[0].Name != "alsa_card.pci-0000_00_1f.3" {
		t.Fatalf("unexpected card name: %s", cards[0].Name)
	}
	if cards[0].ActiveProfile != "output:analog-stereo+input:analog-stereo" {
		t.Fatalf("unexpected active profile: %s", cards[0].ActiveProfile)
	}
	if len(cards[0].Profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(cards[0].Profiles))
	}
	if cards[0].Profiles[0].Name != "output:analog-stereo" {
		t.Fatalf("unexpected profile name: %s", cards[0].Profiles[0].Name)
	}
	if cards[0].Profiles[0].Description != "Analog Stereo Output" {
		t.Fatalf("unexpected profile desc: %s", cards[0].Profiles[0].Description)
	}

	// Card 2
	if cards[1].Name != "bluez_card.AA_BB_CC_DD_EE_FF" {
		t.Fatalf("unexpected card2 name: %s", cards[1].Name)
	}
	if cards[1].ActiveProfile != "a2dp-sink" {
		t.Fatalf("unexpected card2 active profile: %s", cards[1].ActiveProfile)
	}
	if len(cards[1].Profiles) != 3 {
		t.Fatalf("expected 3 profiles for card2, got %d", len(cards[1].Profiles))
	}
}

func TestParsePactlShort(t *testing.T) {
	input := "47\talsa_output.pci-0000_00_1f.3.analog-stereo\tPipeWire\ts32le 2ch 48000Hz\tRUNNING"
	rows, err := ParsePactlShort(input)
	if err != nil {
		t.Fatalf("ParsePactlShort returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].ID != 47 {
		t.Fatalf("unexpected id: %d", rows[0].ID)
	}
	if rows[0].Name != "alsa_output.pci-0000_00_1f.3.analog-stereo" {
		t.Fatalf("unexpected name: %s", rows[0].Name)
	}
	if rows[0].State != "RUNNING" {
		t.Fatalf("unexpected state: %s", rows[0].State)
	}
}
