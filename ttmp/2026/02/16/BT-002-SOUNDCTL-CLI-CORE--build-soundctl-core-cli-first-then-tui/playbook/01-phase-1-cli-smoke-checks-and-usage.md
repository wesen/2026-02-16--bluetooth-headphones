---
Title: Phase 1 CLI Smoke Checks and Usage
Ticket: BT-002-SOUNDCTL-CLI-CORE
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:24:30.809065554-05:00
WhatFor: ""
WhenToUse: ""
---

# Phase 1 CLI Smoke Checks and Usage

## Purpose

Validate the CLI-first SoundCtl Phase 1 implementation and provide copy/paste examples for core Bluetooth/audio operations.

## Environment Assumptions

- Go 1.25+ installed
- `bluetoothctl` and `pactl` available in `PATH`
- Built from repo root `/home/manuel/code/wesen/2026-02-16--bluetooth-headphones`

## Commands

### 1) Unit tests

```bash
go test ./...
```

### 2) CLI help sanity

```bash
go run ./cmd/soundctl --help
go run ./cmd/soundctl devices --help
go run ./cmd/soundctl sinks --help
```

### 3) Read-path smoke commands

```bash
go run ./cmd/soundctl devices list --output json
go run ./cmd/soundctl sinks list --output json
go run ./cmd/soundctl sources list --output json
go run ./cmd/soundctl profiles list --output json
```

### 4) Mutation-path smoke commands (safe validation path)

Use a validation-failure path to verify command plumbing without applying system changes:

```bash
go run ./cmd/soundctl volume set --target invalid --name foo --percent 10 --output json
```

Expected result: non-zero exit with validation error about target.

### 5) Live mutation examples (real system effect)

Run only when you intend to change system state:

```bash
# bluetooth
go run ./cmd/soundctl devices connect --addr 08:FF:44:2B:4C:90 --output json
go run ./cmd/soundctl devices disconnect --addr 08:FF:44:2B:4C:90 --output json
go run ./cmd/soundctl scan pair --addr 08:FF:44:2B:4C:90 --trust --connect --output json

# audio
go run ./cmd/soundctl sinks set-default --sink alsa_output.pci-0000_00_1f.3.analog-stereo --output json
go run ./cmd/soundctl volume set --target sink --name alsa_output.pci-0000_00_1f.3.analog-stereo --percent 35 --output json
go run ./cmd/soundctl mute toggle --target sink --name alsa_output.pci-0000_00_1f.3.analog-stereo --output json
```

## Exit Criteria

- `go test ./...` passes.
- CLI help works for root and representative groups.
- Read-path commands return structured data rows.
- Validation errors return clear non-zero failures.

## Notes

- Core logic is in `pkg/soundctl/*`; CLI commands are thin wrappers in `pkg/cmd/*`.
- TUI work remains deferred to Phase 2 and should reuse the same `pkg/soundctl/*` services.
