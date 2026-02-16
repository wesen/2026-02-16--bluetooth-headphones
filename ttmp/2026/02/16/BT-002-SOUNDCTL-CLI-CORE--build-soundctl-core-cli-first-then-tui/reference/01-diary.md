---
Title: Diary
Ticket: BT-002-SOUNDCTL-CLI-CORE
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/soundctl/audio/service.go
      Note: Audio service API and validation
    - Path: pkg/soundctl/audio/service_test.go
      Note: Audio service tests with fake runner
    - Path: pkg/soundctl/bluetooth/service.go
      Note: Bluetooth service API and command mapping
    - Path: pkg/soundctl/bluetooth/service_test.go
      Note: Bluetooth service tests with fake runner
    - Path: pkg/soundctl/exec/runner.go
      Note: Command runner abstraction and fake test double
    - Path: pkg/soundctl/parse/bluetooth.go
      Note: Bluetooth output parser
    - Path: pkg/soundctl/parse/bluetooth_test.go
      Note: Parser behavior tests
    - Path: pkg/soundctl/parse/pactl.go
      Note: Pactl short output parser
    - Path: pkg/soundctl/parse/pactl_test.go
      Note: Parser behavior tests
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/index.md
      Note: Ticket navigation and status context
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md
      Note: Task-by-task execution plan tracked in diary
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T14:12:41.375423356-05:00
WhatFor: ""
WhenToUse: ""
---




# Diary

## Goal

Track task-by-task implementation of SoundCtl with a CLI-first core architecture, including commands run, failures, tests, and commit checkpoints.

## Step 1: Create BT-002 ticket and define two-phase execution plan

I started by creating a dedicated implementation ticket separate from the exploratory Bluetooth setup ticket. This establishes a clean execution lane for building actual code and committing in increments.

I then defined detailed, ordered tasks with explicit phase boundaries so implementation can proceed deterministically: CLI/core now, TUI later.

### Prompt Context

**User prompt (verbatim):** "Alright, let's create a new ticket for building the tool in two phases, starting with the CLI such that we can easily test the core functionality (in pkg/ and just wrapped by the CLI verbs). Add detailed tasks. Then work on that new ticket task by task, committing at opportune intervals, keeping a detailed diary. Make sure to test things as you go too.

We'll tackle the TUI after that, once the core functionality works."

**Assistant interpretation:** Create a fresh ticket for implementation, define a detailed two-phase plan, and begin executing Phase 1 with tested, incremental commits documented in a detailed diary.

**Inferred user intent:** Build production code iteratively with strong traceability and test discipline, postponing TUI work until backend correctness is validated.

**Commit (code):** `3e18d23` — "phase1: scaffold module runner and parser foundation"

### What I did
- Created ticket `BT-002-SOUNDCTL-CLI-CORE`.
- Created diary doc for the ticket.
- Added design doc `01-two-phase-build-plan-cli-core-first-tui-second.md`.
- Replaced placeholder task list with detailed phased tasks for CLI/core and deferred TUI tasks.
- Updated ticket index links and overview to reflect execution strategy.

### Why
- Separating planning/execution from earlier troubleshooting docs keeps implementation history clean and reviewable.
- Detailed tasks provide explicit commit/test checkpoints.

### What worked
- Ticket scaffolding and docs were created successfully via `docmgr`.
- Task list now reflects concrete deliverables in sequence.

### What didn't work
- N/A

### What I learned
- Starting from an empty repo is simpler when task boundaries are enforced early; otherwise CLI, core services, and TUI concerns blend too quickly.

### What was tricky to build
- The main challenge was balancing detailed planning with immediate implementation momentum; the task list was structured to be specific enough for incremental commits without over-constraining execution.

### What warrants a second pair of eyes
- Confirm phase boundaries are acceptable (especially that TUI remains explicitly deferred until Phase 1 completion).

### What should be done in the future
- Begin Phase 1.1 immediately with module/layout scaffolding and commit once baseline tests pass.

### Code review instructions
- Review task sequencing in:
  - `ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md`
- Review phase rationale in:
  - `ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/design-doc/01-two-phase-build-plan-cli-core-first-tui-second.md`

### Technical details
- Commands used:
  - `docmgr ticket create-ticket --ticket BT-002-SOUNDCTL-CLI-CORE --title "Build SoundCtl core + CLI first, then TUI" --topics bluetooth,ubuntu,audio`
  - `docmgr doc add --ticket BT-002-SOUNDCTL-CLI-CORE --doc-type reference --title "Diary"`
  - `docmgr doc add --ticket BT-002-SOUNDCTL-CLI-CORE --doc-type design-doc --title "Two-Phase Build Plan: CLI Core First, TUI Second"`

## Step 2: Implement module scaffolding, runner abstraction, and parser foundation

I started execution by establishing the Go module and directory layout for core services and CLI wrappers. Then I implemented the command runner abstraction with both OS and fake implementations, so all service logic can be tested without real hardware dependencies.

I also added initial parsers for `bluetoothctl` and `pactl` short outputs with unit tests, and confirmed the baseline test suite passes.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Begin Phase 1 implementation immediately, keeping work testable and organized for CLI-first delivery.

**Inferred user intent:** Make concrete technical progress quickly while preserving testability and clean layering.

**Commit (code):** N/A

### What I did
- Created Go module `soundctl`.
- Added package layout:
  - `pkg/soundctl/exec`
  - `pkg/soundctl/parse`
  - `pkg/soundctl/bluetooth`
  - `pkg/soundctl/audio`
  - `pkg/cmd/*` group folders
  - `cmd/soundctl`
- Implemented `Runner` interface with:
  - `OSRunner` for real command execution
  - `FakeRunner` for deterministic command stubbing in tests
- Implemented parser functions:
  - `ParseBluetoothDevices`
  - `ParseBluetoothInfo`
  - `ParsePactlShort`
- Added parser tests and ran `go test ./...`.

### Why
- Runner abstraction is the core seam required to validate Bluetooth/audio behavior without relying on local daemon/device state.
- Parsers are the highest-risk normalization layer, so they should be validated early.

### What worked
- Module initialized and package layout created.
- Parser tests passed.
- Baseline `go test ./...` passed.

### What didn't work
- N/A

### What I learned
- `bluetoothctl` and `pactl` outputs are parseable with lightweight line-based parsing for the initial CLI feature set.

### What was tricky to build
- Balancing speed with future extensibility required defining abstractions early (runner + parse) before implementing service methods. This adds small up-front cost but avoids later refactors across many CLI verbs.

### What warrants a second pair of eyes
- Confirm parser strictness level (currently rejects malformed lines instead of silently skipping them).

### What should be done in the future
- Implement Bluetooth and audio services on top of the runner/parsers and add service-level tests next.

### Code review instructions
- Start with:
  - `pkg/soundctl/exec/runner.go`
  - `pkg/soundctl/parse/bluetooth.go`
  - `pkg/soundctl/parse/pactl.go`
- Validate tests:
  - `pkg/soundctl/parse/bluetooth_test.go`
  - `pkg/soundctl/parse/pactl_test.go`
- Re-run:
  - `go test ./...`

### Technical details
- Commands used:
  - `go mod init soundctl`
  - `gofmt -w ...`
  - `go test ./...`

## Step 3: Implement Bluetooth and audio core services with fake-runner tests

With the runner/parsing layer in place, I implemented service packages that provide the actual operational API for the future CLI and TUI layers. The Bluetooth service now covers list/info/connect/disconnect/trust/remove/pair/scan, and the audio service covers list/set-default/move-stream/profile/volume/mute operations.

I validated service behavior through unit tests using the fake runner, ensuring command construction and parser integration can be verified deterministically.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue task-by-task implementation in `pkg/` and test each slice before moving to CLI wrappers.

**Inferred user intent:** Establish a stable, reusable backend surface before command/TUI layers.

**Commit (code):** N/A

### What I did
- Added Bluetooth service:
  - `pkg/soundctl/bluetooth/service.go`
  - methods: `ListDevices`, `Info`, `Connect`, `Disconnect`, `Trust`, `Remove`, `Pair`, `StartScan`, `StopScan`
- Added Bluetooth tests:
  - `pkg/soundctl/bluetooth/service_test.go`
- Added audio service:
  - `pkg/soundctl/audio/service.go`
  - methods: `ListSinks`, `ListSources`, `ListCards`, `SetDefaultSink`, `SetDefaultSource`, `MoveSinkInput`, `SetCardProfile`, `SetVolume`, `ToggleMute`
- Added audio tests:
  - `pkg/soundctl/audio/service_test.go`
- Ran formatting + full test suite.

### Why
- These services are the core contract that CLI verbs and future TUI components should depend on.

### What worked
- Service methods compile and execute against fake-runner stubs.
- Unit tests passed for command behavior and validation.

### What didn't work
- N/A

### What I learned
- Separating parser utilities from service orchestration kept service tests focused on command behavior and input validation.

### What was tricky to build
- Audio operations have target-dependent command variants (`sink` vs `source`) and numeric/value validation constraints. Encoding these as shared helper functions (`volumeCommand`, `muteCommand`) reduced duplication and made validation behavior explicit.

### What warrants a second pair of eyes
- Validate whether the current volume range guard (`0..150`) matches expected operational policy for this tool.

### What should be done in the future
- Implement Glazed CLI wrappers on top of these services, ensuring no command execution logic leaks into CLI packages.

### Code review instructions
- Start with:
  - `pkg/soundctl/bluetooth/service.go`
  - `pkg/soundctl/audio/service.go`
- Validate behavior via tests:
  - `pkg/soundctl/bluetooth/service_test.go`
  - `pkg/soundctl/audio/service_test.go`
- Re-run:
  - `go test ./...`

### Technical details
- Commands used:
  - `gofmt -w pkg/soundctl/bluetooth/service.go pkg/soundctl/bluetooth/service_test.go pkg/soundctl/audio/service.go pkg/soundctl/audio/service_test.go`
  - `go test ./...`
