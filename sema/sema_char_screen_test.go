package sema

import (
	"testing"

	"peddle/lexer"
	"peddle/parser"
)

func checkProgramForCharScreenTest(t *testing.T, src string) error {
	t.Helper()

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return New().Check(prog)
}

func TestSemaAllowsCharLiteralAssignment(t *testing.T) {
	err := checkProgramForCharScreenTest(t, `
fn main() {
    var ch char

    ch = 'P'
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsCharLiteralInPutChar(t *testing.T) {
	err := checkProgramForCharScreenTest(t, `
fn main() {
    cls()

    border(6)
    background(0)
    textcolor(1)

    putchar(0, 0, 'P')
    putchar(1, 0, 'E')
    putchar(2, 0, 'D')
    putchar(3, 0, 'D')
    putchar(4, 0, 'L')
    putchar(5, 0, 'E')

    putscreen(0, 1, 16)
    putcolor(0, 0, 2)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsStringLiteralInPutChar(t *testing.T) {
	err := checkProgramForCharScreenTest(t, `
fn main() {
    putchar(0, 0, "P")
}
`)

	if err == nil {
		t.Fatalf("expected sema error, got nil")
	}
}
