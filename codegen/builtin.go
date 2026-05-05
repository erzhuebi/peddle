package codegen

import (
	"fmt"
	"strconv"

	"peddle/ast"
)

func (g *Generator) genPrint(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("print expects one argument")
	}

	switch expr := args[0].(type) {
	case *ast.StringExpr:
		label := g.addLiteral(expr.Value)

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    jsr peddle_print_string")
		g.usedPrint = true
		return ast.Type{}, nil

	case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexFieldExpr:
		if err := g.genCharArrayAddress(args[0]); err != nil {
			return ast.Type{}, err
		}

		g.emit("    ldy #2")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    iny")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta peddle_tmp_int0+1")

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc #4")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    jsr peddle_print_counted_string")
		g.usedPrint = true
		g.usedTmp16 = true
		return ast.Type{}, nil
	}

	return ast.Type{}, fmt.Errorf("unsupported print argument")
}

func (g *Generator) genPoke(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("poke expects two arguments")
	}

	if addr, ok := args[0].(*ast.NumberExpr); ok {
		n, err := strconv.Atoi(addr.Value)
		if err != nil {
			return ast.Type{}, err
		}

		if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
			return ast.Type{}, err
		}

		g.emit(fmt.Sprintf("    sta $%04x", n&0xffff))
		return ast.Type{}, nil
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genPeek(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("peek expects one argument")
	}

	if addr, ok := args[0].(*ast.NumberExpr); ok {
		n, err := strconv.Atoi(addr.Value)
		if err != nil {
			return ast.Type{}, err
		}

		g.emit(fmt.Sprintf("    lda $%04x", n&0xffff))
		return ast.Type{Name: "byte"}, nil
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.usedTmp16 = true
	return ast.Type{Name: "byte"}, nil
}

func (g *Generator) genLen(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("len expects one argument")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genSize(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("size expects one argument")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genAppend(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("append expects two arguments")
	}

	if src, ok := args[1].(*ast.StringExpr); ok {
		arrayType, err := g.arrayExprType(args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !(arrayType.IsArray && arrayType.Name == "char") {
			return ast.Type{}, fmt.Errorf("append string literal requires char array destination")
		}

		if err := g.genAppendStringLiteralToCharArray(args[0], src.Value); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{}, nil
	}

	arrayType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	elemType := ast.Type{Name: arrayType.Name}

	if _, ok := g.structs[elemType.Name]; ok {
		return ast.Type{}, fmt.Errorf("append does not support struct elements yet")
	}

	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	} else {
		g.emit("    sta peddle_tmp_int0")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")

	if err := g.genAddElementOffsetToPtr(elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda peddle_tmp_int0")
		g.emit("    ldy #0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    lda peddle_tmp_int0+1")
		g.emit("    iny")
		g.emit("    sta (ZP_PTR0_LO), y")
	} else {
		g.emit("    lda peddle_tmp_int0")
		g.emit("    ldy #0")
		g.emit("    sta (ZP_PTR0_LO), y")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    clc")
	g.emit("    adc #1")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    adc #0")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genStringAppend(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("stradd expects two arguments")
	}

	src, ok := args[1].(*ast.StringExpr)
	if !ok {
		return ast.Type{}, fmt.Errorf("stradd source must be string literal for now")
	}

	if err := g.genAppendStringLiteralToCharArray(args[0], src.Value); err != nil {
		return ast.Type{}, err
	}

	return ast.Type{}, nil
}

func (g *Generator) genCopy(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("copy expects two arguments")
	}

	if src, ok := args[1].(*ast.StringExpr); ok {
		if err := g.genCopyStringLiteralToCharArray(args[0], src.Value); err != nil {
			return ast.Type{}, err
		}
		return ast.Type{}, nil
	}

	dstType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	srcType, err := g.arrayExprType(args[1])
	if err != nil {
		return ast.Type{}, err
	}

	if dstType.Name != srcType.Name {
		return ast.Type{}, fmt.Errorf("copy requires arrays with same element type")
	}

	elemType := ast.Type{Name: dstType.Name}
	elemSize := g.sizeof(elemType)

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta ZP_PTR1_HI")

	if err := g.genArrayAddress(args[1]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    sta (ZP_PTR1_LO), y")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    sta (ZP_PTR1_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda ZP_PTR1_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR1_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR1_HI")

	g.genLengthTimesElemSizeToCounter(elemSize)

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta (ZP_PTR1_LO), y")

	g.emit("    inc ZP_PTR0_LO")
	g.emit("    bne " + loop + "_src_no_carry")
	g.emit("    inc ZP_PTR0_HI")
	g.emit(loop + "_src_no_carry:")

	g.emit("    inc ZP_PTR1_LO")
	g.emit("    bne " + loop + "_dst_no_carry")
	g.emit("    inc ZP_PTR1_HI")
	g.emit(loop + "_dst_no_carry:")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne " + loop + "_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit(loop + "_dec_low:")
	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genFill(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("fill expects two arguments")
	}

	arrayType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	elemType := ast.Type{Name: arrayType.Name}

	if _, ok := g.structs[elemType.Name]; ok {
		return ast.Type{}, fmt.Errorf("fill does not support struct elements yet")
	}

	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta ZP_TMP1")
	} else {
		g.emit("    sta ZP_TMP0")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    tax")
	g.emit("    ldy #2")
	g.emit("    txa")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    ldy #1")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    tax")
	g.emit("    ldy #3")
	g.emit("    txa")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit(fmt.Sprintf("    lda #<%d", arrayType.ArrayLen))
	g.emit("    sta peddle_tmp_int0")
	g.emit(fmt.Sprintf("    lda #>%d", arrayType.ArrayLen))
	g.emit("    sta peddle_tmp_int0+1")

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))

	if elemType.Name == "int" {
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    iny")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta (ZP_PTR0_LO), y")

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc #2")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR0_HI")
	} else {
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")

		g.emit("    inc ZP_PTR0_LO")
		g.emit("    bne " + loop + "_ptr_no_carry")
		g.emit("    inc ZP_PTR0_HI")
		g.emit(loop + "_ptr_no_carry:")
	}

	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne " + loop + "_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit(loop + "_dec_low:")
	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genCharArrayAddress(expr ast.Expr) error {
	t, err := g.arrayExprType(expr)
	if err != nil {
		return err
	}

	if !(t.IsArray && t.Name == "char") {
		return fmt.Errorf("expected char array")
	}

	return g.genArrayAddress(expr)
}

func (g *Generator) genArrayAddress(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", e.Name)
		}

		if !sym.Type.IsArray {
			return fmt.Errorf("expected array")
		}

		g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
		g.emit("    sta ZP_PTR0_HI")
		return nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(e.Base)
		if !ok {
			return fmt.Errorf("unknown variable %q", e.Base)
		}

		fieldType, offset, err := g.fieldInfo(baseSym.Type, e.Field)
		if err != nil {
			return err
		}

		if !fieldType.IsArray {
			return fmt.Errorf("expected array")
		}

		g.emit(fmt.Sprintf("    lda #<%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_HI")
		return nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(e.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", e.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", e.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, offset, err := g.fieldInfo(elemType, e.Field)
		if err != nil {
			return err
		}

		if !fieldType.IsArray {
			return fmt.Errorf("expected array")
		}

		if err := g.genArrayIndexToY(arraySym, e.Index); err != nil {
			return err
		}

		if offset != 0 {
			g.emit("    lda ZP_PTR0_LO")
			g.emit("    clc")
			g.emit(fmt.Sprintf("    adc #%d", offset))
			g.emit("    sta ZP_PTR0_LO")
			g.emit("    lda ZP_PTR0_HI")
			g.emit("    adc #0")
			g.emit("    sta ZP_PTR0_HI")
		}

		return nil
	}

	return fmt.Errorf("expected array")
}

func (g *Generator) arrayExprType(expr ast.Expr) (ast.Type, error) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Name)
		}
		if !sym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return sym.Type, nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(e.Base)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Base)
		}

		fieldType, _, err := g.fieldInfo(baseSym.Type, e.Field)
		if err != nil {
			return ast.Type{}, err
		}
		if !fieldType.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return fieldType, nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown array %q", e.Name)
		}
		if !arraySym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", e.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, _, err := g.fieldInfo(elemType, e.Field)
		if err != nil {
			return ast.Type{}, err
		}
		if !fieldType.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return fieldType, nil
	}

	return ast.Type{}, fmt.Errorf("expected array")
}

func (g *Generator) genCopyStringLiteralToCharArray(dst ast.Expr, value string) error {
	if err := g.genCharArrayAddress(dst); err != nil {
		return err
	}

	label := g.addLiteral(value)
	loop := g.newLabel()
	done := g.newLabel()

	g.emit(fmt.Sprintf("    lda #<%d", len(value)))
	g.emit("    ldy #2")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit(fmt.Sprintf("    lda #>%d", len(value)))
	g.emit("    ldy #3")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldy #0")
	g.emit(loop + ":")
	g.emit(fmt.Sprintf("    cpy #%d", len(value)))
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit(fmt.Sprintf("    lda %s, y", label))
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit(fmt.Sprintf("    jmp %s", loop))
	g.emit(done + ":")
	return nil
}

func (g *Generator) genAppendStringLiteralToCharArray(dst ast.Expr, value string) error {
	if err := g.genCharArrayAddress(dst); err != nil {
		return err
	}

	label := g.addLiteral(value)
	loop := g.newLabel()
	done := g.newLabel()

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_tmp_int0+1")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    clc")
	g.emit(fmt.Sprintf("    adc #<%d", len(value)))
	g.emit("    sta ZP_TMP0")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    adc #>%d", len(value)))
	g.emit("    sta ZP_TMP1")

	g.emit("    ldy #2")
	g.emit("    lda ZP_TMP0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc peddle_tmp_int0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc peddle_tmp_int0+1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldy #0")
	g.emit(loop + ":")
	g.emit(fmt.Sprintf("    cpy #%d", len(value)))
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit(fmt.Sprintf("    lda %s, y", label))
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit(fmt.Sprintf("    jmp %s", loop))
	g.emit(done + ":")

	g.usedTmp16 = true
	return nil
}

func (g *Generator) genUpdateArrayLenFromIndex(arraySym Symbol, index ast.Expr) error {
	if err := g.genExprTo(index, ast.Type{Name: "int"}); err != nil {
		return err
	}

	noCarry := g.newLabel()
	update := g.newLabel()
	done := g.newLabel()

	g.emit("    inc ZP_TMP0")
	g.emit(fmt.Sprintf("    bne %s", noCarry))
	g.emit("    inc ZP_TMP1")
	g.emit(noCarry + ":")

	g.emit(fmt.Sprintf("    lda %s+3", arraySym.Label))
	g.emit("    cmp ZP_TMP1")
	g.emit(fmt.Sprintf("    bcc %s", update))
	g.emit(fmt.Sprintf("    bne %s", done))
	g.emit(fmt.Sprintf("    lda %s+2", arraySym.Label))
	g.emit("    cmp ZP_TMP0")
	g.emit(fmt.Sprintf("    bcc %s", update))
	g.emit(fmt.Sprintf("    jmp %s", done))

	g.emit(update + ":")
	g.emit("    lda ZP_TMP0")
	g.emit(fmt.Sprintf("    sta %s+2", arraySym.Label))
	g.emit("    lda ZP_TMP1")
	g.emit(fmt.Sprintf("    sta %s+3", arraySym.Label))

	g.emit(done + ":")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genAddElementOffsetToPtr(elemType ast.Type) error {
	elemSize := g.sizeof(elemType)

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	if elemSize == 1 {
		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc ZP_TMP0")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc ZP_TMP1")
		g.emit("    sta ZP_PTR0_HI")
		return nil
	}

	if elemSize == 2 {
		g.emit("    asl ZP_TMP0")
		g.emit("    rol ZP_TMP1")
	} else {
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
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genLengthTimesElemSizeToCounter(elemSize int) {
	if elemSize == 1 {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
		g.usedTmp16 = true
		return
	}

	g.emit("    lda #0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    sta peddle_tmp_int0+1")

	for i := 0; i < elemSize; i++ {
		g.emit("    clc")
		g.emit("    lda peddle_tmp_int0")
		g.emit("    adc ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda peddle_tmp_int0+1")
		g.emit("    adc ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	}

	g.usedTmp16 = true
}
