# Sound

Peddle sound is driven by a small IRQ runtime. Programs build a timed stream in
a `byte[]`, give the runtime a separate memory pool, load the stream into that
pool, then start playback by handle.

The public stream kind is:

```peddle
const SOUND_STREAM = 1
```

The old raw register stream bytecode is replaced by the unified `SOUND_STREAM`
bytecode. Use command `9` for raw SID register writes.

The runtime supports up to four active stream players. Each player owns an
explicit voice mask, and priority decides whether overlay playback can take a
voice that is already busy.

---

# Constants

```peddle
const SOUND_STREAM = 1

const SOUND_VOICE1 = 1
const SOUND_VOICE2 = 2
const SOUND_VOICE3 = 4
const SOUND_ALL = 7

const SOUND_REPLACE = 1
const SOUND_OVERLAY = 2
const SOUND_LOOP = 4

const VOICE1 = 0
const VOICE2 = 1
const VOICE3 = 2

const WAVE_TRIANGLE = 16
const WAVE_SAW = 32
const WAVE_PULSE = 64
const WAVE_NOISE = 128
const GATE = 1

const SID_VOLUME = 24
```

Useful raw SID offsets:

```peddle
const SID_VOICE1 = 0
const SID_VOICE2 = 7
const SID_VOICE3 = 14

const SID_FREQ_LO = 0
const SID_FREQ_HI = 1
const SID_PULSE_LO = 2
const SID_PULSE_HI = 3
const SID_CONTROL = 4
const SID_ATTACK_DECAY = 5
const SID_SUSTAIN_RELEASE = 6

const SID_FILTER_CUTOFF_LO = 21
const SID_FILTER_CUTOFF_HI = 22
const SID_FILTER_RESONANCE = 23
const SID_MODE_VOLUME = 24
```

Logical voices are mapped internally:

```text
VOICE1 -> SID base offset 0
VOICE2 -> SID base offset 7
VOICE3 -> SID base offset 14
```

---

# Runtime API

| Function | Description |
|---|---|
| `sound_init(pool)` | initialize sound runtime with a user-provided `byte[]` pool |
| `sound_reset()` | stop playback, clear loaded sounds, and reset the pool length |
| `sound_load(data, kind)` | load sound bytes and return `(uint, int)` |
| `sound_play(id, voices, priority, flags)` | start a loaded sound and return `int` |
| `sound_stop(id)` | stop active players using the handle |
| `sound_num()` | return number of loaded sounds |
| `sound_memfree()` | return remaining pool bytes |

`sound_init(pool)` installs the IRQ player and clears the pool. Calling it again
with the same or another pool stops playback and clears all loaded sounds.

`sound_load(data, SOUND_STREAM)` copies `len(data)` bytes into the pool. It
returns a sound handle and an error code:

```peddle
id, err = sound_load(data, SOUND_STREAM)
```

There is no per-sound free operation. Reclaim sound memory with `sound_reset()`
or another `sound_init(pool)`.

`sound_play(id, voices, priority, flags)` starts a loaded stream. `voices` is a
mask of `SOUND_VOICE1`, `SOUND_VOICE2`, and `SOUND_VOICE3`; it selects which
logical voices from the stream may write SID voice registers. `SOUND_REPLACE`
stops all active players before starting. `SOUND_OVERLAY` starts alongside
current players when voices are free, or takes busy voices from lower/equal
priority players. `SOUND_LOOP` restarts the stream when it reaches `0`.

---

# `sound_load` Error Codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | sound runtime was not initialized |
| `2` | unsupported sound kind |
| `3` | empty sound data |
| `4` | no free sound slot |
| `5` | not enough pool memory |
| `6` | invalid stream data |

# `sound_play` Error Codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | sound runtime was not initialized |
| `2` | invalid sound handle |
| `3` | invalid voice mask |
| `4` | no free active player slot |
| `5` | requested voice is busy |
| `6` | invalid flags |

---

# Stream Commands

`SOUND_STREAM` is a timed event stream:

```text
0                     end
1, frames             wait frames
2, voice, note        set note frequency for logical voice
3, voice              gate off logical voice
4, voice, waveform    set waveform/control for logical voice
5, voice, ad          set attack/decay for logical voice
6, voice, sr          set sustain/release for logical voice
7, volume             set global volume
8, voice, lo, hi      set raw frequency for logical voice
9, reg, value         raw SID register write
```

Time advances only when the stream contains a wait command. All commands before
the next wait command execute in the same logical sound frame.

The runtime validates loaded streams for unsupported commands, truncated
commands, invalid voices, invalid note numbers, invalid raw SID registers,
empty data, and missing end commands.

---

# Notes

Command `2` uses a small runtime note table. The convention is:

```text
NOTE_C4  = 48
NOTE_CS4 = 49
NOTE_D4  = 50
...
```

The current table covers notes `0..84`, from C0 through C7.

For precise frequencies or values outside the table, use command `8` with raw
frequency low/high bytes.

---

# Helper Functions

Build streams with `append()`:

```peddle
fn wait(data byte[128], frames byte) {
    append(data, 1)
    append(data, frames)
}

fn note(data byte[128], voice byte, n byte) {
    append(data, 2)
    append(data, voice)
    append(data, n)
}

fn gateOff(data byte[128], voice byte) {
    append(data, 3)
    append(data, voice)
}

fn waveform(data byte[128], voice byte, value byte) {
    append(data, 4)
    append(data, voice)
    append(data, value)
}

fn env(data byte[128], voice byte, ad byte, sr byte) {
    append(data, 5)
    append(data, voice)
    append(data, ad)

    append(data, 6)
    append(data, voice)
    append(data, sr)
}

fn volume(data byte[128], value byte) {
    append(data, 7)
    append(data, value)
}

fn freq(data byte[128], voice byte, lo byte, hi byte) {
    append(data, 8)
    append(data, voice)
    append(data, lo)
    append(data, hi)
}

fn raw(data byte[128], reg byte, value byte) {
    append(data, 9)
    append(data, reg)
    append(data, value)
}

fn gateOn(data byte[128], voice byte, wave byte) {
    waveform(data, voice, wave + GATE)
}
```

Adjust the array size in helper signatures when the target stream is larger
than `byte[128]`.

---

# Three-Voice Chord

These voice setup commands occur before the first wait, so all three voices
start in the same logical sound frame.

```peddle
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1

const VOICE1 = 0
const VOICE2 = 1
const VOICE3 = 2

const WAVE_TRIANGLE = 16
const GATE = 1

fn volume(data byte[128], value byte) {
    append(data, 7)
    append(data, value)
}

fn wait(data byte[128], frames byte) {
    append(data, 1)
    append(data, frames)
}

fn waveform(data byte[128], voice byte, value byte) {
    append(data, 4)
    append(data, voice)
    append(data, value)
}

fn env(data byte[128], voice byte, ad byte, sr byte) {
    append(data, 5)
    append(data, voice)
    append(data, ad)

    append(data, 6)
    append(data, voice)
    append(data, sr)
}

fn freq(data byte[128], voice byte, lo byte, hi byte) {
    append(data, 8)
    append(data, voice)
    append(data, lo)
    append(data, hi)
}

fn gateOn(data byte[128], voice byte, wave byte) {
    waveform(data, voice, wave + GATE)
}

fn gateOff(data byte[128], voice byte, wave byte) {
    waveform(data, voice, wave)
}

fn main() {
    var pool byte[512]
    var song byte[128]
    var id uint
    var err int

    volume(song, 15)

    env(song, VOICE1, 9, 240)
    env(song, VOICE2, 9, 240)
    env(song, VOICE3, 9, 240)

    freq(song, VOICE1, 103, 17)
    freq(song, VOICE2, 237, 21)
    freq(song, VOICE3, 69, 29)

    gateOn(song, VOICE1, WAVE_TRIANGLE)
    gateOn(song, VOICE2, WAVE_TRIANGLE)
    gateOn(song, VOICE3, WAVE_TRIANGLE)

    wait(song, 60)

    gateOff(song, VOICE1, WAVE_TRIANGLE)
    gateOff(song, VOICE2, WAVE_TRIANGLE)
    gateOff(song, VOICE3, WAVE_TRIANGLE)

    append(song, 0)

    sound_init(pool)
    id, err = sound_load(song, SOUND_STREAM)

    if err == 0 {
        err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    }
}
```

---

# Raw SID Writes

Command `9` keeps low-level SID access available. It writes `value` to
`$D400 + reg`, where `reg` must be `0..24`.

```peddle
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1
const SID_VOLUME = 24

fn raw(data byte[32], reg byte, value byte) {
    append(data, 9)
    append(data, reg)
    append(data, value)
}

fn main() {
    var pool byte[64]
    var data byte[32]
    var id uint
    var err int

    raw(data, SID_VOLUME, 15)
    append(data, 0)

    sound_init(pool)
    id, err = sound_load(data, SOUND_STREAM)

    if err == 0 {
        err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    }
}
```

---

# Multiple Sounds

You can load several streams and keep their handles. Use `SOUND_REPLACE` for
exclusive playback, or `SOUND_OVERLAY` to layer streams on different voices.

```peddle
const SOUND_VOICE1 = 1
const SOUND_VOICE2 = 2
const SOUND_REPLACE = 1
const SOUND_OVERLAY = 2

sound_init(pool)
shotId, shotErr = sound_load(shotData, SOUND_STREAM)
hitId, hitErr = sound_load(hitData, SOUND_STREAM)

if shotErr == 0 {
    shotErr = sound_play(shotId, SOUND_VOICE1, 4, SOUND_OVERLAY)
}

if hitErr == 0 {
    hitErr = sound_play(hitId, SOUND_VOICE2, 8, SOUND_OVERLAY)
}
```

Use `sound_stop(id)` to stop active players that use the same handle:

```peddle
sound_stop(shotId)
```

Use `sound_reset()` to stop playback, unload every sound, and reuse the whole
pool.
