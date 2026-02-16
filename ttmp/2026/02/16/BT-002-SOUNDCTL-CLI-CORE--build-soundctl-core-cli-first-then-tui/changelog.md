# Changelog

## 2026-02-16

- Initial workspace created


## 2026-02-16

Step 1: Created BT-002 ticket, defined detailed two-phase tasks, and initialized execution diary for task-by-task implementation.

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/reference/01-diary.md — Step-by-step execution log
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/tasks.md — Detailed phase/task breakdown


## 2026-02-16

Step 2: Bootstrapped Go module/layout, added runner abstraction (OS + fake), implemented Bluetooth/Pactl parsers, and passed baseline tests.

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/exec/runner.go — Runner abstraction
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/parse/bluetooth.go — Bluetooth parser
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/parse/pactl.go — Pactl parser


## 2026-02-16

Step 3: Implemented Bluetooth and audio core services with fake-runner-backed unit tests and validation guards.

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/audio/service.go — Audio core methods
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/bluetooth/service.go — Bluetooth core methods


## 2026-02-16

Step 4: Implemented full Glazed CLI command tree over pkg services, passed tests/smoke runs, and documented Phase 1 usage/validation playbook.

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/cmd/devices/commands.go — Device command wrappers
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/cmd/root.go — Root/group registration
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/playbook/01-phase-1-cli-smoke-checks-and-usage.md — Smoke checks and known limitations
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/reference/01-diary.md — Detailed implementation narrative


## 2026-02-16

Step 5: Added bluetooth visibility improvements with devices status and mode/scanning fields in devices list.

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/cmd/devices/commands.go — Command output enhancements
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/soundctl/bluetooth/service.go — Controller and per-device state composition
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-002-SOUNDCTL-CLI-CORE--build-soundctl-core-cli-first-then-tui/reference/01-diary.md — Implementation and test narrative


## 2026-02-16

Phase 2.1: Built Bubble Tea TUI shell with lipgloss-styled panes (Devices/Sinks/Profiles), scanner overlay, typed message routing, and 15 unit tests. Wired as 'soundctl tui' subcommand. (commit 4e3210e)

### Related Files

- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/tui/app.go — Root TUI model with tab bar and pane routing
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/tui/app_test.go — 15 TUI unit tests
- /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/pkg/tui/style.go — Lipgloss colour palette and style system

