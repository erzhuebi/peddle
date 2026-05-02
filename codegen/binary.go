package codegen

import (
	"fmt"

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
	if err := g.genExprTo(b.Right, ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta ZP_TMP0")

	if err := g.genExprTo(b.Left, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	trueLabel := g.newLabel()
	endLabel := g.newLabel()

	g.emit("    cmp ZP_TMP0")

	switch b.Op {
	case "==":
		g.emit(fmt.Sprintf("    beq %s", trueLabel))
	case "!=":
		g.emit(fmt.Sprintf("    bne %s", trueLabel))
	case "<":
		g.emit(fmt.Sprintf("    bcc %s", trueLabel))
	case "<=":
		g.emit(fmt.Sprintf("    bcc %s", trueLabel))
		g.emit(fmt.Sprintf("    beq %s", trueLabel))
	case ">":
		g.emit(fmt.Sprintf("    beq %s", endLabel))
		g.emit(fmt.Sprintf("    bcs %s", trueLabel))
	case ">=":
		g.emit(fmt.Sprintf("    bcs %s", trueLabel))
	default:
		return fmt.Errorf("unsupported comparison operator %q", b.Op)
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

	if err := g.genExprTo(b.Right, ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta ZP_TMP0")

	if err := g.genExprTo(b.Left, ast.Type{Name: "byte"}); err != nil {
		return err
	}

	g.emit("    cmp ZP_TMP0")

	switch b.Op {
	case "==":
		g.emit(fmt.Sprintf("    bne %s", falseLabel))
	case "!=":
		g.emit(fmt.Sprintf("    beq %s", falseLabel))
	case "<":
		g.emit(fmt.Sprintf("    bcs %s", falseLabel))
	case ">=":
		g.emit(fmt.Sprintf("    bcc %s", falseLabel))
	case "<=":
		after := g.newLabel()
		g.emit(fmt.Sprintf("    bcc %s", after))
		g.emit(fmt.Sprintf("    beq %s", after))
		g.emit(fmt.Sprintf("    jmp %s", falseLabel))
		g.emit(after + ":")
	case ">":
		g.emit(fmt.Sprintf("    beq %s", falseLabel))
		g.emit(fmt.Sprintf("    bcc %s", falseLabel))
	default:
		return fmt.Errorf("unsupported condition operator %q", b.Op)
	}

	return nil
}
