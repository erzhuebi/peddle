package sema

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestKeyBuiltinCanAssignToChar(t *testing.T) {
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

	if err := New().Check(p); err != nil {
		t.Fatalf("key() should be valid when assigned to char: %v", err)
	}
}

func TestKeyBuiltinCanAssignToInt(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "k",
						Type: ast.Type{Name: "int"},
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

	if err := New().Check(p); err != nil {
		t.Fatalf("key() should be assignable to int: %v", err)
	}
}

func TestKeyBuiltinRejectsArguments(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
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
				Locals: []*ast.VarDecl{
					{
						Name: "k",
						Type: ast.Type{Name: "char"},
					},
				},
			},
		},
	}

	err := New().Check(p)
	if err == nil {
		t.Fatalf("key(1) should fail")
	}

	if !strings.Contains(err.Error(), "key expects no arguments") {
		t.Fatalf("expected key argument error, got: %v", err)
	}
}
