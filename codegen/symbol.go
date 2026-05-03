package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) loadSymbol(sym Symbol) {
	if sym.Type.Name == "int" && !sym.Type.IsArray {
		g.emit(fmt.Sprintf("    lda %s", sym.Label))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda %s+1", sym.Label))
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    lda %s", sym.Label))
}

func (g *Generator) storeAIntoSymbol(sym Symbol) {
	if sym.Type.Name == "int" && !sym.Type.IsArray {
		g.emit("    lda ZP_TMP0")
		g.emit(fmt.Sprintf("    sta %s", sym.Label))
		g.emit("    lda ZP_TMP1")
		g.emit(fmt.Sprintf("    sta %s+1", sym.Label))
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    sta %s", sym.Label))
}

func (g *Generator) loadField(base Symbol, fieldType ast.Type, offset int) error {
	if fieldType.Name == "int" && !fieldType.IsArray {
		g.emit(fmt.Sprintf("    lda %s+%d", base.Label, offset))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda %s+%d", base.Label, offset+1))
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil
	}

	if fieldType.IsArray {
		return fmt.Errorf("array field reads are not implemented yet")
	}

	if _, ok := g.structs[fieldType.Name]; ok {
		return fmt.Errorf("struct field reads are not implemented yet")
	}

	g.emit(fmt.Sprintf("    lda %s+%d", base.Label, offset))
	return nil
}

func (g *Generator) storeIntoField(base Symbol, fieldType ast.Type, offset int) error {
	if fieldType.Name == "int" && !fieldType.IsArray {
		g.emit("    lda ZP_TMP0")
		g.emit(fmt.Sprintf("    sta %s+%d", base.Label, offset))
		g.emit("    lda ZP_TMP1")
		g.emit(fmt.Sprintf("    sta %s+%d", base.Label, offset+1))
		g.usedTmp16 = true
		return nil
	}

	if fieldType.IsArray {
		return fmt.Errorf("array field assignment is not implemented yet")
	}

	if _, ok := g.structs[fieldType.Name]; ok {
		return fmt.Errorf("struct field assignment is not implemented yet")
	}

	g.emit(fmt.Sprintf("    sta %s+%d", base.Label, offset))
	return nil
}
