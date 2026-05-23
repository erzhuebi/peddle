package codegen

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestKeyBuiltinEmitsGetIn(t *testing.T) {
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
							Name: "key",
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
		t.Fatalf("expected key() to emit KERNAL GETIN call, asm:\n%s", asm)
	}
}

func TestKeyBuiltinRejectsArgumentsInCodegen(t *testing.T) {
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
							Name: "key",
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
		t.Fatalf("key(1) should fail")
	}

	if !strings.Contains(err.Error(), "key expects no arguments") {
		t.Fatalf("expected key argument error, got: %v", err)
	}
}
