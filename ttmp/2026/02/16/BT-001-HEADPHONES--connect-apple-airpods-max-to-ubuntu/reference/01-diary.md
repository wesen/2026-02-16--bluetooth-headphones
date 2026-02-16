---
Title: Diary
Ticket: BT-001-HEADPHONES
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md
      Note: Implementation architecture produced from imported spec
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md
      Note: Glazed CLI command set and wiring plan
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/index.md
      Note: Ticket status summary
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/playbook/01-airpods-max-pairing-and-recovery.md
      Note: Recovery sequence and commands
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md
      Note: Imported source material analyzed in Step 6
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/tasks.md
      Note: Operational task checklist
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T13:54:01.840358393-05:00
WhatFor: ""
WhenToUse: ""
---




# Diary

## Goal

Capture the exact implementation and troubleshooting sequence for connecting Apple AirPods Max to this Ubuntu host, including command outputs, failures, and the resulting operator runbook.

## Step 1: Initialize docmgr workspace and ticket

I started by setting up the documentation workspace because this repository was empty except for `.git`. The goal in this step was to create a stable ticket structure first, then add the diary immediately so every action could be logged.

This established a reproducible documentation baseline (`ttmp`, ticket workspace, and ticket-scoped docs) before touching the Bluetooth stack. That ordering makes later troubleshooting reviewable and keeps operational notes tied to one ticket ID.

### Prompt Context

**User prompt (verbatim):** "I want to connect my bluetooth headphones (apple airpods max) to this ubuntu computer.

Create a new ticket with docmgr BT-001-HEADPHONES and keep a frequent diary as you work."

**Assistant interpretation:** Create a `docmgr` ticket named `BT-001-HEADPHONES`, then run and document a practical Bluetooth connection workflow for AirPods Max on Ubuntu.

**Inferred user intent:** Pair and use AirPods Max on this Ubuntu machine with a documented, repeatable process and clear progress log.

**Commit (code):** N/A

### What I did
- Ran `docmgr init --seed-vocabulary`.
- Created the ticket with `docmgr ticket create-ticket --ticket BT-001-HEADPHONES --title "Connect Apple AirPods Max to Ubuntu" --topics bluetooth,ubuntu,audio`.
- Added diary document `reference/01-diary.md`.
- Added playbook document `playbook/01-airpods-max-pairing-and-recovery.md`.
- Added vocabulary slugs for `bluetooth`, `ubuntu`, and `audio`.

### Why
- The ticket scaffolding is required for structured progress tracking and future reuse.
- Vocabulary normalization prevents topic drift and keeps `docmgr` validation clean.

### What worked
- `docmgr` initialized successfully.
- Ticket workspace and both initial docs were created.
- Topic vocabulary additions succeeded.

### What didn't work
- Initial `docmgr status --summary-only` failed before init:
  `Error: root directory does not exist: /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp`

### What I learned
- This repo had no pre-existing doc workspace, so initialization needed to happen before any status/list workflows.

### What was tricky to build
- The generated ticket skeleton included a `playbooks/` directory, while `docmgr doc add --doc-type playbook` created docs under `playbook/`. The command output confirmed this is currently expected behavior from the tool; I kept the generated path and linked to that concrete file to avoid ambiguity.

### What warrants a second pair of eyes
- Low risk: only doc scaffolding and metadata creation occurred in this step.

### What should be done in the future
- Keep using ticket-local docs for each troubleshooting pass so future headset or adapter issues can be compared across runs.

### Code review instructions
- Start at `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md`.
- Confirm ticket bootstrap artifacts exist:
  `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/index.md`
  `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/tasks.md`
  `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/changelog.md`

### Technical details
- Commands used:
  - `docmgr init --seed-vocabulary`
  - `docmgr ticket create-ticket --ticket BT-001-HEADPHONES --title "Connect Apple AirPods Max to Ubuntu" --topics bluetooth,ubuntu,audio`
  - `docmgr doc add --ticket BT-001-HEADPHONES --doc-type reference --title "Diary"`
  - `docmgr doc add --ticket BT-001-HEADPHONES --doc-type playbook --title "AirPods Max Pairing and Recovery"`
  - `docmgr vocab add --category topics --slug bluetooth --description "Bluetooth device setup and troubleshooting"`
  - `docmgr vocab add --category topics --slug ubuntu --description "Ubuntu operating system tasks"`
  - `docmgr vocab add --category topics --slug audio --description "Audio output and headset configuration"`

## Step 2: Collect Bluetooth baseline and attempt direct connection

I then checked whether the local Bluetooth stack was healthy before changing pairing state. The intent was to determine if the issue was host-side (service/controller/rfkill) or device-side (headphones unavailable/sleeping/in-use elsewhere).

After confirming host readiness, I attempted direct connection to the known AirPods Max MAC address and retried multiple times. This isolated the current blocker to connection establishment, not local Bluetooth service startup.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Diagnose and connect AirPods Max from the terminal while recording command-level evidence.

**Inferred user intent:** Get the headphones actively connected now, with enough logs to recover quickly if connection fails.

**Commit (code):** N/A

### What I did
- Verified Bluetooth userspace and service state:
  - `bluetoothctl --version`
  - `systemctl is-active bluetooth`
  - `systemctl status bluetooth --no-pager --lines=40`
- Verified hardware unblock state with `rfkill list`.
- Queried controller details with `bluetoothctl show`.
- Listed known devices with `bluetoothctl devices`.
- Inspected AirPods entry with `bluetoothctl info 08:FF:44:2B:4C:90`.
- Enabled trust and attempted connect:
  - `bluetoothctl trust 08:FF:44:2B:4C:90`
  - `bluetoothctl connect 08:FF:44:2B:4C:90`
- Ran a timed discovery scan `bluetoothctl --timeout 20 scan on`.
- Retried `bluetoothctl connect 08:FF:44:2B:4C:90` three times in a loop.

### Why
- This sequence validates local readiness first, then verifies whether the known remote device is reachable for ACL/page response.

### What worked
- Bluetooth service is active.
- Adapter `hci0` is present and not soft/hard blocked.
- AirPods Max entry exists and is paired/bonded.
- `trust` operation succeeded (`Trusted: yes`).
- Discovery scan returned nearby devices, proving the adapter can scan.

### What didn't work
- `paired-devices` command failed on this `bluetoothctl` build:
  `Invalid command in menu main: paired-devices`
- Direct connection repeatedly failed:
  `Failed to connect: org.bluez.Error.Failed br-connection-page-timeout`

### What I learned
- The host stack is healthy; current failure mode is remote device not responding to page requests (commonly asleep, out of range, or attached elsewhere).

### What was tricky to build
- The failure is intermittent and depends on physical headset state, which cannot be fully controlled from this shell alone. The symptom (`br-connection-page-timeout`) appears after the controller attempts normal BR/EDR connection, so a successful fix likely requires user-side device state changes (wake/pairing mode/disconnect from other hosts) rather than local daemon restarts.

### What warrants a second pair of eyes
- Check whether AirPods Max are currently connected to another Apple device and therefore unavailable for this host.
- Confirm proximity and battery state before the next connect attempt.

### What should be done in the future
- If direct connect fails after trust is set, immediately move to a deterministic remove-and-repair flow instead of repeating connect indefinitely.

### Code review instructions
- Review this step in `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md`.
- Re-run the baseline command set and compare outputs, especially:
  - `systemctl is-active bluetooth`
  - `rfkill list`
  - `bluetoothctl info 08:FF:44:2B:4C:90`

### Technical details
- Key observed state from `bluetoothctl info 08:FF:44:2B:4C:90`:
  - `Paired: yes`
  - `Bonded: yes`
  - `Trusted: yes`
  - `Connected: no`
- Discovery and connect commands:
  - `bluetoothctl --timeout 20 scan on`
  - `bluetoothctl connect 08:FF:44:2B:4C:90`

## Step 3: Build repeatable pairing recovery playbook

Given repeated connection timeouts, I documented a deterministic operator playbook that includes both quick reconnect and full reset/re-pair paths. This turns the live troubleshooting session into a reusable runbook instead of one-off terminal output.

The playbook is intended to be run while physically placing AirPods Max into pairing mode so discovery and pairing can complete reliably. It also includes verification commands for post-connect confirmation.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Deliver practical next actions that can complete pairing even when direct reconnect fails.

**Inferred user intent:** Leave this ticket with a reliable procedure that works now and can be reused later.

**Commit (code):** N/A

### What I did
- Authored `playbook/01-airpods-max-pairing-and-recovery.md` with:
  - quick reconnect path
  - remove-and-repair path
  - post-connect validation checks
- Prepared task/changelog updates to reflect completed diagnostics and pending user-assisted pairing.

### Why
- Physical device state is the likely blocker, so the best next move is a clear, minimal command sequence tied to explicit user actions.

### What worked
- A concrete runbook now exists in the ticket and maps directly to observed failure mode.

### What didn't work
- Full end-to-end connection could not be completed from shell-only automation because the headset did not answer page requests during direct reconnect attempts.

### What I learned
- For AirPods Max on Linux, pre-existing bond data can still fail to reconnect if the headset is not actively available; quick escalation to reset/re-pair guidance is useful.

### What was tricky to build
- The system can discover many nearby devices but still fail to connect to one known address. Distinguishing scan success from target connect success is important; they exercise different parts of the Bluetooth flow.

### What warrants a second pair of eyes
- Confirm whether the remove-and-repair path should become the default first step in this environment.

### What should be done in the future
- After successful pairing, capture one follow-up diary step with the exact successful command outputs and audio sink verification.

### Code review instructions
- Start at `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/playbook/01-airpods-max-pairing-and-recovery.md`.
- Validate that commands in the playbook are copy/paste ready and reference the correct device MAC or discovery workflow.

### Technical details
- Target device observed in session:
  - `Device 08:FF:44:2B:4C:90 Manuel's AirPods Max`
- Failure signature captured repeatedly:
  - `org.bluez.Error.Failed br-connection-page-timeout`

## Step 4: Confirm discoverability gap before user-assisted pairing

I ran a final timed scan filtered to AirPods identifiers to verify whether the headset was advertising at all during this session. This provides a direct signal for whether shell-only reconnect attempts can reasonably succeed.

The filtered scan returned no AirPods lines, which reinforces that the next successful step requires headset-side action (pairing mode and likely disconnect from other hosts).

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Keep progressing toward a real connection while documenting each diagnostic checkpoint.

**Inferred user intent:** Reach a confirmed connection quickly without losing troubleshooting history.

**Commit (code):** N/A

### What I did
- Ran `bluetoothctl --timeout 20 scan on | rg -i "08:FF:44:2B:4C:90|AirPods" || true`.

### Why
- If the headset is not discoverable/advertising, continued direct `connect` attempts are low-value and operator action should be prioritized.

### What worked
- The scan command completed cleanly and produced a definitive filtered result.

### What didn't work
- No matching output for AirPods identifiers appeared during the 20s scan window.

### What I learned
- Current blocker is availability/discoverability of AirPods Max rather than host adapter health.

### What was tricky to build
- Discovery outcomes are time-window dependent; a short scan can miss transient advertisements. I used this as a confidence indicator, not absolute proof.

### What warrants a second pair of eyes
- Confirm whether increasing scan window (for example 45-60 seconds) changes results once AirPods are explicitly placed in pairing mode.

### What should be done in the future
- Record one additional diary step immediately after successful pair/connect, including post-connect audio sink validation.

### Code review instructions
- Re-run the filtered scan command in pairing mode and confirm that an AirPods match appears before pairing attempts.

### Technical details
- Command used:
  - `bluetoothctl --timeout 20 scan on | rg -i "08:FF:44:2B:4C:90|AirPods" || true`

## Step 5: Final state capture before handoff

I captured the current connection and audio sink state at the end of this session so there is a clean handoff point before user-assisted pairing actions. This locks in exactly what remains unresolved.

The AirPods record is still paired and trusted but not connected, and only the built-in analog sink is currently active.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Leave the ticket in a clear end-of-session state with explicit next action.

**Inferred user intent:** Minimize ambiguity so the next attempt can complete quickly.

**Commit (code):** N/A

### What I did
- Ran `bluetoothctl info 08:FF:44:2B:4C:90`.
- Ran `pactl list short sinks`.
- Ran `docmgr ticket list --ticket BT-001-HEADPHONES`.

### Why
- A final state checkpoint makes it obvious that pairing is incomplete and audio routing has not switched.

### What worked
- Ticket/task state is now tracked in docmgr.
- AirPods trust + bond state persisted.

### What didn't work
- `Connected: no` remains for AirPods Max.
- No AirPods sink appeared in the current `pactl` sink list.

### What I learned
- The next successful attempt depends on headset-side availability changes, not additional local daemon setup.

### What was tricky to build
- It is easy to conflate "paired/trusted" with "connected"; explicit final checks prevent this mistake.

### What warrants a second pair of eyes
- Verify that no other nearby Apple device auto-connects to the headset during the next pairing attempt.

### What should be done in the future
- Execute playbook Step 3 (remove and re-pair in pairing mode), then append a success entry with resulting `bluetoothctl info` and `pactl` outputs.

### Code review instructions
- Validate unresolved state in:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md`
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/tasks.md`

### Technical details
- Final observed state:
  - `Paired: yes`
  - `Trusted: yes`
  - `Connected: no`
  - `pactl` sinks: `alsa_output.pci-0000_00_1f.3.analog-stereo` only

## Step 6: Import UI spec and produce Bubble Tea/Bubbles implementation guide

I imported the provided `/tmp/headphones.md` source into the ticket and read it end-to-end. The document defines a concrete multi-screen TUI spec plus a message/command DSL, so the output in this step focused on an implementation-grade architecture guide rather than high-level advice.

The guide maps the spec to concrete Bubble Tea/Bubbles model boundaries, message taxonomy, subscription loops, command adapters, testing strategy, and phased delivery milestones.

### Prompt Context

**User prompt (verbatim):** "Import the document /tmp/headphones.md with docmgr import file --file /tmp/headphones.md --ticket $TICKETNAME and read it, and then make a detailed analysis / implementation guide on how to build this with bubbletea / bubbles."

**Assistant interpretation:** Import the requested source file via `docmgr`, read it, and generate a detailed design/implementation document for building the described app with Bubble Tea and Bubbles.

**Inferred user intent:** Turn the drafted UI/DSL concept into an executable engineering plan with clear architecture and implementation steps.

**Commit (code):** N/A

### What I did
- Checked ticket context and found `$TICKETNAME` unset in shell.
- Ran the requested import command pattern and captured the initial failure.
- Re-ran import with explicit variable assignment and successfully imported to ticket sources.
- Read `/tmp/headphones.md` (including the model/message DSL, subscriptions, and keybindings sections).
- Created `design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md`.
- Wrote a detailed guide covering:
  - layered architecture (`domain`, `services`, `app`, `runtime`)
  - root/child model decomposition matching the DSL
  - typed message contracts and command wrappers
  - long-lived subscription pattern for Bluetooth and audio events
  - phased implementation and test plan

### Why
- The imported spec already contains strong structure; the most valuable next artifact is an actionable build blueprint that avoids ad-hoc implementation drift.

### What worked
- Source import succeeded and was indexed by `docmgr`.
- Design document was created and populated with implementation-ready detail.
- The ticket index now includes the imported source in `ExternalSources`.

### What didn't work
- First import attempt failed due shell variable expansion order:
  `Error: flag needs an argument: --ticket`

### What I learned
- In shell one-liners, `VAR=value cmd ... $VAR` can expand `$VAR` before assignment for that invocation; explicit assignment + quoted use is safer.

### What was tricky to build
- The imported DSL expresses commands as shell strings, while a robust Bubble Tea implementation benefits from typed service interfaces and DBus-first adapters. The guide resolves this by preserving command intent at the UI level while moving execution concerns into runtime/service layers.

### What warrants a second pair of eyes
- Validate whether v1 should be DBus-first or start with command-wrapper adapters for faster initial delivery.

### What should be done in the future
- Create code scaffolding from the file layout and phase plan in the new design doc, then implement Phase 1 static UI skeleton.

### Code review instructions
- Start with:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md`
- Cross-check source requirements at:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md`
- Confirm diary step consistency at:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md`

### Technical details
- Failing command:
  - `TICKETNAME=BT-001-HEADPHONES docmgr import file --file /tmp/headphones.md --ticket $TICKETNAME`
- Successful command:
  - `TICKETNAME=BT-001-HEADPHONES; docmgr import file --file /tmp/headphones.md --ticket "$TICKETNAME"`

## Step 7: Normalize imported source frontmatter for docmgr validation

After importing the source markdown, `docmgr doctor` flagged a frontmatter parse error because the imported file starts directly with content and no YAML block. I normalized the file by adding valid frontmatter while preserving all original spec content.

This change keeps ticket hygiene checks useful and prevents a non-semantic parser error from masking real documentation issues.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Ensure the imported artifact is not only present but also managed cleanly under `docmgr` conventions.

**Inferred user intent:** Have a usable imported source and implementation guide inside a healthy ticket workspace.

**Commit (code):** N/A

### What I did
- Ran `docmgr doctor --ticket BT-001-HEADPHONES --stale-after 30`.
- Ran `docmgr validate frontmatter --doc /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md --suggest-fixes`.
- Added valid frontmatter to `sources/local/headphones.md`.
- Re-ran `docmgr doctor` and `docmgr validate frontmatter`.

### Why
- Keeping the imported source parseable avoids persistent doctor errors and improves long-term maintainability of this ticket.

### What worked
- Frontmatter validation now passes for the imported source file.
- `docmgr doctor` error was resolved.

### What didn't work
- Initial validation command used a root-relative path with duplicate `ttmp/` segment and failed with:
  `open /home/manuel/code/wesen/2026-02-16--bluetooth-headphones/ttmp/ttmp/...: no such file or directory`

### What I learned
- `docmgr validate frontmatter --doc` is safest with absolute paths to avoid root-relative path confusion.

### What was tricky to build
- Imported raw markdown can violate house frontmatter rules even when content itself is correct; normalization must preserve source semantics while satisfying parser requirements.

### What warrants a second pair of eyes
- Decide whether to keep source imports exempt from numeric filename prefix checks in this workspace.

### What should be done in the future
- Optionally rename source files to numeric-prefixed names (for example `01-headphones.md`) if strict warning-free `doctor` output is required.

### Code review instructions
- Review frontmatter at:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md`
- Re-run:
  - `docmgr doctor --ticket BT-001-HEADPHONES --stale-after 30`

### Technical details
- Current `doctor` state: warning only (`missing_numeric_prefix`) on `sources/local/headphones.md`.

## Step 8: Study glazed-command-authoring and define CLI verbs alongside TUI

I studied the `glazed-command-authoring` skill and translated it into a concrete SoundCtl CLI command surface that complements the TUI. The output is a dedicated design doc that specifies command groups, verb contracts, wiring conventions, and phased implementation.

This step intentionally focused on execution structure, not just command naming, so the future code implementation can follow Glazed conventions without re-designing parser/middleware behavior later.

### Prompt Context

**User prompt (verbatim):** "also study $glazed-command-authoring so that we can have a set of command line CLI verb in addition to the TUI."

**Assistant interpretation:** Use the Glazed command authoring skill and produce a concrete CLI verb plan that works alongside the TUI architecture.

**Inferred user intent:** Ensure SoundCtl has both interactive and scriptable interfaces with consistent behavior.

**Commit (code):** N/A

### What I did
- Loaded and reviewed `/home/manuel/.codex/skills/glazed-command-authoring/SKILL.md`.
- Created `design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md`.
- Defined grouped verb taxonomy (`devices`, `scan`, `sinks`, `sources`, `profiles`, `watch`, plus volume/mute commands).
- Added Glazed command skeleton with:
  - `*cmds.CommandDescription` embedding
  - settings struct with `glazed` tags
  - `vals.DecodeSectionInto(schema.DefaultSlug, settings)`
  - `cli.BuildCobraCommandFromCommand` integration guidance
- Documented streaming output defaults and watch-command behavior.
- Updated `tasks.md` and `index.md` to include this new CLI plan.

### Why
- The project now has a single source of truth for both TUI behavior and CLI parity, reducing future drift and design rework.

### What worked
- Skill-guided conventions were captured in a reusable implementation plan.
- The ticket now links both the TUI and CLI design docs together.

### What didn't work
- N/A

### What I learned
- The Glazed discipline (typed settings decode + standardized sections + row output contracts) gives a clean path to script-friendly commands without sacrificing UX discoverability.

### What was tricky to build
- Balancing terse CLI ergonomics with explicit grouped verbs required careful mapping from TUI interactions to scriptable operations. The chosen structure keeps grouped discoverability while preserving one-to-one parity for core actions.

### What warrants a second pair of eyes
- Validate whether `scan start/stop` should be exposed directly or only managed through higher-level commands like `scan pair`.

### What should be done in the future
- Implement Phase A scaffolding from the new CLI design doc, then start with read-only list commands for fast feedback.

### Code review instructions
- Start with:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md`
- Confirm updated ticket pointers:
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/index.md`
  - `ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/tasks.md`

### Technical details
- Skill source used:
  - `/home/manuel/.codex/skills/glazed-command-authoring/SKILL.md`
