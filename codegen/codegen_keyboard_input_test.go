package codegen

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestCodegenWaitKeyEmitsBlockingGetInLoop(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "k",
						Type: ast.Type{Name: "char"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "k"},
						Value: &ast.CallExpr{
							Name: "waitkey",
						},
					},
				},
			},
		},
	}

	asm, err := New().Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(asm, "    jsr $ffe4") {
		t.Fatalf("expected waitkey() to call KERNAL GETIN, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    cmp #0") {
		t.Fatalf("expected waitkey() to compare against zero, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    beq ") {
		t.Fatalf("expected waitkey() to branch while no key is available, asm:\n%s", asm)
	}
}

func TestCodegenWaitKeyRejectsArguments(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "k",
						Type: ast.Type{Name: "char"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "k"},
						Value: &ast.CallExpr{
							Name: "waitkey",
							Args: []ast.Expr{
								&ast.NumberExpr{Value: "1"},
							},
						},
					},
				},
			},
		},
	}

	_, err := New().Generate(p)
	if err == nil {
		t.Fatalf("waitkey(1) should fail")
	}

	if !strings.Contains(err.Error(), "waitkey expects no arguments") {
		t.Fatalf("expected waitkey argument error, got: %v", err)
	}
}

func TestCodegenReadLineEmitsRuntimeCallAndRuntime(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "line",
						Type: ast.Type{Name: "char", IsArray: true, ArrayLen: 32},
					},
					{
						Name: "n",
						Type: ast.Type{Name: "int"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "n"},
						Value: &ast.CallExpr{
							Name: "readline",
							Args: []ast.Expr{
								&ast.IdentExpr{Name: "line"},
								&ast.BoolExpr{Value: true},
								&ast.NumberExpr{Value: "16"},
							},
						},
					},
				},
			},
		},
	}

	asm, err := New().Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expected := []string{
		"    sta peddle_readline_echo",
		"    sta peddle_readline_max",
		"    sta peddle_readline_max+1",
		"    jsr peddle_readline",
		"peddle_readline:",
		"peddle_readline_loop:",
		"    jsr $ffe4",
		"    cmp #13",
		"    beq peddle_readline_done",
		"    jsr $ffd2",
		"peddle_readline_done:",
	}

	for _, needle := range expected {
		if !strings.Contains(asm, needle) {
			t.Fatalf("expected readline asm to contain %q, asm:\n%s", needle, asm)
		}
	}
}

func TestCodegenReadLineRejectsWrongArgumentCount(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "line",
						Type: ast.Type{Name: "char", IsArray: true, ArrayLen: 32},
					},
					{
						Name: "n",
						Type: ast.Type{Name: "int"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "n"},
						Value: &ast.CallExpr{
							Name: "readline",
							Args: []ast.Expr{
								&ast.IdentExpr{Name: "line"},
							},
						},
					},
				},
			},
		},
	}

	_, err := New().Generate(p)
	if err == nil {
		t.Fatalf("readline(line) should fail")
	}

	if !strings.Contains(err.Error(), "readline expects three arguments") {
		t.Fatalf("expected readline argument count error, got: %v", err)
	}
}
