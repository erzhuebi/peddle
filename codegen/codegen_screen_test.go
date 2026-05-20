package codegen

import "testing"

func TestCodegenClsClearsScreenAndResetsKernelCursor(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()
    print("HELLO")
}
`)

	requireASM(t, asm,
		"sta $0400, x",
		"sta $0500, x",
		"sta $0600, x",
		"sta $0700, x",
		"clc",
		"ldx #0",
		"ldy #0",
		"jsr $fff0",
		"jsr peddle_print_counted_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPrintNewlineUsesC64CarriageReturn(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    print("LINE 1\nLINE 2")
}
`)

	requireASM(t, asm,
		"literal_0:",
		".byte 76,73,78,69,32,49,13,76,73,78,69,32,50",
		"lda #<13",
		"sta peddle_tmp_int0",
		"lda #>13",
		"sta peddle_tmp_int0+1",
		"jsr peddle_print_counted_string",
	)

	requireNoASM(t, asm,
		".byte 76,73,78,69,32,49,10,76,73,78,69,32,50",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutCharUsesDirectCoordinatesAfterCls(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()

    print("LINE 1\n")
    print("LINE 2\n")

    putchar(0, 5, 'P')
    putchar(1, 5, 'E')
    putcolor(0, 5, 2)
    putcolor(1, 5, 3)
}
`)

	requireASM(t, asm,
		"jsr $fff0",
		"jsr peddle_print_counted_string",
		"lda #5",
		"sta peddle_tmp_int0+1",
		"adc #40",
		"sta (ZP_PTR0_LO), y",
		"lda #$20",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutScreenIsRawAndPutCharConvertsLetters(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()

    putscreen(0, 0, 16)
    putchar(1, 0, 'P')
}
`)

	requireASM(t, asm,
		"lda #16",
		"sta ZP_TMP0",
		"lda #80",
		"sta ZP_TMP0",
		"sbc #64",
		"sta ZP_TMP0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
