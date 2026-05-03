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

	case *ast.IdentExpr:
		sym, ok := g.resolve(expr.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}

		if sym.Type.IsArray && sym.Type.Name == "char" {
			g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
			g.emit("    sta ZP_PTR0_LO")
			g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
			g.emit("    sta ZP_PTR0_HI")

			g.emit("    jsr peddle_print_string")
			g.usedPrint = true
			return ast.Type{}, nil
		}

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(expr.Base)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Base)
		}

		fieldType, offset, err := g.fieldInfo(baseSym.Type, expr.Field)
		if err != nil {
			return ast.Type{}, err
		}

		if !(fieldType.IsArray && fieldType.Name == "char") {
			return ast.Type{}, fmt.Errorf("unsupported print argument")
		}

		g.emit(fmt.Sprintf("    lda #<%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    jsr peddle_print_string")
		g.usedPrint = true
		return ast.Type{}, nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(expr.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown array %q", expr.Name)
		}
		if !arraySym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", expr.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, offset, err := g.fieldInfo(elemType, expr.Field)
		if err != nil {
			return ast.Type{}, err
		}

		if !(fieldType.IsArray && fieldType.Name == "char") {
			return ast.Type{}, fmt.Errorf("unsupported print argument")
		}

		if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
			return ast.Type{}, err
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

		g.emit("    jsr peddle_print_string")
		g.usedPrint = true
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
