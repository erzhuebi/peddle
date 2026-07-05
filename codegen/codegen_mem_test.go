package codegen

import "testing"

func TestCodegenMemWindowLocalReadWriteAndAddressOf(t *testing.T) {
	asm := compileSource(t, `
fn set(x *byte) {
    x = 99
}

fn main() {
    var screen mem[1000] at $0400
    var b byte
    var addr uint

    screen[0] = 65
    b = screen[0]
    addr = &screen
    addr = &screen[2]
    set(&screen[3])
}
`)

	requireNoASM(t, asm,
		"main_screen:",
		".fill 1000",
		"+4",
	)
	requireASM(t, asm,
		"lda #<1024",
		"lda #>1024",
		"pha",
		"pla",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
		"sta main_b",
		"sta main_addr",
		"sta main_addr+1",
		"jsr set",
		"set_x:",
		".fill 2, 0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMemParameterPassingAndForwarding(t *testing.T) {
	asm := compileSource(t, `
fn put(buf mem[1000], index byte, value byte) {
    buf[index] = value
}

fn forward(buf mem[1000]) {
    put(buf, 2, 7)
}

fn main() {
    var screen mem[1000] at $0400

    put(screen, 1, 32)
    forward(screen)
}
`)

	requireASM(t, asm,
		"put_buf:",
		".fill 2, 0",
		"forward_buf:",
		".fill 2, 0",
		"lda #<1024",
		"sta put_buf",
		"sta put_buf+1",
		"jsr put",
		"lda forward_buf",
		"sta ZP_TMP0",
		"lda forward_buf+1",
		"sta ZP_TMP1",
		"sta put_buf+1",
		"jsr forward",
	)
	requireNoASM(t, asm,
		"main_screen:",
		"+4",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMemStorePreservesValueAcrossDynamicAddressCalculation(t *testing.T) {
	asm := compileSource(t, `
fn put(buf mem[4], index byte, value byte) {
    buf[index] = value
}

fn main() {
    var window mem[4] at $c000

    put(window, 1, 23)
}
`)

	requireASMOrder(t, asm,
		"put:",
		"lda put_value",
		"pha",
		"lda put_index",
		"ldy #0",
		"pla",
		"sta (ZP_PTR0_LO), y",
	)
	requireNoASM(t, asm,
		"lda put_value\n    sta peddle_tmp_int0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMemReadWriteWithComplexExpressions(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var window mem[16] at $c000
    var i byte = 1
    var j byte = 2
    var v byte

    window[i] = 10
    window[j] = 20
    window[i + j] = window[i] + window[j]
    v = window[i + j]

    if v == 30 {
        print("MEM OK")
    }
}
`)

	requireASM(t, asm,
		"lda #<49152",
		"lda #>49152",
		"adc peddle_tmp_int0",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
		"sta main_v",
	)

	requireASMOrder(t, asm,
		"lda (ZP_PTR0_LO), y",
		"pha",
		"lda main_j",
		"lda main_i",
		"sta peddle_tmp_int0",
		"pla",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMemLenAndSizeAreConstants(t *testing.T) {
	asm := compileSource(t, `
fn count(buf mem[1000]) int {
    return len(buf)
}

fn main() {
    var screen mem[1000] at $0400
    var n int

    n = len(screen)
    n = size(screen)
    n = count(screen)
}
`)

	requireASM(t, asm,
		"lda #<1000",
		"lda #>1000",
		"sta main_n",
		"sta main_n+1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
