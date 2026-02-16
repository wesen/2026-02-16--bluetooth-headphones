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

