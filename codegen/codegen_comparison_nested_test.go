package codegen

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestCodegenComparisonPreservesRightSideAcrossBitwiseLeftExpression(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "j",
						Type: ast.Type{Name: "byte"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "j"},
						Value:  &ast.NumberExpr{Value: "27"},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							Op: "==",
							Left: &ast.BinaryExpr{
								Op:    "&",
								Left:  &ast.IdentExpr{Name: "j"},
								Right: &ast.NumberExpr{Value: "4"},
							},
							Right: &ast.NumberExpr{Value: "0"},
						},
						Then: []ast.Stmt{
							&ast.CallStmt{
								Name: "putstr",
								Args: []ast.Expr{
									&ast.NumberExpr{Value: "0"},
									&ast.NumberExpr{Value: "0"},
									&ast.StringExpr{Value: "LEFT"},
								},
							},
						},
						Else: []ast.Stmt{
							&ast.CallStmt{
								Name: "putstr",
								Args: []ast.Expr{
									&ast.NumberExpr{Value: "0"},
									&ast.NumberExpr{Value: "0"},
									&ast.StringExpr{Value: "NO"},
								},
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
		"    lda ZP_TMP0",
		"    pha",
		"    lda ZP_TMP1",
		"    pha",
		"    and peddle_tmp_int0",
		"    pla",
		"    sta peddle_tmp_int0+1",
		"    pla",
		"    sta peddle_tmp_int0",
	}

	for _, needle := range expected {
		if !strings.Contains(asm, needle) {
			t.Fatalf("expected ASM to contain %q\n\nASM:\n%s", needle, asm)
		}
	}
}
