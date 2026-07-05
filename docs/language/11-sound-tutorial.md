# How to Program Sound Effects and Music

This tutorial shows how to build C64 sound in Peddle from small pieces:

- one beep
- envelopes
- note numbers
- short sound effects
- three simultaneous voices
- overlay playback
- stopping selected voices
- a small three-voice song

For the full API reference, see [Sound](09-sound.md). This tutorial focuses on
how to grow working sound code step by step.

Every step has a complete runnable checkpoint file:

| Step | Checkpoint |
|---|---|
| 1 | [sound_step01_beep.ped](../../examples/tutorial/sound_step01_beep.ped) |
| 2 | [sound_step02_envelope.ped](../../examples/tutorial/sound_step02_envelope.ped) |
| 3 | [sound_step03_notes.ped](../../examples/tutorial/sound_step03_notes.ped) |
| 4 | [sound_step04_effects.ped](../../examples/tutorial/sound_step04_effects.ped) |
| 5 | [sound_step05_three_voices.ped](../../examples/tutorial/sound_step05_three_voices.ped) |
| 6 | [sound_step06_overlay.ped](../../examples/tutorial/sound_step06_overlay.ped) |
| 7 | [sound_step07_stop_voices.ped](../../examples/tutorial/sound_step07_stop_voices.ped) |
| 8 | [sound_step08_small_song.ped](../../examples/tutorial/sound_step08_small_song.ped) |

---

# Step 1: Build One Beep

Peddle sound starts with a byte stream. A program creates a `byte[]`, appends
sound commands to it, loads it into a sound pool, then plays the loaded handle.

The minimum setup looks like this:

```peddle
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1
const ERR_OK = 0
```

You also need logical voice and waveform constants:

```peddle
const VOICE1 = 0
const WAVE_TRIANGLE = 16
const GATE = 1
```

A sound stream must end with command `0`:

```peddle
append(data, 0)
```

To play it:

```peddle
sound_init(pool)
id, err = sound_load(data, SOUND_STREAM)

if err == ERR_OK {
    err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
}
```

The first checkpoint builds one triangle-wave beep with raw SID frequency
values.

**Checkpoint**

Complete runnable code for this step:
[sound_step01_beep.ped](../../examples/tutorial/sound_step01_beep.ped)

```sh
./peddle.sh examples/tutorial/sound_step01_beep.ped
```

---

# Step 2: Shape the Sound With Envelopes

The SID envelope controls how a note starts, holds, and fades. Peddle exposes
the attack/decay and sustain/release registers as stream commands:

```text
5, voice, ad
6, voice, sr
```

A helper keeps the stream readable:

```peddle
fn env(data byte[96], voice byte, ad byte, sr byte) {
    append(data, 5)
    append(data, voice)
    append(data, ad)
    append(data, 6)
    append(data, voice)
    append(data, sr)
}
```

The checkpoint plays two notes at the same frequency: first with a fast attack,
then with a slower attack. This makes the envelope easy to hear.

**Checkpoint**

Complete runnable code for this step:
[sound_step02_envelope.ped](../../examples/tutorial/sound_step02_envelope.ped)

```sh
./peddle.sh examples/tutorial/sound_step02_envelope.ped
```

---

# Step 3: Use Note Numbers

Raw frequency values are useful for effects, but music is easier with note
numbers. Command `2` sets a note for a logical voice:

```text
2, voice, note
```

Peddle uses the convention:

```text
NOTE_C4  = 48
NOTE_CS4 = 49
NOTE_D4  = 50
```

A small helper:

```peddle
fn note(data byte[128], voice byte, n byte) {
    append(data, 2)
    append(data, voice)
    append(data, n)
}
```

The checkpoint turns note numbers into a tiny melody.

**Checkpoint**

Complete runnable code for this step:
[sound_step03_notes.ped](../../examples/tutorial/sound_step03_notes.ped)

```sh
./peddle.sh examples/tutorial/sound_step03_notes.ped
```

---

# Step 4: Build Game Sound Effects

Game effects are often more about motion than pitch accuracy. A laser can sweep
frequency downward. An explosion can use the noise waveform.

Useful waveforms:

```peddle
const WAVE_TRIANGLE = 16
const WAVE_SAW = 32
const WAVE_PULSE = 64
const WAVE_NOISE = 128
```

For a sweep, keep the gate open and change frequency between waits:

```peddle
gateOn(data, VOICE1, WAVE_TRIANGLE)
freq(data, VOICE1, 40, 40)
soundWait(data, 3)
freq(data, VOICE1, 10, 32)
soundWait(data, 3)
freq(data, VOICE1, 220, 23)
```

For a hit or explosion, switch to `WAVE_NOISE` and use a short envelope.

**Checkpoint**

Complete runnable code for this step:
[sound_step04_effects.ped](../../examples/tutorial/sound_step04_effects.ped)

```sh
./peddle.sh examples/tutorial/sound_step04_effects.ped
```

---

# Step 5: Use Three Voices Together

The SID has three voices. Peddle streams use logical voices:

```text
VOICE1 = 0
VOICE2 = 1
VOICE3 = 2
```

Commands before a wait happen in the same sound frame. That means you can set
up all three voices, gate them on, and then wait:

```peddle
note(data, VOICE1, NOTE_C4)
note(data, VOICE2, NOTE_E4)
note(data, VOICE3, NOTE_G4)

gateOn(data, VOICE1, WAVE_TRIANGLE)
gateOn(data, VOICE2, WAVE_TRIANGLE)
gateOn(data, VOICE3, WAVE_SAW)

soundWait(data, 90)
```

Play the stream with all voices enabled:

```peddle
err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
```

**Checkpoint**

Complete runnable code for this step:
[sound_step05_three_voices.ped](../../examples/tutorial/sound_step05_three_voices.ped)

```sh
./peddle.sh examples/tutorial/sound_step05_three_voices.ped
```

---

# Step 6: Overlay Sounds

Most games need more than one sound at once. For example, a looping background
drone can run on voice 2 while a short ping plays on voice 1.

Use a voice mask when playing:

```peddle
err = sound_play(droneId, SOUND_VOICE2, 1, SOUND_REPLACE + SOUND_LOOP)
err = sound_play(pingId, SOUND_VOICE1, 8, SOUND_OVERLAY)
```

`SOUND_OVERLAY` starts another player without stopping unrelated voices.
Priority decides what happens if the requested voice is already busy. Higher
priority can take a voice from lower priority playback.

The checkpoint starts a loop on voice 2 and overlays two pings on voice 1.

**Checkpoint**

Complete runnable code for this step:
[sound_step06_overlay.ped](../../examples/tutorial/sound_step06_overlay.ped)

```sh
./peddle.sh examples/tutorial/sound_step06_overlay.ped
```

---

# Step 7: Stop Selected Voices

Sometimes you do not want to stop every sound. `sound_stop_voices(mask)` stops
only the selected SID voices:

```peddle
sound_stop_voices(SOUND_VOICE2)
sound_stop_voices(SOUND_VOICE1 + SOUND_VOICE3)
```

This is useful for layered playback. A game can stop a background drone while
leaving a short effect alone, or remove one part of a chord.

**Checkpoint**

Complete runnable code for this step:
[sound_step07_stop_voices.ped](../../examples/tutorial/sound_step07_stop_voices.ped)

```sh
./peddle.sh examples/tutorial/sound_step07_stop_voices.ped
```

---

# Step 8: Make a Small Song

A song is just repeated sound patterns. The final checkpoint defines a `step()`
helper that starts three notes, waits, gates them off, then adds a small gap:

```peddle
fn step(song byte[1024], melody byte, bass byte, harmony byte) {
    note(song, VOICE1, melody)
    note(song, VOICE2, bass)
    note(song, VOICE3, harmony)
    gateOn(song, VOICE1, WAVE_TRIANGLE)
    gateOn(song, VOICE2, WAVE_TRIANGLE)
    gateOn(song, VOICE3, WAVE_SAW)
    soundWait(song, NOTE_ON)
    gateOff(song, VOICE1, WAVE_TRIANGLE)
    gateOff(song, VOICE2, WAVE_TRIANGLE)
    gateOff(song, VOICE3, WAVE_SAW)
    soundWait(song, NOTE_GAP)
}
```

Then the song builder is a readable list of musical steps:

```peddle
step(song, NOTE_C4, NOTE_C3, NOTE_E4)
step(song, NOTE_D4, NOTE_G3, NOTE_G4)
step(song, NOTE_E4, NOTE_C3, NOTE_C5)
```

This is the same pattern used by larger examples such as
[fuer_elise.ped](../../examples/fuer_elise.ped), only smaller.

**Checkpoint**

Complete runnable code for this step:
[sound_step08_small_song.ped](../../examples/tutorial/sound_step08_small_song.ped)

```sh
./peddle.sh examples/tutorial/sound_step08_small_song.ped
```

---

# Design Checklist

When adding sound to a game, decide these things first:

- Which sounds are short effects?
- Which sounds should loop?
- Which SID voice should each sound use?
- Which sounds may overlay?
- Which sounds should replace existing playback?
- Which priority should important effects get?
- How much pool memory do you need?

For many games, a simple voice plan works well:

| Voice | Use |
|---|---|
| Voice 1 | player actions, shots, menu pings |
| Voice 2 | enemy actions or bass/drone |
| Voice 3 | explosions, noise, warnings |

Keep effect streams short and explicit. Build music from small step helpers.
Use `SOUND_OVERLAY` for independent effects and `SOUND_REPLACE` when a sound
should clearly take over.

---

# Next Improvements

Good next steps after this tutorial:

- move repeated sound helpers into an imported `lib/sound.ped`
- add named note constants for more octaves
- add a small table of ADSR presets
- build reusable effects like `laser`, `hit`, `miss`, and `powerup`
- add support for tracker/module imports later

Keep each sound small while tuning it. The SID rewards tiny changes.
