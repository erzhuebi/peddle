package codegen

import "testing"

func TestCodegenFoldsConstantArithmeticExpression(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int

    x = 2 + 3 * 4
}
`)

	requireASM(t, asm,
		"lda #<14",
		"lda #>14",
		"sta main_x",
	)

	requireNoASM(t, asm,
		"adc peddle_tmp_int0",
		"jsr peddle_mul_int",
	)
}

func TestCodegenFoldsConstantsInExpressions(t *testing.T) {
	asm := compileSource(t, `
const A = 10
const B = %0000_1111
const C = 0xff

fn main() {
    var x int
    var b byte

    x = A * 10 + B
    b = C & B
}
`)

	requireASM(t, asm,
		"lda #<115",
		"lda #>115",
		"lda #15",
		"sta main_b",
	)

	requireNoASM(t, asm,
		"lda #<10",
		"lda #255",
		"and ZP_TMP0",
	)
}

func TestCodegenFoldsByteWidthArithmetic(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b byte

    b = 300 / 2
}
`)

	requireASM(t, asm,
		"lda #22",
		"sta main_b",
	)

	requireNoASM(t, asm,
		"peddle_divmod_byte",
		"cmp ZP_TMP0",
	)
}

func TestCodegenFoldsBitwiseShiftsUnaryAndComparisons(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int
    var b byte
    var ok bool

    x = 1 << 4 | 3
    b = %1111_0000 & 15
    x = -2 + 1
    ok = !0
    ok = 2 + 3 == 5
}
`)

	requireASM(t, asm,
		"lda #<19",
		"lda #>19",
		"lda #0",
		"lda #<-1",
		"lda #>-1",
		"lda #1",
	)

	requireNoASM(t, asm,
		"asl",
		"and ZP_TMP0",
		"cmp peddle_tmp_int0",
	)
}

func TestCodegenFoldsDivisionAndModuloByZero(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int
    var y int

    x = 100 / 0
    y = 100 % 0
}
`)

	requireASM(t, asm,
		"lda #<0",
		"lda #>0",
		"lda #<100",
		"lda #>100",
	)

	requireNoASM(t, asm,
		"peddle_divmod_int",
		"cmp peddle_tmp_int0+1",
	)
}

func TestCodegenPartiallyFoldsConstantSubExpressions(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a int
    var x int

    a = 1
    x = a + 2 * 3
}
`)

	requireASM(t, asm,
		"lda #<6",
		"lda #>6",
		"adc peddle_tmp_int0",
	)

	requireNoASM(t, asm,
		"jsr peddle_mul_int",
		"dec peddle_tmp_int0",
	)
}

func TestCodegenFoldsShiftCountsForVariableShifts(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a byte
    var b byte

    b = a << 1 + 2
}
`)

	requireASM(t, asm,
		"asl",
	)

	requireNoASM(t, asm,
		"ldx ZP_TMP0",
		"dec ZP_TMP0",
	)
}

func TestCodegenFoldsClampedShiftCounts(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b byte
    var x int

    b = 1 << -1
    b = 1 << 20
    x = 1 << 20
}
`)

	requireASM(t, asm,
		"lda #1",
		"lda #0",
		"lda #<0",
		"lda #>0",
	)

	requireNoASM(t, asm,
		"asl",
		"rol ZP_TMP1",
	)
}

func TestCodegenFoldsPeekPokeAddresses(t *testing.T) {
	asm := compileSource(t, `
const BORDER = $d020

fn main() {
    var b byte

    b = peek(BORDER + 1)
    poke(BORDER + 1, b)
}
`)

	requireASM(t, asm,
		"lda $d021",
		"sta $d021",
	)

	requireNoASM(t, asm,
		"(ZP_PTR0_LO), y",
	)
}
