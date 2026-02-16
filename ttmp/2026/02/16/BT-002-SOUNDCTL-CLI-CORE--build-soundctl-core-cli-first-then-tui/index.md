---
Title: Build SoundCtl core + CLI first, then TUI
Ticket: BT-002-SOUNDCTL-CLI-CORE
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/soundctl/main.go
      Note: Executable CLI binary entrypoint
    - Path: pkg/cmd/devices/commands.go
      Note: |-
        Representative command wrapper implementation
        Improved device visibility and status command
    - Path: pkg/cmd/root.go
      Note: CLI command tree wiring
    - Path: pkg/soundctl/audio/service.go
      Note: Core audio operations for CLI/TUI
    - Path: pkg/soundctl/bluetooth/service.go
      Note: |-
        Core bluetooth operations for CLI/TUI
        Aggregates show/info output into CLI-friendly status
    - Path: pkg/soundctl/exec/runner.go
      Note: Core runner abstraction used by all services
    - Path: pkg/soundctl/parse/bluetooth.go
      Note: Bluetooth parsing for service layer
    - Path: pkg/soundctl/parse/pactl.go
      Note: Audio parsing for service layer
    - Path: pkg/tui/app.go
      Note: TUI root model
    - Path: pkg/tui/subscriptions.go
      Note: Live event subscriptions
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/design-doc/01-two-phase-build-plan-cli-core-first-tui-second.md
      Note: Two-phase design and execution constraints
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/reference/01-diary.md
      Note: Implementation chronology
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md
      Note: Detailed phased task list
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:12:41.247973418-05:00
WhatFor: ""
WhenToUse: ""
---







# Build SoundCtl core + CLI first, then TUI

## Overview

This ticket implements SoundCtl in two phases.
**Both phases are complete.** Phase 1 delivers a testable `pkg/` core with Glazed CLI wrappers.
Phase 2 delivers a Bubble Tea TUI with lipgloss-styled panes matching all 4 spec screens,
live event subscriptions, and 42 total tests.

## Key Links

- [Tasks](./tasks.md)
- [Diary](./reference/01-diary.md)
- [Two-Phase Build Plan: CLI Core First, TUI Second](./design-doc/01-two-phase-build-plan-cli-core-first-tui-second.md)
- [Phase 1 CLI Smoke Checks and Usage](./playbook/01-phase-1-cli-smoke-checks-and-usage.md)
- [Changelog](./changelog.md)

## Status

Current status: **active**

## Topics

- bluetooth
- ubuntu
- audio

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
