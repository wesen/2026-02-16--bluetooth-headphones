package parse

import "testing"

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
