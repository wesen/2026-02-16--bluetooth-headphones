---
Title: AirPods Max Pairing and Recovery
Ticket: BT-001-HEADPHONES
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-16T13:56:05.530958558-05:00
WhatFor: ""
WhenToUse: ""
---

# AirPods Max Pairing and Recovery

## Purpose

Connect Apple AirPods Max to this Ubuntu host and recover from `br-connection-page-timeout` failures using a deterministic command flow.

## Environment Assumptions

- Ubuntu host with BlueZ (`bluetoothctl`) and active `bluetooth.service`.
- Local adapter is visible as `hci0`.
- AirPods Max are charged and physically nearby.
- If reconnect fails, you can put AirPods Max into pairing mode (press and hold the noise control button until the LED flashes white).

## Commands

### 1) Baseline checks

```bash
bluetoothctl --version
systemctl is-active bluetooth
rfkill list
bluetoothctl show
bluetoothctl devices
```

### 2) Quick reconnect (existing pairing)

```bash
MAC="08:FF:44:2B:4C:90"
bluetoothctl trust "$MAC"
bluetoothctl connect "$MAC"
bluetoothctl info "$MAC"
```

If `Connected: yes`, stop here.

### 3) Full reset and re-pair (when reconnect times out)

Put AirPods Max in pairing mode first, then run:

```bash
MAC="08:FF:44:2B:4C:90"

# optional cleanup of stale bond
bluetoothctl disconnect "$MAC" || true
bluetoothctl remove "$MAC" || true

# discover device while in pairing mode
bluetoothctl --timeout 20 scan on
bluetoothctl devices | rg -i "airpods max|airpods"

# if MAC changed, replace MAC below with the discovered value
bluetoothctl pair "$MAC"
bluetoothctl trust "$MAC"
bluetoothctl connect "$MAC"
bluetoothctl info "$MAC"
```

### 4) Audio routing verification

```bash
pactl list short sinks
pactl list short cards
```

If needed, select the AirPods sink in your desktop audio settings.

## Exit Criteria

- `bluetoothctl info <MAC>` reports:
  - `Paired: yes`
  - `Trusted: yes`
  - `Connected: yes`
- Audio output is available via AirPods Max sink/card in PipeWire/PulseAudio tools.

## Notes

- During this session, repeated quick reconnect attempts failed with:
  `org.bluez.Error.Failed br-connection-page-timeout`
- That error usually indicates headset availability/state, not local adapter startup problems.
