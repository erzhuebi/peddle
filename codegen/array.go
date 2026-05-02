package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genArrayIndexToY(arraySym Symbol, index ast.Expr) error {
	g.emit(fmt.Sprintf("    lda #<%s", arraySym.Label))
	g.emit("    sta ZP_PTR0_LO")
	g.emit(fmt.Sprintf("    lda #>%s", arraySym.Label))
	g.emit("    sta ZP_PTR0_HI")

	if err := g.genExprTo(index, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	if arraySym.Type.Name == "int" {
		g.emit("    asl")
	}

	g.emit("    tay")
	return nil
}
