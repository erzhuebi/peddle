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

		g.emit(fmt.Sprintf("    lda #<%d", len(expr.Value)))
		g.emit("    sta peddle_tmp_int0")
		g.emit(fmt.Sprintf("    lda #>%d", len(expr.Value)))
		g.emit("    sta peddle_tmp_int0+1")

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    jsr peddle_print_counted_string")
		g.usedPrint = true
		g.usedTmp16 = true
		return ast.Type{}, nil

	case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexFieldExpr:
		if err := g.genCharArrayAddress(args[0]); err != nil {
			return ast.Type{}, err
		}

		g.emit("    ldy #2")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    iny")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    sta peddle_tmp_int0+1")

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc #4")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR0_HI")

		g.emit("    jsr peddle_print_counted_string")
		g.usedPrint = true
		g.usedTmp16 = true
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

	if addr, ok := args[0].(*ast.IdentExpr); ok {
		if n, ok := g.constants[addr.Name]; ok {
			if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
				return ast.Type{}, err
			}

			g.emit(fmt.Sprintf("    sta $%04x", n&0xffff))
			return ast.Type{}, nil
		}
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

	if addr, ok := args[0].(*ast.IdentExpr); ok {
		if n, ok := g.constants[addr.Name]; ok {
			g.emit(fmt.Sprintf("    lda $%04x", n&0xffff))
			return ast.Type{Name: "byte"}, nil
		}
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

func (g *Generator) genCls(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("cls expects no arguments")
	}

	loopFull := g.newLabel()
	loopLast := g.newLabel()
	done := g.newLabel()

	g.emit("    lda #$20")
	g.emit("    ldx #0")
	g.emit(loopFull + ":")
	g.emit("    sta $0400, x")
	g.emit("    sta $0500, x")
	g.emit("    sta $0600, x")
	g.emit("    inx")
	g.emit(fmt.Sprintf("    bne %s", loopFull))

	g.emit("    ldx #0")
	g.emit(loopLast + ":")
	g.emit("    cpx #232")
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit("    sta $0700, x")
	g.emit("    inx")
	g.emit(fmt.Sprintf("    jmp %s", loopLast))
	g.emit(done + ":")

	// Reset KERNAL text cursor to row 0, column 0.
	// KERNAL PLOT ($fff0), carry clear = set cursor position.
	g.emit("    clc")
	g.emit("    ldx #0")
	g.emit("    ldy #0")
	g.emit("    jsr $fff0")

	return ast.Type{}, nil
}

func (g *Generator) genBorder(args []ast.Expr) (ast.Type, error) {
	return g.genStoreByteBuiltin("border", args, 0xd020)
}

func (g *Generator) genBackground(args []ast.Expr) (ast.Type, error) {
	return g.genStoreByteBuiltin("background", args, 0xd021)
}

func (g *Generator) genTextColor(args []ast.Expr) (ast.Type, error) {
	return g.genStoreByteBuiltin("textcolor", args, 0x0286)
}

func (g *Generator) genStoreByteBuiltin(name string, args []ast.Expr, addr int) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("%s expects one argument", name)
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}

	g.emit(fmt.Sprintf("    sta $%04x", addr&0xffff))
	return ast.Type{}, nil
}

func (g *Generator) genPutScreen(args []ast.Expr) (ast.Type, error) {
	return g.genPutScreenByte("putscreen", args, 0x0400, false)
}

func (g *Generator) genPutChar(args []ast.Expr) (ast.Type, error) {
	return g.genPutScreenByte("putchar", args, 0x0400, true)
}

func (g *Generator) genPutColor(args []ast.Expr) (ast.Type, error) {
	return g.genPutScreenByte("putcolor", args, 0xd800, false)
}

func (g *Generator) genPutScreenByte(name string, args []ast.Expr, base int, convertChar bool) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("%s expects three arguments", name)
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_tmp_int0")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_tmp_int0+1")

	if err := g.genExprTo(args[2], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta ZP_TMP0")

	if convertChar {
		g.genCharCodeToScreenCode()
	}

	clip := g.newLabel()
	addRow := g.newLabel()
	rowDone := g.newLabel()

	// Clip invalid coordinates.
	g.emit("    lda peddle_tmp_int0")
	g.emit("    cmp #40")
	g.emit(fmt.Sprintf("    bcs %s", clip))
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emit(fmt.Sprintf("    bcs %s", clip))

	g.emit(fmt.Sprintf("    lda #<$%04x", base&0xffff))
	g.emit("    sta ZP_PTR0_LO")
	g.emit(fmt.Sprintf("    lda #>$%04x", base&0xffff))
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldx peddle_tmp_int0+1")
	g.emit(addRow + ":")
	g.emit(fmt.Sprintf("    beq %s", rowDone))
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #40")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    dex")
	g.emit(fmt.Sprintf("    jmp %s", addRow))

	g.emit(rowDone + ":")
	g.emit("    ldy peddle_tmp_int0")
	g.emit("    lda ZP_TMP0")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit(clip + ":")

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genCharCodeToScreenCode() {
	checkLower := g.newLabel()
	done := g.newLabel()

	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #65")
	g.emit(fmt.Sprintf("    bcc %s", checkLower))
	g.emit("    cmp #91")
	g.emit(fmt.Sprintf("    bcs %s", checkLower))
	g.emit("    sec")
	g.emit("    sbc #64")
	g.emit("    sta ZP_TMP0")
	g.emit(fmt.Sprintf("    jmp %s", done))

	g.emit(checkLower + ":")
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #97")
	g.emit(fmt.Sprintf("    bcc %s", done))
	g.emit("    cmp #123")
	g.emit(fmt.Sprintf("    bcs %s", done))
	g.emit("    sec")
	g.emit("    sbc #96")
	g.emit("    sta ZP_TMP0")

	g.emit(done + ":")
}

func (g *Generator) genPutStr(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("putstr expects three arguments")
	}

	text, ok := args[2].(*ast.StringExpr)
	if !ok {
		return ast.Type{}, fmt.Errorf("putstr currently expects a string literal")
	}

	return g.genPutStringLiteral(args[0], args[1], text.Value, nil)
}

func (g *Generator) genPutStrColor(args []ast.Expr) (ast.Type, error) {
	if len(args) != 4 {
		return ast.Type{}, fmt.Errorf("putstrcolor expects four arguments")
	}

	text, ok := args[2].(*ast.StringExpr)
	if !ok {
		return ast.Type{}, fmt.Errorf("putstrcolor currently expects a string literal")
	}

	return g.genPutStringLiteral(args[0], args[1], text.Value, args[3])
}

func (g *Generator) genPutStringLiteral(xExpr ast.Expr, yExpr ast.Expr, value string, colorExpr ast.Expr) (ast.Type, error) {
	if err := g.genExprTo(xExpr, ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_tmp_int0") // start x
	g.emit("    sta ZP_TMP0")         // current x

	if err := g.genExprTo(yExpr, ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_tmp_int0+1") // current y

	if colorExpr != nil {
		if err := g.genExprTo(colorExpr, ast.Type{Name: "byte"}); err != nil {
			return ast.Type{}, err
		}
		g.emit("    sta ZP_TMP1") // color
	}

	done := g.newLabel()

	// If starting coordinates are outside the screen, clip the whole string.
	g.emit("    lda peddle_tmp_int0")
	g.emit("    cmp #40")
	g.emitJumpIfCarrySet(done)

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emitJumpIfCarrySet(done)

	for i := 0; i < len(value); i++ {
		ch := value[i]

		if ch == 13 {
			g.emit("    lda peddle_tmp_int0")
			g.emit("    sta ZP_TMP0")
			g.emit("    inc peddle_tmp_int0+1")
			g.emit("    lda peddle_tmp_int0+1")
			g.emit("    cmp #25")
			g.emitJumpIfCarrySet(done)
			continue
		}

		screenCode := peddleScreenCode(ch)

		g.emitPutLiteralAtCurrentScreenPosition(done, 0x0400, screenCode)

		if colorExpr != nil {
			g.emitPutCurrentColorAtCurrentScreenPosition(done)
		}

		g.emitAdvanceCurrentScreenPosition(done)
	}

	g.emit(done + ":")

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) emitPutLiteralAtCurrentScreenPosition(done string, base int, value byte) {
	addRow := g.newLabel()
	rowDone := g.newLabel()

	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emitJumpIfCarrySet(done)

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emitJumpIfCarrySet(done)

	g.emit(fmt.Sprintf("    lda #<$%04x", base&0xffff))
	g.emit("    sta ZP_PTR0_LO")
	g.emit(fmt.Sprintf("    lda #>$%04x", base&0xffff))
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldx peddle_tmp_int0+1")
	g.emit(addRow + ":")
	g.emit(fmt.Sprintf("    beq %s", rowDone))
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #40")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    dex")
	g.emit(fmt.Sprintf("    jmp %s", addRow))

	g.emit(rowDone + ":")
	g.emit("    ldy ZP_TMP0")
	g.emit(fmt.Sprintf("    lda #%d", int(value)))
	g.emit("    sta (ZP_PTR0_LO), y")
}

func (g *Generator) emitPutCurrentColorAtCurrentScreenPosition(done string) {
	addRow := g.newLabel()
	rowDone := g.newLabel()

	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emitJumpIfCarrySet(done)

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emitJumpIfCarrySet(done)

	g.emit("    lda #<$d800")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda #>$d800")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldx peddle_tmp_int0+1")
	g.emit(addRow + ":")
	g.emit(fmt.Sprintf("    beq %s", rowDone))
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #40")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    dex")
	g.emit(fmt.Sprintf("    jmp %s", addRow))

	g.emit(rowDone + ":")
	g.emit("    ldy ZP_TMP0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta (ZP_PTR0_LO), y")
}

func (g *Generator) emitAdvanceCurrentScreenPosition(done string) {
	noWrap := g.newLabel()

	g.emit("    inc ZP_TMP0")
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emit(fmt.Sprintf("    bcc %s", noWrap))

	g.emit("    lda #0")
	g.emit("    sta ZP_TMP0")
	g.emit("    inc peddle_tmp_int0+1")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emitJumpIfCarrySet(done)

	g.emit(noWrap + ":")
}

func peddleScreenCode(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch - 64
	}

	if ch >= 'a' && ch <= 'z' {
		return ch - 96
	}

	return ch
}

func (g *Generator) genLen(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("len expects one argument")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genSize(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("size expects one argument")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genAppend(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("append expects two arguments")
	}

	if src, ok := args[1].(*ast.StringExpr); ok {
		arrayType, err := g.arrayExprType(args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !(arrayType.IsArray && arrayType.Name == "char") {
			return ast.Type{}, fmt.Errorf("append string literal requires char array destination")
		}

		if err := g.genAppendStringLiteralToCharArray(args[0], src.Value); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{}, nil
	}

	arrayType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	elemType := ast.Type{Name: arrayType.Name}

	if _, ok := g.structs[elemType.Name]; ok {
		return ast.Type{}, fmt.Errorf("append does not support struct elements yet")
	}

	if g.options.OptMode == OptModeSize {
		return g.genAppendRuntime(args, elemType)
	}

	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	} else {
		g.emit("    sta peddle_tmp_int0")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")

	if err := g.genAddElementOffsetToPtr(elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda peddle_tmp_int0")
		g.emit("    ldy #0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    lda peddle_tmp_int0+1")
		g.emit("    iny")
		g.emit("    sta (ZP_PTR0_LO), y")
	} else {
		g.emit("    lda peddle_tmp_int0")
		g.emit("    ldy #0")
		g.emit("    sta (ZP_PTR0_LO), y")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    clc")
	g.emit("    adc #1")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    adc #0")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genAppendRuntime(args []ast.Expr, elemType ast.Type) (ast.Type, error) {
	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	} else {
		g.emit("    sta peddle_tmp_int0")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	switch elemType.Name {
	case "int":
		g.emit("    jsr peddle_append_int")
		g.usedAppendIntRuntime = true
	default:
		g.emit("    jsr peddle_append_byte")
		g.usedAppendByteRuntime = true
	}

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genCopy(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("copy expects two arguments")
	}

	if src, ok := args[1].(*ast.StringExpr); ok {
		if err := g.genCopyStringLiteralToCharArray(args[0], src.Value); err != nil {
			return ast.Type{}, err
		}
		return ast.Type{}, nil
	}

	dstType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	srcType, err := g.arrayExprType(args[1])
	if err != nil {
		return ast.Type{}, err
	}

	if dstType.Name != srcType.Name {
		return ast.Type{}, fmt.Errorf("copy requires arrays with same element type")
	}

	elemType := ast.Type{Name: dstType.Name}
	elemSize := g.sizeof(elemType)

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta ZP_PTR1_HI")

	if err := g.genArrayAddress(args[1]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    sta (ZP_PTR1_LO), y")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    sta (ZP_PTR1_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda ZP_PTR1_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR1_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR1_HI")

	g.genLengthTimesElemSizeToCounter(elemSize)

	if g.options.OptMode == OptModeSize {
		g.emit("    jsr peddle_array_copy")
		g.usedArrayCopyRuntime = true
		g.usedTmp16 = true
		return ast.Type{}, nil
	}

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta (ZP_PTR1_LO), y")

	g.emit("    inc ZP_PTR0_LO")
	g.emit("    bne " + loop + "_src_no_carry")
	g.emit("    inc ZP_PTR0_HI")
	g.emit(loop + "_src_no_carry:")

	g.emit("    inc ZP_PTR1_LO")
	g.emit("    bne " + loop + "_dst_no_carry")
	g.emit("    inc ZP_PTR1_HI")
	g.emit(loop + "_dst_no_carry:")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne " + loop + "_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit(loop + "_dec_low:")
	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genFill(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("fill expects two arguments")
	}

	arrayType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	elemType := ast.Type{Name: arrayType.Name}

	if _, ok := g.structs[elemType.Name]; ok {
		return ast.Type{}, fmt.Errorf("fill does not support struct elements yet")
	}

	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if elemType.Name == "int" {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta ZP_TMP1")
	} else {
		g.emit("    sta ZP_TMP0")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    tax")
	g.emit("    ldy #2")
	g.emit("    txa")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    ldy #1")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    tax")
	g.emit("    ldy #3")
	g.emit("    txa")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit(fmt.Sprintf("    lda #<%d", arrayType.ArrayLen))
	g.emit("    sta peddle_tmp_int0")
	g.emit(fmt.Sprintf("    lda #>%d", arrayType.ArrayLen))
	g.emit("    sta peddle_tmp_int0+1")

	if g.options.OptMode == OptModeSize {
		if elemType.Name == "int" {
			g.emit("    jsr peddle_fill_int")
			g.usedFillIntRuntime = true
		} else {
			g.emit("    jsr peddle_fill_byte")
			g.usedFillByteRuntime = true
		}

		g.usedTmp16 = true
		return ast.Type{}, nil
	}

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(loop + ":")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    beq %s", done))

	if elemType.Name == "int" {
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    iny")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta (ZP_PTR0_LO), y")

		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc #2")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR0_HI")
	} else {
		g.emit("    ldy #0")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")

		g.emit("    inc ZP_PTR0_LO")
		g.emit("    bne " + loop + "_ptr_no_carry")
		g.emit("    inc ZP_PTR0_HI")
		g.emit(loop + "_ptr_no_carry:")
	}

	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne " + loop + "_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit(loop + "_dec_low:")
	g.emit("    dec peddle_tmp_int0")
	g.emit(fmt.Sprintf("    jmp %s", loop))

	g.emit(done + ":")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genCopyStringLiteralToCharArray(dst ast.Expr, value string) error {
	if err := g.genCharArrayAddress(dst); err != nil {
		return err
	}

	label := g.addLiteral(value)

	if g.options.OptMode == OptModeSize {
		g.emit(fmt.Sprintf("    lda #<%d", len(value)))
		g.emit("    sta peddle_tmp_int0")
		g.emit(fmt.Sprintf("    lda #>%d", len(value)))
		g.emit("    sta peddle_tmp_int0+1")

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR1_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR1_HI")

		g.emit("    jsr peddle_string_copy_literal")
		g.usedStringCopyRuntime = true
		g.usedTmp16 = true
		return nil
	}

	loop := g.newLabel()
	done := g.newLabel()

	g.emit(fmt.Sprintf("    lda #<%d", len(value)))
	g.emit("    ldy #2")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit(fmt.Sprintf("    lda #>%d", len(value)))
	g.emit("    ldy #3")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldy #0")
	g.emit(loop + ":")
	g.emit(fmt.Sprintf("    cpy #%d", len(value)))
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit(fmt.Sprintf("    lda %s, y", label))
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit(fmt.Sprintf("    jmp %s", loop))
	g.emit(done + ":")
	return nil
}

func (g *Generator) genAppendStringLiteralToCharArray(dst ast.Expr, value string) error {
	if err := g.genCharArrayAddress(dst); err != nil {
		return err
	}

	label := g.addLiteral(value)

	if g.options.OptMode == OptModeSize {
		g.emit(fmt.Sprintf("    lda #<%d", len(value)))
		g.emit("    sta peddle_tmp_int0")
		g.emit(fmt.Sprintf("    lda #>%d", len(value)))
		g.emit("    sta peddle_tmp_int0+1")

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR1_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR1_HI")

		g.emit("    jsr peddle_string_append_literal")
		g.usedStringAppendRuntime = true
		g.usedTmp16 = true
		return nil
	}

	loop := g.newLabel()
	done := g.newLabel()

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_tmp_int0+1")

	g.emit("    lda peddle_tmp_int0")
	g.emit("    clc")
	g.emit(fmt.Sprintf("    adc #<%d", len(value)))
	g.emit("    sta ZP_TMP0")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    adc #>%d", len(value)))
	g.emit("    sta ZP_TMP1")

	g.emit("    ldy #2")
	g.emit("    lda ZP_TMP0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc peddle_tmp_int0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc peddle_tmp_int0+1")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldy #0")
	g.emit(loop + ":")
	g.emit(fmt.Sprintf("    cpy #%d", len(value)))
	g.emit(fmt.Sprintf("    beq %s", done))
	g.emit(fmt.Sprintf("    lda %s, y", label))
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit(fmt.Sprintf("    jmp %s", loop))
	g.emit(done + ":")

	g.usedTmp16 = true
	return nil
}

func (g *Generator) genClear(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("clear expects one argument")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    ldy #2")
	g.emit("    lda #0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    sta (ZP_PTR0_LO), y")

	return ast.Type{}, nil
}

func (g *Generator) genCharArrayAddress(expr ast.Expr) error {
	t, err := g.arrayExprType(expr)
	if err != nil {
		return err
	}

	if !(t.IsArray && t.Name == "char") {
		return fmt.Errorf("expected char array")
	}

	return g.genArrayAddress(expr)
}

func (g *Generator) genArrayAddress(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return fmt.Errorf("unknown variable %q", e.Name)
		}

		if !sym.Type.IsArray {
			return fmt.Errorf("expected array")
		}

		g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
		g.emit("    sta ZP_PTR0_HI")
		return nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(e.Base)
		if !ok {
			return fmt.Errorf("unknown variable %q", e.Base)
		}

		fieldType, offset, err := g.fieldInfo(baseSym.Type, e.Field)
		if err != nil {
			return err
		}

		if !fieldType.IsArray {
			return fmt.Errorf("expected array")
		}

		g.emit(fmt.Sprintf("    lda #<%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_LO")
		g.emit(fmt.Sprintf("    lda #>%s+%d", baseSym.Label, offset))
		g.emit("    sta ZP_PTR0_HI")
		return nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(e.Name)
		if !ok {
			return fmt.Errorf("unknown array %q", e.Name)
		}
		if !arraySym.Type.IsArray {
			return fmt.Errorf("%q is not an array", e.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, offset, err := g.fieldInfo(elemType, e.Field)
		if err != nil {
			return err
		}

		if !fieldType.IsArray {
			return fmt.Errorf("expected array")
		}

		if err := g.genArrayIndexToY(arraySym, e.Index); err != nil {
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

		return nil
	}

	return fmt.Errorf("expected array")
}

func (g *Generator) arrayExprType(expr ast.Expr) (ast.Type, error) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Name)
		}
		if !sym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return sym.Type, nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(e.Base)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Base)
		}

		fieldType, _, err := g.fieldInfo(baseSym.Type, e.Field)
		if err != nil {
			return ast.Type{}, err
		}
		if !fieldType.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return fieldType, nil

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown array %q", e.Name)
		}
		if !arraySym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", e.Name)
		}

		elemType := ast.Type{Name: arraySym.Type.Name}

		fieldType, _, err := g.fieldInfo(elemType, e.Field)
		if err != nil {
			return ast.Type{}, err
		}
		if !fieldType.IsArray {
			return ast.Type{}, fmt.Errorf("expected array")
		}
		return fieldType, nil
	}

	return ast.Type{}, fmt.Errorf("expected array")
}

func (g *Generator) genUpdateArrayLenFromIndex(arraySym Symbol, index ast.Expr) error {
	if err := g.genExprTo(index, ast.Type{Name: "int"}); err != nil {
		return err
	}

	noCarry := g.newLabel()
	update := g.newLabel()
	done := g.newLabel()

	g.emit("    inc ZP_TMP0")
	g.emit(fmt.Sprintf("    bne %s", noCarry))
	g.emit("    inc ZP_TMP1")
	g.emit(noCarry + ":")

	g.emit(fmt.Sprintf("    lda %s+3", arraySym.Label))
	g.emit("    cmp ZP_TMP1")
	g.emit(fmt.Sprintf("    bcc %s", update))
	g.emit(fmt.Sprintf("    bne %s", done))
	g.emit(fmt.Sprintf("    lda %s+2", arraySym.Label))
	g.emit("    cmp ZP_TMP0")
	g.emit(fmt.Sprintf("    bcc %s", update))
	g.emit(fmt.Sprintf("    jmp %s", done))

	g.emit(update + ":")
	g.emit("    lda ZP_TMP0")
	g.emit(fmt.Sprintf("    sta %s+2", arraySym.Label))
	g.emit("    lda ZP_TMP1")
	g.emit(fmt.Sprintf("    sta %s+3", arraySym.Label))

	g.emit(done + ":")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genAddElementOffsetToPtr(elemType ast.Type) error {
	elemSize := g.sizeof(elemType)

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")

	if elemSize == 1 {
		g.emit("    lda ZP_PTR0_LO")
		g.emit("    clc")
		g.emit("    adc ZP_TMP0")
		g.emit("    sta ZP_PTR0_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc ZP_TMP1")
		g.emit("    sta ZP_PTR0_HI")
		return nil
	}

	if elemSize == 2 {
		g.emit("    asl ZP_TMP0")
		g.emit("    rol ZP_TMP1")
	} else {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")

		for i := 1; i < elemSize; i++ {
			g.emit("    clc")
			g.emit("    lda ZP_TMP0")
			g.emit("    adc peddle_tmp_int0")
			g.emit("    sta ZP_TMP0")
			g.emit("    lda ZP_TMP1")
			g.emit("    adc peddle_tmp_int0+1")
			g.emit("    sta ZP_TMP1")
		}
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc ZP_TMP0")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc ZP_TMP1")
	g.emit("    sta ZP_PTR0_HI")
	g.usedTmp16 = true
	return nil
}

func (g *Generator) genLengthTimesElemSizeToCounter(elemSize int) {
	if elemSize == 1 {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
		g.usedTmp16 = true
		return
	}

	g.emit("    lda #0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    sta peddle_tmp_int0+1")

	for i := 0; i < elemSize; i++ {
		g.emit("    clc")
		g.emit("    lda peddle_tmp_int0")
		g.emit("    adc ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda peddle_tmp_int0+1")
		g.emit("    adc ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	}

	g.usedTmp16 = true
}

func (g *Generator) emitJumpIfCarrySet(label string) {
	skip := g.newLabel()

	g.emit(fmt.Sprintf("    bcc %s", skip))
	g.emit(fmt.Sprintf("    jmp %s", label))
	g.emit(skip + ":")
}
