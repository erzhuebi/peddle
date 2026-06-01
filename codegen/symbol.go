package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) loadSymbol(sym Symbol) {
	if isWordSymbol(sym) {
		g.emit(fmt.Sprintf("    lda %s", sym.Label))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda %s+1", sym.Label))
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    lda %s", sym.Label))
}

func (g *Generator) loadSymbolAs(sym Symbol, target ast.Type) {
	if !isWordType(target) && isWordSymbol(sym) {
		g.emit(fmt.Sprintf("    lda %s", sym.Label))
		return
	}

	g.loadSymbol(sym)

	if isWordType(target) && !isWordSymbol(sym) {
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
	}
}

func (g *Generator) storeAIntoSymbol(sym Symbol) {
	if isWordSymbol(sym) {
		g.emit("    lda ZP_TMP0")
		g.emit(fmt.Sprintf("    sta %s", sym.Label))
		g.emit("    lda ZP_TMP1")
		g.emit(fmt.Sprintf("    sta %s+1", sym.Label))
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    sta %s", sym.Label))
}

func isWordSymbol(sym Symbol) bool {
	return sym.IsReference || isWordType(sym.Type)
}

func isWordType(t ast.Type) bool {
	return t.IsPointer || (!t.IsArray && (t.Name == "int" || t.Name == "uint"))
}

func pointerPointeeType(t ast.Type) ast.Type {
	return ast.Type{Name: t.Name}
}

func isScalarPointerType(t ast.Type) bool {
	return t.IsPointer && !t.IsArray && isScalarTypeName(t.Name)
}

func isScalarTypeName(name string) bool {
	switch name {
	case "byte", "char", "bool", "int", "uint":
		return true
	default:
		return false
	}
}

func (g *Generator) loadScalarPointer(sym Symbol) error {
	pointee := pointerPointeeType(sym.Type)

	g.loadPointerFieldAddress(sym, 0)

	if isWordType(pointee) {
		g.emit("    ldy #0")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta ZP_TMP0")
		g.emit("    iny")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	return nil
}

func (g *Generator) storeIntoScalarPointer(sym Symbol) error {
	pointee := pointerPointeeType(sym.Type)

	if isWordType(pointee) {
		g.loadPointerFieldAddress(sym, 0)
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    iny")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.usedTmp16 = true
		return nil
	}

	g.emit("    sta peddle_tmp_int0")
	g.usedTmp16 = true

	g.loadPointerFieldAddress(sym, 0)
	g.emit("    ldy #0")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    sta (ZP_PTR0_LO), y")
	return nil
}

func (g *Generator) loadField(base Symbol, fieldType ast.Type, offset int) error {
	if isWordType(fieldType) {
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
	if isWordType(fieldType) {
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

func (g *Generator) loadPointerField(base Symbol, fieldType ast.Type, offset int) error {
	if fieldType.IsArray {
		return fmt.Errorf("array field reads through pointer are not implemented yet")
	}

	if _, ok := g.structs[fieldType.Name]; ok {
		return fmt.Errorf("struct field reads through pointer are not implemented yet")
	}

	g.loadPointerFieldAddress(base, offset)

	if isWordType(fieldType) {
		g.emit("    ldy #0")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta ZP_TMP0")
		g.emit("    iny")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	return nil
}

func (g *Generator) storeIntoPointerField(base Symbol, fieldType ast.Type, offset int) error {
	if fieldType.IsArray {
		return fmt.Errorf("array field assignment through pointer is not implemented yet")
	}

	if _, ok := g.structs[fieldType.Name]; ok {
		return fmt.Errorf("struct field assignment through pointer is not implemented yet")
	}

	if isWordType(fieldType) {
		g.loadPointerFieldAddress(base, offset)
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    iny")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.usedTmp16 = true
		return nil
	}

	g.emit("    sta peddle_tmp_int0")
	g.usedTmp16 = true

	g.loadPointerFieldAddress(base, offset)
	g.emit("    ldy #0")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    sta (ZP_PTR0_LO), y")
	return nil
}

func (g *Generator) loadPointerFieldAddress(base Symbol, offset int) {
	g.emit(fmt.Sprintf("    lda %s", base.Label))
	g.emit("    sta ZP_PTR0_LO")
	g.emit(fmt.Sprintf("    lda %s+1", base.Label))
	g.emit("    sta ZP_PTR0_HI")

	if offset == 0 {
		return
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit(fmt.Sprintf("    adc #<%d", offset))
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit(fmt.Sprintf("    adc #>%d", offset))
	g.emit("    sta ZP_PTR0_HI")
}
