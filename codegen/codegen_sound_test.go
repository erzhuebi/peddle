package codegen

import "testing"

func TestCodegenSoundRuntimeAndBuiltinReturns(t *testing.T) {
	asm := compileSource(t, `
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1

fn main() {
    var pool byte[64]
    var data byte[8]
    var id uint
    var err int
    var n int

    data[0] = 9
    data[1] = 24
    data[2] = 15
    data[3] = 0

    sound_init(pool)
    id, err = sound_load(data, SOUND_STREAM)
    err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    sound_stop(id)
    n = sound_num()
    n = sound_memfree()
    sound_reset()
}
`)

	requireASM(t, asm,
		"jsr peddle_sound_shutdown",
		"jsr peddle_sound_init",
		"jsr peddle_sound_load",
		"jsr peddle_sound_play",
		"jsr peddle_sound_stop",
		"jsr peddle_sound_num",
		"jsr peddle_sound_memfree",
		"jsr peddle_sound_reset",
		"peddle_sound_load_return_id:",
		"peddle_sound_load_return_err:",
		"peddle_sound_play_return_err:",
		"lda peddle_sound_load_return_id",
		"sta main_id",
		"lda peddle_sound_load_return_err",
		"sta main_err",
		"sta peddle_sound_play_voices_lo",
		"sta peddle_sound_play_priority_lo",
		"sta peddle_sound_play_flags_lo",
		"peddle_sound_slot_inuse:",
		".fill 16, 0",
		"peddle_sound_player_inuse:",
		".fill 4, 0",
		"lda $0314",
		"sta peddle_sound_old_irq_lo",
		"sta $0314",
		"jmp (peddle_sound_old_irq_lo)",
		"PEDDLE_SOUND_STREAM     = 1",
		"peddle_sound_validate_stream:",
		"peddle_sound_irq_raw:",
		"sta $d400, x",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenSoundStreamSupportsThreeVoicesBeforeOneWait(t *testing.T) {
	asm := compileSource(t, `
const SOUND_STREAM = 1
const VOICE1 = 0
const VOICE2 = 1
const VOICE3 = 2
const WAVE_TRIANGLE = 16
const GATE = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1

fn volume(data byte[128], value byte) {
    append(data, 7)
    append(data, value)
}

fn wait(data byte[128], frames byte) {
    append(data, 1)
    append(data, frames)
}

fn note(data byte[128], voice byte, n byte) {
    append(data, 2)
    append(data, voice)
    append(data, n)
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
    var playErr int

    volume(song, 15)
    note(song, VOICE1, 48)

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
        playErr = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    }
}
`)

	requireASM(t, asm,
		"peddle_sound_irq_command_loop:",
		"peddle_sound_irq_wait:",
		"peddle_sound_irq_note:",
		"peddle_sound_irq_gate_off:",
		"peddle_sound_irq_waveform:",
		"peddle_sound_irq_ad:",
		"peddle_sound_irq_sr:",
		"peddle_sound_irq_volume:",
		"peddle_sound_irq_freq:",
		"peddle_sound_irq_player_loop:",
		"peddle_sound_play_check_conflicts:",
		"peddle_sound_play_stop_conflicts:",
		"peddle_sound_voice_base_1:",
		"lda #7",
		"peddle_sound_voice_base_2:",
		"lda #14",
	)

	requireASMOrder(t, asm,
		"peddle_sound_irq_volume:",
		"jsr peddle_sound_irq_advance_2",
		"jmp peddle_sound_irq_command_loop",
	)
	requireASMOrder(t, asm,
		"peddle_sound_irq_freq:",
		"jsr peddle_sound_irq_advance_4",
		"jmp peddle_sound_irq_command_loop",
	)
	requireASMOrder(t, asm,
		"peddle_sound_irq_waveform:",
		"jsr peddle_sound_irq_advance_3",
		"jmp peddle_sound_irq_command_loop",
	)
	requireASMOrder(t, asm,
		"peddle_sound_irq_wait:",
		"ldx peddle_sound_irq_player_index",
		"sta peddle_sound_player_wait, x",
		"jsr peddle_sound_irq_advance_2",
		"rts",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenSoundStreamValidationRejectsInvalidVoiceAndRawRegister(t *testing.T) {
	asm := compileSource(t, `
const SOUND_STREAM = 1

fn main() {
    var pool byte[64]
    var data byte[8]
    var id uint
    var err int

    data[0] = 8
    data[1] = 3
    data[2] = 0
    data[3] = 0
    data[4] = 0

    sound_init(pool)
    id, err = sound_load(data, SOUND_STREAM)
}
`)

	requireASMOrder(t, asm,
		"peddle_sound_validate_voice:",
		"ldy #1",
		"lda (ZP_PTR0_LO), y",
		"cmp #3",
		"rts",
	)
	requireASMOrder(t, asm,
		"peddle_sound_validate_raw:",
		"jsr peddle_sound_validate_need_3",
		"ldy #1",
		"lda (ZP_PTR0_LO), y",
		"cmp #25",
		"bcs peddle_sound_validate_bad",
	)
	requireASMOrder(t, asm,
		"peddle_sound_validate_loop:",
		"lda peddle_sound_validate_remaining_lo",
		"ora peddle_sound_validate_remaining_hi",
		"bne peddle_sound_validate_has_data",
		"sec",
		"rts",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenDoesNotEmitUnusedSoundRuntime(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x byte
    x = 1
}
`)

	requireNoASM(t, asm,
		"peddle_sound_init:",
		"peddle_sound_irq:",
		"peddle_sound_load_return_id:",
		"jsr peddle_sound_shutdown",
	)
}
