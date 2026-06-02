package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genMemBaseToTmp(sym Symbol) error {
	if !sym.Type.IsMem {
		return fmt.Errorf("%q is not mem", sym.SourceName)
	}

	if sym.IsReference {
		g.loadSymbol(sym)
		return nil
	}

	if !sym.HasAtAddress {
		return fmt.Errorf("mem %q has no fixed address", sym.SourceName)
	}

	g.emit(fmt.Sprintf("    lda #<%d", sym.AtAddress))
	g.emit("    sta ZP_TMP0")
	g.emit(fmt.Sprintf("    lda #>%d", sym.AtAddress))
	g.emit("    sta ZP_TMP1")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genMemIndexToPtr(sym Symbol, index ast.Expr) error {
	if err := g.genExprTo(index, ast.Type{Name: "int"}); err != nil {
		return err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_tmp_int0+1")

	if err := g.genMemBaseToTmp(sym); err != nil {
		return err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    clc")
	g.emit("    adc peddle_tmp_int0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    adc peddle_tmp_int0+1")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    ldy #0")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genMemAddress(expr ast.Expr) error {
	ident, ok := expr.(*ast.IdentExpr)
	if !ok {
		return fmt.Errorf("mem parameter requires mem storage")
	}

	sym, ok := g.resolve(ident.Name)
	if !ok {
		return fmt.Errorf("unknown variable %q", ident.Name)
	}

	return g.genMemBaseToTmp(sym)
}
