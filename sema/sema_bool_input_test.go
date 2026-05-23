package sema

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestBoolLiteralsCanAssignToBool(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "enabled",
						Type: ast.Type{Name: "bool"},
					},
					{
						Name: "done",
						Type: ast.Type{Name: "bool"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "enabled"},
						Value:  &ast.BoolExpr{Value: true},
					},
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "done"},
						Value:  &ast.BoolExpr{Value: false},
					},
				},
			},
		},
	}

	if err := New().Check(p); err != nil {
		t.Fatalf("true/false should be assignable to bool: %v", err)
	}
}

func TestBoolLiteralsCanBeUsedInConditions(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Body: []ast.Stmt{
					&ast.IfStmt{
						Cond: &ast.BoolExpr{Value: true},
						Then: []ast.Stmt{},
						Else: []ast.Stmt{},
					},
					&ast.WhileStmt{
						Cond: &ast.BoolExpr{Value: false},
						Body: []ast.Stmt{},
					},
				},
			},
		},
	}

	if err := New().Check(p); err != nil {
		t.Fatalf("true/false should be valid conditions: %v", err)
	}
}

func TestWaitKeyBuiltinCanAssignToChar(t *testing.T) {
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

	if err := New().Check(p); err != nil {
		t.Fatalf("waitkey() should be valid when assigned to char: %v", err)
	}
}

func TestWaitKeyBuiltinRejectsArguments(t *testing.T) {
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

	err := New().Check(p)
	if err == nil {
		t.Fatalf("waitkey(1) should fail")
	}

	if !strings.Contains(err.Error(), "waitkey expects no arguments") {
		t.Fatalf("expected waitkey argument error, got: %v", err)
	}
}

func TestReadLineBuiltinCanAssignReturnLengthToInt(t *testing.T) {
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

	if err := New().Check(p); err != nil {
		t.Fatalf("readline(line, true, 16) should be valid: %v", err)
	}
}

func TestReadLineBuiltinRejectsNonArrayBuffer(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "line",
						Type: ast.Type{Name: "char"},
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

	err := New().Check(p)
	if err == nil {
		t.Fatalf("readline() should reject non-array buffer")
	}

	if !strings.Contains(err.Error(), "readline buffer must be char array") {
		t.Fatalf("expected readline buffer error, got: %v", err)
	}
}

func TestReadLineBuiltinRejectsWrongArgumentCount(t *testing.T) {
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

	err := New().Check(p)
	if err == nil {
		t.Fatalf("readline(line) should fail")
	}

	if !strings.Contains(err.Error(), "readline expects three arguments") {
		t.Fatalf("expected readline argument count error, got: %v", err)
	}
}
