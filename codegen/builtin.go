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

func (g *Generator) genStrlen(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("strlen expects one argument")
	}

	switch expr := args[0].(type) {
	case *ast.StringExpr:
		g.emit(fmt.Sprintf("    lda #<%d", len(expr.Value)))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda #>%d", len(expr.Value)))
		g.emit("    sta ZP_TMP1")
		g.emit("    lda ZP_TMP0")
		g.usedTmp16 = true
		return ast.Type{Name: "int"}, nil

	case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexFieldExpr:
		if err := g.genCharArrayAddress(args[0]); err != nil {
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

	return ast.Type{}, fmt.Errorf("unsupported strlen argument")
}

func (g *Generator) genStrcpy(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("strcpy expects two arguments")
	}

	src, ok := args[1].(*ast.StringExpr)
	if !ok {
		return ast.Type{}, fmt.Errorf("strcpy source must be string literal for now")
	}

	if err := g.genCopyStringLiteralToCharArray(args[0], src.Value); err != nil {
		return ast.Type{}, err
	}

	return ast.Type{}, nil
}

func (g *Generator) genStradd(args []ast.Expr) (ast.Type, error) {
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

func (g *Generator) genCharArrayAddress(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", e.Name)
		}

		if !(sym.Type.IsArray && sym.Type.Name == "char") {
			return fmt.Errorf("expected char array")
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

		if !(fieldType.IsArray && fieldType.Name == "char") {
			return fmt.Errorf("expected char array")
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

		if !(fieldType.IsArray && fieldType.Name == "char") {
			return fmt.Errorf("expected char array")
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

	return fmt.Errorf("expected char array")
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
