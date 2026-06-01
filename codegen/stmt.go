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
	case *ast.ForStmt:
		return g.genFor(stmt)
	case *ast.IfStmt:
		return g.genIf(stmt)
	case *ast.ReturnStmt:
		return g.genReturn(stmt)
	case *ast.BreakStmt:
		return g.genBreak(stmt)
	case *ast.ContinueStmt:
		return g.genContinue(stmt)
	default:
		return fmt.Errorf("unsupported statement")
	}
}

type loopLabels struct {
	continueLabel string
	breakLabel    string
}

var generatorLoopLabels = map[*Generator][]loopLabels{}

func (g *Generator) pushLoopLabels(continueLabel string, breakLabel string) {
	generatorLoopLabels[g] = append(generatorLoopLabels[g], loopLabels{
		continueLabel: continueLabel,
		breakLabel:    breakLabel,
	})
}

func (g *Generator) popLoopLabels() {
	stack := generatorLoopLabels[g]
	if len(stack) == 0 {
		return
	}

	stack = stack[:len(stack)-1]
	if len(stack) == 0 {
		delete(generatorLoopLabels, g)
		return
	}

	generatorLoopLabels[g] = stack
}

func (g *Generator) currentLoopLabels() (loopLabels, bool) {
	stack := generatorLoopLabels[g]
	if len(stack) == 0 {
		return loopLabels{}, false
	}

	return stack[len(stack)-1], true
}

func (g *Generator) genBreak(_ *ast.BreakStmt) error {
	labels, ok := g.currentLoopLabels()
	if !ok {
		return fmt.Errorf("break outside loop")
	}

	g.emit(fmt.Sprintf("    jmp %s", labels.breakLabel))
	return nil
}

func (g *Generator) genContinue(_ *ast.ContinueStmt) error {
	labels, ok := g.currentLoopLabels()
	if !ok {
		return fmt.Errorf("continue outside loop")
	}

	g.emit(fmt.Sprintf("    jmp %s", labels.continueLabel))
	return nil
}

func (g *Generator) genAssign(a *ast.AssignStmt) error {
	if len(a.Targets) > 0 {
		return g.genMultiAssign(a)
	}

	switch target := a.Target.(type) {
	case *ast.VarLValue:
		sym, ok := g.resolve(target.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", target.Name)
		}

		if isScalarPointerType(sym.Type) {
			if err := g.genExprTo(a.Value, pointerPointeeType(sym.Type)); err != nil {
				return err
			}
			return g.storeIntoScalarPointer(sym)
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

		if isWordType(elemType) {
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

		if isWordType(fieldType) {
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

		if baseSym.Type.IsPointer {
			if err := g.genExprTo(a.Value, fieldType); err != nil {
				return err
			}
			return g.storeIntoPointerField(baseSym, fieldType, offset)
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

func (g *Generator) genMultiAssign(a *ast.AssignStmt) error {
	call, ok := a.Value.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("multi-assignment requires a function call")
	}

	if _, err := g.genCall(call.Name, call.Args); err != nil {
		return err
	}

	fnFrame := g.frames[call.Name]
	if fnFrame == nil {
		return fmt.Errorf("multi-assignment requires a user function call")
	}
	if len(fnFrame.Returns) != len(a.Targets) {
		return fmt.Errorf("multi-assignment has %d targets but %s returns %d values", len(a.Targets), call.Name, len(fnFrame.Returns))
	}

	for i, target := range a.Targets {
		if target == "_" {
			continue
		}

		sym, ok := g.resolve(target)
		if !ok {
			return fmt.Errorf("unknown variable %q", target)
		}

		targetType := sym.Type
		if isScalarPointerType(sym.Type) {
			targetType = pointerPointeeType(sym.Type)
		}

		g.loadSymbolAs(fnFrame.Returns[i], targetType)
		if isScalarPointerType(sym.Type) {
			if err := g.storeIntoScalarPointer(sym); err != nil {
				return err
			}
			continue
		}

		g.storeAIntoSymbol(sym)
	}

	return nil
}

func (g *Generator) genCallStmt(c *ast.CallStmt) error {
	if fnFrame := g.frames[c.Name]; fnFrame != nil && len(fnFrame.Returns) > 1 {
		return fmt.Errorf("function %s returns multiple values", c.Name)
	}

	_, err := g.genCall(c.Name, c.Args)
	return err
}

func (g *Generator) genWhile(w *ast.WhileStmt) error {
	start := g.newLabel()
	end := g.newLabel()

	g.pushLoopLabels(start, end)
	defer g.popLoopLabels()

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

func (g *Generator) genFor(f *ast.ForStmt) error {
	if f.IsCounted {
		return g.genCountedFor(f)
	}

	start := g.newLabel()
	end := g.newLabel()

	g.pushLoopLabels(start, end)
	defer g.popLoopLabels()

	g.emit(start + ":")

	if f.Cond != nil {
		if err := g.genConditionFalseJump(f.Cond, end); err != nil {
			return err
		}
	}

	for _, stmt := range f.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit(fmt.Sprintf("    jmp %s", start))
	g.emit(end + ":")
	return nil
}

func (g *Generator) genCountedFor(f *ast.ForStmt) error {
	counterSym, ok := g.resolve(f.Counter)
	if !ok {
		return fmt.Errorf("unknown loop variable %q", f.Counter)
	}

	endSym := g.newForLoopEndSymbol(counterSym.Type)
	check := g.newLabel()
	continueLabel := g.newLabel()
	endLabel := g.newLabel()

	if err := g.genExprTo(f.Start, counterSym.Type); err != nil {
		return err
	}
	g.storeAIntoSymbol(counterSym)

	if err := g.genExprTo(f.End, counterSym.Type); err != nil {
		return err
	}
	g.storeAIntoSymbol(endSym)

	g.pushLoopLabels(continueLabel, endLabel)
	defer g.popLoopLabels()

	g.emit(check + ":")
	if err := g.genForCounterPastEndJump(counterSym, endSym, endLabel); err != nil {
		return err
	}

	for _, stmt := range f.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit(continueLabel + ":")
	if err := g.genForCounterAtOrPastEndJump(counterSym, endSym, endLabel); err != nil {
		return err
	}

	g.emitIncrementSymbol(counterSym)
	g.emit(fmt.Sprintf("    jmp %s", check))
	g.emit(endLabel + ":")
	return nil
}

func (g *Generator) newForLoopEndSymbol(t ast.Type) Symbol {
	g.labelCounter++
	sym := Symbol{
		SourceName: "for_end",
		Label:      fmt.Sprintf("%s_for_end_%d", g.currentFn.Name, g.labelCounter),
		Type:       t,
		Size:       g.sizeof(t),
	}
	g.forLoopTemps = append(g.forLoopTemps, sym)
	return sym
}

func (g *Generator) genForCounterPastEndJump(counter Symbol, end Symbol, endLabel string) error {
	if counter.Type.Name == "int" {
		past := g.newLabel()
		skip := g.newLabel()

		g.loadForIntOperands(counter, end)
		// Signed comparison branches only to nearby labels. The far exit from
		// the user-controlled loop body is always an absolute jmp to endLabel.
		if err := g.genSignedGreaterThanJump(past); err != nil {
			return err
		}

		g.emit(fmt.Sprintf("    jmp %s", skip))
		g.emit(past + ":")
		g.emit(fmt.Sprintf("    jmp %s", endLabel))
		g.emit(skip + ":")
		return nil
	}

	skip := g.newLabel()
	g.emit(fmt.Sprintf("    lda %s", counter.Label))
	g.emit(fmt.Sprintf("    cmp %s", end.Label))
	// Byte comparisons branch to a nearby skip label, then use an absolute jmp
	// for the potentially far loop exit.
	g.emit(fmt.Sprintf("    beq %s", skip))
	g.emit(fmt.Sprintf("    bcc %s", skip))
	g.emit(fmt.Sprintf("    jmp %s", endLabel))
	g.emit(skip + ":")
	return nil
}

func (g *Generator) genForCounterAtOrPastEndJump(counter Symbol, end Symbol, endLabel string) error {
	if counter.Type.Name == "int" {
		less := g.newLabel()

		g.loadForIntOperands(counter, end)
		// Keep signed comparison branch targets local; use jmp for the
		// potentially far counted-loop exit.
		if err := g.genSignedLessThanJump(less); err != nil {
			return err
		}

		g.emit(fmt.Sprintf("    jmp %s", endLabel))
		g.emit(less + ":")
		return nil
	}

	skip := g.newLabel()
	g.emit(fmt.Sprintf("    lda %s", counter.Label))
	g.emit(fmt.Sprintf("    cmp %s", end.Label))
	// The branch target is local. The far exit is the following absolute jmp.
	g.emit(fmt.Sprintf("    bcc %s", skip))
	g.emit(fmt.Sprintf("    jmp %s", endLabel))
	g.emit(skip + ":")
	return nil
}

func (g *Generator) loadForIntOperands(counter Symbol, end Symbol) {
	g.loadSymbol(counter)
	g.emit(fmt.Sprintf("    lda %s", end.Label))
	g.emit("    sta peddle_tmp_int0")
	g.emit(fmt.Sprintf("    lda %s+1", end.Label))
	g.emit("    sta peddle_tmp_int0+1")
	g.usedTmp16 = true
}

func (g *Generator) emitIncrementSymbol(sym Symbol) {
	if sym.Type.Name == "int" {
		noCarry := g.newLabel()
		g.emit(fmt.Sprintf("    inc %s", sym.Label))
		g.emit(fmt.Sprintf("    bne %s", noCarry))
		g.emit(fmt.Sprintf("    inc %s+1", sym.Label))
		g.emit(noCarry + ":")
		return
	}

	g.emit(fmt.Sprintf("    inc %s", sym.Label))
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
	returnSymbols := g.currentFrame.Returns
	if len(returnSymbols) == 0 {
		g.emit("    rts")
		return nil
	}

	values := returnValues(r)
	if len(values) == 0 {
		return fmt.Errorf("missing return value")
	}
	if len(values) != len(returnSymbols) {
		return fmt.Errorf("function returns %d values, return statement has %d", len(returnSymbols), len(values))
	}

	for i, value := range values {
		if err := g.genExprTo(value, returnSymbols[i].Type); err != nil {
			return err
		}
		g.storeAIntoSymbol(returnSymbols[i])
	}
	g.emit("    rts")
	return nil
}
