---
Title: SoundCtl Glazed CLI Verb Set and Integration Plan
Ticket: BT-001-HEADPHONES
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md
      Note: TUI architecture counterpart for CLI parity
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md
      Note: Implementation history and rationale
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md
      Note: Imported source spec driving command parity
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:09:01.116083267-05:00
WhatFor: ""
WhenToUse: ""
---


# SoundCtl Glazed CLI Verb Set and Integration Plan

## Executive Summary

This document defines a Glazed-based CLI surface for SoundCtl that mirrors the TUI capabilities and is script-friendly.

The CLI is organized into explicit command groups (`devices`, `sinks`, `profiles`, `scan`, `watch`) and built with the Glazed patterns from the `glazed-command-authoring` skill:
- command struct embedding `*cmds.CommandDescription`
- typed settings struct with `glazed` tags
- decode via `vals.DecodeSectionInto(schema.DefaultSlug, settings)`
- standardized Glazed sections and Cobra wiring

Goal: provide reliable non-interactive verbs and machine-readable output (`--output json|yaml|table`) while keeping behavior aligned with the TUI state model.

## Problem Statement

Current work defines a TUI and runtime behavior, but there is no complementary CLI contract for:
- automation (scripts, cron, CI diagnostics),
- deterministic one-shot operations (`connect`, `set-default`, `set-profile`),
- streaming observability (`watch bluetooth`, `watch audio`).

Without a defined verb set and Glazed command architecture, CLI behavior will diverge from the TUI and become difficult to maintain.

## Proposed Solution

### CLI Shape

Use root command `soundctl` with grouped verbs:

1. `soundctl devices list`
2. `soundctl devices connect --addr <MAC>`
3. `soundctl devices disconnect --addr <MAC>`
4. `soundctl devices trust --addr <MAC>`
5. `soundctl devices forget --addr <MAC>`
6. `soundctl scan start|stop`
7. `soundctl scan pair --addr <MAC> [--trust] [--connect]`
8. `soundctl sinks list`
9. `soundctl sinks set-default --sink <name-or-id>`
10. `soundctl sinks move-stream --stream-id <id> --sink <name-or-id>`
11. `soundctl sources list`
12. `soundctl sources set-default --source <name-or-id>`
13. `soundctl profiles list`
14. `soundctl profiles set --card <card> --profile <profile>`
15. `soundctl volume get|set --target <sink|source> [--name X] [--percent N]`
16. `soundctl mute toggle --target <sink|source> --name <id>`
17. `soundctl watch bluetooth`
18. `soundctl watch audio`
19. `soundctl watch all`

### Verb Mapping to TUI Actions

- TUI device row actions -> `devices connect|disconnect|trust|forget`
- scanner overlay actions -> `scan start|stop|pair`
- sinks pane actions -> `sinks set-default`, `sources set-default`, `sinks move-stream`
- profiles pane action -> `profiles set`
- volume controls -> `volume set`, `mute toggle`
- subscription feeds -> `watch *`

### Command Contract Conventions

- All list/watch commands output rows suitable for Glaze processors.
- Mutating commands return one result row with status fields:
  - `ok` (bool)
  - `operation`
  - `target`
  - `error` (string, empty on success)
- Common flags on relevant commands:
  - `--timeout` (duration/int seconds)
  - `--watch` where continuous mode is meaningful
  - `--json-errors` for machine-readable failures (optional v2)

### Glazed Authoring Pattern (Canonical for this project)

```go
type DevicesConnectCommand struct {
    *cmds.CommandDescription
    bt services.BluetoothService
}

type DevicesConnectSettings struct {
    Addr    string `glazed:"addr"`
    Timeout int    `glazed:"timeout"`
}

func NewDevicesConnectCommand(bt services.BluetoothService) (*DevicesConnectCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    commandSettingsSection, err := cli.NewCommandSettingsSection()
    if err != nil {
        return nil, err
    }

    cmdDesc := cmds.NewCommandDescription(
        "connect",
        cmds.WithShort("Connect a Bluetooth device"),
        cmds.WithLong("Connect a device by MAC address.\n\nExamples:\n  soundctl devices connect --addr 08:FF:44:2B:4C:90"),
        cmds.WithFlags(
            fields.New("addr", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Bluetooth MAC address")),
            fields.New("timeout", fields.TypeInteger, fields.WithDefault(15), fields.WithHelp("Timeout in seconds")),
        ),
        cmds.WithSections(glazedSection, commandSettingsSection),
        cmds.WithParents("devices"),
    )

    return &DevicesConnectCommand{CommandDescription: cmdDesc, bt: bt}, nil
}

func (c *DevicesConnectCommand) RunIntoGlazeProcessor(
    ctx context.Context, vals *values.Values, gp middlewares.Processor,
) error {
    s := &DevicesConnectSettings{}
    if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
        return err
    }
    err := c.bt.Connect(ctx, s.Addr)
    return gp.AddRow(ctx, types.NewRow(
        types.MRP("operation", "devices.connect"),
        types.MRP("target", s.Addr),
        types.MRP("ok", err == nil),
        types.MRP("error", errorString(err)),
    ))
}
```

### Streaming Defaults for Watch Commands

Use output defaults for watch-style commands:

```go
glazedSection, err := settings.NewGlazedSchema(
    settings.WithOutputSectionOptions(
        schema.WithDefaults(map[string]interface{}{
            "output": "yaml",
            "stream": true,
        }),
    ),
)
```

This makes `soundctl watch bluetooth` immediately useful in terminal pipelines.

### Cobra Integration

- Build each Glazed command with `cli.BuildCobraCommandFromCommand`.
- Use explicit Cobra parent groups (`devices`, `scan`, `sinks`, `sources`, `profiles`, `watch`) for discoverability.
- Root command should initialize logging via `logging.InitLoggerFromCobra`.

## Design Decisions

1. Explicit grouped verbs over a flat command list.
Reason: easier discovery and parity with TUI panes.

2. One file per verb and one `root.go` per group.
Reason: aligns with skill guidance and reduces merge conflicts.

3. Decode settings from Glazed section only.
Reason: avoids split-brain between Cobra flags and Glazed values.

4. DBus/Pulse service interfaces reused from TUI architecture.
Reason: single domain/runtime layer supports both TUI and CLI.

5. Structured result rows for mutating commands.
Reason: script consumers can branch on `ok` and parse `error`.

6. Dedicated `watch` group with stream-oriented defaults.
Reason: separates long-running event feeds from one-shot commands.

## Alternatives Considered

1. Implement CLI directly in Cobra without Glazed.
Rejected: loses uniform output/middleware model and schema-driven flags.

2. Add a single `exec` command that proxies arbitrary `bluetoothctl`/`pactl`.
Rejected: poor UX, no stable contract, high parser/security risk.

3. TUI-only product surface.
Rejected: blocks automation and repeatable diagnostics workflows.

4. Metadata-based parent grouping only (`cmds.WithParents`) without explicit Cobra groups.
Rejected for v1: explicit groups give clearer help and structured registration for larger command sets.

## Implementation Plan

### Phase A: CLI scaffolding

1. Create root package and groups:
   - `cmd/soundctl/main.go`
   - `pkg/cmd/root.go`
   - `pkg/cmd/devices/root.go`
   - `pkg/cmd/scan/root.go`
   - `pkg/cmd/sinks/root.go`
   - `pkg/cmd/sources/root.go`
   - `pkg/cmd/profiles/root.go`
   - `pkg/cmd/watch/root.go`
2. Wire logging section and persistent pre-run logger init.
3. Wire help system at root.

### Phase B: Read-first commands

1. Implement:
   - `devices list`
   - `sinks list`
   - `sources list`
   - `profiles list`
2. Validate `--output table|json|yaml` parity.
3. Add snapshot tests for row schemas.

### Phase C: Mutating commands

1. Implement:
   - `devices connect|disconnect|trust|forget`
   - `scan pair`
   - `sinks set-default`
   - `sources set-default`
   - `sinks move-stream`
   - `profiles set`
   - `volume set`
   - `mute toggle`
2. Add timeout handling and deterministic error rows.

### Phase D: Streaming commands

1. Implement `watch bluetooth|audio|all` using event channels.
2. Ensure clean cancellation on SIGINT/context cancel.
3. Add integration tests with mocked event streams.

### Phase E: UX hardening

1. Improve long help examples per command.
2. Add command-specific output defaults where needed.
3. Add command aliases only where unambiguous.

### Suggested command package layout

```text
pkg/cmd/root.go
pkg/cmd/devices/root.go
pkg/cmd/devices/list.go
pkg/cmd/devices/connect.go
pkg/cmd/devices/disconnect.go
pkg/cmd/devices/trust.go
pkg/cmd/devices/forget.go
pkg/cmd/scan/root.go
pkg/cmd/scan/start.go
pkg/cmd/scan/stop.go
pkg/cmd/scan/pair.go
pkg/cmd/sinks/root.go
pkg/cmd/sinks/list.go
pkg/cmd/sinks/set_default.go
pkg/cmd/sinks/move_stream.go
pkg/cmd/sources/root.go
pkg/cmd/sources/list.go
pkg/cmd/sources/set_default.go
pkg/cmd/profiles/root.go
pkg/cmd/profiles/list.go
pkg/cmd/profiles/set.go
pkg/cmd/watch/root.go
pkg/cmd/watch/bluetooth.go
pkg/cmd/watch/audio.go
pkg/cmd/watch/all.go
```

## Open Questions

1. Should `scan start` be exposed publicly, or should `scan pair` own scan lifecycle internally?
2. Should `devices connect` optionally auto-`trust` on success (`--auto-trust`)?
3. Do we want stable numeric IDs for sinks/sources in command UX, or only names and native IDs?
4. Should we include a `doctor` command for host checks (`rfkill`, service status, adapter state`) in v1?
5. What output format should be default for list commands (`table` vs `yaml`)?

## References

- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md`
- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md`
- `/home/manuel/.codex/skills/glazed-command-authoring/SKILL.md`
