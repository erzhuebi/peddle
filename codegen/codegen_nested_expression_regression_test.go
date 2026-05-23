package codegen

import (
	"strings"
	"testing"

	"peddle/ast"
)

func TestCodegenComparisonPreservesRightSideAcrossNestedBitwiseLeftExpression(t *testing.T) {
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

	assertASMContainsAll(t, asm, []string{
		"    lda ZP_TMP0",
		"    pha",
		"    lda ZP_TMP1",
		"    pha",
		"    and peddle_tmp_int0",
		"    pla",
		"    sta peddle_tmp_int0+1",
		"    pla",
		"    sta peddle_tmp_int0",
	})
}

func TestCodegenComparisonPreservesRightSideAcrossNestedArithmeticLeftExpression(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "a",
						Type: ast.Type{Name: "byte"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "a"},
						Value:  &ast.NumberExpr{Value: "1"},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							Op: "==",
							Left: &ast.BinaryExpr{
								Op:    "+",
								Left:  &ast.IdentExpr{Name: "a"},
								Right: &ast.NumberExpr{Value: "1"},
							},
							Right: &ast.NumberExpr{Value: "2"},
						},
						Then: []ast.Stmt{
							&ast.CallStmt{
								Name: "putstr",
								Args: []ast.Expr{
									&ast.NumberExpr{Value: "0"},
									&ast.NumberExpr{Value: "0"},
									&ast.StringExpr{Value: "YES"},
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

	assertASMContainsAll(t, asm, []string{
		"    pha",
		"    pla",
	})
}

func TestCodegenByteBinaryPreservesRightSideAcrossNestedLeftExpression(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "a",
						Type: ast.Type{Name: "byte"},
					},
					{
						Name: "result",
						Type: ast.Type{Name: "byte"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "a"},
						Value:  &ast.NumberExpr{Value: "27"},
					},
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "result"},
						Value: &ast.BinaryExpr{
							Op: "&",
							Left: &ast.BinaryExpr{
								Op:    "+",
								Left:  &ast.IdentExpr{Name: "a"},
								Right: &ast.NumberExpr{Value: "1"},
							},
							Right: &ast.NumberExpr{Value: "4"},
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

	// This protects against the bug pattern:
	//
	//   right side 4 is saved in ZP_TMP0
	//   left side (a + 1) reuses ZP_TMP0
	//   outer "&" accidentally uses 1 instead of 4
	//
	// A correct implementation must preserve the outer right side while
	// generating the nested left expression.
	assertASMContainsAll(t, asm, []string{
		"    pha",
		"    pla",
		"    and ZP_TMP0",
	})
}

func TestCodegenByteShiftPreservesRightSideAcrossNestedLeftExpression(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "a",
						Type: ast.Type{Name: "byte"},
					},
					{
						Name: "shift",
						Type: ast.Type{Name: "byte"},
					},
					{
						Name: "result",
						Type: ast.Type{Name: "byte"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "a"},
						Value:  &ast.NumberExpr{Value: "3"},
					},
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "shift"},
						Value:  &ast.NumberExpr{Value: "2"},
					},
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "result"},
						Value: &ast.BinaryExpr{
							Op: "<<",
							Left: &ast.BinaryExpr{
								Op:    "+",
								Left:  &ast.IdentExpr{Name: "a"},
								Right: &ast.NumberExpr{Value: "1"},
							},
							Right: &ast.IdentExpr{Name: "shift"},
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

	// Variable shifts use the right side as the shift count.
	// If the nested left expression overwrites that count, the shift amount
	// becomes wrong.
	assertASMContainsAll(t, asm, []string{
		"    pha",
		"    pla",
		"    ldx ZP_TMP0",
	})
}

func TestCodegenIntBinaryPreservesRightSideAcrossNestedLeftExpression(t *testing.T) {
	p := &ast.Program{
		Functions: []*ast.FunctionDecl{
			{
				Name: "main",
				Locals: []*ast.VarDecl{
					{
						Name: "a",
						Type: ast.Type{Name: "int"},
					},
					{
						Name: "result",
						Type: ast.Type{Name: "int"},
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "a"},
						Value:  &ast.NumberExpr{Value: "27"},
					},
					&ast.AssignStmt{
						Target: &ast.VarLValue{Name: "result"},
						Value: &ast.BinaryExpr{
							Op: "&",
							Left: &ast.BinaryExpr{
								Op:    "+",
								Left:  &ast.IdentExpr{Name: "a"},
								Right: &ast.NumberExpr{Value: "1"},
							},
							Right: &ast.NumberExpr{Value: "4"},
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

	// Int binary ops save the right side in peddle_tmp_int0.
	// A nested left expression can reuse peddle_tmp_int0, so the outer right
	// side must be preserved before generating the left expression.
	assertASMContainsAll(t, asm, []string{
		"    pha",
		"    pla",
		"    and peddle_tmp_int0",
		"    and peddle_tmp_int0+1",
	})
}

func assertASMContainsAll(t *testing.T, asm string, needles []string) {
	t.Helper()

	for _, needle := range needles {
		if !strings.Contains(asm, needle) {
			t.Fatalf("expected ASM to contain %q\n\nASM:\n%s", needle, asm)
		}
	}
}
