package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genArrayIndexToY(arraySym Symbol, index ast.Expr) error {
	elemType := ast.Type{Name: arraySym.Type.Name}
	elemSize := g.sizeof(elemType)

	if elemSize <= 0 {
		return fmt.Errorf("invalid element size for array %q", arraySym.SourceName)
	}

	if err := g.genExprTo(index, ast.Type{Name: "int"}); err != nil {
		return err
	}

	if elemSize == 2 {
		g.emit("    asl ZP_TMP0")
		g.emit("    rol ZP_TMP1")
		g.usedTmp16 = true
	} else if elemSize > 2 {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")

		for i := 1; i < elemSize; i++ {
			g.emit("    clc")
			g.emit("    lda ZP_TMP0")
			g.emit("    adc peddle_tmp_int0")
			g.emit("    sta ZP_TMP0")
			g.emit("    lda ZP_TMP1")
			g.emit("    adc peddle_tmp_int0+1")
			g.emit("    sta ZP_TMP1")
		}

		g.usedTmp16 = true
	}

	g.emit(fmt.Sprintf("    lda #<%s+4", arraySym.Label))
	g.emit("    clc")
	g.emit("    adc ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")

	g.emit(fmt.Sprintf("    lda #>%s+4", arraySym.Label))
	g.emit("    adc ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldy #0")
	return nil
}
