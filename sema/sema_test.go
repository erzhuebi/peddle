package sema

import (
	"testing"

	"peddle/lexer"
	"peddle/parser"
)

func checkSource(t *testing.T, src string) error {
	t.Helper()

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return New().Check(prog)
}

func TestSemaAllowsCoreFeatures(t *testing.T) {
	src := `
fn main() {
    var b: byte
    var i: int
    var done: bool
    var a: int[4]
    var s: char[6]

    b = 42
    i = b
    i = -i
    done = !done
    a[0] = i
    i = a[0]
    s = "HELLO"

    if i >= 0 {
        poke(53280, b)
    }

    b = peek(53280)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsUnknownVariable(t *testing.T) {
	src := `
fn main() {
    x = 1
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsStringToByte(t *testing.T) {
	src := `
fn main() {
    var b: byte
    b = "HELLO"
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaUserFunctionCall(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    var x: int
    x = add(1, 2)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsWrongArgCount(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    var x: int
    x = add(1)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsWrongArgType(t *testing.T) {
	src := `
fn takesByteArray(a: byte[4]) {
    return
}

fn main() {
    var x: int
    takesByteArray(x)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsNonArrayIndex(t *testing.T) {
	src := `
fn main() {
    var x: int
    x[0] = 1
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsInvalidReturnValue(t *testing.T) {
	src := `
fn f() -> int {
    return "NO"
}

fn main() {
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaStage1Operators(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    a = 3 * 4
    a = a & 15
    a = a | 64
    a = a ^ 255

    x = 300 * 4
    y = x & 1023
    y = x | 4096
    y = x ^ 65535
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	checker := New()
	if err := checker.Check(prog); err != nil {
		t.Fatalf("sema check failed: %v", err)
	}
}

func TestSemaStage2ShiftOperators(t *testing.T) {
	input := `
fn main() {
    var a, b, s: byte
    var x, y, n: int

    a = 1 << 3
    b = 128 >> 2
    a = b << s
    b = a >> s

    x = 1 << 12
    y = 4096 >> 4
    x = y << n
    y = x >> n
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	checker := New()
	if err := checker.Check(prog); err != nil {
		t.Fatalf("sema check failed: %v", err)
	}
}

func TestSemaStage3DivisionModuloOperators(t *testing.T) {
	input := `
fn main() {
    var a, b: byte
    var x, y: int

    a = 100 / 5
    b = 100 % 7

    x = 1000 / 10
    y = 1000 % 33

    x = a / b
    y = x % 10
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	checker := New()
	if err := checker.Check(prog); err != nil {
		t.Fatalf("sema check failed: %v", err)
	}
}
