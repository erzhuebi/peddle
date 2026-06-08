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

func TestCodegenPutRawIsRawAndPutCharUsesCharToScreenTable(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()

    putraw(0, 0, 16)
    putchar(1, 0, 'P')
}
`)

	// putraw() is raw screen-code output.
	requireASM(t, asm,
		"lda #16",
		"sta ZP_TMP0",
	)

	// putchar() is character output and must convert through the shared table.
	requireASM(t, asm,
		"lda #80",
		"sta ZP_TMP0",
		"lda ZP_TMP0",
		"tax",
		"lda peddle_char_to_screen_table, x",
		"sta ZP_TMP0",
		"peddle_char_to_screen_table:",
	)

	// Old branch-based character conversion should no longer be emitted.
	requireNoASM(t, asm,
		"sbc #64",
		"sbc #96",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenScreenBuiltinsPreserveArgsAcrossUserFunctionArgs(t *testing.T) {
	asm := compileSource(t, `
fn xpos() byte {
    return 3
}

fn ypos() byte {
    return 4
}

fn alienChar(row byte) char {
    if row == 0 {
        return 'A'
    }
    return 'B'
}

fn alienCode(row byte) int {
    return 16 + row
}

fn alienColor(row byte) byte {
    return row + 1
}

fn main() {
    var ax byte[2]
    var ay byte[2]
    var i byte
    var row byte
    var title char[8]

    ax[0] = 3
    ay[0] = 4
    i = 0
    row = 0
    copy(title, "ALIEN")

    putchar(ax[i], ay[i], alienChar(row))
    putraw(xpos(), ypos(), alienCode(row))
    putcolor(xpos(), ypos(), alienColor(row))
    gotoxy(xpos(), ypos())
    putstrcolor(xpos(), ypos(), title, alienColor(row))
}
`)

	requireASM(t, asm,
		"jsr alienChar",
		"jsr alienCode",
		"lda alienCode_return",
		"jsr alienColor",
		"jsr xpos",
		"jsr ypos",
		"jsr peddle_putstrcolor",
	)

	requireASMOrder(t, asm,
		"jsr alienChar",
		"sta ZP_TMP0",
		"lda ZP_TMP0",
		"tax",
		"lda peddle_char_to_screen_table, x",
		"sta ZP_TMP0",
		"pla",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
		"sta (ZP_PTR0_LO), y",
	)

	requireASMOrder(t, asm,
		"jsr alienCode",
		"lda alienCode_return",
		"sta ZP_TMP0",
		"pla",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
	)

	requireASMOrder(t, asm,
		"jsr alienColor",
		"sta peddle_putstr_color",
		"pla",
		"sta ZP_PTR1_HI",
		"pla",
		"sta ZP_PTR1_LO",
		"pla",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
		"pla",
		"sta peddle_putstr_y",
		"pla",
		"sta peddle_putstr_x",
		"jsr peddle_putstrcolor",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
