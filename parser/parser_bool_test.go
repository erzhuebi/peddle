package parser

import (
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func TestParserBoolLiterals(t *testing.T) {
	input := `
fn main() {
    var enabled bool
    var done bool

    enabled = true
    done = false
}
`

	l := lexer.New(input)
	p := New(l)

	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if len(fn.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(fn.Body))
	}

	first, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("first statement should be AssignStmt, got %T", fn.Body[0])
	}

	firstBool, ok := first.Value.(*ast.BoolExpr)
	if !ok {
		t.Fatalf("first assignment value should be BoolExpr, got %T", first.Value)
	}

	if firstBool.Value != true {
		t.Fatalf("first bool value should be true")
	}

	second, ok := fn.Body[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("second statement should be AssignStmt, got %T", fn.Body[1])
	}

	secondBool, ok := second.Value.(*ast.BoolExpr)
	if !ok {
		t.Fatalf("second assignment value should be BoolExpr, got %T", second.Value)
	}

	if secondBool.Value != false {
		t.Fatalf("second bool value should be false")
	}
}
