package parser

import (
	"testing"

	"peddle/ast"
)

func TestParseMemWindowDeclarationAndParameter(t *testing.T) {
	prog := parseProgramForTest(t, `
fn clear(buf mem[1000]) {
    buf[0] = 32
}

fn main() {
    var screen mem[1000] at $0400

    clear(screen)
}
`)

	clearFn := prog.Functions[0]
	if len(clearFn.Params) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(clearFn.Params))
	}
	paramType := clearFn.Params[0].Type
	if !paramType.IsMem || paramType.ArrayLen != 1000 {
		t.Fatalf("expected mem[1000] parameter, got %#v", paramType)
	}

	mainFn := prog.Functions[1]
	if len(mainFn.Locals) != 1 {
		t.Fatalf("expected 1 local, got %d", len(mainFn.Locals))
	}
	local := mainFn.Locals[0]
	if local.Name != "screen" {
		t.Fatalf("got local name %q, want screen", local.Name)
	}
	if !local.Type.IsMem || local.Type.ArrayLen != 1000 {
		t.Fatalf("expected mem[1000] local, got %#v", local.Type)
	}
	if !local.HasAtAddress || local.AtAddress != 1024 {
		t.Fatalf("expected at address 1024, got has=%t address=%d", local.HasAtAddress, local.AtAddress)
	}

	assign, ok := clearFn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", clearFn.Body[0])
	}
	if _, ok := assign.Target.(*ast.IndexLValue); !ok {
		t.Fatalf("expected IndexLValue target, got %T", assign.Target)
	}
}

func TestParseRejectsInvalidMemAtDeclarations(t *testing.T) {
	tests := []string{
		`
fn main() {
    var a, b mem[10] at $0400
}
`,
		`
fn main() {
    var data byte[10] at $0400
}
`,
		`
fn main() {
    var screen mem[10] at screen
}
`,
		`
fn main() {
    var screen mem at $0400
}
`,
	}

	for _, src := range tests {
		requireParseError(t, src)
	}
}
