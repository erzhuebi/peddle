package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genBinaryTo(b *ast.BinaryExpr, target ast.Type) error {
	switch b.Op {
	case "+", "-", "*", "/", "%", "&", "|", "^", "<<", ">>":
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
	if b.Op == "<<" || b.Op == ">>" {
		if count, ok, err := g.constShiftCount(b.Right, 8); err != nil {
			return err
		} else if ok {
			if err := g.genExprTo(b.Left, ast.Type{Name: "byte"}); err != nil {
				return err
			}
			g.emitConstShiftByte(b.Op, count)
			return nil
		}
	}

	if err := g.genExprTo(b.Right, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	// Preserve the right-hand byte while generating the left side.
	//
	// Nested left expressions also use ZP_TMP0/ZP_TMP1 internally. Without
	// preserving the right side, expressions like:
	//
	//     (a + 1) & 4
	//     (a + 1) << shift
	//
	// can accidentally use the nested expression's temporary value instead of
	// the outer right-hand operand.
	g.emit("    pha")

	if err := g.genExprTo(b.Left, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	// Save left in ZP_TMP1, restore right into ZP_TMP0, and put left back in A.
	g.emit("    sta ZP_TMP1")
	g.emit("    pla")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda ZP_TMP1")

	switch b.Op {
	case "+":
		g.emit("    clc")
		g.emit("    adc ZP_TMP0")

	case "-":
		g.emit("    sec")
		g.emit("    sbc ZP_TMP0")

	case "&":
		g.emit("    and ZP_TMP0")

	case "|":
		g.emit("    ora ZP_TMP0")

	case "^":
		g.emit("    eor ZP_TMP0")

	case "*":
		if g.options.OptMode == OptModeSize {
			g.emit("    ldx ZP_TMP0")
			g.emit("    stx peddle_tmp_int0")
			g.emit("    jsr peddle_mul_byte")
			g.usedMulByteRuntime = true
			g.usedTmp16 = true
			return nil
		}

		g.emit("    sta ZP_TMP1")
		g.emit("    lda #0")
		g.emit("    tax")

		loop := g.newLabel()
		done := g.newLabel()

		g.emit(loop + ":")
		g.emit("    lda ZP_TMP0")
		g.emit(fmt.Sprintf("    beq %s", done))
		g.emit("    txa")
		g.emit("    clc")
		g.emit("    adc ZP_TMP1")
		g.emit("    tax")
		g.emit("    dec ZP_TMP0")
		g.emit(fmt.Sprintf("    jmp %s", loop))
		g.emit(done + ":")
		g.emit("    txa")

	case "/", "%":
		if g.options.OptMode == OptModeSize {
			g.emit("    ldx ZP_TMP0")
			g.emit("    stx peddle_tmp_int0")
			g.emit("    jsr peddle_divmod_byte")
			if b.Op == "%" {
				g.emit("    lda ZP_TMP1")
			}
			g.usedDivModByteRuntime = true
			g.usedTmp16 = true
			return nil
		}

		g.emitInlineDivModByte(b.Op)

	case "<<", ">>":
		if g.options.OptMode == OptModeSize {
			g.emit("    ldx ZP_TMP0")
			g.emit("    stx peddle_tmp_int0")

			if b.Op == "<<" {
				g.emit("    jsr peddle_shl_byte")
				g.usedShlByteRuntime = true
			} else {
				g.emit("    jsr peddle_shr_byte")
				g.usedShrByteRuntime = true
			}

			g.usedTmp16 = true
			return nil
		}

		g.emitVariableShiftByte(b.Op)
	}

	return nil
}

func (g *Generator) genBinaryInt(b *ast.BinaryExpr) error {
	if b.Op == "<<" || b.Op == ">>" {
		if count, ok, err := g.constShiftCount(b.Right, 16); err != nil {
			return err
		} else if ok {
			if err := g.genExprTo(b.Left, ast.Type{Name: "int"}); err != nil {
				return err
			}
			g.emitConstShiftInt(b.Op, count)
			g.usedTmp16 = true
			return nil
		}
	}

	if err := g.genExprTo(b.Right, ast.Type{Name: "int"}); err != nil {
		return err
	}

	// Preserve the right-hand int while generating the left side.
	//
	// Nested left expressions also use peddle_tmp_int0 and ZP_TMP0/ZP_TMP1.
	// Without preserving the right side, expressions like:
	//
	//     (a + 1) & 4
	//
	// can overwrite the outer right-hand operand while generating (a + 1).
	g.emit("    lda ZP_TMP0")
	g.emit("    pha")
	g.emit("    lda ZP_TMP1")
	g.emit("    pha")

	if err := g.genExprTo(b.Left, ast.Type{Name: "int"}); err != nil {
		return err
	}

	// Save left temporarily, restore right into peddle_tmp_int0, then restore
	// left into ZP_TMP0/ZP_TMP1 for the operation below.
	g.emit("    lda ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta ZP_TMP1")

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

	case "&":
		g.emit("    lda ZP_TMP0")
		g.emit("    and peddle_tmp_int0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    and peddle_tmp_int0+1")
		g.emit("    sta ZP_TMP1")

	case "|":
		g.emit("    lda ZP_TMP0")
		g.emit("    ora peddle_tmp_int0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    ora peddle_tmp_int0+1")
		g.emit("    sta ZP_TMP1")

	case "^":
		g.emit("    lda ZP_TMP0")
		g.emit("    eor peddle_tmp_int0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    eor peddle_tmp_int0+1")
		g.emit("    sta ZP_TMP1")

	case "*":
		if g.options.OptMode == OptModeSize {
			g.emit("    jsr peddle_mul_int")
			g.usedMulIntRuntime = true
			g.usedTmp16 = true
			return nil
		}

		return g.genInlineMulInt()

	case "/", "%":
		if g.options.OptMode == OptModeSize {
			g.emit("    jsr peddle_divmod_int")
			if b.Op == "%" {
				g.emit("    lda ZP_PTR0_LO")
				g.emit("    sta ZP_TMP0")
				g.emit("    lda ZP_PTR0_HI")
				g.emit("    sta ZP_TMP1")
			}
			g.usedDivModIntRuntime = true
			g.usedTmp16 = true
			return nil
		}

		g.emitInlineDivModInt(b.Op)

	case "<<", ">>":
		if g.options.OptMode == OptModeSize {
			if b.Op == "<<" {
				g.emit("    jsr peddle_shl_int")
				g.usedShlIntRuntime = true
			} else {
				g.emit("    jsr peddle_shr_int")
				g.usedShrIntRuntime = true
			}

			g.usedTmp16 = true
			return nil
		}

		g.emitVariableShiftInt(b.Op)
	}

	g.usedTmp16 = true
	return nil
}

func (g *Generator) emitConstShiftByte(op string, count int) {
	if count > 8 {
		count = 8
	}

	for i := 0; i < count; i++ {
		if op == "<<" {
			g.emit("    asl")
		} else {
			g.emit("    lsr")
		}
	}
}

func (g *Generator) emitConstShiftInt(op string, count int) {
	if count > 16 {
		count = 16
	}

	for i := 0; i < count; i++ {
		if op == "<<" {
			g.emit("    asl ZP_TMP0")
			g.emit("    rol ZP_TMP1")
		} else {
			g.emit("    lsr ZP_TMP1")
			g.emit("    ror ZP_TMP0")
		}
	}
}

func (g *Generator) emitVariableShiftByte(op string) {
	loop := g.newLabel()
	done := g.newLabel()

	g.emit("    sta ZP_TMP1")
	g.emit("    ldx ZP_TMP0")

	g.emit(loop + ":")
	g.emit(fmt.Sprintf("    beq %s", done))

	if op == "<<" {
		g.emit("    asl ZP_TMP1")
	} else {
		g.emit("    lsr ZP_TMP1")
	}

	g.emit("    dex")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.emit("    lda ZP_TMP1")
}

func (g *Generator) emitVariableShiftInt(op string) {
	loop := g.newLabel()
	done := g.newLabel()

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit(fmt.Sprintf("    beq %s", done))

	if op == "<<" {
		g.emit("    asl ZP_TMP0")
		g.emit("    rol ZP_TMP1")
	} else {
		g.emit("    lsr ZP_TMP1")
		g.emit("    ror ZP_TMP0")
	}

	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.usedTmp16 = true
}

func (g *Generator) genInlineMulInt() error {
	loop := g.newLabel()
	skipAdd := g.newLabel()
	done := g.newLabel()

	g.emit("    lda ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda #0")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    sta ZP_PTR1_HI")

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))

	g.emit("    lda peddle_tmp_int0")
	g.emit("    and #1")
	g.emit(fmt.Sprintf("    beq %s", skipAdd))

	g.emit("    clc")
	g.emit("    lda ZP_PTR1_LO")
	g.emit("    adc ZP_PTR0_LO")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR1_HI")
	g.emit("    adc ZP_PTR0_HI")
	g.emit("    sta ZP_PTR1_HI")

	g.emit(skipAdd + ":")
	g.emit("    asl ZP_PTR0_LO")
	g.emit("    rol ZP_PTR0_HI")

	g.emit("    lsr peddle_tmp_int0+1")
	g.emit("    ror peddle_tmp_int0")

	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.emit("    lda ZP_PTR1_LO")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda ZP_PTR1_HI")
	g.emit("    sta ZP_TMP1")

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
	if !ok || !isComparisonOp(b.Op) {
		if err := g.genExprTo(cond, ast.Type{Name: "byte"}); err != nil {
			return err
		}

		g.emit("    cmp #0")
		// The false label can be after an arbitrary user-controlled statement
		// body, so this conditional exit must use a long branch sequence.
		if err := g.emitLongBranch("beq", falseLabel); err != nil {
			return err
		}

		return nil
	}

	trueLabel := g.newLabel()

	// Comparison helpers only branch to trueLabel, which is emitted a few
	// instructions below. The far path is the false case, handled by an
	// absolute jmp so large if/while/for bodies cannot exceed 6502 branch
	// range.
	if err := g.genComparisonJumpTrue(b, trueLabel); err != nil {
		return err
	}

	g.emit(fmt.Sprintf("    jmp %s", falseLabel))
	g.emit(trueLabel + ":")

	return nil
}

func isComparisonOp(op string) bool {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
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

	// Preserve the right-hand comparison value while generating the left side.
	//
	// genExprToIntValue() and nested binary expressions use peddle_tmp_int0
	// internally. If we stored the right side there before generating the left
	// side, an expression like:
	//
	//     (j & 4) == 0
	//
	// would overwrite the saved 0 while generating (j & 4), because bitwise
	// code also uses peddle_tmp_int0. The hardware stack is safe here because
	// calls made during left-side generation balance their own JSR/RTS use.
	g.emit("    lda ZP_TMP0")
	g.emit("    pha")
	g.emit("    lda ZP_TMP1")
	g.emit("    pha")

	if err := g.genExprToIntValue(b.Left); err != nil {
		return err
	}

	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")

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
	if n, ok, err := g.foldConstExpr(e, ast.Type{Name: "int"}); err != nil {
		return err
	} else if ok {
		g.emitConstExprTo(n, ast.Type{Name: "int"})
		return nil
	}

	switch expr := e.(type) {
	case *ast.NumberExpr:
		return fmt.Errorf("unfoldable number literal %q", expr.Value)

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

	case *ast.BinaryExpr:
		if err := g.genBinaryTo(expr, ast.Type{Name: "int"}); err != nil {
			return err
		}
		g.usedTmp16 = true
		return nil

	default:
		return g.genExprTo(e, ast.Type{Name: "int"})
	}
}

func (g *Generator) emitInlineDivModByte(op string) {
	divZero := g.newLabel()
	loop := g.newLabel()
	done := g.newLabel()

	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.emit(fmt.Sprintf("    beq %s", divZero))
	g.emit("    ldx #0")

	g.emit(loop + ":")
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp ZP_TMP0")
	g.emit(fmt.Sprintf("    bcc %s", done))
	g.emit("    sec")
	g.emit("    sbc ZP_TMP0")
	g.emit("    sta ZP_TMP1")
	g.emit("    inx")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(divZero + ":")
	if op == "/" {
		g.emit("    lda #0")
	} else {
		g.emit("    lda ZP_TMP1")
	}
	g.emit(fmt.Sprintf("    jmp %s", done+"_return"))

	g.emit(done + ":")
	if op == "/" {
		g.emit("    txa")
	} else {
		g.emit("    lda ZP_TMP1")
	}

	g.emit(done + "_return:")
}

func (g *Generator) emitInlineDivModInt(op string) {
	divZero := g.newLabel()
	loop := g.newLabel()
	subtract := g.newLabel()
	done := g.newLabel()

	g.emit("    lda ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", divZero))

	g.emit("    lda #0")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    sta ZP_PTR1_HI")

	g.emit(loop + ":")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    cmp peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    bcc %s", done))
	g.emit(fmt.Sprintf("    bne %s", subtract))
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    cmp peddle_tmp_int0")
	g.emit(fmt.Sprintf("    bcc %s", done))

	g.emit(subtract + ":")
	g.emit("    sec")
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sbc peddle_tmp_int0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sbc peddle_tmp_int0+1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    inc ZP_PTR1_LO")
	g.emit(fmt.Sprintf("    bne %s", loop))
	g.emit("    inc ZP_PTR1_HI")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(divZero + ":")
	if op == "/" {
		g.emit("    lda #0")
		g.emit("    sta ZP_TMP0")
		g.emit("    sta ZP_TMP1")
	} else {
		g.emit("    lda ZP_PTR0_LO")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    sta ZP_TMP1")
	}
	g.emit(fmt.Sprintf("    jmp %s", done+"_return"))

	g.emit(done + ":")
	if op == "/" {
		g.emit("    lda ZP_PTR1_LO")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_PTR1_HI")
		g.emit("    sta ZP_TMP1")
	} else {
		g.emit("    lda ZP_PTR0_LO")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    sta ZP_TMP1")
	}

	g.emit(done + "_return:")
	g.usedTmp16 = true
}
