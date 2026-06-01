package parser

import (
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func TestParseMultipleReturnTypesAndAssignments(t *testing.T) {
	prog := parseProgramForTest(t, `
fn load() (uint, int) {
    var id uint
    var err int

    return id, err
}

fn main() {
    var id uint
    var err int

    id, err = load()
    _, err = load()
}
`)

	load := prog.Functions[0]
	if len(load.ReturnTypes) != 2 {
		t.Fatalf("expected 2 return types, got %d", len(load.ReturnTypes))
	}
	if load.ReturnTypes[0].Name != "uint" || load.ReturnTypes[1].Name != "int" {
		t.Fatalf("got return types %s, %s; want uint, int", load.ReturnTypes[0].String(), load.ReturnTypes[1].String())
	}

	ret, ok := load.Body[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", load.Body[0])
	}
	if len(ret.Values) != 2 {
		t.Fatalf("expected 2 return values, got %d", len(ret.Values))
	}

	mainFn := prog.Functions[1]
	if len(mainFn.Body) != 2 {
		t.Fatalf("expected 2 statements in main, got %d", len(mainFn.Body))
	}

	assign, ok := mainFn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", mainFn.Body[0])
	}
	if got := assign.Targets; len(got) != 2 || got[0] != "id" || got[1] != "err" {
		t.Fatalf("got targets %#v, want [id err]", got)
	}

	ignore, ok := mainFn.Body[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", mainFn.Body[1])
	}
	if got := ignore.Targets; len(got) != 2 || got[0] != "_" || got[1] != "err" {
		t.Fatalf("got targets %#v, want [_ err]", got)
	}
}

func TestParseRejectsGroupedSingleReturnType(t *testing.T) {
	requireParseError(t, `
fn f() (int) {
    return 1
}

fn main() {
}
`)
}

func TestParseRejectsUngroupedMultipleReturnTypes(t *testing.T) {
	requireParseError(t, `
fn f() uint, int {
    return 1, 2
}

fn main() {
}
`)
}

func TestParseRejectsUnderscoreOutsideMultiAssignment(t *testing.T) {
	requireParseError(t, `
fn main() {
    var x int
    _ = x
}
`)

	requireParseError(t, `
fn main() {
    var x int
    x = _
}
`)
}

func requireParseError(t *testing.T, src string) {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	_ = p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatalf("expected parser error")
	}
}
