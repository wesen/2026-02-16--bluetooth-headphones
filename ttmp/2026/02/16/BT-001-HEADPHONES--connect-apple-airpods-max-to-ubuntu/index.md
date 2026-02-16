---
Title: Connect Apple AirPods Max to Ubuntu
Ticket: BT-001-HEADPHONES
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md
      Note: Detailed Bubble Tea/Bubbles architecture and phased implementation plan
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md
      Note: Glazed command groups and verb contracts
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/playbook/01-airpods-max-pairing-and-recovery.md
      Note: Runbook for reconnect and re-pair
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/reference/01-diary.md
      Note: Implementation diary and troubleshooting chronology
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/sources/local/headphones.md
      Note: Original product/UI spec imported from /tmp/headphones.md
    - Path: ttmp/2026/02/16/BT-001-HEADPHONES--connect-apple-airpods-max-to-ubuntu/tasks.md
      Note: |-
        Current completion and pending pairing steps
        Tracks CLI planning completion and pairing follow-ups
ExternalSources:
    - local:headphones.md
Summary: ""
LastUpdated: 2026-02-16T14:04:34.173853351-05:00
WhatFor: ""
WhenToUse: ""
---





# Connect Apple AirPods Max to Ubuntu

## Overview

This ticket tracks end-to-end setup of Apple AirPods Max on this Ubuntu machine.
The host Bluetooth stack is healthy, but direct reconnect attempts currently fail with `org.bluez.Error.Failed br-connection-page-timeout`.
Use the playbook for a deterministic remove-and-repair flow while the headset is in pairing mode.

## Key Links

- [Diary](./reference/01-diary.md)
- [SoundCtl Bubble Tea/Bubbles Implementation Guide](./design-doc/01-soundctl-bubble-tea-bubbles-implementation-guide.md)
- [SoundCtl Glazed CLI Verb Set and Integration Plan](./design-doc/02-soundctl-glazed-cli-verb-set-and-integration-plan.md)
- [AirPods Max Pairing and Recovery](./playbook/01-airpods-max-pairing-and-recovery.md)
- [Tasks](./tasks.md)
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
