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
		sym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Name)
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

		if arraySym.Type.Name == "int" {
			return fmt.Errorf("int array reads are not implemented yet")
		}

		if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
			return err
		}

		g.emit("    lda (ZP_PTR0_LO), y")
		return nil

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
			return fmt.Errorf("missing return slot for %s", expr.Name)
		}

		g.loadSymbol(*fnFrame.Return)
		return nil

	case *ast.StringExpr:
		return fmt.Errorf("string expression cannot be used here directly")

	default:
		return fmt.Errorf("unsupported expression")
	}
}
