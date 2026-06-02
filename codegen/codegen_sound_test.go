package codegen

import "testing"

func TestCodegenSoundRuntimeAndBuiltinReturns(t *testing.T) {
	asm := compileSource(t, `
const SOUND_REGSTREAM = 1

fn main() {
    var pool byte[64]
    var data byte[8]
    var id uint
    var err int
    var n int

    data[0] = 2
    data[1] = 24
    data[2] = 15
    data[3] = 0

    sound_init(pool)
    id, err = sound_load(data, SOUND_REGSTREAM)
    sound_play(id)
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
		"lda peddle_sound_load_return_id",
		"sta main_id",
		"lda peddle_sound_load_return_err",
		"sta main_err",
		"peddle_sound_slot_inuse:",
		".fill 16, 0",
		"lda $0314",
		"sta peddle_sound_old_irq_lo",
		"sta $0314",
		"jmp (peddle_sound_old_irq_lo)",
		"sta $d400,x",
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
