package codegen

import "testing"

func TestCodegenItoxByteAndIntWidths(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b byte
    var i int
    var h2 char[2]
    var h4 char[4]

    b = 27
    i = 27
    copy(h2, itox(b))
    copy(h4, itox(i))
    putstr(0, 0, itox(b))
    putstr(0, 1, itox(i))
}
`)

	requireASM(t, asm,
		"jsr peddle_itox_byte",
		"jsr peddle_itox_int",
		"peddle_itox_byte_buffer:",
		"    .word 2",
		"peddle_itox_int_buffer:",
		"    .word 4",
		"peddle_itox_nibble:",
		"adc #55",
		"adc #48",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenItoxWorksAsAppendTemporarySource(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var rx byte[4]
    var line char[32]

    rx[0] = 27
    append(line, itox(rx[0]))
}
`)

	requireASM(t, asm,
		"jsr peddle_itox_byte",
		"lda #<peddle_itox_byte_buffer",
		"jsr peddle_string_append_literal",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPrintAcceptsTemporaryCharArrays(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var score int
    var b byte

    score = -123
    b = 27
    print(itoa(score))
    print(itox(b))
}
`)

	requireASM(t, asm,
		"jsr peddle_itoa",
		"lda #<peddle_itoa_buffer",
		"jsr peddle_itox_byte",
		"lda #<peddle_itox_byte_buffer",
		"jsr peddle_print_counted_string",
		"peddle_itoa_buffer:",
		"peddle_itox_byte_buffer:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenItoxRuntimeEmittedOnlyWhenUsed(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var line char[8]

    copy(line, "OK")
}
`)

	requireNotContains(t, asm,
		"peddle_itox_byte_buffer:",
		"peddle_itox_int_buffer:",
		"peddle_itox_nibble:",
	)
}
