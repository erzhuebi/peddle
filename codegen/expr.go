package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genExprTo(e ast.Expr, target ast.Type) error {
	if n, ok, err := g.foldConstExpr(e, target); err != nil {
		return err
	} else if ok {
		g.emitConstExprTo(n, target)
		return nil
	}

	switch expr := e.(type) {
	case *ast.NumberExpr:
		return fmt.Errorf("unfoldable number literal %q", expr.Value)

	case *ast.CharExpr:
		return fmt.Errorf("unfoldable char literal %q", expr.Value)

	case *ast.BoolExpr:
		return fmt.Errorf("unfoldable bool literal %t", expr.Value)

	case *ast.IdentExpr:
		sym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Name)
		}

		if sym.Type.IsPointer {
			if isScalarPointerType(sym.Type) {
				pointee := pointerPointeeType(sym.Type)
				if err := g.loadScalarPointer(sym); err != nil {
					return err
				}
				if !isWordType(target) && isWordType(pointee) {
					g.emit("    lda ZP_TMP0")
					return nil
				}
				if isWordType(target) && !isWordType(pointee) {
					g.emit("    sta ZP_TMP0")
					g.emit("    lda #0")
					g.emit("    sta ZP_TMP1")
					g.usedTmp16 = true
				}
				return nil
			}
			if target.IsPointer {
				g.loadSymbol(sym)
				return nil
			}
			return fmt.Errorf("pointer value cannot be used as %s", target.String())
		}
		if target.IsPointer {
			return fmt.Errorf("pointer value requires address-of expression")
		}

		if !isWordType(target) && isWordType(sym.Type) {
			g.emit(fmt.Sprintf("    lda %s", sym.Label))
			return nil
		}

		if isWordType(target) && !isWordType(sym.Type) && !sym.Type.IsArray {
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
		if target.IsPointer {
			return fmt.Errorf("pointer value requires address-of expression")
		}

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

		if isWordType(ast.Type{Name: arraySym.Type.Name}) {
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP0")
			g.emit("    iny")
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    lda (ZP_PTR0_LO), y")
		if isWordType(target) {
			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
		}
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

		if isWordType(fieldType) {
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP0")
			g.emit("    iny")
			g.emit("    lda (ZP_PTR0_LO), y")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		g.emit("    lda (ZP_PTR0_LO), y")
		if isWordType(target) {
			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
		}
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

		if baseSym.Type.IsPointer {
			if err := g.loadPointerField(baseSym, fieldType, offset); err != nil {
				return err
			}
			if isWordType(target) && !isWordType(fieldType) {
				g.emit("    sta ZP_TMP0")
				g.emit("    lda #0")
				g.emit("    sta ZP_TMP1")
				g.usedTmp16 = true
			}
			return nil
		}

		if fieldType.IsArray {
			return fmt.Errorf("array field reads are not implemented yet")
		}

		if _, ok := g.structs[fieldType.Name]; ok {
			return fmt.Errorf("struct field reads are not implemented yet")
		}

		if err := g.loadField(baseSym, fieldType, offset); err != nil {
			return err
		}
		if isWordType(target) && !isWordType(fieldType) {
			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
		}
		return nil

	case *ast.UnaryExpr:
		switch expr.Op {
		case "&":
			return g.genAddressOfTo(expr.Expr, target)

		case "-":
			if isWordType(target) {
				if err := g.genExprTo(expr.Expr, target); err != nil {
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
		if fnFrame := g.frames[expr.Name]; fnFrame != nil && len(fnFrame.Returns) > 1 {
			return fmt.Errorf("function %s returns multiple values", expr.Name)
		}

		retType, err := g.genCall(expr.Name, expr.Args)
		if err != nil {
			return err
		}

		if retType.Name == "" {
			return fmt.Errorf("function %s does not return a value", expr.Name)
		}

		fnFrame := g.frames[expr.Name]
		if fnFrame == nil || fnFrame.Return == nil {
			if isWordType(target) && !isWordType(retType) {
				g.emit("    sta ZP_TMP0")
				g.emit("    lda #0")
				g.emit("    sta ZP_TMP1")
				g.usedTmp16 = true
			}
			return nil
		}

		g.loadSymbol(*fnFrame.Return)
		if isWordType(target) && !isWordType(retType) {
			g.emit("    sta ZP_TMP0")
			g.emit("    lda #0")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
		}
		return nil

	case *ast.StringExpr:
		return fmt.Errorf("string expression cannot be used here directly")

	default:
		return fmt.Errorf("unsupported expression")
	}
}

func (g *Generator) genAddressOfTo(e ast.Expr, target ast.Type) error {
	if !target.IsPointer && target.Name != "uint" {
		return fmt.Errorf("address-of expression cannot be used as %s", target.String())
	}

	switch expr := e.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(expr.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", expr.Name)
		}
		if target.Name == "uint" {
			if sym.Type.IsPointer {
				return fmt.Errorf("cannot take address of pointer parameter")
			}
			if sym.Type.IsArray && sym.IsReference {
				g.loadSymbol(sym)
				return nil
			}

			g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
			g.emit("    sta ZP_TMP0")
			g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}

		if sym.Type.IsPointer || sym.Type.IsArray {
			return fmt.Errorf("cannot take address of %s as %s", sym.Type.String(), target.String())
		}
		if sym.Type.Name != target.Name {
			return fmt.Errorf("cannot take address of %s as %s", sym.Type.String(), target.String())
		}
		if _, ok := g.structs[sym.Type.Name]; !ok && !isScalarTypeName(sym.Type.Name) {
			return fmt.Errorf("cannot take address of non-struct %q", expr.Name)
		}

		g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
		g.emit("    sta ZP_TMP0")
		g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
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
		if target.Name == "uint" {
			if err := g.genUpdateArrayLenFromIndex(arraySym, expr.Index); err != nil {
				return err
			}
			if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
				return err
			}

			g.emit("    lda ZP_PTR0_LO")
			g.emit("    sta ZP_TMP0")
			g.emit("    lda ZP_PTR0_HI")
			g.emit("    sta ZP_TMP1")
			g.usedTmp16 = true
			return nil
		}
		if arraySym.Type.Name != target.Name {
			return fmt.Errorf("cannot take address of %s[] element as %s", arraySym.Type.Name, target.String())
		}
		if _, ok := g.structs[arraySym.Type.Name]; !ok && !isScalarTypeName(arraySym.Type.Name) {
			return fmt.Errorf("cannot take address of non-struct array element %q", expr.Name)
		}

		if err := g.genUpdateArrayLenFromIndex(arraySym, expr.Index); err != nil {
			return err
		}
		if err := g.genArrayIndexToY(arraySym, expr.Index); err != nil {
			return err
		}

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    sta ZP_TMP1")
		g.usedTmp16 = true
		return nil

	default:
		return fmt.Errorf("can only take address of variables or array elements")
	}
}
