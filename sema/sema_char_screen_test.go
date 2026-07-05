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

    putraw(0, 1, 16)
    putcolor(0, 0, 2)
    putcharcolor(6, 0, '!', 7)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsUserFunctionCallsInScreenBuiltinArgs(t *testing.T) {
	err := checkProgramForCharScreenTest(t, `
fn xpos() byte {
    return 3
}

fn ypos() byte {
    return 4
}

fn alienChar(row byte) char {
    if row == 0 {
        return 'A'
    }
    return 'B'
}

fn alienColor(row byte) byte {
    return row + 1
}

fn main() {
    var ax byte[2]
    var ay byte[2]
    var i byte
    var row byte
    var title char[8]

    ax[0] = 3
    ay[0] = 4
    i = 0
    row = 0
    copy(title, "ALIEN")

    putchar(ax[i], ay[i], alienChar(row))
    putraw(xpos(), ypos(), alienColor(row))
    putcolor(xpos(), ypos(), alienColor(row))
    putcharcolor(xpos(), ypos(), alienChar(row), alienColor(row))
    gotoxy(xpos(), ypos())
    putstrcolor(xpos(), ypos(), title, alienColor(row))
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

func TestSemaRejectsStringLiteralInPutCharColor(t *testing.T) {
	err := checkProgramForCharScreenTest(t, `
fn main() {
    putcharcolor(0, 0, "P", 2)
}
`)

	if err == nil {
		t.Fatalf("expected sema error, got nil")
	}
}
