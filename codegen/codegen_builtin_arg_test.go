package codegen

import "testing"

func TestCodegenLowLevelBuiltinsPreserveArgsAcrossUserFunctionArgs(t *testing.T) {
	asm := compileSource(t, `
fn addr() uint {
    var a uint
    a = 1024
    return a
}

fn value() byte {
    return 90
}

fn interval() int {
    return 5
}

fn main() {
    var last int

    last = ticks()
    poke(addr(), value())
    if tickdue(last, interval()) {
        border(value())
    }
}
`)

	requireASMOrder(t, asm,
		"jsr addr",
		"lda addr_return",
		"sta ZP_TMP0",
		"lda addr_return+1",
		"sta ZP_TMP1",
		"lda ZP_TMP0",
		"pha",
		"lda ZP_TMP1",
		"pha",
		"jsr value",
		"sta ZP_TMP0",
		"pla",
		"sta ZP_PTR0_HI",
		"pla",
		"sta ZP_PTR0_LO",
		"lda ZP_TMP0",
		"sta (ZP_PTR0_LO), y",
	)

	requireASMOrder(t, asm,
		"lda main_last",
		"sta ZP_TMP0",
		"lda main_last+1",
		"sta ZP_TMP1",
		"lda ZP_TMP0",
		"pha",
		"lda ZP_TMP1",
		"pha",
		"jsr interval",
		"sta peddle_tmp_int0",
		"lda ZP_TMP1",
		"sta peddle_tmp_int0+1",
		"pla",
		"sta ZP_PTR0_HI",
		"pla",
		"sta ZP_PTR0_LO",
		"lda $00a2",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
