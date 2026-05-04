package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genStmt(s ast.Stmt) error {
	switch stmt := s.(type) {
	case *ast.AssignStmt:
		return g.genAssign(stmt)
	case *ast.CallStmt:
		return g.genCallStmt(stmt)
	case *ast.WhileStmt:
		return g.genWhile(stmt)
	case *ast.IfStmt:
		return g.genIf(stmt)
	case *ast.ReturnStmt:
		return g.genReturn(stmt)
	default:
		return fmt.Errorf("unsupported statement")
	}
}

func (g *Generator) genAssign(a *ast.AssignStmt) error {
	switch target := a.Target.(type) {
	case *ast.VarLValue:
		sym, ok := g.resolve(target.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", target.Name)
		}

		if str, ok := a.Value.(*ast.StringExpr); ok {
			if !sym.Type.IsArray || sym.Type.Name != "char" {
				return fmt.Errorf("cannot assign string to %s", sym.Type.String())
			}

			return g.genCopyStringLiteralToCharArray(&ast.IdentExpr{Name: target.Name}, str.Value)
		}

		if err := g.genExprTo(a.Value, sym.Type); err != nil {
			return err
		}

		g.storeAIntoSymbol(sym)
		return nil

	case *ast.IndexLValue:
		arraySym, ok := g.resolve(target.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", target.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", target.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		if err := g.genExprTo(a.Value, elemType); err != nil {
			return err
		}

		if elemType.Name == "int" {
			g.emit("    lda ZP_TMP0")
			g.emit("    sta peddle_tmp_int0")
			g.emit("    lda ZP_TMP1")
			g.emit("    sta peddle_tmp_int0+1")

			if err := g.genUpdateArrayLenFromIndex(arraySym, target.Index); err != nil {
				return err
			}

			if err := g.genArrayIndexToY(arraySym, target.Index); err != nil {
				return err
			}

			g.emit("    lda peddle_tmp_int0")
			g.emit("    sta (ZP_PTR0_LO), y")
			g.emit("    iny")
			g.emit("    lda peddle_tmp_int0+1")
			g.emit("    sta (ZP_PTR0_LO), y")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    sta peddle_tmp_int0")
		g.usedTmp16 = true

		if err := g.genUpdateArrayLenFromIndex(arraySym, target.Index); err != nil {
			return err
		}

		if err := g.genArrayIndexToY(arraySym, target.Index); err != nil {
			return err
		}

		g.emit("    lda peddle_tmp_int0")
		g.emit("    sta (ZP_PTR0_LO), y")
		return nil

	case *ast.IndexFieldLValue:
		arraySym, ok := g.resolve(target.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", target.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", target.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, offset, err := g.fieldInfo(elemType, target.Field)
		if err != nil {
			return err
		}

		if fieldType.IsArray {
			str, ok := a.Value.(*ast.StringExpr)
			if !ok || fieldType.Name != "char" {
				return fmt.Errorf("array field assignment is not implemented yet")
			}

			if err := g.genUpdateArrayLenFromIndex(arraySym, target.Index); err != nil {
				return err
			}

			return g.genCopyStringLiteralToCharArray(&ast.IndexFieldExpr{
				Name:  target.Name,
				Index: target.Index,
				Field: target.Field,
			}, str.Value)
		}

		if _, ok := g.structs[fieldType.Name]; ok {
			return fmt.Errorf("struct field assignment is not implemented yet")
		}

		if err := g.genUpdateArrayLenFromIndex(arraySym, target.Index); err != nil {
			return err
		}

		if err := g.genArrayIndexToY(arraySym, target.Index); err != nil {
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

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    sta ZP_PTR1_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    sta ZP_PTR1_HI")

		if err := g.genExprTo(a.Value, fieldType); err != nil {
			return err
		}

		if fieldType.Name == "int" {
			g.emit("    lda ZP_PTR1_LO")
			g.emit("    sta ZP_PTR0_LO")
			g.emit("    lda ZP_PTR1_HI")
			g.emit("    sta ZP_PTR0_HI")

			g.emit("    lda ZP_TMP0")
			g.emit("    ldy #0")
			g.emit("    sta (ZP_PTR0_LO), y")
			g.emit("    lda ZP_TMP1")
			g.emit("    iny")
			g.emit("    sta (ZP_PTR0_LO), y")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    sta peddle_tmp_int0")
		g.usedTmp16 = true

		g.emit("    lda ZP_PTR1_LO")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR1_HI")
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    lda peddle_tmp_int0")
		g.emit("    ldy #0")
		g.emit("    sta (ZP_PTR0_LO), y")
		return nil

	case *ast.FieldLValue:
		baseSym, ok := g.resolve(target.Base)
		if !ok {
			return fmt.Errorf("unknown variable %q", target.Base)
		}

		fieldType, offset, err := g.fieldInfo(baseSym.Type, target.Field)
		if err != nil {
			return err
		}

		if fieldType.IsArray {
			str, ok := a.Value.(*ast.StringExpr)
			if !ok || fieldType.Name != "char" {
				return fmt.Errorf("array field assignment is not implemented yet")
			}

			return g.genCopyStringLiteralToCharArray(&ast.FieldExpr{
				Base:  target.Base,
				Field: target.Field,
			}, str.Value)
		}

		if _, ok := g.structs[fieldType.Name]; ok {
			return fmt.Errorf("struct field assignment is not implemented yet")
		}

		if err := g.genExprTo(a.Value, fieldType); err != nil {
			return err
		}

		return g.storeIntoField(baseSym, fieldType, offset)

	default:
		return fmt.Errorf("unsupported assignment target")
	}
}

func (g *Generator) genCallStmt(c *ast.CallStmt) error {
	_, err := g.genCall(c.Name, c.Args)
	return err
}

func (g *Generator) genWhile(w *ast.WhileStmt) error {
	start := g.newLabel()
	end := g.newLabel()

	g.emit(start + ":")

	if err := g.genConditionFalseJump(w.Cond, end); err != nil {
		return err
	}

	for _, stmt := range w.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit(fmt.Sprintf("    jmp %s", start))
	g.emit(end + ":")
	return nil
}

func (g *Generator) genIf(i *ast.IfStmt) error {
	elseLabel := g.newLabel()
	endLabel := g.newLabel()

	if err := g.genConditionFalseJump(i.Cond, elseLabel); err != nil {
		return err
	}

	for _, stmt := range i.Then {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit(fmt.Sprintf("    jmp %s", endLabel))
	g.emit(elseLabel + ":")

	for _, stmt := range i.Else {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit(endLabel + ":")
	return nil
}

func (g *Generator) genReturn(r *ast.ReturnStmt) error {
	if g.currentFrame.Return == nil {
		g.emit("    rts")
		return nil
	}

	if r.Value == nil {
		return fmt.Errorf("missing return value")
	}

	if err := g.genExprTo(r.Value, g.currentFrame.Return.Type); err != nil {
		return err
	}

	g.storeAIntoSymbol(*g.currentFrame.Return)
	g.emit("    rts")
	return nil
}
