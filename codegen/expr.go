package codegen

import (
	"fmt"
	"strconv"

	"peddle/ast"
)

func (g *Generator) genExprTo(e ast.Expr, target ast.Type) error {
	switch expr := e.(type) {
	case *ast.NumberExpr:
		n, err := strconv.Atoi(expr.Value)
		if err != nil {
			return err
		}

		if target.Name == "int" {
			g.emit(fmt.Sprintf("    lda #<%d", n))
			g.emit("    sta ZP_TMP0")
			g.emit(fmt.Sprintf("    lda #>%d", n))
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit(fmt.Sprintf("    lda #%d", n&0xff))
		return nil

	case *ast.IdentExpr:
		if n, ok := g.constants[expr.Name]; ok {
			if target.Name == "int" {
				g.emit(fmt.Sprintf("    lda #<%d", n))
				g.emit("    sta ZP_TMP0")
				g.emit(fmt.Sprintf("    lda #>%d", n))
				g.emit("    sta ZP_TMP1")
				g.usedTmp16 = true
				return nil
			}

			g.emit(fmt.Sprintf("    lda #%d", n&0xff))
			return nil
		}

		sym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Name)
		}

		if target.Name == "int" && sym.Type.Name != "int" && !sym.Type.IsArray {
			g.loadSymbol(sym)
			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.loadSymbol(sym)
		return nil

	case *ast.IndexExpr:
		arraySym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", expr.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", expr.Name)
		}

		if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
			return err
		}

		if arraySym.Type.Name == "int" {
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP0")
			g.emit("    iny")
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    lda (ZP_PTR0_LO), y")
		return nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", expr.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", expr.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, offset, err := g.fieldInfo(elemType, expr.Field)
		if err != nil {
			return err
		}

		if fieldType.IsArray {
			return fmt.Errorf("array field reads are not implemented yet")
		}

		if _, ok := g.structs[fieldType.Name]; ok {
			return fmt.Errorf("struct field reads are not implemented yet")
		}

		if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
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

		if fieldType.Name == "int" {
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP0")
			g.emit("    iny")
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    lda (ZP_PTR0_LO), y")
		return nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(expr.Base)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Base)
		}

		fieldType, offset, err := g.fieldInfo(baseSym.Type, expr.Field)
		if err != nil {
			return err
		}

		if fieldType.IsArray {
			return fmt.Errorf("array field reads are not implemented yet")
		}

		if _, ok := g.structs[fieldType.Name]; ok {
			return fmt.Errorf("struct field reads are not implemented yet")
		}

		return g.loadField(baseSym, fieldType, offset)

	case *ast.UnaryExpr:
		switch expr.Op {
		case "-":
			if target.Name == "int" {
				if err := g.genExprTo(expr.Expr, ast.Type{Name: "int"}); err != nil {
					return err
				}

				g.emit("    lda ZP_TMP0")
				g.emit("    eor #$ff")
				g.emit("    clc")
				g.emit("    adc #1")
				g.emit("    sta ZP_TMP0")

				g.emit("    lda ZP_TMP1")
				g.emit("    eor #$ff")
				g.emit("    adc #0")
				g.emit("    sta ZP_TMP1")

				g.usedTmp16 = true
				return nil
			}

			if err := g.genExprTo(expr.Expr, ast.Type{Name: "byte"}); err != nil {
				return err
			}

			g.emit("    eor #$ff")
			g.emit("    clc")
			g.emit("    adc #1")
			return nil

		case "!":
			if err := g.genExprTo(expr.Expr, ast.Type{Name: "byte"}); err != nil {
				return err
			}

			trueLabel := g.newLabel()
			endLabel := g.newLabel()

			g.emit("    cmp #0")
			g.emit(fmt.Sprintf("    beq %s", trueLabel))
			g.emit("    lda #0")
			g.emit(fmt.Sprintf("    jmp %s", endLabel))
			g.emit(trueLabel + ":")
			g.emit("    lda #1")
			g.emit(endLabel + ":")
			return nil

		default:
			return fmt.Errorf("unsupported unary operator %q", expr.Op)
		}

	case *ast.BinaryExpr:
		return g.genBinaryTo(expr, target)

	case *ast.CallExpr:
		retType, err := g.genCall(expr.Name, expr.Args)
		if err != nil {
			return err
		}

		if retType.Name == "" {
			return fmt.Errorf("function %s does not return a value", expr.Name)
		}

		fnFrame := g.frames[expr.Name]
		if fnFrame == nil || fnFrame.Return == nil {
			if target.Name == "int" && retType.Name != "int" {
				g.emit("    sta ZP_TMP0")
				g.emit("    lda #0")
				g.emit("    sta ZP_TMP1")
				g.usedTmp16 = true
			}
			return nil
		}

		g.loadSymbol(*fnFrame.Return)
		return nil

	case *ast.StringExpr:
		return fmt.Errorf("string expression cannot be used here directly")

	default:
		return fmt.Errorf("unsupported expression")
	}
}
