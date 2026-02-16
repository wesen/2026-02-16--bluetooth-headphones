---
Title: SoundCtl Bubble Tea/Bubbles Implementation Guide
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
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/index.md
      Note: Ticket-level navigation to implementation guide
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md
      Note: Work log and command outcomes for import + analysis
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md
      Note: Imported source spec with screens and message DSL
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:04:53.295362912-05:00
WhatFor: ""
WhenToUse: ""
---


# SoundCtl Bubble Tea/Bubbles Implementation Guide

## Executive Summary

This guide maps the imported `SoundCtl` spec (`sources/local/headphones.md`) to a production-ready Bubble Tea + Bubbles terminal app architecture.

The proposed design uses:
- A root `AppModel` that owns shared state and routes messages.
- Child pane models for Devices, Sinks, Profiles, Scanner overlay, and Status.
- Typed service adapters for Bluetooth and audio, with long-lived event subscriptions.
- A command orchestration layer that keeps shell/DBus I/O outside UI update logic.

The result is a responsive TUI that matches the four target screens, supports live updates from BlueZ/PulseAudio, and remains testable without real hardware.

## Problem Statement

The imported spec defines rich multi-pane behavior, real-time subscriptions, and command-triggered side effects. Directly implementing this in one Bubble Tea model would create tight coupling between:
- view rendering,
- key handling,
- side effects (`bluetoothctl`, `pactl`, DBus monitors),
- and asynchronous updates.

Without clear boundaries, you get fragile parsing, hard-to-test update logic, and regressions when adding features like scanning overlays and app routing.

## Proposed Solution

### 1) Architecture: UI, Domain, Services, Runtime

Use four layers:

1. `domain`: pure structs and enums (`BluetoothDevice`, `Sink`, `CardProfile`, `Tab`).
2. `services`: interfaces for Bluetooth and audio operations/events.
3. `app`: Bubble Tea models, message types, update/view logic.
4. `runtime`: concrete adapters (DBus + `pactl`/`bluetoothctl` fallback), command execution, parser utilities.

This gives strict dependency direction: `app -> services interfaces -> runtime impl`.

### 2) Root Model and Child Models

Match the DSL in `headphones.md`:

- `AppModel`
  - `tabBar TabBarModel`
  - `devices DevicesPaneModel`
  - `sinks SinksPaneModel`
  - `profiles ProfilesPaneModel`
  - `scanner ScanOverlayModel`
  - `status StatusBarModel`
  - `width`, `height`, `ready`

Root responsibilities:
- own active tab + layout metrics,
- route global keybindings,
- fan-out domain events to panes,
- aggregate `tea.Cmd` from children,
- render composition with overlay priority.

### 3) Use Bubbles for Stateful Widgets

Recommended Bubbles usage:
- `list.Model` for device, sink, source, profile, and discovered-device collections.
- `spinner.Model` for scan overlay indicator.
- `progress.Model` for volume bars.
- `help.Model` and `key.Map` for contextual shortcuts.
- `viewport.Model` only if lists need scrollable detail panes/logs.

Keep Bubbles state inside child models. Root only coordinates.

### 4) Message Taxonomy

Define explicit app messages; avoid passing raw monitor lines into UI:

```go
type Msg interface{}

type BluetoothEventMsg struct{ Event BluetoothEvent }
type PulseAudioEventMsg struct{ Event AudioEvent }
type StatusMsg struct{ Text string }
type ErrorMsg struct{ Err error }
type TabChangedMsg struct{ Delta int }
type OpenScannerMsg struct{}
type CloseScannerMsg struct{}
type ConnectResultMsg struct{ Addr string; Err error }
type PairResultMsg struct{ Addr string; Err error }
type SetProfileResultMsg struct{ Card, Profile string; Err error }
```

Map key events to intent messages first, then to commands. This keeps keymaps configurable and testable.

### 5) Commands and Side Effects

Use typed commands through service interfaces instead of embedding shell strings in `Update`:

```go
type BluetoothService interface {
    Connect(ctx context.Context, addr string) error
    Disconnect(ctx context.Context, addr string) error
    Remove(ctx context.Context, addr string) error
    Pair(ctx context.Context, addr string) error
    Trust(ctx context.Context, addr string) error
    StartScan(ctx context.Context) error
    StopScan(ctx context.Context) error
    Events(ctx context.Context) (<-chan BluetoothEvent, error)
}
```

In UI models, wrap calls as `tea.Cmd`:

```go
func connectCmd(bt BluetoothService, addr string) tea.Cmd {
    return func() tea.Msg {
        err := bt.Connect(context.Background(), addr)
        return ConnectResultMsg{Addr: addr, Err: err}
    }
}
```

### 6) Long-Lived Subscriptions (Critical Pattern)

For monitor streams (`dbus-monitor`, `pactl subscribe`, or native DBus callbacks), use the standard Bubble Tea listen loop:

```go
func waitBluetooth(ch <-chan BluetoothEvent) tea.Cmd {
    return func() tea.Msg {
        ev, ok := <-ch
        if !ok {
            return ErrorMsg{Err: errors.New("bluetooth event channel closed")}
        }
        return BluetoothEventMsg{Event: ev}
    }
}
```

In `Update`, after handling `BluetoothEventMsg`, always return another `waitBluetooth(...)` to keep the subscription alive.

### 7) Overlay and Focus Rules

Behavioral contract:
- When scanner is visible, it captures key events (`enter`, `esc`, cursor movement).
- Underlying pane still receives data updates but not interactive keys.
- `esc` closes overlay and restores focus to prior pane/cursor.

This avoids focus bugs and matches Screen 4 expectations.

### 8) Rendering Strategy

Use `lipgloss` for layout and borders; render in deterministic order:

1. Header (`SoundCtl`, tabs)
2. Active pane
3. Status bar/help
4. Overlay on top (if visible)

Use width-aware truncation for long device names to avoid broken box drawing.

### 9) Backend Choice: DBus-first, CLI fallback

Prefer:
- BlueZ DBus API for device state/scanning/pairing.
- PulseAudio/PipeWire API (or `pactl` wrappers when fast iteration is preferred).

Use `bluetoothctl`/`pactl` command wrappers as fallback only.
Rationale: stable structured data, less parser breakage, easier test stubbing.

## Design Decisions

1. Typed message bus over raw strings.
Reason: clearer contracts and simpler tests.

2. Child models own their Bubbles controls.
Reason: keeps each pane locally coherent and reusable.

3. Event subscriptions implemented as channel-backed recurring commands.
Reason: idiomatic Bubble Tea streaming pattern.

4. Service interfaces for side effects.
Reason: hardware-free unit testing and runtime swap (mock vs real DBus).

5. Overlay-first key precedence.
Reason: avoids accidental background actions while scanning/pairing.

6. DBus-first integration.
Reason: richer signals and reduced shell parsing fragility.

## Alternatives Considered

1. Single mega-model with direct shell execution in `Update`.
Rejected: high coupling, hard to test, difficult to extend safely.

2. Polling-only architecture (periodic refresh).
Rejected: delayed UI updates and unnecessary process churn versus subscriptions.

3. Pure CLI wrappers (`bluetoothctl`/`pactl`) for everything.
Rejected: text parsing brittleness and missing structured event semantics.

4. One model per tab without root coordinator.
Rejected: global behaviors (status, overlay, help, tab changes) become duplicated.

## Implementation Plan

### Phase 0: Repo bootstrap

1. Create Go module and command entrypoint (`cmd/soundctl/main.go`).
2. Add dependencies: `bubbletea`, `bubbles`, `lipgloss`.
3. Add `internal/` layout and service interfaces.

### Phase 1: Static UI skeleton

1. Build root `AppModel` with tab bar + placeholder panes.
2. Implement global keybindings (`q`, `tab`, `shift+tab`, `?`).
3. Add responsive sizing via `tea.WindowSizeMsg`.

Exit criteria:
- All four screen layouts can be approximated with static data.

### Phase 2: Devices pane + scanner overlay

1. Implement `DevicesPaneModel` list and actions (connect/disconnect/forget).
2. Implement `ScanOverlayModel` with spinner + discovered list.
3. Wire overlay focus rules and key handling.
4. Add status message flow for result feedback.

Exit criteria:
- Can open scan overlay, select device, trigger pair/connect commands.

### Phase 3: Audio panes

1. Implement `SinksPaneModel` (outputs, inputs, app routes).
2. Implement `ProfilesPaneModel` (cards/profiles selection + apply).
3. Add volume controls + mute toggles in devices view.

Exit criteria:
- Can set defaults, move app streams, and switch profiles from TUI.

### Phase 4: Event subscriptions

1. Add Bluetooth event monitor adapter and parser.
2. Add PulseAudio subscription adapter and parser.
3. Implement recurring `waitBluetooth` / `waitPulse` commands.
4. Ensure event fan-out updates only affected panes.

Exit criteria:
- UI updates live when devices/sinks/profiles change externally.

### Phase 5: Hardening and UX polish

1. Standardize error/status timeout behavior.
2. Add contextual help panel and empty states.
3. Add reconnect/pairing conflict hints (for auto-switch contention).
4. Add graceful shutdown with context cancellation.

Exit criteria:
- Stable under rapid event bursts, no goroutine leaks, predictable focus behavior.

### Phase 6: Test strategy

1. Unit-test each pane `Update` with synthetic messages.
2. Snapshot-test key view states.
3. Integration-test root model with mocked services + synthetic event channels.
4. Add parser tests for monitor line fixtures.

Exit criteria:
- Deterministic tests for key workflows and regressions.

### Suggested file layout

```text
cmd/soundctl/main.go
internal/app/model.go
internal/app/messages.go
internal/app/keys.go
internal/app/layout.go
internal/app/status/model.go
internal/app/tabs/model.go
internal/app/panes/devices/model.go
internal/app/panes/sinks/model.go
internal/app/panes/profiles/model.go
internal/app/overlay/scanner/model.go
internal/domain/bluetooth.go
internal/domain/audio.go
internal/services/bluetooth.go
internal/services/audio.go
internal/runtime/bluez_dbus.go
internal/runtime/pulse_client.go
internal/runtime/exec_fallback.go
internal/runtime/parse/bluetooth_events.go
internal/runtime/parse/pactl_events.go
```

## Open Questions

1. Should the first release use DBus directly or start with `bluetoothctl`/`pactl` wrappers for speed?
2. Are per-app routing controls required in v1, or can they be deferred after core devices/sinks/profiles?
3. Should scanner pairing auto-run `trust + connect` after successful `pair`, or prompt user confirmation?
4. What minimum terminal sizes must be officially supported?

## References

- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md`
- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md`
- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/playbook/01-airpods-max-pairing-and-recovery.md`
