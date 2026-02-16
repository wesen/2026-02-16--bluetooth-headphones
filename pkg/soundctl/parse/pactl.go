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
