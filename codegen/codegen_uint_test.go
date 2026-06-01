package codegen

import "testing"

func TestCodegenUintConstantsAndAddressOf(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    hp byte
}

fn main() {
    var max uint
    var border uint
    var addr uint
    var x byte
    var player Player
    var data byte[4]

    max = 65535
    border = 0xd020
    addr = &x
    addr = &player
    addr = &data
}
`)

	requireASM(t, asm,
		"main_max:",
		".fill 2, 0",
		"lda #<65535",
		"lda #>65535",
		"sta main_max+1",
		"lda #<53280",
		"lda #>53280",
		"sta main_border+1",
		"lda #<main_x",
		"lda #>main_x",
		"lda #<main_player",
		"lda #>main_player",
		"lda #<main_data",
		"lda #>main_data",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenUintComparisonUsesUnsignedOrdering(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var hi uint
    var mid uint
    var ok byte

    hi = 65535
    mid = 32768

    if hi > mid {
        ok = 1
    }

    if mid < hi {
        ok = 2
    }
}
`)

	requireASM(t, asm,
		"cmp ZP_TMP1",
		"cmp peddle_tmp_int0+1",
		"bcc",
	)
	requireNoASM(t, asm, "bmi")

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenUintArithmeticStaysWordSized(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a uint
    var b uint
    var c uint

    a = 65535
    b = 2
    c = a + b
}
`)

	requireASM(t, asm,
		"main_a:",
		".fill 2, 0",
		"main_c:",
		".fill 2, 0",
		"adc peddle_tmp_int0",
		"sta main_c",
		"sta main_c+1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
