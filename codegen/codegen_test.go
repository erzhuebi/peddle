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

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if err := sema.New().Check(prog); err != nil {
		t.Fatalf("sema error: %v", err)
	}

	asm, err := New().Generate(prog)
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
		"jsr peddle_print_string",
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
		"jsr peddle_print_string",
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
		"jsr peddle_print_string",
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
