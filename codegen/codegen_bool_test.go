package codegen

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestCodegenBoolLiteralsEmitOneAndZero(t *testing.T) {
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

	asm, err := New().Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(asm, "    lda #1") {
		t.Fatalf("expected true to emit lda #1, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    lda #0") {
		t.Fatalf("expected false to emit lda #0, asm:\n%s", asm)
	}
}

func TestCodegenBoolLiteralToIntUsesTmp16(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "n",
						Type: ast.Type{Name: "int"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "n"},
						Value:  &ast.BoolExpr{Value: true},
					},
				},
			},
		},
	}

	asm, err := New().Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(asm, "    lda #<1") {
		t.Fatalf("expected true assigned to int to emit low byte, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    lda #>1") {
		t.Fatalf("expected true assigned to int to emit high byte, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    sta ZP_TMP0") {
		t.Fatalf("expected int assignment to use ZP_TMP0, asm:\n%s", asm)
	}

	if !strings.Contains(asm, "    sta ZP_TMP1") {
		t.Fatalf("expected int assignment to use ZP_TMP1, asm:\n%s", asm)
	}
}
