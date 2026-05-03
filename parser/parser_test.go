package parser

import (
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func parseExprFromMain(t *testing.T, src string) ast.Expr {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	return assign.Value
}

func TestParseUnaryMinus(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: int
    x = -1
}
`)

	u, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}

	if u.Op != "-" {
		t.Fatalf("got op %q, want -", u.Op)
	}
}

func TestParseUnaryBang(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: bool
    x = !0
}
`)

	u, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}

	if u.Op != "!" {
		t.Fatalf("got op %q, want !", u.Op)
	}
}

func parseProgramForTest(t *testing.T, src string) *ast.Program {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return prog
}

func TestParseIfElse(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var x: byte

    if x == 0 {
        x = 1
    } else {
        x = 2
    }
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	stmt, ok := fn.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body[0])
	}

	if len(stmt.Then) != 1 {
		t.Fatalf("expected 1 then stmt, got %d", len(stmt.Then))
	}

	if len(stmt.Else) != 1 {
		t.Fatalf("expected 1 else stmt, got %d", len(stmt.Else))
	}
}

func TestParseWhile(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var i: byte

    while i < 10 {
        i = i + 1
    }
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	_, ok := fn.Body[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", fn.Body[0])
	}
}

func TestParseUserFunctionCall(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: int
    x = add(1, 2)
}
`)

	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", expr)
	}

	if call.Name != "add" {
		t.Fatalf("got call name %q, want add", call.Name)
	}

	if len(call.Args) != 2 {
		t.Fatalf("got %d args, want 2", len(call.Args))
	}
}

func TestParseArrayIndexAssignment(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var a: int[4]

    a[0] = 1
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	_, ok = assign.Target.(*ast.IndexLValue)
	if !ok {
		t.Fatalf("expected IndexLValue, got %T", assign.Target)
	}
}
