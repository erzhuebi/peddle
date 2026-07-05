package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genInitSymbol(sym Symbol, init ast.Expr) error {
	if sym.IsReference {
		return fmt.Errorf("cannot initialize reference parameter")
	}

	if sym.Type.IsArray || g.isStructType(sym.Type) {
		g.genResetForInitializer(sym.Type, sym.Label, 0)
		return g.genInitAt(sym.Type, sym.Label, 0, init)
	}

	if err := g.genExprTo(init, sym.Type); err != nil {
		return err
	}
	g.storeAAt(sym.Label, 0, sym.Type)
	return nil
}

func (g *Generator) genInitAt(t ast.Type, label string, offset int, init ast.Expr) error {
	if t.IsArray {
		if str, ok := init.(*ast.StringExpr); ok {
			if t.Name != "char" {
				return fmt.Errorf("string initializer requires char array")
			}
			g.genSetArrayLenAt(label, offset, len(str.Value))
			g.genStoreStringBytesAt(label, offset+4, str.Value)
			return nil
		}

		lit, ok := init.(*ast.ArrayLiteralExpr)
		if !ok {
			return fmt.Errorf("array initializer must be an array literal")
		}

		g.genSetArrayLenAt(label, offset, len(lit.Values))

		elemType := ast.Type{Name: t.Name}
		elemSize := g.sizeof(elemType)
		for i, value := range lit.Values {
			elemOffset := offset + 4 + i*elemSize
			if err := g.genInitAt(elemType, label, elemOffset, value); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
		return nil
	}

	if _, ok := g.structs[t.Name]; ok {
		lit, ok := init.(*ast.StructLiteralExpr)
		if !ok {
			return fmt.Errorf("struct initializer for %s must be a struct literal", t.Name)
		}

		for _, initField := range lit.Fields {
			fieldType, fieldOffset, err := g.fieldInfo(t, initField.Name)
			if err != nil {
				return err
			}
			if err := g.genInitAt(fieldType, label, offset+fieldOffset, initField.Value); err != nil {
				return fmt.Errorf("field %q: %w", initField.Name, err)
			}
		}

		return nil
	}

	if err := g.genExprTo(init, t); err != nil {
		return err
	}
	g.storeAAt(label, offset, t)
	return nil
}

func (g *Generator) genResetForInitializer(t ast.Type, label string, offset int) {
	if t.IsArray {
		g.genSetArrayLenAt(label, offset, 0)

		elemType := ast.Type{Name: t.Name}
		elemSize := g.sizeof(elemType)
		dataOffset := offset + 4

		if g.isStructType(elemType) {
			for i := 0; i < t.ArrayLen; i++ {
				g.genResetForInitializer(elemType, label, dataOffset+i*elemSize)
			}
			return
		}

		g.genZeroRange(label, dataOffset, elemSize*t.ArrayLen)
		return
	}

	if s, ok := g.structs[t.Name]; ok {
		fieldOffset := offset
		for _, field := range s.Fields {
			g.genResetForInitializer(field.Type, label, fieldOffset)
			fieldOffset += g.sizeof(field.Type)
		}
		return
	}

	g.genZeroRange(label, offset, g.sizeof(t))
}

func (g *Generator) genSetArrayLenAt(label string, offset int, length int) {
	g.emit(fmt.Sprintf("    lda #<%d", length))
	g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset+2)))
	g.emit(fmt.Sprintf("    lda #>%d", length))
	g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset+3)))
}

func (g *Generator) genStoreStringBytesAt(label string, offset int, value string) {
	for i := 0; i < len(value); i++ {
		g.emit(fmt.Sprintf("    lda #%d", value[i]))
		g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset+i)))
	}
}

func (g *Generator) storeAAt(label string, offset int, t ast.Type) {
	if isWordType(t) {
		g.emit("    lda ZP_TMP0")
		g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset)))
		g.emit("    lda ZP_TMP1")
		g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset+1)))
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset)))
}

func (g *Generator) genZeroRange(label string, offset int, size int) {
	if size <= 0 {
		return
	}

	if size <= 16 {
		g.emit("    lda #0")
		for i := 0; i < size; i++ {
			g.emit(fmt.Sprintf("    sta %s", asmLabelOffset(label, offset+i)))
		}
		return
	}

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(fmt.Sprintf("    lda #<%s", asmLabelOffset(label, offset)))
	g.emit("    sta ZP_PTR0_LO")
	g.emit(fmt.Sprintf("    lda #>%s", asmLabelOffset(label, offset)))
	g.emit("    sta ZP_PTR0_HI")
	g.emit(fmt.Sprintf("    lda #<%d", size))
	g.emit("    sta peddle_tmp_int0")
	g.emit(fmt.Sprintf("    lda #>%d", size))
	g.emit("    sta peddle_tmp_int0+1")

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit("    lda #0")
	g.emit("    ldy #0")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    inc ZP_PTR0_LO")
	g.emit("    bne " + loop + "_ptr_no_carry")
	g.emit("    inc ZP_PTR0_HI")
	g.emit(loop + "_ptr_no_carry:")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne " + loop + "_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit(loop + "_dec_low:")
	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))
	g.emit(done + ":")

	g.usedTmp16 = true
}

func (g *Generator) emitStaticValueInit(t ast.Type, init ast.Expr) error {
	if t.IsArray {
		if str, ok := init.(*ast.StringExpr); ok {
			if t.Name != "char" {
				return fmt.Errorf("string initializer requires char array")
			}
			g.emitStaticArrayHeader(t.ArrayLen, len(str.Value))
			g.emitStaticStringData(t.ArrayLen, str.Value)
			return nil
		}

		lit, ok := init.(*ast.ArrayLiteralExpr)
		if !ok {
			return fmt.Errorf("array initializer must be an array literal")
		}

		g.emitStaticArrayHeader(t.ArrayLen, len(lit.Values))

		elemType := ast.Type{Name: t.Name}
		for i := 0; i < t.ArrayLen; i++ {
			if i < len(lit.Values) {
				if err := g.emitStaticValueInit(elemType, lit.Values[i]); err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
			} else {
				g.emitStaticValue(elemType)
			}
		}
		return nil
	}

	if s, ok := g.structs[t.Name]; ok {
		lit, ok := init.(*ast.StructLiteralExpr)
		if !ok {
			return fmt.Errorf("struct initializer for %s must be a struct literal", t.Name)
		}

		fields := map[string]ast.Expr{}
		for _, field := range lit.Fields {
			fields[field.Name] = field.Value
		}

		for _, field := range s.Fields {
			if value, ok := fields[field.Name]; ok {
				if err := g.emitStaticValueInit(field.Type, value); err != nil {
					return fmt.Errorf("field %q: %w", field.Name, err)
				}
			} else {
				g.emitStaticValue(field.Type)
			}
		}
		return nil
	}

	value, ok, err := g.foldConstExpr(init, t)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("initializer must be constant")
	}

	g.emitStaticScalarValue(t, value)
	return nil
}

func (g *Generator) emitStaticArrayHeader(capacity int, length int) {
	g.emit(fmt.Sprintf("    .word %d", capacity))
	g.emit(fmt.Sprintf("    .word %d", length))
}

func (g *Generator) emitStaticStringData(capacity int, value string) {
	if len(value) > 0 {
		g.emit(fmt.Sprintf("    .byte %s", asmStringBytes(value)))
	}
	if remaining := capacity - len(value); remaining > 0 {
		g.emit(fmt.Sprintf("    .fill %d, 0", remaining))
	}
}

func (g *Generator) emitStaticScalarValue(t ast.Type, value int) {
	if isWordType(t) {
		g.emit(fmt.Sprintf("    .word %d", normalizeConstValue(value, t)))
		return
	}

	g.emit(fmt.Sprintf("    .byte %d", normalizeConstValue(value, t)))
}

func (g *Generator) isStructType(t ast.Type) bool {
	if t.IsArray || t.IsMem || t.IsPointer {
		return false
	}
	_, ok := g.structs[t.Name]
	return ok
}

func asmLabelOffset(label string, offset int) string {
	if offset == 0 {
		return label
	}
	return fmt.Sprintf("%s+%d", label, offset)
}
