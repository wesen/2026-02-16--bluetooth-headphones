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
    - Path: cmd/soundctl/main.go
      Note: CLI entrypoint and dependency wiring
    - Path: pkg/cmd/common/common.go
      Note: Glazed section/parser helpers
    - Path: pkg/cmd/devices/commands.go
      Note: |-
        Devices CLI verbs
        devices status verb + richer list output
    - Path: pkg/cmd/mute/commands.go
      Note: Mute CLI verbs
    - Path: pkg/cmd/profiles/commands.go
      Note: Profiles CLI verbs
    - Path: pkg/cmd/root.go
      Note: Root command and group registration
    - Path: pkg/cmd/scan/commands.go
      Note: Scan CLI verbs
    - Path: pkg/cmd/sinks/commands.go
      Note: Sinks CLI verbs
    - Path: pkg/cmd/sources/commands.go
      Note: Sources CLI verbs
    - Path: pkg/cmd/volume/commands.go
      Note: Volume CLI verbs
    - Path: pkg/soundctl/audio/service.go
      Note: Audio service API and validation
    - Path: pkg/soundctl/audio/service_test.go
      Note: Audio service tests with fake runner
    - Path: pkg/soundctl/bluetooth/service.go
      Note: |-
        Bluetooth service API and command mapping
        ControllerStatus + enriched ListDevices output
    - Path: pkg/soundctl/bluetooth/service_test.go
      Note: |-
        Bluetooth service tests with fake runner
        Service behavior tests for mode/scanning
    - Path: pkg/soundctl/exec/runner.go
      Note: Command runner abstraction and fake test double
    - Path: pkg/soundctl/parse/bluetooth.go
      Note: |-
        Bluetooth output parser
        Controller status parser from bluetoothctl show
    - Path: pkg/soundctl/parse/bluetooth_test.go
      Note: |-
        Parser behavior tests
        Controller parser tests
    - Path: pkg/soundctl/parse/pactl.go
      Note: Pactl short output parser
    - Path: pkg/soundctl/parse/pactl_test.go
      Note: Parser behavior tests
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/index.md
      Note: Ticket navigation and status context
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/playbook/01-phase-1-cli-smoke-checks-and-usage.md
      Note: Updated usage for status and mode fields
    - Path: ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md
      Note: |-
        Task-by-task execution plan tracked in diary
        Track Phase 1.11 completion
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

**Commit (code):** `70198d7` — "phase1: add bluetooth and audio core services with tests"

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

## Step 4: Build Glazed CLI command tree and complete Phase 1 test/smoke/documentation pass

I implemented the SoundCtl CLI root command and all planned group/verb wrappers, keeping business logic inside `pkg/soundctl/*` services only. This preserves the intended architecture where CLI is orchestration and output formatting, while core functionality remains reusable for the future TUI.

After wiring commands, I ran full unit tests and live CLI smoke commands (help, read-path commands, and a validation-failure mutation path). I also added a Phase 1 playbook documenting exact validation commands and known boundaries.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the CLI-first phase task-by-task, with tested command wrappers around core services and detailed documentation.

**Inferred user intent:** Reach a stable, testable Phase 1 completion before starting any TUI work.

**Commit (code):** `b7b0f48` — "phase1: add glazed CLI verbs and complete validation docs"

### What I did
- Added CLI root and dependency wiring:
  - `cmd/soundctl/main.go`
  - `pkg/cmd/root.go`
  - `pkg/cmd/common/common.go`
- Implemented command groups with Glazed wrappers:
  - `pkg/cmd/devices/commands.go`
  - `pkg/cmd/scan/commands.go`
  - `pkg/cmd/sinks/commands.go`
  - `pkg/cmd/sources/commands.go`
  - `pkg/cmd/profiles/commands.go`
  - `pkg/cmd/volume/commands.go`
  - `pkg/cmd/mute/commands.go`
- Pulled module dependencies via `go mod tidy`.
- Executed tests and smoke checks:
  - `go test ./...`
  - `go run ./cmd/soundctl --help`
  - `go run ./cmd/soundctl devices list --output json`
  - `go run ./cmd/soundctl sinks list --output json`
  - `go run ./cmd/soundctl volume set --target invalid --name foo --percent 10 --output json` (expected failure path)
- Added Phase 1 usage/smoke playbook:
  - `playbook/01-phase-1-cli-smoke-checks-and-usage.md`

### Why
- Completing all Phase 1 CLI tasks now gives a concrete platform for validating core behavior and unblocks TUI implementation later without backend churn.

### What worked
- Command groups/verbs compile and run.
- Help output and structured read commands return expected data.
- Validation error path returns clear non-zero result.
- Full test suite passes after CLI integration.

### What didn't work
- One `go mod tidy` tool session appeared to hang without output in the terminal wrapper; rerunning `go mod tidy` completed successfully.

### What I learned
- Glazed command wrappers are straightforward when each command is mapped to one service call + a small typed settings struct.

### What was tricky to build
- Avoiding logic leakage into CLI wrappers required discipline: it is tempting to add parsing/validation in command files. The chosen implementation keeps CLI-level validation minimal and pushes operational logic to services.

### What warrants a second pair of eyes
- Validate command UX naming/semantics (especially `profiles list` currently based on `pactl list short cards` output) before freezing external interface.

### What should be done in the future
- Start Phase 2 by creating Bubble Tea shell models that call the same `pkg/soundctl/*` services.

### Code review instructions
- Start with command wiring:
  - `pkg/cmd/root.go`
  - `cmd/soundctl/main.go`
- Review one representative command per group:
  - `pkg/cmd/devices/commands.go`
  - `pkg/cmd/sinks/commands.go`
  - `pkg/cmd/profiles/commands.go`
- Confirm CLI wrappers call services only:
  - `pkg/soundctl/bluetooth/service.go`
  - `pkg/soundctl/audio/service.go`
- Re-run:
  - `go test ./...`
  - `go run ./cmd/soundctl --help`

### Technical details
- Smoke output evidence captured:
  - `devices list --output json` returned known paired device rows
  - `sinks list --output json` returned current sink rows
  - invalid `volume set` target returned `invalid target \"invalid\": expected sink or source`

## Step 5: Improve Bluetooth status visibility in CLI output

After using the new CLI, a gap was identified: `devices list` did not expose controller scan state or clear device mode information. I implemented a focused enhancement so scanning and mode are visible directly from CLI output.

The update adds a dedicated `devices status` command (controller-level state) and enriches `devices list` with per-device state fields.

### Prompt Context

**User prompt (verbatim):** "device slist doesn't show me if they're scanning and which mode they are in..."

**Assistant interpretation:** Extend the CLI to make scan state and device mode explicit, without requiring manual `bluetoothctl` inspection.

**Inferred user intent:** Get immediate, scriptable visibility into whether scanning is active and whether each device is connected/paired/saved.

**Commit (code):** N/A

### What I did
- Added parser support for `bluetoothctl show`:
  - `ParseBluetoothShow` in `pkg/soundctl/parse/bluetooth.go`
- Added controller status support in Bluetooth service:
  - `ControllerStatus(ctx)` in `pkg/soundctl/bluetooth/service.go`
- Enriched `ListDevices` response to include:
  - `paired`, `trusted`, `connected`, `mode`
- Updated devices CLI:
  - `devices list` now outputs mode and scan state fields
  - new `devices status` command outputs `powered`, `pairable`, `scanning`
- Added/updated tests:
  - parser test for `ParseBluetoothShow`
  - service tests for enriched list behavior and controller status
- Ran full tests and live smoke commands.

### Why
- The CLI should be sufficient for operational visibility; users should not need to drop into `bluetoothctl` for basic status checks.

### What worked
- `devices status --output json` now reports controller scan state directly.
- `devices list --output json` now includes mode and boolean status flags.
- Tests pass after changes.

### What didn't work
- N/A

### What I learned
- Combining controller-level and device-level data in list output is useful, but an explicit `devices status` verb is still needed for no-device scenarios.

### What was tricky to build
- `bluetoothctl` splits state across separate commands (`show`, `devices`, `info`). Surfacing a single useful CLI output required composing these calls in the service layer without leaking parsing details into command wrappers.

### What warrants a second pair of eyes
- Confirm if `mode` classification should introduce additional states in future (for example `trusted-only`).

### What should be done in the future
- If desired, add an optional `--include-controller` switch to `devices list` to emit a dedicated controller row in single-stream output.

### Code review instructions
- Review parser/service changes:
  - `pkg/soundctl/parse/bluetooth.go`
  - `pkg/soundctl/bluetooth/service.go`
- Review CLI changes:
  - `pkg/cmd/devices/commands.go`
- Re-run:
  - `go test ./...`
  - `go run ./cmd/soundctl devices status --output json`
  - `go run ./cmd/soundctl devices list --output json`

### Technical details
- Live verification showed:
  - `devices status` now returns `powered/pairable/scanning`
  - `devices list` now returns `mode/paired/trusted/connected/scanning`

## Step 6: Add dual-mode scan/pair primitives with wait-based discovery and initialize help system

After confirming that one-shot `scan start` was not a reliable verification primitive, I added a timed discovery primitive and refactored scan commands to dual mode. I also wired the Glazed help system at root and verified `help build-first-command` is available from `soundctl`.

This gives two practical operation modes:
- default normal mode for human-readable troubleshooting
- `--with-glaze-output` mode for structured rows and stream-friendly pipelines

### Prompt Context

**User prompt (verbatim):** "ok this turned on scanning for sure. but the go version didn't... Also read `glaze help build-first-command` and initialize the help system properly and also make scan / pair commands dual mode with normal mode first, such that we can have a --wait and a more streaming output version of it all."

**Assistant interpretation:** Make scanning verifiable from the Go CLI, initialize help-system integration correctly, and provide dual-mode scan/pair commands with timed waiting and structured output support.

**Inferred user intent:** Improve practical operability and debugging confidence for pairing flows, while keeping CLI scriptability.

**Commit (code):** N/A

### What I did
- Read and validated `glaze help build-first-command` content.
- Initialized help system in root command:
  - `help.NewHelpSystem()`
  - `doc.AddDocToHelpSystem(...)`
  - `help_cmd.SetupCobraRootCommand(...)`
- Added dual-mode support helper for Cobra command builds:
  - `BuildCobraDual` in `pkg/cmd/common/common.go`
- Refactored `scan` group commands to dual mode:
  - `start` / `stop`
  - new `discover --wait N [--name-filter ...]`
  - enhanced `pair` with `--wait` and optional auto-targeting from discovery
- Added timed discovery service primitive:
  - `Discover(ctx, seconds)` in `pkg/soundctl/bluetooth/service.go`
- Improved pairing resilience:
  - `Pair` treats `org.bluez.Error.AlreadyExists` as non-fatal and continues trust/connect flow
- Improved diagnostics:
  - command runner now preserves stdout error text when stderr is empty
- Added parser support for scan output and tests:
  - `ParseBluetoothScanOutput` in `pkg/soundctl/parse/bluetooth.go`
  - updated parser/service tests

### Why
- A deterministic `--wait` scan primitive is necessary to verify that scanning actually finds target devices before pairing attempts.
- Dual mode is needed for both human-first workflows and structured automation workflows.

### What worked
- `scan discover --wait 5` now shows discovered devices in normal mode.
- `scan discover --wait 8 --with-glaze-output --output json` emits structured discovery rows + summary.
- `scan pair --help` shows dual-mode and new `--wait`/`--name-filter` options.
- Pair failure output is now descriptive (includes BlueZ failure detail).

### What didn't work
- Loading Glazed docs into the help system emits a debug-level parse warning from an upstream glazed tutorial file (`migrating-to-facade-packages.md`) when log level is debug. Help still works and `build-first-command` loads correctly.

### What I learned
- For Bluetooth in this environment, timed scan-and-collect is a more reliable primitive than trying to model persistent scanning as a stateless CLI toggle.

### What was tricky to build
- `bluetoothctl` emits mixed event lines (`[NEW]`, `[CHG]`, ANSI-colored output). Correct discovery output required filtering only true `[NEW] Device ...` events and excluding status-change noise.

### What warrants a second pair of eyes
- Decide whether help-system doc loading should be narrowed to a curated subset to avoid upstream debug-noise from unrelated doc files.

### What should be done in the future
- Add a dedicated long-running `scan watch` streaming command if continuous discovery events are needed beyond timed windows.

### Code review instructions
- Help integration:
  - `pkg/cmd/root.go`
- Dual-mode command plumbing:
  - `pkg/cmd/common/common.go`
  - `pkg/cmd/scan/commands.go`
- Discovery/pair primitives:
  - `pkg/soundctl/bluetooth/service.go`
  - `pkg/soundctl/parse/bluetooth.go`
  - `pkg/soundctl/exec/runner.go`
- Re-run:
  - `go run ./cmd/soundctl help build-first-command`
  - `go run ./cmd/soundctl scan discover --wait 8`
  - `go run ./cmd/soundctl scan discover --wait 8 --with-glaze-output --output json`
  - `go run ./cmd/soundctl scan pair --addr 08:FF:44:2B:4C:90 --trust --connect --with-glaze-output --output json`

### Technical details
- Verified outputs now include:
  - human-readable scan discovery summary in normal mode
  - structured `kind=discovered` rows plus summary row in glaze mode
  - detailed pair failure text instead of generic `exit status 1`
