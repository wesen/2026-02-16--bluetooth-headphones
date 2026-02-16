## Screen 5: Presets Tab

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  Devices     Sinks     Profiles     ▸ Presets           │
│                                                         │
│  ┌─ Saved Presets ────────────────────────────────┐    │
│  │  ★ Music Mode                          [active] │    │
│  │    XM5→A2DP  Master:80%  Route:all→XM5          │    │
│  │                                                  │    │
│  │    Video Call                                    │    │
│  │    XM5→HSP/HFP  Master:60%  Discord→XM5         │    │
│  │                                                  │    │
│  │    Speakers + Low                                │    │
│  │    XM5→Off  Master:40%  Route:all→Built-in       │    │
│  │                                                  │    │
│  │    Late Night                                    │    │
│  │    XM5→A2DP  Master:25%  Alerts:0%               │    │
│  └──────────────────────────────────────────────────┘    │
│                                                         │
│  enter apply  n new  e edit  d delete  c clone  q quit  │
└─────────────────────────────────────────────────────────┘
```

## Screen 6: Create/Edit Preset

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  ┌─ Edit Preset ──────────────────────────────────┐    │
│  │                                                  │    │
│  │  Name: [ Late Night Gaming___________]           │    │
│  │                                                  │    │
│  │  ── Profiles ──────────────────────────────      │    │
│  │  Sony WH-1000XM5:   ◀ A2DP Sink ▶               │    │
│  │  Built-in Audio:     ◀ Off ▶                     │    │
│  │                                                  │    │
│  │  ── Volumes ───────────────────────────────      │    │
│  │  Master  ▓▓▓▓▓▓░░░░░░░░░░░░░░░░  25%            │    │
│  │  Media   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░  90%            │    │
│  │  Alerts  ░░░░░░░░░░░░░░░░░░░░░░   0%   🔇       │    │
│  │                                                  │    │
│  │  ── App Routing ───────────────────────────      │    │
│  │  Default sink:  Sony WH-1000XM5                  │    │
│  │  + Firefox      ◀ follow default ▶               │    │
│  │  + Discord      ◀ follow default ▶               │    │
│  │  + Spotify      ◀ follow default ▶               │    │
│  │                                                  │    │
│  │        [ 💾 Save ]          [ Cancel ]            │    │
│  └──────────────────────────────────────────────────┘    │
│                                                         │
│  tab next-field  ←→ cycle option  ↑↓ navigate          │
└─────────────────────────────────────────────────────────┘
```

## Screen 7: Apply Confirmation

```
┌─ SoundCtl ──────────────────────────────────────────────┐
│                                                         │
│  Devices     Sinks     Profiles     ▸ Presets           │
│                                                         │
│  ┌─ Saved Presets ──────────────────────────┐          │
│  │  ★ Music Mode                   [active] │          │
│  │  ┌─ Apply "Video Call"? ──────────────┐  │          │
│  │  │                                    │  │          │
│  │  │  Changes:                          │  │          │
│  │  │   XM5 profile  A2DP → HSP/HFP     │  │          │
│  │  │   Master vol    80% → 60%          │  │          │
│  │  │   Firefox      XM5 → Built-in      │  │          │
│  │  │                                    │  │          │
│  │  │      [ Apply ]    [ Cancel ]       │  │          │
│  │  └────────────────────────────────────┘  │          │
│  │    Speakers + Low                        │          │
│  │    Late Night                            │          │
│  └──────────────────────────────────────────┘          │
│                                                         │
│  ✓ Preset "Music Mode" applied                          │
└─────────────────────────────────────────────────────────┘
```

---

## Updated YAML DSL

```yaml
# ── New tab added to AppModel ──
app:
  children:
    tabs:     TabBar          # now 4 tabs
    devices:  DevicesPane
    sinks:    SinksPane
    profiles: ProfilesPane
    presets:  PresetsPane     # ← new
    scanner:  ScanOverlay
    status:   StatusBar

# ── Preset Data Structures ──
types:
  Preset:
    name:           string
    card_profiles:  map[card_id → profile_name]
    volumes:        map[channel_name → {level: int, muted: bool}]
    default_sink:   sink_id
    app_routes:     map[app_name → sink_id | "follow_default"]
    created_at:     time
    updated_at:     time

  PresetDiff:
    changes:  []DiffLine     # {field, from, to}

# ── Presets Pane ──
PresetsPane:
  model:
    presets:       []Preset
    cursor:        int
    active_preset: *string       # name of currently applied preset
    editor:        *PresetEditor # nil when closed
    confirm:       *ConfirmModal # nil when closed
  messages:
    handles:
      - KeyMsg
      - ApplyPresetResultMsg
      - SavePresetResultMsg
      - DeletePresetResultMsg
      - SnapshotCurrentMsg       # captures live state into editor
    emits:
      - ApplyPresetCmd
      - SavePresetCmd
      - DeletePresetCmd
      - OpenEditorMsg
      - CloseEditorMsg
      - OpenConfirmMsg
      - CloseConfirmMsg

PresetEditor:
  model:
    preset:        Preset
    is_new:        bool
    focused_field: enum[name, profiles, volumes, routes, buttons]
    profile_cursors:  map[card_id → int]
    volume_cursor:    int
    route_cursor:     int
  messages:
    handles:
      - KeyMsg
      - SnapshotCurrentMsg     # populates fields from live state
    emits:
      - SavePresetCmd
      - CloseEditorMsg
  interactions:
    name_field:    "text input, printable keys"
    profile_sel:   "←→ cycles through available profiles"
    volume_bars:   "←→ adjusts ±5%,  m toggles mute"
    route_sel:     "←→ cycles sink or 'follow default'"

ConfirmModal:
  model:
    preset:     Preset
    diff:       PresetDiff
    cursor:     enum[apply, cancel]
  messages:
    handles:  [KeyMsg]
    emits:    [ApplyPresetCmd, CloseConfirmMsg]

# ── Apply Preset Command Sequence ──
ApplyPresetCmd:
  description: "Batch applies all preset settings"
  steps:
    - for card, profile in preset.card_profiles:
        "pactl set-card-profile {card} {profile}"
    - for channel in preset.volumes:
        "pactl set-sink-volume {sink} {level}%"
        "pactl set-sink-mute {sink} {muted}"
    - "pactl set-default-sink {preset.default_sink}"
    - for app, sink in preset.app_routes:
        if sink != "follow_default":
          "pactl move-sink-input {stream_id} {sink}"
  emits: ApplyPresetResultMsg

# ── Persistence ──
persistence:
  file: "$XDG_CONFIG_HOME/soundctl/presets.json"
  commands:
    SavePresetCmd:   "write preset to JSON file"
    DeletePresetCmd: "remove preset from JSON file"
    LoadPresetsCmd:  "read all presets on startup"

# ── New Key Bindings ──
keybindings:
  presets:
    enter:  OpenConfirmMsg        # show diff + confirm
    n:      OpenEditorMsg{new}    # blank editor
    e:      OpenEditorMsg{edit}   # edit selected
    c:      OpenEditorMsg{clone}  # clone selected
    d:      DeletePresetCmd
    S:      SnapshotCurrentMsg    # capture current state as new preset
  editor:
    tab:    NextField
    S-tab:  PrevField
    "←/→":  CycleOption / AdjustVolume
    m:      ToggleMute
    enter:  SavePresetCmd         # on save button
    esc:    CloseEditorMsg
  confirm:
    enter:  ApplyPresetCmd
    esc:    CloseConfirmMsg
    "←/→":  ToggleConfirmCursor
```

The key addition is `S` (Snapshot) — it captures your **current live state** (active profiles, volumes, routing) directly into the editor so you can name it and save, rather than building presets from scratch.
