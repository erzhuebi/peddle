package codegen

import (
	"fmt"
	"strconv"

	"peddle/ast"
)

func (g *Generator) foldConstExpr(e ast.Expr, target ast.Type) (int, bool, error) {
	switch expr := e.(type) {
	case *ast.NumberExpr:
		n, err := strconv.Atoi(expr.Value)
		if err != nil {
			return 0, false, err
		}
		return normalizeConstValue(n, target), true, nil

	case *ast.CharExpr:
		n, err := strconv.Atoi(expr.Value)
		if err != nil {
			return 0, false, err
		}
		return normalizeConstValue(n, target), true, nil

	case *ast.BoolExpr:
		if expr.Value {
			return normalizeConstValue(1, target), true, nil
		}
		return normalizeConstValue(0, target), true, nil

	case *ast.IdentExpr:
		n, ok := g.constants[expr.Name]
		if !ok {
			return 0, false, nil
		}
		return normalizeConstValue(n, target), true, nil

	case *ast.UnaryExpr:
		return g.foldConstUnary(expr, target)

	case *ast.BinaryExpr:
		return g.foldConstBinary(expr, target)

	default:
		return 0, false, nil
	}
}

func (g *Generator) foldConstUnary(expr *ast.UnaryExpr, target ast.Type) (int, bool, error) {
	switch expr.Op {
	case "-":
		v, ok, err := g.foldConstExpr(expr.Expr, target)
		if err != nil || !ok {
			return 0, ok, err
		}
		return normalizeConstValue(-v, target), true, nil

	case "!":
		v, ok, err := g.foldConstExpr(expr.Expr, ast.Type{Name: "byte"})
		if err != nil || !ok {
			return 0, ok, err
		}
		if v == 0 {
			return normalizeConstValue(1, target), true, nil
		}
		return normalizeConstValue(0, target), true, nil

	default:
		return 0, false, fmt.Errorf("unsupported unary operator %q", expr.Op)
	}
}

func (g *Generator) foldConstBinary(expr *ast.BinaryExpr, target ast.Type) (int, bool, error) {
	if isComparisonOp(expr.Op) {
		left, ok, err := g.foldConstExpr(expr.Left, ast.Type{Name: "int"})
		if err != nil || !ok {
			return 0, ok, err
		}

		right, ok, err := g.foldConstExpr(expr.Right, ast.Type{Name: "int"})
		if err != nil || !ok {
			return 0, ok, err
		}

		if evalConstComparison(expr.Op, left, right) {
			return normalizeConstValue(1, target), true, nil
		}
		return normalizeConstValue(0, target), true, nil
	}

	left, ok, err := g.foldConstExpr(expr.Left, target)
	if err != nil || !ok {
		return 0, ok, err
	}

	rightTarget := target
	if expr.Op == "<<" || expr.Op == ">>" {
		rightTarget = ast.Type{Name: "int"}
	}

	right, ok, err := g.foldConstExpr(expr.Right, rightTarget)
	if err != nil || !ok {
		return 0, ok, err
	}

	v, err := evalConstBinary(expr.Op, left, right, target)
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}

func evalConstComparison(op string, left int, right int) bool {
	switch op {
	case "==":
		return left == right
	case "!=":
		return left != right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

func evalConstBinary(op string, left int, right int, target ast.Type) (int, error) {
	width := constWidth(target)
	leftUnsigned := unsignedConstValue(left, width)
	rightUnsigned := unsignedConstValue(right, width)

	var result int

	switch op {
	case "+":
		result = leftUnsigned + rightUnsigned
	case "-":
		result = leftUnsigned - rightUnsigned
	case "*":
		result = leftUnsigned * rightUnsigned
	case "/":
		if rightUnsigned == 0 {
			result = 0
		} else {
			result = leftUnsigned / rightUnsigned
		}
	case "%":
		if rightUnsigned == 0 {
			result = leftUnsigned
		} else {
			result = leftUnsigned % rightUnsigned
		}
	case "&":
		result = leftUnsigned & rightUnsigned
	case "|":
		result = leftUnsigned | rightUnsigned
	case "^":
		result = leftUnsigned ^ rightUnsigned
	case "<<":
		count := clampShiftCount(right, width)
		result = leftUnsigned << count
	case ">>":
		count := clampShiftCount(right, width)
		result = leftUnsigned >> count
	default:
		return 0, fmt.Errorf("unsupported binary op %q", op)
	}

	return normalizeConstValue(result, target), nil
}

func (g *Generator) emitConstExprTo(value int, target ast.Type) {
	if target.Name == "int" {
		g.emit(fmt.Sprintf("    lda #<%d", value))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda #>%d", value))
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return
	}

	g.emit(fmt.Sprintf("    lda #%d", value&0xff))
}

func (g *Generator) constShiftCount(e ast.Expr, width int) (int, bool, error) {
	v, ok, err := g.foldConstExpr(e, ast.Type{Name: "int"})
	if err != nil || !ok {
		return 0, ok, err
	}
	return clampShiftCount(v, width), true, nil
}

func normalizeConstValue(v int, target ast.Type) int {
	if target.Name == "int" {
		return sign16(v)
	}
	return v & 0xff
}

func constWidth(target ast.Type) int {
	if target.Name == "int" {
		return 16
	}
	return 8
}

func unsignedConstValue(v int, width int) int {
	if width == 16 {
		return v & 0xffff
	}
	return v & 0xff
}

func sign16(v int) int {
	v &= 0xffff
	if v&0x8000 != 0 {
		return v - 0x10000
	}
	return v
}

func clampShiftCount(count int, width int) int {
	if count < 0 {
		return 0
	}
	if count > width {
		return width
	}
	return count
}
