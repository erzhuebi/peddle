package codegen

import "fmt"

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
