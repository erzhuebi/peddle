package parser

import (
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func parseProgramForCharTest(t *testing.T, src string) *ast.Program {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return prog
}

func TestParseCharLiteralInPutCharCall(t *testing.T) {
	prog := parseProgramForCharTest(t, `
fn main() {
    putchar(0, 0, 'P')
}
`)

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	call, ok := fn.Body[0].(*ast.CallStmt)
	if !ok {
		t.Fatalf("expected CallStmt, got %T", fn.Body[0])
	}

	if call.Name != "putchar" {
		t.Fatalf("got call name %q, want putchar", call.Name)
	}

	if len(call.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(call.Args))
	}

	ch, ok := call.Args[2].(*ast.CharExpr)
	if !ok {
		t.Fatalf("expected third arg CharExpr, got %T", call.Args[2])
	}

	if ch.Value != "80" {
		t.Fatalf("got char literal value %q, want 80", ch.Value)
	}
}

func TestParseCharLiteralAssignment(t *testing.T) {
	prog := parseProgramForCharTest(t, `
fn main() {
    var ch char

    ch = 'A'
}
`)

	fn := prog.Functions[0]

	if len(fn.Locals) != 1 {
		t.Fatalf("expected 1 local, got %d", len(fn.Locals))
	}

	if fn.Locals[0].Name != "ch" {
		t.Fatalf("got local name %q, want ch", fn.Locals[0].Name)
	}

	if fn.Locals[0].Type.Name != "char" {
		t.Fatalf("got local type %q, want char", fn.Locals[0].Type.Name)
	}

	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	ch, ok := assign.Value.(*ast.CharExpr)
	if !ok {
		t.Fatalf("expected CharExpr assignment value, got %T", assign.Value)
	}

	if ch.Value != "65" {
		t.Fatalf("got char literal value %q, want 65", ch.Value)
	}
}
