package codegen

import (
	"strings"
	"testing"
)

func requireASMOrder(t *testing.T, asm string, parts ...string) {
	t.Helper()

	pos := 0

	for _, part := range parts {
		idx := strings.Index(asm[pos:], part)
		if idx < 0 {
			t.Fatalf("ASM does not contain %q after offset %d\n\nASM:\n%s", part, pos, asm)
		}

		pos += idx + len(part)
	}
}

func TestCodegenGotoXYEmitsKernelPlotCall(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    gotoxy(10, 8)
}
`)

	requireASM(t, asm,
		"lda #10",
		"pha",
		"lda #8",
		"sta peddle_tmp_int0",
		"sta peddle_tmp_int0+1",
		"cmp #40",
		"cmp #25",
		"clc",
		"ldx peddle_tmp_int0+1",
		"ldy peddle_tmp_int0",
		"jsr $fff0",
	)

	requireASMOrder(t, asm,
		"lda #10",
		"pha",
		"lda #8",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
		"clc",
		"ldx peddle_tmp_int0+1",
		"ldy peddle_tmp_int0",
		"jsr $fff0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenGotoXYUsesRuntimeInSizeMode(t *testing.T) {
	asm := compileSourceWithOptions(t, `
fn main() {
    gotoxy(10, 8)
}
`, Options{OptMode: OptModeSize})

	requireASM(t, asm,
		"jsr peddle_gotoxy",
		"peddle_gotoxy:",
		"jsr $fff0",
	)

	requireASMOrder(t, asm,
		"lda #10",
		"pha",
		"lda #8",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
		"jsr peddle_gotoxy",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenGotoXYBeforePrint(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()

    gotoxy(0, 5)
    print("HELLO")

    gotoxy(0, 24)
}
`)

	requireASM(t, asm,
		"jsr $fff0",
		"jsr peddle_print_counted_string",
		"literal_0:",
	)

	requireASMOrder(t, asm,
		"lda #0",
		"pha",
		"lda #5",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta peddle_tmp_int0",
		"clc",
		"ldx peddle_tmp_int0+1",
		"ldy peddle_tmp_int0",
		"jsr $fff0",
		"jsr peddle_print_counted_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenGotoXYEmitsClippingChecks(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    gotoxy(40, 25)
}
`)

	requireASM(t, asm,
		"lda #40",
		"pha",
		"lda #25",
		"sta peddle_tmp_int0",
		"sta peddle_tmp_int0+1",
		"lda peddle_tmp_int0",
		"cmp #40",
		"bcs",
		"lda peddle_tmp_int0+1",
		"cmp #25",
		"bcs",
		"jsr $fff0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
