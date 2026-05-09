package codegen

import (
	"strings"
	"testing"

	"peddle/lexer"
	"peddle/parser"
	"peddle/sema"
)

func compileSource(t *testing.T, src string) string {
	t.Helper()

	return compileSourceWithOptions(t, src, Options{
		OptMode: OptModeSpeed,
	})
}

func compileSourceWithOptions(t *testing.T, src string, options Options) string {
	t.Helper()

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if err := sema.New().Check(prog); err != nil {
		t.Fatalf("sema error: %v", err)
	}

	asm, err := NewWithOptions(options).Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return asm
}

func requireASM(t *testing.T, asm string, parts ...string) {
	t.Helper()

	for _, part := range parts {
		if !strings.Contains(asm, part) {
			t.Fatalf("ASM does not contain %q\n\nASM:\n%s", part, asm)
		}
	}
}

func TestCodegenHelloWorld(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    print("HELLO WORLD")
}
`)

	requireASM(t, asm,
		"jsr main",
		"jsr peddle_print_counted_string",
		"literal_0:",
	)
}

func TestCodegenPeekPoke(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b: byte
    b = peek(53280)
    poke(53281, b)
}
`)

	requireASM(t, asm,
		"lda $d020",
		"sta $d021",
	)
}

func TestCodegenIntArithmetic(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int
    var b: int
    var c: int

    a = 1000
    b = 2000
    c = a + b
    c = c - a
}
`)

	requireASM(t, asm,
		"adc peddle_tmp_int0",
		"sbc peddle_tmp_int0",
		"peddle_tmp_int0:",
	)
}

func TestCodegenSignedIntComparison(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int
    var b: int

    a = -1
    b = 1

    if a < b {
        print("OK")
    }
}
`)

	requireASM(t, asm,
		"bmi",
		"cmp peddle_tmp_int0+1",
		"jsr peddle_print_counted_string",
	)
}

func TestCodegenIntArrayReadWrite(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int[4]
    var x: int

    a[0] = 1234
    x = a[0]
}
`)

	requireASM(t, asm,
		"asl",
		"sta (ZP_PTR0_LO), y",
		"iny",
		"lda (ZP_PTR0_LO), y",
	)
}

func TestCodegenStringAssignment(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var s: char[6]

    s = "HELLO"
    print(s)
}
`)

	requireASM(t, asm,
		"literal_0:",
		"ldy #2",
		"ldy #3",
		"sta (ZP_PTR0_LO), y",
		"beq",
		"jsr peddle_print_counted_string",
	)
}

func TestCodegenUnaryOperators(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x: int
    var b: bool

    x = -1
    b = !b
}
`)

	requireASM(t, asm,
		"eor #$ff",
		"adc #1",
		"cmp #0",
		"lda #1",
	)
}

func TestCodegenUserFunctionCallAndReturn(t *testing.T) {
	asm := compileSource(t, `
fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    var x: int

    x = add(1, 2)
}
`)

	requireASM(t, asm,
		"sta add_a",
		"sta add_b",
		"jsr add",
		"lda add_return",
		"sta main_x",
	)
}

func TestCodegenIfElse(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int
    var b: int

    a = 1
    b = 2

    if a < b {
        print("YES")
    } else {
        print("NO")
    }
}
`)

	requireASM(t, asm,
		"jsr peddle_print_counted_string",
		"literal_0:",
		"literal_1:",
	)
}

func TestCodegenWhileLoop(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var i: byte

    i = 0

    while i < 3 {
        i = i + 1
    }
}
`)

	requireASM(t, asm,
		"cmp peddle_tmp_int0",
		"jmp L",
		"adc ZP_TMP0",
	)
}

func TestCodegenByteIntConversions(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var b: byte
    var i: int

    b = 10
    i = b
    b = i
}
`)

	requireASM(t, asm,
		"sta main_b",
		"sta main_i",
		"sta main_i+1",
	)
}

func TestCodegenArrayIndexVariable(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[4]
    var i: byte
    var x: byte

    i = 2
    a[i] = 7
    x = a[i]
}
`)

	requireASM(t, asm,
		"lda main_i",
		"ldy #0",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
	)
}

func TestCodegenNestedExpressionSimple(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int
    var b: int
    var c: int

    a = 1
    b = 2
    c = (a + b) - 1
}
`)

	requireASM(t, asm,
		"adc peddle_tmp_int0",
		"sbc peddle_tmp_int0",
	)
}

func TestCodegenArraySizeReturnsCapacity(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int[10]
    var n: int

    n = size(a)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		"ldy #0",
		"lda (ZP_PTR0_LO), y",
		"sta ZP_TMP0",
		"sta main_n",
		"sta main_n+1",
	)
}

func TestCodegenArrayLenReadsRuntimeLength(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]
    var n: int

    n = len(a)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"ldy #2",
		"lda (ZP_PTR0_LO), y",
		"sta ZP_TMP0",
		"sta main_n",
		"sta main_n+1",
	)
}

func TestCodegenArrayIndexWriteUpdatesLength(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]
    var n: int

    a[5] = 1
    n = len(a)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"inc ZP_TMP0",
		"sta main_a+2",
		"sta main_a+3",
		"ldy #2",
		"sta main_n",
		"sta main_n+1",
	)
}

func TestCodegenAppendByteArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]

    append(a, 1)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"ldy #2",
		"sta peddle_tmp_int0",
		"sta (ZP_PTR0_LO), y",
		"adc #1",
	)
}

func TestCodegenAppendIntArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int[10]

    append(a, 1000)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"lda #<1000",
		"sta peddle_tmp_int0",
		"lda #>1000",
		"sta peddle_tmp_int0+1",
		"sta (ZP_PTR0_LO), y",
		"iny",
	)
}

func TestCodegenCopyByteArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x: byte[10]
    var y: byte[10]

    append(x, 1)
    append(x, 2)
    copy(y, x)
}
`)

	requireASM(t, asm,
		"main_x:",
		".word 10",
		"main_y:",
		".word 10",
		"sta ZP_PTR1_LO",
		"sta ZP_PTR1_HI",
		"sta (ZP_PTR1_LO), y",
		"inc ZP_PTR0_LO",
		"inc ZP_PTR1_LO",
	)
}

func TestCodegenFillByteArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]

    fill(a, 1)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"ldy #0",
		"ldy #2",
		"sta (ZP_PTR0_LO), y",
		"sta peddle_tmp_int0",
		"lda ZP_TMP0",
	)
}

func TestCodegenFillIntArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: int[10]

    fill(a, 1)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"lda #<1",
		"sta ZP_TMP0",
		"lda #>1",
		"sta ZP_TMP1",
		"adc #2",
	)
}

func TestCodegenCopyStringLiteral(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var s: char[10]

    copy(s, "ABC")
    print(s)
}
`)

	requireASM(t, asm,
		"literal_0:",
		".byte 65,66,67",
		"ldy #2",
		"ldy #3",
		"sta (ZP_PTR0_LO), y",
		"jsr peddle_print_counted_string",
	)
}

func TestCodegenClearArray(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]

    append(a, 1)
    clear(a)
}
`)

	requireASM(t, asm,
		"main_a:",
		".word 10",
		".word 0",
		"ldy #2",
		"lda #0",
		"sta (ZP_PTR0_LO), y",
		"iny",
		"sta (ZP_PTR0_LO), y",
	)
}

func TestCodegenAppendStringLiteral(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var s: char[10]

    copy(s, "ABC")
    append(s, "D")
}
`)

	requireASM(t, asm,
		"literal_0:",
		"literal_1:",
		"adc #<1",
		"adc #>1",
		"lda literal_1, y",
		"sta (ZP_PTR0_LO), y",
	)
}

func TestCodegenDoesNotEmitUnusedPrintRuntime(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x: byte

    x = 1
}
`)

	requireNoASM(t, asm,
		"peddle_print_counted_string:",
		"peddle_print_string:",
		"jsr peddle_print_counted_string",
		"jsr peddle_print_string",
	)
}

func TestCodegenNonPrintBuiltinsDoNotEmitPrintRuntime(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]
    var n: int
    var b: byte

    append(a, 1)
    clear(a)
    fill(a, 2)
    n = len(a)
    n = size(a)
    b = peek(53280)
    poke(53281, b)
}
`)

	requireNoASM(t, asm,
		"peddle_print_counted_string:",
		"peddle_print_string:",
		"jsr peddle_print_counted_string",
		"jsr peddle_print_string",
	)
}

func TestCodegenStringLiteralsAreNotZeroTerminated(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    print("ABC")
}
`)

	requireASM(t, asm,
		"literal_0:",
		".byte 65,66,67",
		"jsr peddle_print_counted_string",
	)

	requireNoASM(t, asm,
		".byte 65,66,67,0",
		"peddle_print_string:",
		"jsr peddle_print_string",
	)
}

func TestCodegenOptSpeedDoesNotEmitSizeRuntimeHelpers(t *testing.T) {
	asm := compileSourceWithOptions(t, `
fn main() {
    var a: byte[10]
    var b: byte[10]
    var s: char[10]

    append(a, 1)
    fill(a, 2)
    copy(b, a)
    copy(s, "ABC")
    append(s, "D")
}
`, Options{OptMode: OptModeSpeed})

	requireNoASM(t, asm,
		"peddle_array_copy:",
		"peddle_fill_byte:",
		"peddle_fill_int:",
		"peddle_append_byte:",
		"peddle_append_int:",
		"peddle_string_copy_literal:",
		"peddle_string_append_literal:",
		"jsr peddle_array_copy",
		"jsr peddle_fill_byte",
		"jsr peddle_append_byte",
		"jsr peddle_string_copy_literal",
		"jsr peddle_string_append_literal",
	)
}

func TestCodegenOptSizeEmitsRuntimeHelpersForLargeBuiltins(t *testing.T) {
	asm := compileSourceWithOptions(t, `
fn main() {
    var a: byte[10]
    var b: byte[10]
    var s: char[10]

    append(a, 1)
    fill(a, 2)
    copy(b, a)
    copy(s, "ABC")
    append(s, "D")
}
`, Options{OptMode: OptModeSize})

	requireASM(t, asm,
		"jsr peddle_append_byte",
		"jsr peddle_fill_byte",
		"jsr peddle_array_copy",
		"jsr peddle_string_copy_literal",
		"jsr peddle_string_append_literal",
		"peddle_append_byte:",
		"peddle_fill_byte:",
		"peddle_array_copy:",
		"peddle_string_copy_literal:",
		"peddle_string_append_literal:",
	)
}

func TestCodegenStage1OperatorsSpeedMode(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    a = 3 * 4
    b = a & 15
    b = b | 64
    b = b ^ 255

    x = 300 * 4
    y = x & 1023
    y = y | 4096
    y = y ^ 65535
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm, "and ZP_TMP0")
	requireASM(t, asm, "ora ZP_TMP0")
	requireASM(t, asm, "eor ZP_TMP0")

	requireASM(t, asm, "and peddle_tmp_int0")
	requireASM(t, asm, "ora peddle_tmp_int0")
	requireASM(t, asm, "eor peddle_tmp_int0")

	requireNoASM(t, asm, "jsr peddle_mul_byte")
	requireNoASM(t, asm, "jsr peddle_mul_int")

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage1OperatorsSizeModeUsesMulRuntime(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var x: int

    a = 3 * 4
    x = 300 * 4
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm, "jsr peddle_mul_byte")
	requireASM(t, asm, "peddle_mul_byte:")

	requireASM(t, asm, "jsr peddle_mul_int")
	requireASM(t, asm, "peddle_mul_int:")

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage1BitwiseOperatorsRemainInlineInSizeMode(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    b = a & 15
    b = b | 64
    b = b ^ 255

    y = x & 1023
    y = y | 4096
    y = y ^ 65535
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm, "and ZP_TMP0")
	requireASM(t, asm, "ora ZP_TMP0")
	requireASM(t, asm, "eor ZP_TMP0")

	requireASM(t, asm, "and peddle_tmp_int0")
	requireASM(t, asm, "ora peddle_tmp_int0")
	requireASM(t, asm, "eor peddle_tmp_int0")

	requireNoASM(t, asm, "peddle_and")
	requireNoASM(t, asm, "peddle_or")
	requireNoASM(t, asm, "peddle_xor")

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage2ConstantShiftsAreInlineInSpeedMode(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    a = 1 << 3
    b = 128 >> 2

    x = 1 << 12
    y = 4096 >> 4
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm,
		"asl",
		"lsr",
		"asl ZP_TMP0",
		"rol ZP_TMP1",
		"lsr ZP_TMP1",
		"ror ZP_TMP0",
	)

	requireNoASM(t, asm,
		"jsr peddle_shl_byte",
		"jsr peddle_shr_byte",
		"jsr peddle_shl_int",
		"jsr peddle_shr_int",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage2ConstantShiftsAreInlineInSizeMode(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    a = 1 << 3
    b = 128 >> 2

    x = 1 << 12
    y = 4096 >> 4
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm,
		"asl",
		"lsr",
		"asl ZP_TMP0",
		"rol ZP_TMP1",
		"lsr ZP_TMP1",
		"ror ZP_TMP0",
	)

	requireNoASM(t, asm,
		"jsr peddle_shl_byte",
		"jsr peddle_shr_byte",
		"jsr peddle_shl_int",
		"jsr peddle_shr_int",
		"peddle_shl_byte:",
		"peddle_shr_byte:",
		"peddle_shl_int:",
		"peddle_shr_int:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage2VariableShiftsInlineInSpeedMode(t *testing.T) {
	input := `
fn main() {
    var a, b, s: byte
    var x, y, n: int

    a = 8
    s = 1
    b = a << s
    b = b >> s

    x = 1024
    n = 2
    y = x << n
    y = y >> n
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm,
		"dex",
		"dec peddle_tmp_int0",
		"asl ZP_TMP1",
		"lsr ZP_TMP1",
		"asl ZP_TMP0",
		"rol ZP_TMP1",
		"ror ZP_TMP0",
	)

	requireNoASM(t, asm,
		"jsr peddle_shl_byte",
		"jsr peddle_shr_byte",
		"jsr peddle_shl_int",
		"jsr peddle_shr_int",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage2VariableShiftsUseRuntimeInSizeMode(t *testing.T) {
	input := `
fn main() {
    var a, b, s: byte
    var x, y, n: int

    a = 8
    s = 1
    b = a << s
    b = b >> s

    x = 1024
    n = 2
    y = x << n
    y = y >> n
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm,
		"jsr peddle_shl_byte",
		"jsr peddle_shr_byte",
		"jsr peddle_shl_int",
		"jsr peddle_shr_int",
		"peddle_shl_byte:",
		"peddle_shr_byte:",
		"peddle_shl_int:",
		"peddle_shr_int:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage3DivisionAndModuloSpeedMode(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var b: byte
    var x: int
    var y: int

    a = 100 / 5
    b = 100 % 7

    x = 1000 / 10
    y = 1000 % 33
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm,
		"cmp ZP_TMP0",
		"sbc ZP_TMP0",
		"sta ZP_TMP1",
		"cmp peddle_tmp_int0+1",
		"sbc peddle_tmp_int0",
		"sbc peddle_tmp_int0+1",
	)

	requireNoASM(t, asm,
		"jsr peddle_divmod_byte",
		"jsr peddle_divmod_int",
		"peddle_divmod_byte:",
		"peddle_divmod_int:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage3DivisionAndModuloSizeMode(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var b: byte
    var x: int
    var y: int

    a = 100 / 5
    b = 100 % 7

    x = 1000 / 10
    y = 1000 % 33
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm,
		"jsr peddle_divmod_byte",
		"jsr peddle_divmod_int",
		"peddle_divmod_byte:",
		"peddle_divmod_int:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage3ModuloUsesRemainderRegister(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var x: int

    a = 13 % 5
    x = 1000 % 256
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm,
		"sta ZP_TMP1",
		"sta ZP_PTR0_LO",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage3DivisionByZeroSpeedModeIsInline(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var x: int

    a = 10 / 0
    x = 100 / 0
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSpeed,
	})

	requireASM(t, asm,
		"lda #0",
		"beq",
		"jmp L",
		"_return:",
	)

	requireNoASM(t, asm,
		"peddle_divmod_byte_div_zero:",
		"peddle_divmod_int_div_zero:",
		"jsr peddle_divmod_byte",
		"jsr peddle_divmod_int",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStage3DivisionByZeroSizeModeRuntimeExists(t *testing.T) {
	input := `
fn main() {
    var a: byte
    var x: int

    a = 10 / 0
    x = 100 / 0
}
`

	asm := compileSourceWithOptions(t, input, Options{
		OptMode: OptModeSize,
	})

	requireASM(t, asm,
		"jsr peddle_divmod_byte",
		"jsr peddle_divmod_int",
		"peddle_divmod_byte_div_zero:",
		"peddle_divmod_int_div_zero:",
		"lda #0",
		"rts",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
