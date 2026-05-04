package codegen

import (
	"fmt"
	"strconv"

	"peddle/ast"
)

func (g *Generator) genBinaryTo(b *ast.BinaryExpr, target ast.Type) error {
	switch b.Op {
	case "+", "-":
		if target.Name == "int" {
			return g.genBinaryInt(b)
		}
		return g.genBinaryByte(b)

	case "==", "!=", "<", "<=", ">", ">=":
		return g.genComparisonToBool(b)

	default:
		return fmt.Errorf("unsupported binary op %q", b.Op)
	}
}

func (g *Generator) genBinaryByte(b *ast.BinaryExpr) error {
	if err := g.genExprTo(b.Right, ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta ZP_TMP0")

	if err := g.genExprTo(b.Left, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	switch b.Op {
	case "+":
		g.emit("    clc")
		g.emit("    adc ZP_TMP0")
	case "-":
		g.emit("    sec")
		g.emit("    sbc ZP_TMP0")
	}

	return nil
}

func (g *Generator) genBinaryInt(b *ast.BinaryExpr) error {
	if err := g.genExprTo(b.Right, ast.Type{Name: "int"}); err != nil {
		return err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_tmp_int0+1")

	if err := g.genExprTo(b.Left, ast.Type{Name: "int"}); err != nil {
		return err
	}

	switch b.Op {
	case "+":
		g.emit("    clc")
		g.emit("    lda ZP_TMP0")
		g.emit("    adc peddle_tmp_int0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    adc peddle_tmp_int0+1")
		g.emit("    sta ZP_TMP1")
	case "-":
		g.emit("    sec")
		g.emit("    lda ZP_TMP0")
		g.emit("    sbc peddle_tmp_int0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sbc peddle_tmp_int0+1")
		g.emit("    sta ZP_TMP1")
	}

	g.usedTmp16 = true
	return nil
}

func (g *Generator) genComparisonToBool(b *ast.BinaryExpr) error {
	trueLabel := g.newLabel()
	endLabel := g.newLabel()

	if err := g.genComparisonJumpTrue(b, trueLabel); err != nil {
		return err
	}

	g.emit("    lda #0")
	g.emit(fmt.Sprintf("    jmp %s", endLabel))
	g.emit(trueLabel + ":")
	g.emit("    lda #1")
	g.emit(endLabel + ":")

	return nil
}

func (g *Generator) genConditionFalseJump(cond ast.Expr, falseLabel string) error {
	b, ok := cond.(*ast.BinaryExpr)
	if !ok {
		if err := g.genExprTo(cond, ast.Type{Name: "byte"}); err != nil {
			return err
		}
		g.emit("    cmp #0")
		g.emit(fmt.Sprintf("    beq %s", falseLabel))
		return nil
	}

	trueLabel := g.newLabel()
	endLabel := g.newLabel()

	if err := g.genComparisonJumpTrue(b, trueLabel); err != nil {
		return err
	}

	g.emit(fmt.Sprintf("    jmp %s", falseLabel))
	g.emit(trueLabel + ":")
	g.emit(endLabel + ":")

	return nil
}

func (g *Generator) genComparisonJumpTrue(b *ast.BinaryExpr, trueLabel string) error {
	switch b.Op {
	case "==", "!=", "<", "<=", ">", ">=":
	default:
		return fmt.Errorf("unsupported comparison operator %q", b.Op)
	}

	if err := g.genExprToIntValue(b.Right); err != nil {
		return err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_tmp_int0+1")

	if err := g.genExprToIntValue(b.Left); err != nil {
		return err
	}

	switch b.Op {
	case "==":
		g.emit("    lda ZP_TMP1")
		g.emit("    cmp peddle_tmp_int0+1")
		g.emit("    bne " + trueLabel + "_skip")
		g.emit("    lda ZP_TMP0")
		g.emit("    cmp peddle_tmp_int0")
		g.emit(fmt.Sprintf("    beq %s", trueLabel))
		g.emit(trueLabel + "_skip:")

	case "!=":
		g.emit("    lda ZP_TMP1")
		g.emit("    cmp peddle_tmp_int0+1")
		g.emit(fmt.Sprintf("    bne %s", trueLabel))
		g.emit("    lda ZP_TMP0")
		g.emit("    cmp peddle_tmp_int0")
		g.emit(fmt.Sprintf("    bne %s", trueLabel))

	case "<":
		return g.genSignedLessThanJump(trueLabel)

	case ">=":
		skip := g.newLabel()
		if err := g.genSignedLessThanJump(skip); err != nil {
			return err
		}
		g.emit(fmt.Sprintf("    jmp %s", trueLabel))
		g.emit(skip + ":")

	case ">":
		return g.genSignedGreaterThanJump(trueLabel)

	case "<=":
		skip := g.newLabel()
		if err := g.genSignedGreaterThanJump(skip); err != nil {
			return err
		}
		g.emit(fmt.Sprintf("    jmp %s", trueLabel))
		g.emit(skip + ":")
	}

	g.usedTmp16 = true
	return nil
}

func (g *Generator) genSignedLessThanJump(trueLabel string) error {
	leftNeg := g.newLabel()
	sameSign := g.newLabel()
	compareLow := g.newLabel()
	done := g.newLabel()

	g.emit("    lda ZP_TMP1")
	g.emit("    bmi " + leftNeg)

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    bmi " + done)
	g.emit("    jmp " + sameSign)

	g.emit(leftNeg + ":")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    bmi " + sameSign)
	g.emit(fmt.Sprintf("    jmp %s", trueLabel))

	g.emit(sameSign + ":")
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    bcc %s", trueLabel))
	g.emit("    bne " + done)

	g.emit(compareLow + ":")
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp peddle_tmp_int0")
	g.emit(fmt.Sprintf("    bcc %s", trueLabel))

	g.emit(done + ":")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genSignedGreaterThanJump(trueLabel string) error {
	leftNeg := g.newLabel()
	sameSign := g.newLabel()
	compareLow := g.newLabel()
	done := g.newLabel()

	g.emit("    lda ZP_TMP1")
	g.emit("    bmi " + leftNeg)

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    bmi " + trueLabel)
	g.emit("    jmp " + sameSign)

	g.emit(leftNeg + ":")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    bmi " + sameSign)
	g.emit("    jmp " + done)

	g.emit(sameSign + ":")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp ZP_TMP1")
	g.emit(fmt.Sprintf("    bcc %s", trueLabel))
	g.emit("    bne " + done)

	g.emit(compareLow + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    cmp ZP_TMP0")
	g.emit(fmt.Sprintf("    bcc %s", trueLabel))

	g.emit(done + ":")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genExprToIntValue(e ast.Expr) error {
	switch expr := e.(type) {
	case *ast.NumberExpr:
		n, err := strconv.Atoi(expr.Value)
		if err != nil {
			return err
		}
		g.emit(fmt.Sprintf("    lda #<%d", n))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda #>%d", n))
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil

	case *ast.IdentExpr:
		sym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Name)
		}

		if sym.Type.Name == "int" && !sym.Type.IsArray {
			g.loadSymbol(sym)
			return nil
		}

		g.loadSymbol(sym)
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
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
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
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

		if fieldType.Name == "int" {
			g.emit(fmt.Sprintf("    lda %s+%d", baseSym.Label, offset))
			g.emit("    sta ZP_TMP0")
			g.emit(fmt.Sprintf("    lda %s+%d", baseSym.Label, offset+1))
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit(fmt.Sprintf("    lda %s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
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
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil

	case *ast.BinaryExpr:
		if err := g.genBinaryTo(expr, ast.Type{Name: "int"}); err != nil {
			return err
		}
		g.usedTmp16 = true
		return nil

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
			if retType.Name == "int" {
				return nil
			}

			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		if retType.Name == "int" {
			g.loadSymbol(*fnFrame.Return)
			return nil
		}

		g.loadSymbol(*fnFrame.Return)
		g.emit("    sta ZP_TMP0")
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil

	default:
		return fmt.Errorf("unsupported int comparison expression")
	}
}
