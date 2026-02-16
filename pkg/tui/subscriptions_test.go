package tui

import (
	"testing"
)

func TestParsePactlSubscribeLine(t *testing.T) {
	tests := []struct {
		line     string
		wantType string
		wantFac  string
	}{
		{
			line:     "Event 'change' on sink #47",
			wantType: "change",
			wantFac:  "sink",
		},
		{
			line:     "Event 'new' on sink-input #63",
			wantType: "new",
			wantFac:  "sink-input",
		},
		{
			line:     "Event 'remove' on source #2",
			wantType: "remove",
			wantFac:  "source",
		},
		{
			line:     "Event 'change' on card #0",
			wantType: "change",
			wantFac:  "card",
		},
		{
			line:     "not an event line",
			wantType: "",
			wantFac:  "",
		},
		{
			line:     "",
			wantType: "",
			wantFac:  "",
		},
	}

	for _, tt := range tests {
		ev := parsePactlSubscribeLine(tt.line)
		if ev.EventType != tt.wantType {
			t.Errorf("line=%q: EventType=%q, want=%q", tt.line, ev.EventType, tt.wantType)
		}
		if ev.Facility != tt.wantFac {
			t.Errorf("line=%q: Facility=%q, want=%q", tt.line, ev.Facility, tt.wantFac)
		}
	}
}

func TestParseDbusMonitorLine(t *testing.T) {
	tests := []struct {
		line     string
		wantType string
	}{
		{
			line:     "signal time=1234 sender=:1.4 -> destination=(null) serial=42 path=/org/bluez/hci0/dev_AA_BB; interface=org.freedesktop.DBus.Properties; member=PropertiesChanged",
			wantType: "property-changed",
		},
		{
			line:     "signal sender=:1.4 path=/; interface=org.freedesktop.DBus.ObjectManager; member=InterfacesAdded",
			wantType: "device-added",
		},
		{
			line:     "signal sender=:1.4 path=/; interface=org.freedesktop.DBus.ObjectManager; member=InterfacesRemoved",
			wantType: "device-removed",
		},
		{
			line:     "just some random output",
			wantType: "",
		},
	}

	for _, tt := range tests {
		ev := parseDbusMonitorLine(tt.line)
		if ev.EventType != tt.wantType {
			t.Errorf("line=%q: EventType=%q, want=%q", tt.line, ev.EventType, tt.wantType)
		}
	}
}
