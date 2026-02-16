# Tasks

## TODO

- [x] Create BT-002 ticket scaffold and diary
- [x] Define two-phase plan (Phase 1 CLI/core, Phase 2 TUI)
- [x] Phase 1.1: Initialize Go module and project layout (`pkg/` core + `cmd/` CLI wrappers)
- [x] Phase 1.2: Implement command runner abstraction for shell invocations with unit-test stubs
- [x] Phase 1.3: Implement Bluetooth core service in `pkg/` (list/info/connect/disconnect/trust/remove/scan/pair)
- [x] Phase 1.4: Implement audio core service in `pkg/` (sinks/sources/cards list + set default + set profile + move stream + volume/mute)
- [x] Phase 1.5: Implement parsing utilities + test fixtures for `bluetoothctl` and `pactl` outputs
- [x] Phase 1.6: Implement Glazed CLI root and command groups (`devices`, `scan`, `sinks`, `sources`, `profiles`, `volume`, `mute`)
- [x] Phase 1.7: Wire CLI verbs to `pkg/` services only (no business logic in command layer)
- [x] Phase 1.8: Add integration-style tests for core services using fake runner
- [x] Phase 1.9: Run `go test ./...` and execute smoke commands (`--help`, representative read/write verbs)
- [x] Phase 1.10: Finalize Phase 1 docs (usage examples, known limitations, next-step handoff to TUI phase)
- [x] Phase 1.11: Improve Bluetooth visibility (`devices status` + mode/scanning fields in `devices list`)
- [ ] Phase 2.1 (deferred): Build Bubble Tea shell consuming the same `pkg/` services
- [ ] Phase 2.2 (deferred): Implement panes/overlay/keymap parity with spec
- [ ] Phase 2.3 (deferred): Add live event subscriptions and TUI integration tests
