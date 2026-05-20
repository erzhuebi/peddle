package codegen

import "testing"

func TestCodegenPutStrUsesRuntimeHelperForStringLiteral(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(0, 0, "PEDDLE")
}
`)

	requireASM(t, asm,
		"sta peddle_putstr_x",
		"sta peddle_putstr_y",
		"lda #<literal_0",
		"sta ZP_PTR1_LO",
		"lda #>literal_0",
		"sta ZP_PTR1_HI",
		"lda #<6",
		"sta peddle_tmp_int0",
		"lda #>6",
		"sta peddle_tmp_int0+1",
		"jsr peddle_putstr",
		"peddle_putstr:",
		"peddle_putstr_common:",
		"peddle_putstr_loop:",
		"peddle_putstr_done:",
		"literal_0:",
	)

	requireNoASM(t, asm,
		"jsr peddle_print_counted_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrColorUsesRuntimeHelperForStringLiteral(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstrcolor(0, 1, "OK", 2)
}
`)

	requireASM(t, asm,
		"sta peddle_putstr_x",
		"sta peddle_putstr_y",
		"lda #<literal_0",
		"sta ZP_PTR1_LO",
		"lda #>literal_0",
		"sta ZP_PTR1_HI",
		"lda #<2",
		"sta peddle_tmp_int0",
		"lda #>2",
		"sta peddle_tmp_int0+1",
		"lda #2",
		"sta peddle_putstr_color",
		"jsr peddle_putstrcolor",
		"peddle_putstrcolor:",
		"lda #1",
		"sta peddle_putstr_write_color",
		"lda #<$d800",
		"sta ZP_PTR0_LO",
		"lda #>$d800",
		"sta ZP_PTR0_HI",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrRuntimeHandlesNewline(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(5, 3, "A\nB")
}
`)

	requireASM(t, asm,
		"lda #5",
		"sta peddle_putstr_x",
		"lda #3",
		"sta peddle_putstr_y",
		"lda #<3",
		"sta peddle_tmp_int0",
		"jsr peddle_putstr",
		"cmp #13",
		"bne peddle_putstr_not_newline",
		"jmp peddle_putstr_newline",
		"peddle_putstr_not_newline:",
		"peddle_putstr_newline:",
		"lda peddle_putstr_start_x",
		"sta ZP_TMP0",
		"inc ZP_TMP1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrRuntimeClipsAtStartCoordinates(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(40, 25, "NO")
}
`)

	requireASM(t, asm,
		"lda #40",
		"sta peddle_putstr_x",
		"lda #25",
		"sta peddle_putstr_y",
		"jsr peddle_putstr",

		"peddle_putstr_common:",
		"lda peddle_putstr_x",
		"sta peddle_putstr_start_x",
		"sta ZP_TMP0",
		"lda peddle_putstr_y",
		"sta ZP_TMP1",

		"lda ZP_TMP0",
		"cmp #40",
		"bcc peddle_putstr_start_x_ok",
		"jmp peddle_putstr_done",
		"peddle_putstr_start_x_ok:",

		"lda ZP_TMP1",
		"cmp #25",
		"bcc peddle_putstr_start_y_ok",
		"jmp peddle_putstr_done",
		"peddle_putstr_start_y_ok:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrRuntimeClipsAtBottomOfScreen(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(38, 24, "ABCD")
}
`)

	requireASM(t, asm,
		"lda #38",
		"sta peddle_putstr_x",
		"lda #24",
		"sta peddle_putstr_y",
		"jsr peddle_putstr",

		"peddle_putstr_advance:",
		"inc ZP_TMP0",
		"lda ZP_TMP0",
		"cmp #40",
		"bcs peddle_putstr_wrap_line",
		"jmp peddle_putstr_loop",

		"peddle_putstr_wrap_line:",
		"lda #0",
		"sta ZP_TMP0",
		"inc ZP_TMP1",
		"lda ZP_TMP1",
		"cmp #25",
		"bcc peddle_putstr_continue_after_wrap",
		"jmp peddle_putstr_done",

		"peddle_putstr_continue_after_wrap:",
		"jmp peddle_putstr_loop",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrColorNewlineDoesNotWriteColorForNewline(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstrcolor(0, 0, "A\nB", 5)
}
`)

	requireASM(t, asm,
		"lda #5",
		"sta peddle_putstr_color",
		"jsr peddle_putstrcolor",
		"cmp #13",
		"bne peddle_putstr_not_newline",
		"jmp peddle_putstr_newline",
		"peddle_putstr_not_newline:",
		"peddle_putstr_newline:",
		"lda peddle_putstr_start_x",
		"sta ZP_TMP0",
		"inc ZP_TMP1",
	)

	requireASM(t, asm,
		"lda peddle_putstr_write_color",
		"beq peddle_putstr_advance",
		"lda peddle_putstr_color",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrRuntimeConvertsUppercaseAndLowercase(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(0, 0, "abc")
}
`)

	requireASM(t, asm,
		"peddle_putstr_check_lower:",
		"cmp #97",
		"bcc peddle_putstr_converted",
		"cmp #123",
		"bcs peddle_putstr_converted",
		"sec",
		"sbc #96",
		"peddle_putstr_converted:",
	)

	requireASM(t, asm,
		"cmp #65",
		"bcc peddle_putstr_check_lower",
		"cmp #91",
		"bcs peddle_putstr_check_lower",
		"sbc #64",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrUsesRuntimeHelperForCharArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var title char[16]

    copy(title, "HELLO")
    putstr(0, 0, title)
}
`)

	requireASM(t, asm,
		"main_title:",
		".word 16",
		".word 0",
		"ldy #2",
		"lda (ZP_PTR0_LO), y",
		"sta peddle_tmp_int0",
		"iny",
		"lda (ZP_PTR0_LO), y",
		"sta peddle_tmp_int0+1",
		"adc #4",
		"sta ZP_PTR1_LO",
		"sta ZP_PTR1_HI",
		"jsr peddle_putstr",
		"peddle_putstr:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrColorUsesRuntimeHelperForCharArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var title char[16]

    copy(title, "HELLO")
    putstrcolor(0, 0, title, 2)
}
`)

	requireASM(t, asm,
		"main_title:",
		".word 16",
		".word 0",
		"ldy #2",
		"lda (ZP_PTR0_LO), y",
		"sta peddle_tmp_int0",
		"iny",
		"lda (ZP_PTR0_LO), y",
		"sta peddle_tmp_int0+1",
		"adc #4",
		"sta ZP_PTR1_LO",
		"sta ZP_PTR1_HI",
		"lda #2",
		"sta peddle_putstr_color",
		"jsr peddle_putstrcolor",
		"peddle_putstrcolor:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenPutStrRuntimeEmittedOnlyOnce(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    putstr(0, 0, "ONE")
    putstr(0, 1, "TWO")
    putstrcolor(0, 2, "THREE", 2)
}
`)

	if countOccurrences(asm, "peddle_putstr:") != 1 {
		t.Fatalf("expected peddle_putstr runtime label once\n\nASM:\n%s", asm)
	}

	if countOccurrences(asm, "peddle_putstr_common:") != 1 {
		t.Fatalf("expected peddle_putstr_common runtime label once\n\nASM:\n%s", asm)
	}

	if countOccurrences(asm, "peddle_putstrcolor:") != 1 {
		t.Fatalf("expected peddle_putstrcolor runtime label once\n\nASM:\n%s", asm)
	}

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func countOccurrences(s string, sub string) int {
	count := 0
	start := 0

	for {
		idx := indexFrom(s, sub, start)
		if idx < 0 {
			return count
		}

		count++
		start = idx + len(sub)
	}
}

func indexFrom(s string, sub string, start int) int {
	if start >= len(s) {
		return -1
	}

	for i := start; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}

	return -1
}
