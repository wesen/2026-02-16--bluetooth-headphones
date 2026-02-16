package preset

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(filepath.Join(dir, "presets.yaml"))
}

func TestSaveAndList(t *testing.T) {
	store := tempStore(t)

	p := Preset{
		Name:         "Music Mode",
		CardProfiles: map[string]string{"bluez_card.sony": "a2dp-sink"},
		Volumes:      map[string]VolumeSpec{"Master": {Level: 80, Muted: false}},
		DefaultSink:  "bluez_sink.sony",
		AppRoutes:    map[string]string{"Firefox": "follow_default"},
	}

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	presets, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(presets))
	}
	if presets[0].Name != "Music Mode" {
		t.Fatalf("expected Music Mode, got %q", presets[0].Name)
	}
	if presets[0].CardProfiles["bluez_card.sony"] != "a2dp-sink" {
		t.Fatalf("expected a2dp-sink profile")
	}
	if presets[0].Volumes["Master"].Level != 80 {
		t.Fatalf("expected Master volume 80")
	}
	if presets[0].CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestSaveUpdate(t *testing.T) {
	store := tempStore(t)

	p1 := Preset{Name: "Test", DefaultSink: "sink-a"}
	if err := store.Save(p1); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Update same name
	p2 := Preset{Name: "Test", DefaultSink: "sink-b"}
	if err := store.Save(p2); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	presets, _ := store.List()
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset after update, got %d", len(presets))
	}
	if presets[0].DefaultSink != "sink-b" {
		t.Fatalf("expected sink-b, got %q", presets[0].DefaultSink)
	}
}

func TestSavePreservesCreatedAt(t *testing.T) {
	store := tempStore(t)

	p1 := Preset{Name: "Test", DefaultSink: "sink-a"}
	if err := store.Save(p1); err != nil {
		t.Fatalf("Save: %v", err)
	}
	presets, _ := store.List()
	createdAt := presets[0].CreatedAt

	// Update
	p2 := Preset{Name: "Test", DefaultSink: "sink-b"}
	if err := store.Save(p2); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	presets, _ = store.List()
	if !presets[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt preserved, got %v vs %v", presets[0].CreatedAt, createdAt)
	}
}

func TestGet(t *testing.T) {
	store := tempStore(t)

	store.Save(Preset{Name: "Alpha"})
	store.Save(Preset{Name: "Beta"})

	p, err := store.Get("Beta")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Name != "Beta" {
		t.Fatalf("expected Beta, got %q", p.Name)
	}

	_, err = store.Get("Gamma")
	if err == nil {
		t.Fatal("expected error for missing preset")
	}
}

func TestDelete(t *testing.T) {
	store := tempStore(t)

	store.Save(Preset{Name: "A"})
	store.Save(Preset{Name: "B"})
	store.Save(Preset{Name: "C"})

	if err := store.Delete("B"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	presets, _ := store.List()
	if len(presets) != 2 {
		t.Fatalf("expected 2 presets after delete, got %d", len(presets))
	}
	for _, p := range presets {
		if p.Name == "B" {
			t.Fatal("B should have been deleted")
		}
	}
}

func TestDeleteNotFound(t *testing.T) {
	store := tempStore(t)
	store.Save(Preset{Name: "A"})

	if err := store.Delete("Z"); err == nil {
		t.Fatal("expected error deleting non-existent preset")
	}
}

func TestListEmptyFile(t *testing.T) {
	store := tempStore(t)
	presets, err := store.List()
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(presets) != 0 {
		t.Fatalf("expected 0 presets, got %d", len(presets))
	}
}

func TestFileRoundTrip(t *testing.T) {
	store := tempStore(t)

	p := Preset{
		Name:         "Full",
		CardProfiles: map[string]string{"card1": "profile1", "card2": "profile2"},
		Volumes: map[string]VolumeSpec{
			"Master": {Level: 72, Muted: false},
			"Alerts": {Level: 0, Muted: true},
		},
		DefaultSink: "my-sink",
		AppRoutes: map[string]string{
			"Firefox": "my-sink",
			"Discord": "follow_default",
		},
	}
	store.Save(p)

	// Read the raw YAML to verify format
	data, _ := os.ReadFile(store.Path())
	if len(data) == 0 {
		t.Fatal("expected non-empty file")
	}

	// Re-read via store
	presets, _ := store.List()
	if len(presets) != 1 {
		t.Fatalf("expected 1, got %d", len(presets))
	}
	got := presets[0]
	if got.CardProfiles["card1"] != "profile1" {
		t.Fatal("card1 profile mismatch")
	}
	if got.Volumes["Alerts"].Muted != true {
		t.Fatal("Alerts should be muted")
	}
	if got.AppRoutes["Discord"] != "follow_default" {
		t.Fatal("Discord route mismatch")
	}
}

func TestMultiplePresets(t *testing.T) {
	store := tempStore(t)

	store.Save(Preset{Name: "Music", DefaultSink: "bt"})
	store.Save(Preset{Name: "Video Call", DefaultSink: "bt"})
	store.Save(Preset{Name: "Speakers", DefaultSink: "alsa"})
	store.Save(Preset{Name: "Late Night", DefaultSink: "bt"})

	presets, _ := store.List()
	if len(presets) != 4 {
		t.Fatalf("expected 4 presets, got %d", len(presets))
	}
}
