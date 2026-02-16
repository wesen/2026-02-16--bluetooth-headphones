---
Title: SoundCtl Bluetooth and Audio TUI Source Spec
Ticket: BT-001-HEADPHONES
Status: active
Topics:
    - bluetooth
    - ubuntu
    - audio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - /tmp/headphones.md
Summary: Imported source specification for SoundCtl screens, model/message DSL, subscriptions, and keybindings.
LastUpdated: 2026-02-16T14:10:00-05:00
WhatFor: Source blueprint for Bubble Tea/Bubbles implementation planning.
WhenToUse: Use when implementing or reviewing SoundCtl UI architecture and behaviors.
---

# 🎧 SoundCtl — Bluetooth & Audio TUI

---

## Screen 1: Main Dashboard

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  ▸ Devices        Sinks        Profiles                 │
│                                                         │
│  ┌─ Bluetooth ─────────────────────────────────────┐    │
│  │  ● Sony WH-1000XM5          Connected   ■■■ 85% │    │
│  │  ○ AirPods Pro               Paired              │    │
│  │  ○ JBL Flip 6                Saved               │    │
│  │                                                   │    │
│  │  [ Scan ]  [ Disconnect ]  [ Forget ]             │    │
│  └───────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─ Volume ────────────────────────────────────────┐    │
│  │  Master   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░  72%         │    │
│  │  Media    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░  90%         │    │
│  │  Alerts   ▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░  35%         │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  q quit  ↑↓ navigate  enter select  s scan  / search   │
└─────────────────────────────────────────────────────────┘
```

## Screen 2: Sinks Tab

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  Devices        ▸ Sinks        Profiles                 │
│                                                         │
│  ┌─ Output Sinks ─────────────────────────────────┐    │
│  │  ★ Sony WH-1000XM5 (A2DP Sink)      [default]  │    │
│  │    Built-in Audio Analog Stereo                  │    │
│  │    HDMI / DisplayPort                            │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─ Input Sources ────────────────────────────────┐    │
│  │  ★ Sony WH-1000XM5 (HSP/HFP)        [default]  │    │
│  │    Built-in Audio Analog Stereo                  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─ App Routing ──────────────────────────────────┐    │
│  │  Firefox        → Sony WH-1000XM5              │    │
│  │  Spotify        → Sony WH-1000XM5              │    │
│  │  Discord        → Built-in Audio    🔀 reroute  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  d set-default  r reroute  m mute  tab next-tab        │
└─────────────────────────────────────────────────────────┘
```

## Screen 3: Profiles Tab

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  Devices        Sinks        ▸ Profiles                 │
│                                                         │
│  ┌─ Sony WH-1000XM5 ─────────────────────────────┐    │
│  │  ● A2DP Sink (High Fidelity Playback)           │    │
│  │  ○ HSP/HFP (Headset Head Unit)                  │    │
│  │  ○ Off                                           │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─ Built-in Audio ───────────────────────────────┐    │
│  │  ● Analog Stereo Duplex                         │    │
│  │  ○ Analog Stereo Output                         │    │
│  │  ○ Digital Stereo (IEC958) Output               │    │
│  │  ○ Off                                           │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  enter apply  ↑↓ navigate  tab next-tab  q quit        │
└─────────────────────────────────────────────────────────┘
```

## Screen 4: Scanning Overlay

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  ▸ Devices        Sinks        Profiles                 │
│                                                         │
│  ┌─ Bluetooth ───────┬─────────────────────────┐       │
│  │  ● Sony WH-1000X  │  ┌─ Scanning... ⠋ ───┐ │       │
│  │  ○ AirPods Pro     │  │                    │ │       │
│  │  ○ JBL Flip 6      │  │  JBL Charge 5      │ │       │
│  │                    │  │  Bose QC45          │ │       │
│  │                    │  │  Unknown (4A:3F..)  │ │       │
│  │                    │  │                    │ │       │
│  │                    │  │  enter pair         │ │       │
│  │                    │  │  esc   cancel       │ │       │
│  │                    │  └────────────────────┘ │       │
│  └────────────────────┴─────────────────────────┘       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Model & Message DSL

```yaml
app:
  model: AppModel
  children:
    tabs: TabBar
    devices: DevicesPane
    sinks: SinksPane
    profiles: ProfilesPane
    scanner: ScanOverlay
    status: StatusBar

  messages:
    handles:
      - KeyMsg
      - WindowSizeMsg
      - BluetoothEventMsg      # from dbus subscription
      - PulseAudioEventMsg     # from PA subscription
    emits:
      - TabChangedMsg
      - QuitMsg

TabBar:
  model:
    active_tab: int            # 0=Devices 1=Sinks 2=Profiles
  messages:
    handles: [TabChangedMsg, KeyMsg]
    emits:   [TabChangedMsg]

DevicesPane:
  model:
    devices:    []BluetoothDevice
    cursor:     int
    focused:    bool
  children:
    volume: VolumeGroup
  messages:
    handles:
      - KeyMsg
      - BluetoothEventMsg      # device added/removed/changed
      - ConnectResultMsg
      - DisconnectResultMsg
      - ForgetResultMsg
    emits:
      - ConnectCmd             # → bluetoothctl connect
      - DisconnectCmd          # → bluetoothctl disconnect
      - ForgetCmd              # → bluetoothctl remove
      - OpenScannerMsg
    commands:
      - ConnectCmd:      "bluetoothctl connect {addr}"
      - DisconnectCmd:   "bluetoothctl disconnect {addr}"
      - ForgetCmd:       "bluetoothctl remove {addr}"
      - BatteryPollCmd:  "bluetoothctl info {addr} | grep Battery"

VolumeGroup:
  model:
    channels:  []Channel       # {name, level, muted}
    cursor:    int
  messages:
    handles:  [KeyMsg, PulseAudioEventMsg]
    emits:    [SetVolumeCmd, ToggleMuteCmd]
    commands:
      - SetVolumeCmd:    "pactl set-sink-volume {sink} {pct}%"
      - ToggleMuteCmd:   "pactl set-sink-mute {sink} toggle"

SinksPane:
  model:
    outputs:     []Sink
    inputs:      []Source
    app_routes:  []StreamRoute  # {app_name, sink_name, sink_id}
    cursor:      int
    section:     enum[outputs, inputs, routes]
  messages:
    handles:
      - KeyMsg
      - PulseAudioEventMsg
      - SetDefaultResultMsg
      - MoveStreamResultMsg
    emits:
      - SetDefaultSinkCmd
      - SetDefaultSourceCmd
      - MoveStreamCmd
    commands:
      - SetDefaultSinkCmd:    "pactl set-default-sink {sink}"
      - SetDefaultSourceCmd:  "pactl set-default-source {source}"
      - MoveStreamCmd:        "pactl move-sink-input {stream_id} {sink}"

ProfilesPane:
  model:
    cards:    []Card           # {name, profiles[], active_profile}
    cursor:   int
  messages:
    handles:  [KeyMsg, PulseAudioEventMsg, SetProfileResultMsg]
    emits:    [SetProfileCmd]
    commands:
      - SetProfileCmd: "pactl set-card-profile {card} {profile}"

ScanOverlay:
  model:
    visible:     bool
    discovered:  []BluetoothDevice
    cursor:      int
    spinner:     spinner.Model
  messages:
    handles:
      - KeyMsg
      - OpenScannerMsg
      - BluetoothEventMsg     # new device discovered
      - PairResultMsg
      - spinner.TickMsg
    emits:
      - StartScanCmd
      - StopScanCmd
      - PairCmd
      - CloseScannerMsg
    commands:
      - StartScanCmd:  "bluetoothctl scan on"
      - StopScanCmd:   "bluetoothctl scan off"
      - PairCmd:       "bluetoothctl pair {addr} && bluetoothctl trust {addr}"

StatusBar:
  model:
    message:   string
    err:       error
    timeout:   timer.Model
  messages:
    handles:  [StatusMsg, ErrorMsg, timer.TimeoutMsg]

# ── Subscriptions (long-lived) ──
subscriptions:
  - name: BluetoothMonitor
    impl: "dbus-monitor --system org.bluez"
    emits: BluetoothEventMsg

  - name: PulseAudioMonitor
    impl: "pactl subscribe"
    emits: PulseAudioEventMsg

# ── Key Bindings ──
keybindings:
  global:
    q:      QuitMsg
    tab:    TabChangedMsg{+1}
    S-tab:  TabChangedMsg{-1}
    "?":    ToggleHelpMsg
  devices:
    s:      OpenScannerMsg
    enter:  ConnectCmd
    D:      DisconnectCmd
    X:      ForgetCmd
    "←/→":  VolumeAdjust{±5}
  sinks:
    d:      SetDefaultSinkCmd
    r:      MoveStreamCmd
    m:      ToggleMuteCmd
  profiles:
    enter:  SetProfileCmd
  scanner:
    enter:  PairCmd
    esc:    CloseScannerMsg
```
