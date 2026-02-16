// Package preset provides preset data types and YAML persistence
// at ~/.config/soundctl/presets.yaml.
package preset

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// VolumeSpec defines volume+mute state for a single channel.
type VolumeSpec struct {
	Level int  `yaml:"level"`
	Muted bool `yaml:"muted"`
}

// Preset captures a named audio configuration snapshot.
type Preset struct {
	Name         string                `yaml:"name"`
	CardProfiles map[string]string     `yaml:"card_profiles"` // card name → profile name
	Volumes      map[string]VolumeSpec `yaml:"volumes"`       // channel name → {level, muted}
	DefaultSink  string                `yaml:"default_sink"`
	AppRoutes    map[string]string     `yaml:"app_routes"` // app name → sink name | "follow_default"
	CreatedAt    time.Time             `yaml:"created_at"`
	UpdatedAt    time.Time             `yaml:"updated_at"`
}

// DiffLine describes a single change when applying a preset.
type DiffLine struct {
	Field string `yaml:"field"`
	From  string `yaml:"from"`
	To    string `yaml:"to"`
}

// ── Persistence ────────────────────────────────────────────────────────────

// Store manages preset persistence to a YAML file.
type Store struct {
	mu   sync.RWMutex
	path string
}

// NewStore creates a store at the given path.
// If path is "", it defaults to ~/.config/soundctl/presets.yaml.
func NewStore(path string) *Store {
	if path == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			cfgDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
		path = filepath.Join(cfgDir, "soundctl", "presets.yaml")
	}
	return &Store{path: path}
}

// Path returns the file path used by this store.
func (s *Store) Path() string {
	return s.path
}

// fileData is the YAML root structure.
type fileData struct {
	Presets []Preset `yaml:"presets"`
}

// List returns all saved presets.
func (s *Store) List() ([]Preset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readFile()
}

// Get returns a preset by name.
func (s *Store) Get(name string) (Preset, error) {
	presets, err := s.List()
	if err != nil {
		return Preset{}, err
	}
	for _, p := range presets {
		if p.Name == name {
			return p, nil
		}
	}
	return Preset{}, fmt.Errorf("preset %q not found", name)
}

// Save writes or updates a preset. If a preset with the same name exists,
// it is replaced. Otherwise a new entry is appended.
func (s *Store) Save(p Preset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	presets, err := s.readFile()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	found := false
	for i, existing := range presets {
		if existing.Name == p.Name {
			p.CreatedAt = existing.CreatedAt // preserve original creation time
			presets[i] = p
			found = true
			break
		}
	}
	if !found {
		presets = append(presets, p)
	}

	return s.writeFile(presets)
}

// Delete removes a preset by name.
func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	presets, err := s.readFile()
	if err != nil {
		return err
	}

	filtered := make([]Preset, 0, len(presets))
	found := false
	for _, p := range presets {
		if p.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, p)
	}
	if !found {
		return fmt.Errorf("preset %q not found", name)
	}

	return s.writeFile(filtered)
}

func (s *Store) readFile() ([]Preset, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read presets file: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var fd fileData
	if err := yaml.Unmarshal(data, &fd); err != nil {
		return nil, fmt.Errorf("parse presets file: %w", err)
	}
	return fd.Presets, nil
}

func (s *Store) writeFile(presets []Preset) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	fd := fileData{Presets: presets}
	data, err := yaml.Marshal(&fd)
	if err != nil {
		return fmt.Errorf("marshal presets: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write presets file: %w", err)
	}
	return nil
}
