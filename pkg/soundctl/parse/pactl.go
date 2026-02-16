package parse

import (
	"fmt"
	"strconv"
	"strings"
)

// PactlShortRecord matches a row from `pactl list short ...`.
type PactlShortRecord struct {
	ID         int
	Name       string
	Driver     string
	SampleSpec string
	State      string
}

// PactlInfoRecord captures the default sink/source from `pactl info`.
type PactlInfoRecord struct {
	DefaultSinkName   string
	DefaultSourceName string
	ServerName        string
}

// PactlSinkInputRecord captures a sink-input (app stream) from `pactl list sink-inputs`.
type PactlSinkInputRecord struct {
	Index     int
	SinkIndex int
	AppName   string
	MediaName string
	SinkName  string // populated externally
}

// PactlCardRecord captures a card with its available profiles from `pactl list cards`.
type PactlCardRecord struct {
	Index         int
	Name          string
	Driver        string
	Profiles      []PactlProfileRecord
	ActiveProfile string
}

// PactlProfileRecord captures a profile within a card.
type PactlProfileRecord struct {
	Name        string
	Description string
	Available   bool
}

// ParsePactlInfo parses `pactl info` output for default sink/source.
func ParsePactlInfo(output string) (PactlInfoRecord, error) {
	var info PactlInfoRecord
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "Default Sink:") {
			info.DefaultSinkName = strings.TrimSpace(strings.TrimPrefix(line, "Default Sink:"))
		}
		if strings.HasPrefix(line, "Default Source:") {
			info.DefaultSourceName = strings.TrimSpace(strings.TrimPrefix(line, "Default Source:"))
		}
		if strings.HasPrefix(line, "Server Name:") {
			info.ServerName = strings.TrimSpace(strings.TrimPrefix(line, "Server Name:"))
		}
	}
	return info, nil
}

// ParsePactlSinkInputs parses `pactl list sink-inputs` output.
func ParsePactlSinkInputs(output string) ([]PactlSinkInputRecord, error) {
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var records []PactlSinkInputRecord
	var current *PactlSinkInputRecord
	inProperties := false

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)

		if strings.HasPrefix(line, "Sink Input #") {
			if current != nil {
				records = append(records, *current)
			}
			idx, _ := strconv.Atoi(strings.TrimPrefix(line, "Sink Input #"))
			current = &PactlSinkInputRecord{Index: idx}
			inProperties = false
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(line, "Sink:") {
			current.SinkIndex, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Sink:")))
		}
		if line == "Properties:" {
			inProperties = true
			continue
		}
		if inProperties {
			if strings.HasPrefix(line, "application.name = ") {
				current.AppName = trimQuotes(strings.TrimPrefix(line, "application.name = "))
			}
			if strings.HasPrefix(line, "media.name = ") {
				current.MediaName = trimQuotes(strings.TrimPrefix(line, "media.name = "))
			}
			// Properties section ends when next non-indented section starts
			if !strings.HasPrefix(raw, "\t\t") && !strings.HasPrefix(raw, "    ") && line != "" && !strings.Contains(line, "=") {
				inProperties = false
			}
		}
	}
	if current != nil {
		records = append(records, *current)
	}
	return records, nil
}

// ParsePactlCards parses `pactl list cards` output for full card+profile details.
func ParsePactlCards(output string) ([]PactlCardRecord, error) {
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var cards []PactlCardRecord
	var current *PactlCardRecord
	inProfiles := false

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)

		if strings.HasPrefix(line, "Card #") {
			if current != nil {
				cards = append(cards, *current)
			}
			idx, _ := strconv.Atoi(strings.TrimPrefix(line, "Card #"))
			current = &PactlCardRecord{Index: idx}
			inProfiles = false
			continue
		}
		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "Name:") {
			current.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
			inProfiles = false
		} else if strings.HasPrefix(line, "Driver:") {
			current.Driver = strings.TrimSpace(strings.TrimPrefix(line, "Driver:"))
			inProfiles = false
		} else if strings.HasPrefix(line, "Active Profile:") {
			current.ActiveProfile = strings.TrimSpace(strings.TrimPrefix(line, "Active Profile:"))
			inProfiles = false
		} else if line == "Profiles:" {
			inProfiles = true
		} else if inProfiles {
			// Profile lines look like:
			//   output:analog-stereo: Analog Stereo Output (sinks: 1, sources: 0, priority: 6500, available: yes)
			//   off: Off (sinks: 0, sources: 0, priority: 0, available: yes)
			if strings.Contains(line, ": ") && !strings.HasPrefix(line, "Part of") {
				colonIdx := strings.Index(line, ": ")
				if colonIdx > 0 {
					profName := line[:colonIdx]
					rest := line[colonIdx+2:]
					desc := rest
					available := true

					// Extract description (before parenthesized details)
					if parenIdx := strings.Index(rest, " ("); parenIdx > 0 {
						desc = rest[:parenIdx]
						if strings.Contains(rest, "available: no") {
							available = false
						}
					}

					current.Profiles = append(current.Profiles, PactlProfileRecord{
						Name:        profName,
						Description: desc,
						Available:   available,
					})
				}
			} else {
				// Non-profile line: end profiles section
				inProfiles = false
			}
		}
	}
	if current != nil {
		cards = append(cards, *current)
	}
	return cards, nil
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func ParsePactlShort(output string) ([]PactlShortRecord, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []PactlShortRecord{}, nil
	}

	rows := make([]PactlShortRecord, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 2 {
			return nil, fmt.Errorf("invalid pactl short row: %q", line)
		}
		id, err := strconv.Atoi(cols[0])
		if err != nil {
			return nil, fmt.Errorf("invalid pactl id %q: %w", cols[0], err)
		}
		rec := PactlShortRecord{ID: id, Name: cols[1]}
		if len(cols) > 2 {
			rec.Driver = cols[2]
		}
		if len(cols) > 3 {
			rec.SampleSpec = cols[3]
		}
		if len(cols) > 4 {
			rec.State = cols[4]
		}
		rows = append(rows, rec)
	}
	return rows, nil
}
