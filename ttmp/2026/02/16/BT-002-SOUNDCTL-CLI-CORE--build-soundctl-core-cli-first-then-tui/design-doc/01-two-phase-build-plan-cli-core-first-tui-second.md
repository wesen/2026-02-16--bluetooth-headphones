---
Title: 'Two-Phase Build Plan: CLI Core First, TUI Second'
Ticket: BT-002-SOUNDCTL-CLI-CORE
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:12:56.123423176-05:00
WhatFor: ""
WhenToUse: ""
---

# Two-Phase Build Plan: CLI Core First, TUI Second

## Executive Summary

Build SoundCtl in two phases:

1. **Phase 1 (now):** implement testable core functionality in `pkg/` and expose it through Glazed CLI verbs.
2. **Phase 2 (later):** build Bubble Tea TUI on top of the already-tested `pkg/` services.

This sequencing minimizes UI-related noise while proving Bluetooth/audio behavior through fast CLI-driven tests.

## Problem Statement

If TUI and runtime logic are built simultaneously, debugging becomes slow and unclear (UI event issues vs service/parsing issues). We need deterministic, scriptable validation first.

## Proposed Solution

Create layered code:

- `pkg/soundctl/exec`: command runner abstraction
- `pkg/soundctl/bluetooth`: Bluetooth domain/service operations
- `pkg/soundctl/audio`: audio domain/service operations
- `pkg/soundctl/parse`: output parsers and helpers
- `pkg/cmd/*`: Glazed command wrappers that only orchestrate inputs/outputs
- `cmd/soundctl`: binary entrypoint

All CLI verbs call `pkg/` services; no shell command formatting or parsing stays in CLI command files.

## Design Decisions

1. Core logic in `pkg/`, thin CLI wrappers.
Reason: easier unit/integration testing and TUI reuse.

2. Glazed output model for CLI verbs.
Reason: machine-readable outputs with minimal custom formatting code.

3. Fake runner for tests.
Reason: stable tests without requiring local Bluetooth/audio hardware state.

4. Commit by completed task slices.
Reason: clear review boundaries and rollback points.

## Alternatives Considered

1. Build TUI first.
Rejected: slow feedback loops and harder diagnosis of backend bugs.

2. Inline shell command calls in each CLI command file.
Rejected: duplicated logic and poor testability.

3. No parser tests (manual-only validation).
Rejected: brittle against command output changes.

## Implementation Plan

### Phase 1: CLI/Core (active)

1. Initialize module + directory layout.
2. Implement runner abstraction and fake runner.
3. Implement Bluetooth and audio services in `pkg/`.
4. Add parsers and fixture-driven tests.
5. Implement Glazed commands and root registration.
6. Run full test suite + CLI smoke checks.
7. Document usage/limitations and handoff state.

### Phase 2: TUI (deferred)

1. Build Bubble Tea root shell.
2. Add panes and keymaps.
3. Connect subscriptions and polish UX.
4. Reuse `pkg/` services without duplicating logic.

## Open Questions

1. Should v1 include watch/streaming CLI verbs now or after baseline CRUD verbs stabilize?
2. Should we support both PulseAudio and PipeWire-native introspection in phase 1, or start with `pactl` only?

## References

- `ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md`
- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md`
- `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md`
