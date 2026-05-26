package codegen

import "testing"

func TestCodegenMixedByteIntExpressions(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b byte
    var i int
    var x int

    b = 10
    i = 1000

    x = b + i
    x = i + b
    x = i - b
}
`)

	requireASM(t, asm,
		"lda main_b",
		"lda main_i",
		"adc peddle_tmp_int0",
		"sbc peddle_tmp_int0",
	)
}

func TestCodegenNestedExpressionRightSide(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a int
    var b int
    var c int
    var d int
    var x int

    a = 10
    b = 2
    c = 4
    d = 1

    x = (a + b) - (c + d)
}
`)

	requireASM(t, asm,
		"adc peddle_tmp_int0",
		"sbc peddle_tmp_int0",
		"sta main_x",
	)
}

func TestCodegenNestedExpressionChain(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a int
    var b int
    var c int
    var x int

    a = 1
    b = 2
    c = 3

    x = a + b + c
}
`)

	requireASM(t, asm,
		"adc peddle_tmp_int0",
		"sta main_x",
	)
}

func TestCodegenBooleanConditionDirect(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var done bool

    done = 1

    if done {
        print("OK")
    }
}
`)

	requireASM(t, asm,
		"lda main_done",
		"cmp #0",
		"jsr peddle_print_counted_string",
	)
}

func TestCodegenIntConditionDirect(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int

    x = 1

    if x {
        print("OK")
    }
}
`)

	requireASM(t, asm,
		"lda main_x",
		"sta ZP_TMP0",
		"jsr peddle_print_counted_string",
	)
}

func TestCodegenArrayIndexExpression(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a byte[8]
    var i byte
    var x byte

    i = 2
    a[i + 1] = 7
    x = a[i + 1]
}
`)

	requireASM(t, asm,
		"adc ZP_TMP0",
		"ldy #0",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
	)
}

func TestCodegenCharArrayIndexComparisonZeroExtendsElement(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var rx char[16]
    var i int
    var ok bool

    rx[0] = 13
    i = 0

    if rx[i] == 13 {
        ok = true
    } else {
        ok = false
    }
}
`)

	requireASMOrder(t, asm,
		"    lda (ZP_PTR0_LO), y",
		"    sta ZP_TMP0",
		"    lda #0",
		"    sta ZP_TMP1",
		"    pla",
		"    sta peddle_tmp_int0+1",
		"    pla",
		"    sta peddle_tmp_int0",
		"    lda ZP_TMP1",
		"    cmp peddle_tmp_int0+1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMultipleReturns(t *testing.T) {
	asm := compileSource(t, `
fn choose(x int) int {
    if x {
        return 1
    }

    return 2
}

fn main() {
    var y int
    y = choose(1)
}
`)

	requireASM(t, asm,
		"choose_return:",
		"sta choose_return",
		"sta choose_return+1",
		"jsr choose",
	)
}
