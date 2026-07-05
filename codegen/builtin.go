package codegen

import (
	"fmt"

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

	}

	t, err := g.arrayExprType(args[0])
	if err == nil && t.IsArray && t.Name == "char" {
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

	if n, ok, err := g.foldConstExpr(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	} else if ok {
		if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
			return ast.Type{}, err
		}

		g.emit(fmt.Sprintf("    sta $%04x", n&0xffff))
		return ast.Type{}, nil
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    pha")
	g.emit("    lda ZP_TMP1")
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    sta ZP_TMP0")
	g.emit("    pla")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    pla")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_TMP0")
	g.emit("    ldy #0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genPeek(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("peek expects one argument")
	}

	if n, ok, err := g.foldConstExpr(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	} else if ok {
		g.emit(fmt.Sprintf("    lda $%04x", n&0xffff))
		return ast.Type{Name: "byte"}, nil
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

func (g *Generator) genTicks(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("ticks expects no arguments")
	}

	// C64 KERNAL jiffy clock:
	//   $A2 = fastest-changing byte
	//   $A1 = next byte
	//
	// We expose these two bytes as a Peddle int:
	//   low byte  = $A2
	//   high byte = $A1
	//
	// This gives a 16-bit tick counter that advances roughly once per frame
	// while the normal KERNAL interrupt is running.
	g.emit("    lda $00a2")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda $00a1")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")

	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genElapsed(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("elapsed expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	// Store last tick value.
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_tmp_int0+1")

	// Compute current - last modulo 65536.
	//
	// C64 KERNAL jiffy clock byte order for our exposed 16-bit value:
	//   low byte  = $A2
	//   high byte = $A1
	g.emit("    lda $00a2")
	g.emit("    sec")
	g.emit("    sbc peddle_tmp_int0")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda $00a1")
	g.emit("    sbc peddle_tmp_int0+1")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")

	g.usedTmp16 = true
	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genTickDue(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("tickdue expects two arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    pha")
	g.emit("    lda ZP_TMP1")
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	// Store interval in peddle_tmp_int0.
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_tmp_int0+1")

	// Restore last tick value into ZP_PTR0_LO/HI after interval generation.
	g.emit("    pla")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    pla")
	g.emit("    sta ZP_PTR0_LO")

	// Compute elapsed = current - last modulo 65536.
	//
	// C64 KERNAL jiffy clock byte order for our exposed 16-bit value:
	//   low byte  = $A2
	//   high byte = $A1
	g.emit("    lda $00a2")
	g.emit("    sec")
	g.emit("    sbc ZP_PTR0_LO")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda $00a1")
	g.emit("    sbc ZP_PTR0_HI")
	g.emit("    sta ZP_TMP1")

	trueLabel := g.newLabel()
	falseLabel := g.newLabel()
	doneLabel := g.newLabel()

	// Unsigned 16-bit compare:
	//   elapsed >= interval
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp peddle_tmp_int0+1")
	g.emit(fmt.Sprintf("    bcc %s", falseLabel))
	g.emit(fmt.Sprintf("    bne %s", trueLabel))
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp peddle_tmp_int0")
	g.emit(fmt.Sprintf("    bcc %s", falseLabel))

	g.emit(trueLabel + ":")
	g.emit("    lda #1")
	g.emit(fmt.Sprintf("    jmp %s", doneLabel))

	g.emit(falseLabel + ":")
	g.emit("    lda #0")

	g.emit(doneLabel + ":")

	g.usedTmp16 = true
	return ast.Type{Name: "bool"}, nil
}

func (g *Generator) genJoy(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("joy expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}

	port1 := g.newLabel()
	port2 := g.newLabel()
	done := g.newLabel()

	// C64 joystick inputs are active-low.
	// joy(1) reads CIA #1 port B ($dc01).
	// joy(2) reads CIA #1 port A ($dc00).
	// Invalid port numbers return 255, which means no direction/fire pressed.
	g.emit("    cmp #1")
	g.emit(fmt.Sprintf("    beq %s", port1))
	g.emit("    cmp #2")
	g.emit(fmt.Sprintf("    beq %s", port2))
	g.emit("    lda #255")
	g.emit(fmt.Sprintf("    jmp %s", done))

	g.emit(port1 + ":")
	g.emit("    lda $dc01")
	g.emit(fmt.Sprintf("    jmp %s", done))

	g.emit(port2 + ":")
	g.emit("    lda $dc00")

	g.emit(done + ":")
	return ast.Type{Name: "byte"}, nil
}

func (g *Generator) genKey(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("key expects no arguments")
	}

	// KERNAL GETIN ($ffe4): returns one PETSCII character code in A,
	// or 0 when no key is waiting in the keyboard buffer.
	g.emit("    jsr $ffe4")
	return ast.Type{Name: "char"}, nil
}

func (g *Generator) genWaitKey(args []ast.Expr) (ast.Type, error) {
	return g.genWaitKeyNamed("waitkey", args)
}

func (g *Generator) genWaitKeyNamed(name string, args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("%s expects no arguments", name)
	}

	loop := g.newLabel()

	// KERNAL GETIN ($ffe4): returns one PETSCII character code in A,
	// or 0 when no key is waiting in the keyboard buffer.
	g.emit(loop + ":")
	g.emit("    jsr $ffe4")
	g.emit("    cmp #0")
	g.emit(fmt.Sprintf("    beq %s", loop))

	return ast.Type{Name: "char"}, nil
}

func (g *Generator) genReadLine(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("readline expects three arguments")
	}

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_readline_echo")

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_readline_max")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_readline_max+1")

	if err := g.genCharArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    jsr peddle_readline")
	g.usedReadLineRuntime = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genCls(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("cls expects no arguments")
	}

	g.emit("    jsr peddle_cls")
	g.usedClsRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genAsciiFont(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("asciifont expects no arguments")
	}

	g.emit("    jsr peddle_asciifont")
	g.usedAsciiFontRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genToASCII(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("toascii expects one argument")
	}

	if err := g.genCharArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    jsr peddle_toascii")
	g.usedAsciiConvertRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genToPETSCII(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("topetscii expects one argument")
	}

	if err := g.genCharArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    jsr peddle_topetscii")
	g.usedAsciiConvertRuntime = true

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

func (g *Generator) genGotoXY(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("gotoxy expects two arguments")
	}

	if g.options.OptMode == OptModeSize {
		if err := g.genScreenXYArgs(args); err != nil {
			return ast.Type{}, err
		}

		g.emit("    jsr peddle_gotoxy")
		g.usedTmp16 = true
		g.usedGotoXYRuntime = true
		return ast.Type{}, nil
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha") // x / column

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_tmp_int0+1") // y / row
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")

	done := g.newLabel()

	// Clip invalid coordinates.
	g.emit("    lda peddle_tmp_int0")
	g.emit("    cmp #40")
	g.emit(fmt.Sprintf("    bcs %s", done))

	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emit(fmt.Sprintf("    bcs %s", done))

	// KERNAL PLOT ($fff0), carry clear = set cursor position.
	// X register = row, Y register = column.
	g.emit("    clc")
	g.emit("    ldx peddle_tmp_int0+1")
	g.emit("    ldy peddle_tmp_int0")
	g.emit("    jsr $fff0")

	g.emit(done + ":")

	g.usedTmp16 = true
	return ast.Type{}, nil
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

func (g *Generator) genPutRaw(args []ast.Expr) (ast.Type, error) {
	if g.options.OptMode == OptModeSize {
		return g.genPutScreenByteRuntime("putraw", args, "peddle_putraw", false)
	}
	return g.genPutScreenByte("putraw", args, 0x0400, false)
}

func (g *Generator) genPutChar(args []ast.Expr) (ast.Type, error) {
	if g.options.OptMode == OptModeSize {
		return g.genPutScreenByteRuntime("putchar", args, "peddle_putchar", true)
	}
	return g.genPutScreenByte("putchar", args, 0x0400, true)
}

func (g *Generator) genPutColor(args []ast.Expr) (ast.Type, error) {
	if g.options.OptMode == OptModeSize {
		return g.genPutScreenByteRuntime("putcolor", args, "peddle_putcolor", false)
	}
	return g.genPutScreenByte("putcolor", args, 0xd800, false)
}

func (g *Generator) genPutCharColor(args []ast.Expr) (ast.Type, error) {
	if len(args) != 4 {
		return ast.Type{}, fmt.Errorf("putcharcolor expects four arguments")
	}

	if g.options.OptMode == OptModeSize {
		if err := g.genPutCharColorRuntime(args); err != nil {
			return ast.Type{}, err
		}

		g.emit("    jsr peddle_putcharcolor")
		g.usedTmp16 = true
		g.usedPutCharColorRuntime = true
		g.usedCharToScreenTable = true
		return ast.Type{}, nil
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[2], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta ZP_TMP0")
	g.genCharCodeToScreenCode()

	g.emit("    lda ZP_TMP0")
	g.emit("    pha")

	if err := g.genExprTo(args[3], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta ZP_TMP1")
	g.emit("    pla")
	g.emit("    sta ZP_TMP0")

	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")

	clip := g.newLabel()
	addRow := g.newLabel()
	rowDone := g.newLabel()

	g.emit("    lda peddle_tmp_int0")
	g.emit("    cmp #40")
	g.emit(fmt.Sprintf("    bcs %s", clip))
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    cmp #25")
	g.emit(fmt.Sprintf("    bcs %s", clip))

	g.emit("    lda #<$0400")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda #>$0400")
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
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    clc")
	g.emit("    adc #212")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta (ZP_PTR0_LO), y")

	g.emit(clip + ":")

	g.usedTmp16 = true
	return ast.Type{}, nil
}

func (g *Generator) genScreenXYArgs(args []ast.Expr) error {
	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")
	return nil
}

func (g *Generator) genPushScreenXYArgs(args []ast.Expr) error {
	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    pha")
	return nil
}

func (g *Generator) genPopScreenXYArgs() {
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")
}

func (g *Generator) genPutScreenByteRuntime(name string, args []ast.Expr, helper string, convertChar bool) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("%s expects three arguments", name)
	}

	if err := g.genPushScreenXYArgs(args); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta ZP_TMP0")
	g.genPopScreenXYArgs()
	g.emit(fmt.Sprintf("    jsr %s", helper))

	g.usedTmp16 = true
	switch name {
	case "putraw":
		g.usedPutRawRuntime = true
	case "putchar":
		g.usedPutCharRuntime = true
		g.usedCharToScreenTable = true
	case "putcolor":
		g.usedPutColorRuntime = true
	}
	if convertChar {
		g.usedCharToScreenTable = true
	}

	return ast.Type{}, nil
}

func (g *Generator) genPutCharColorRuntime(args []ast.Expr) error {
	if err := g.genPushScreenXYArgs(args); err != nil {
		return err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta ZP_TMP0")
	g.emit("    lda ZP_TMP0")
	g.emit("    pha")

	if err := g.genExprTo(args[3], ast.Type{Name: "byte"}); err != nil {
		return err
	}
	g.emit("    sta ZP_TMP1")
	g.emit("    pla")
	g.emit("    sta ZP_TMP0")
	g.genPopScreenXYArgs()
	return nil
}

func (g *Generator) genPutScreenByte(name string, args []ast.Expr, base int, convertChar bool) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("%s expects three arguments", name)
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[2], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta ZP_TMP0")

	if convertChar {
		g.genCharCodeToScreenCode()
	}

	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")

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
	g.usedCharToScreenTable = true

	g.emit("    lda ZP_TMP0")
	g.emit("    tax")
	g.emit("    lda peddle_char_to_screen_table, x")
	g.emit("    sta ZP_TMP0")
}

func (g *Generator) genPutStr(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("putstr expects three arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genPutStrTextArg("putstr", args[2]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    pla")
	g.emit("    sta peddle_putstr_y")
	g.emit("    pla")
	g.emit("    sta peddle_putstr_x")

	g.emit("    jsr peddle_putstr")
	g.usedTmp16 = true
	g.usedPutStrRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genPutStrColor(args []ast.Expr) (ast.Type, error) {
	if len(args) != 4 {
		return ast.Type{}, fmt.Errorf("putstrcolor expects four arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genExprTo(args[1], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    pha")

	if err := g.genPutStrTextArg("putstrcolor", args[2]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda peddle_tmp_int0")
	g.emit("    pha")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    pha")
	g.emit("    lda ZP_PTR1_LO")
	g.emit("    pha")
	g.emit("    lda ZP_PTR1_HI")
	g.emit("    pha")

	if err := g.genExprTo(args[3], ast.Type{Name: "byte"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    sta peddle_putstr_color")

	g.emit("    pla")
	g.emit("    sta ZP_PTR1_HI")
	g.emit("    pla")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")
	g.emit("    pla")
	g.emit("    sta peddle_putstr_y")
	g.emit("    pla")
	g.emit("    sta peddle_putstr_x")

	g.emit("    jsr peddle_putstrcolor")
	g.usedTmp16 = true
	g.usedPutStrRuntime = true
	g.usedPutStrColorRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genPutStrTextArg(name string, arg ast.Expr) error {
	switch text := arg.(type) {
	case *ast.StringExpr:
		label := g.addLiteral(text.Value)

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR1_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR1_HI")

		g.emit(fmt.Sprintf("    lda #<%d", len(text.Value)))
		g.emit("    sta peddle_tmp_int0")
		g.emit(fmt.Sprintf("    lda #>%d", len(text.Value)))
		g.emit("    sta peddle_tmp_int0+1")

	case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexFieldExpr, *ast.CallExpr:
		arrayType, err := g.arrayExprType(arg)
		if err != nil {
			return err
		}

		if !(arrayType.IsArray && arrayType.Name == "char") {
			return fmt.Errorf("%s expects string literal or char array", name)
		}

		if err := g.genArrayAddress(arg); err != nil {
			return err
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
		g.emit("    sta ZP_PTR1_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR1_HI")

	default:
		return fmt.Errorf("%s expects string literal or char array", name)
	}

	return nil
}

func (g *Generator) genLen(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("len expects one argument")
	}

	t, err := g.exprValueType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if t.IsMem {
		g.emitConstExprTo(t.ArrayLen, ast.Type{Name: "int"})
		return ast.Type{Name: "int"}, nil
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

	t, err := g.exprValueType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if t.IsMem {
		g.emitConstExprTo(t.ArrayLen, ast.Type{Name: "int"})
		return ast.Type{Name: "int"}, nil
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

func (g *Generator) pushTmpInt0() {
	g.emit("    lda peddle_tmp_int0")
	g.emit("    pha")
	g.emit("    lda peddle_tmp_int0+1")
	g.emit("    pha")
}

func (g *Generator) popTmpInt0() {
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0+1")
	g.emit("    pla")
	g.emit("    sta peddle_tmp_int0")
}

func (g *Generator) pushZPTmp0Word() {
	g.emit("    lda ZP_TMP0")
	g.emit("    pha")
	g.emit("    lda ZP_TMP1")
	g.emit("    pha")
}

func (g *Generator) popZPTmp0Word() {
	g.emit("    pla")
	g.emit("    sta ZP_TMP1")
	g.emit("    pla")
	g.emit("    sta ZP_TMP0")
}

func arrayAddressMayClobberScratch(expr ast.Expr) bool {
	_, ok := expr.(*ast.IndexFieldExpr)
	return ok
}

func (g *Generator) genAppend(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("append expects two arguments")
	}

	dstType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	if dstType.IsArray && dstType.Name == "char" {
		if _, ok := args[1].(*ast.StringExpr); ok {
			return g.genAppendCharArraySource(args[0], args[1])
		}

		srcType, err := g.arrayExprType(args[1])
		if err == nil && srcType.IsArray && srcType.Name == "char" {
			return g.genAppendCharArraySource(args[0], args[1])
		}
	}

	elemType := ast.Type{Name: dstType.Name}

	if _, ok := g.structs[elemType.Name]; ok {
		return ast.Type{}, fmt.Errorf("append does not support struct elements yet")
	}

	if g.options.OptMode == OptModeSize {
		return g.genAppendRuntime(args, elemType)
	}

	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if isWordType(elemType) {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	} else {
		g.emit("    sta peddle_tmp_int0")
	}

	preserveValue := arrayAddressMayClobberScratch(args[0])
	if preserveValue {
		g.pushTmpInt0()
	}
	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}
	if preserveValue {
		g.popTmpInt0()
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

	if isWordType(elemType) {
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

func (g *Generator) genAppendCharArraySource(dst ast.Expr, src ast.Expr) (ast.Type, error) {
	dstType, err := g.arrayExprType(dst)
	if err != nil {
		return ast.Type{}, err
	}

	if !(dstType.IsArray && dstType.Name == "char") {
		return ast.Type{}, fmt.Errorf("append char array source requires char array destination")
	}

	if err := g.genCharArraySourceToPtr1AndLen(src); err != nil {
		return ast.Type{}, err
	}

	preserveSource := arrayAddressMayClobberScratch(dst)
	if preserveSource {
		// Destination address generation may use the same scratch word that holds
		// the prepared source length, especially for indexed struct-array fields.
		g.emit("    lda ZP_PTR1_LO")
		g.emit("    pha")
		g.emit("    lda ZP_PTR1_HI")
		g.emit("    pha")
		g.emit("    lda peddle_tmp_int0")
		g.emit("    pha")
		g.emit("    lda peddle_tmp_int0+1")
		g.emit("    pha")
	}

	if err := g.genArrayAddress(dst); err != nil {
		return ast.Type{}, err
	}

	if preserveSource {
		g.emit("    pla")
		g.emit("    sta peddle_tmp_int0+1")
		g.emit("    pla")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    pla")
		g.emit("    sta ZP_PTR1_HI")
		g.emit("    pla")
		g.emit("    sta ZP_PTR1_LO")
	}

	g.emit("    jsr peddle_string_append_literal")

	g.usedTmp16 = true
	g.usedStringAppendRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genCharArraySourceToPtr1AndLen(src ast.Expr) error {
	switch e := src.(type) {
	case *ast.StringExpr:
		label := g.addLiteral(e.Value)

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit("    sta ZP_PTR1_LO")
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit("    sta ZP_PTR1_HI")

		g.emit(fmt.Sprintf("    lda #<%d", len(e.Value)))
		g.emit("    sta peddle_tmp_int0")
		g.emit(fmt.Sprintf("    lda #>%d", len(e.Value)))
		g.emit("    sta peddle_tmp_int0+1")

		return nil

	default:
		srcType, err := g.arrayExprType(src)
		if err != nil {
			return err
		}

		if !(srcType.IsArray && srcType.Name == "char") {
			return fmt.Errorf("expected char array source")
		}

		if err := g.genArrayAddress(src); err != nil {
			return err
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
		g.emit("    sta ZP_PTR1_LO")
		g.emit("    lda ZP_PTR0_HI")
		g.emit("    adc #0")
		g.emit("    sta ZP_PTR1_HI")

		return nil
	}
}

func (g *Generator) genAppendRuntime(args []ast.Expr, elemType ast.Type) (ast.Type, error) {
	if err := g.genExprTo(args[1], elemType); err != nil {
		return ast.Type{}, err
	}

	if isWordType(elemType) {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta peddle_tmp_int0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta peddle_tmp_int0+1")
	} else {
		g.emit("    sta peddle_tmp_int0")
	}

	preserveValue := arrayAddressMayClobberScratch(args[0])
	if preserveValue {
		g.pushTmpInt0()
	}
	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}
	if preserveValue {
		g.popTmpInt0()
	}

	switch elemType.Name {
	case "int", "uint":
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

	if isWordType(elemType) {
		g.emit("    lda ZP_TMP0")
		g.emit("    sta ZP_TMP0")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta ZP_TMP1")
	} else {
		g.emit("    sta ZP_TMP0")
	}

	preserveValue := arrayAddressMayClobberScratch(args[0])
	if preserveValue {
		g.pushZPTmp0Word()
	}
	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}
	if preserveValue {
		g.popZPTmp0Word()
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
		if isWordType(elemType) {
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

	if isWordType(elemType) {
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

		g.genArrayHeaderAddress(sym)
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

		if baseSym.Type.IsPointer {
			return fmt.Errorf("array fields through pointer parameters are not implemented yet")
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

	case *ast.CallExpr:
		switch e.Name {
		case "itoa":
			if _, err := g.genItoa(e.Args); err != nil {
				return err
			}

			g.emit("    lda #<peddle_itoa_buffer")
			g.emit("    sta ZP_PTR0_LO")
			g.emit("    lda #>peddle_itoa_buffer")
			g.emit("    sta ZP_PTR0_HI")
			return nil

		case "itox":
			t, err := g.genItox(e.Args)
			if err != nil {
				return err
			}

			label := "peddle_itox_byte_buffer"
			if t.ArrayLen == 4 {
				label = "peddle_itox_int_buffer"
			}

			g.emit(fmt.Sprintf("    lda #<%s", label))
			g.emit("    sta ZP_PTR0_LO")
			g.emit(fmt.Sprintf("    lda #>%s", label))
			g.emit("    sta ZP_PTR0_HI")
			return nil

		default:
			return fmt.Errorf("only itoa() or itox() can be used as temporary char array expression")
		}
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

	case *ast.CallExpr:
		switch e.Name {
		case "itoa":
			return ast.Type{
				Name:     "char",
				IsArray:  true,
				ArrayLen: 6,
			}, nil

		case "itox":
			return g.itoxReturnType(e.Args)
		}
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

	if arraySym.IsReference {
		g.genArrayHeaderAddress(arraySym)
		g.emit("    ldy #3")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    cmp ZP_TMP1")
		g.emit(fmt.Sprintf("    bcc %s", update))
		g.emit(fmt.Sprintf("    bne %s", done))
		g.emit("    dey")
		g.emit("    lda (ZP_PTR0_LO), y")
		g.emit("    cmp ZP_TMP0")
		g.emit(fmt.Sprintf("    bcc %s", update))
		g.emit(fmt.Sprintf("    jmp %s", done))

		g.emit(update + ":")
		g.emit("    ldy #2")
		g.emit("    lda ZP_TMP0")
		g.emit("    sta (ZP_PTR0_LO), y")
		g.emit("    iny")
		g.emit("    lda ZP_TMP1")
		g.emit("    sta (ZP_PTR0_LO), y")
	} else {
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
	}

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

func (g *Generator) genItoa(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("itoa expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    jsr peddle_itoa")

	g.usedItoaRuntime = true

	return ast.Type{
		Name:     "char",
		IsArray:  true,
		ArrayLen: 6,
	}, nil
}

func (g *Generator) genItox(args []ast.Expr) (ast.Type, error) {
	t, err := g.itoxReturnType(args)
	if err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	if t.ArrayLen == 4 {
		g.emit("    jsr peddle_itox_int")
	} else {
		g.emit("    jsr peddle_itox_byte")
	}

	g.usedItoxRuntime = true

	return t, nil
}

func (g *Generator) itoxReturnType(args []ast.Expr) (ast.Type, error) {
	argType, err := g.itoxArgType(args)
	if err != nil {
		return ast.Type{}, err
	}

	width := 2
	if argType.Name == "int" || argType.Name == "uint" {
		width = 4
	}

	return ast.Type{
		Name:     "char",
		IsArray:  true,
		ArrayLen: width,
	}, nil
}

func (g *Generator) itoxArgType(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("itox expects one argument")
	}

	t, err := g.exprValueType(args[0])
	if err != nil {
		return ast.Type{}, err
	}

	if !codegenNumericType(t) {
		return ast.Type{}, fmt.Errorf("itox argument must be numeric")
	}

	return t, nil
}

func (g *Generator) exprValueType(expr ast.Expr) (ast.Type, error) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		if _, ok := g.constants[e.Name]; ok {
			return ast.Type{Name: "int"}, nil
		}

		sym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Name)
		}
		if isScalarPointerType(sym.Type) {
			return pointerPointeeType(sym.Type), nil
		}
		return sym.Type, nil

	case *ast.IndexExpr:
		sym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Name)
		}
		if !sym.Type.IsArray {
			if sym.Type.IsMem {
				return ast.Type{Name: "byte"}, nil
			}
			return ast.Type{}, fmt.Errorf("%q is not an array or mem", e.Name)
		}
		return ast.Type{Name: sym.Type.Name}, nil

	case *ast.FieldExpr:
		baseSym, ok := g.resolve(e.Base)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", e.Base)
		}
		return g.fieldInfoType(baseSym.Type, e.Field)

	case *ast.IndexFieldExpr:
		arraySym, ok := g.resolve(e.Name)
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown array %q", e.Name)
		}
		if !arraySym.Type.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", e.Name)
		}
		return g.fieldInfoType(ast.Type{Name: arraySym.Type.Name}, e.Field)

	case *ast.NumberExpr:
		return ast.Type{Name: "int"}, nil

	case *ast.CharExpr:
		return ast.Type{Name: "char"}, nil

	case *ast.BoolExpr:
		return ast.Type{Name: "bool"}, nil

	case *ast.StringExpr:
		return ast.Type{Name: "char", IsArray: true, ArrayLen: len(e.Value)}, nil

	case *ast.UnaryExpr:
		t, err := g.exprValueType(e.Expr)
		if err != nil {
			return ast.Type{}, err
		}

		switch e.Op {
		case "-":
			if t.Name == "int" || t.Name == "uint" {
				return ast.Type{Name: "int"}, nil
			}
			return ast.Type{Name: "byte"}, nil

		case "!":
			return ast.Type{Name: "bool"}, nil
		}

	case *ast.BinaryExpr:
		left, err := g.exprValueType(e.Left)
		if err != nil {
			return ast.Type{}, err
		}
		right, err := g.exprValueType(e.Right)
		if err != nil {
			return ast.Type{}, err
		}

		if isComparisonOp(e.Op) {
			return ast.Type{Name: "bool"}, nil
		}

		if left.Name == "uint" || right.Name == "uint" {
			return ast.Type{Name: "uint"}, nil
		}
		if left.Name == "int" || right.Name == "int" {
			return ast.Type{Name: "int"}, nil
		}
		return ast.Type{Name: "byte"}, nil

	case *ast.CallExpr:
		switch e.Name {
		case "itoa":
			return ast.Type{Name: "char", IsArray: true, ArrayLen: 6}, nil
		case "itox":
			return g.itoxReturnType(e.Args)
		case "peek", "ticks", "elapsed", "readline", "netavailable", "netread", "netwrite", "fileload", "filesave", "fileread", "filewrite", "len", "size":
			return ast.Type{Name: "int"}, nil
		case "joy":
			return ast.Type{Name: "byte"}, nil
		case "fileopen":
			return ast.Type{Name: "byte"}, nil
		case "key", "waitkey":
			return ast.Type{Name: "char"}, nil
		case "tickdue", "netconnect", "netreadlf", "netconnected":
			return ast.Type{Name: "bool"}, nil
		default:
			fn, ok := g.functions[e.Name]
			if !ok {
				return ast.Type{}, fmt.Errorf("unknown function %q", e.Name)
			}
			returnTypes := functionReturnTypes(fn)
			if len(returnTypes) > 1 {
				return ast.Type{}, fmt.Errorf("function %s returns multiple values", e.Name)
			}
			if len(returnTypes) == 0 {
				return ast.Type{}, nil
			}
			return returnTypes[0], nil
		}
	}

	return ast.Type{}, fmt.Errorf("unsupported expression")
}

func (g *Generator) fieldInfoType(base ast.Type, field string) (ast.Type, error) {
	t, _, err := g.fieldInfo(base, field)
	return t, err
}

func codegenNumericType(t ast.Type) bool {
	return !t.IsArray && (t.Name == "byte" || t.Name == "char" || t.Name == "int" || t.Name == "uint" || t.Name == "bool")
}

func (g *Generator) emitClsRuntime() {
	if !g.usedClsRuntime {
		return
	}

	g.emit("")
	g.emit("; cls runtime")
	g.emit("peddle_cls:")
	g.emit("    lda #$20")
	g.emit("    ldx #0")

	g.emit("peddle_cls_loop_full:")
	g.emit("    sta $0400, x")
	g.emit("    sta $0500, x")
	g.emit("    sta $0600, x")
	g.emit("    inx")
	g.emit("    bne peddle_cls_loop_full")

	g.emit("    ldx #0")
	g.emit("peddle_cls_loop_last:")
	g.emit("    cpx #232")
	g.emit("    beq peddle_cls_done_clear")
	g.emit("    sta $0700, x")
	g.emit("    inx")
	g.emit("    jmp peddle_cls_loop_last")

	g.emit("peddle_cls_done_clear:")

	// Reset KERNAL text cursor to row 0, column 0.
	// KERNAL PLOT ($fff0), carry clear = set cursor position.
	g.emit("    clc")
	g.emit("    ldx #0")
	g.emit("    ldy #0")
	g.emit("    jsr $fff0")

	g.emit("    rts")
}

func (g *Generator) emitAsciiFontRuntime() {
	if !g.usedAsciiFontRuntime {
		return
	}

	g.emit("")
	g.emit("; ASCII-ish terminal font runtime")
	g.emit("; Copies the C64 ROM charset into RAM at $3800 and patches the")
	g.emit("; PETSCII left-arrow glyph slot so ASCII underscore displays as _.")
	g.emit("peddle_asciifont:")
	g.emit("    php")
	g.emit("    sei")
	g.emit("    lda $01")
	g.emit("    pha")
	g.emit("    lda #$32")
	g.emit("    sta $01")
	g.emit("    ldx #0")
	g.emit("peddle_asciifont_copy:")
	g.emit("    lda $d000, x")
	g.emit("    sta $3800, x")
	g.emit("    lda $d100, x")
	g.emit("    sta $3900, x")
	g.emit("    lda $d200, x")
	g.emit("    sta $3a00, x")
	g.emit("    lda $d300, x")
	g.emit("    sta $3b00, x")
	g.emit("    lda $d400, x")
	g.emit("    sta $3c00, x")
	g.emit("    lda $d500, x")
	g.emit("    sta $3d00, x")
	g.emit("    lda $d600, x")
	g.emit("    sta $3e00, x")
	g.emit("    lda $d700, x")
	g.emit("    sta $3f00, x")
	g.emit("    inx")
	g.emit("    bne peddle_asciifont_copy")
	g.emit("")
	g.emit("    ldx #0")
	g.emit("peddle_asciifont_copy_lowercase:")
	g.emit("    lda $d808, x")
	g.emit("    sta $3a08, x")
	g.emit("    inx")
	g.emit("    cpx #208")
	g.emit("    bne peddle_asciifont_copy_lowercase")
	g.emit("    pla")
	g.emit("    sta $01")
	g.emit("    plp")
	g.emit("")
	g.emit("    ldx #0")
	g.emit("peddle_asciifont_patch_underscore:")
	g.emit("    lda peddle_asciifont_underscore, x")
	g.emit("    sta $38f8, x")
	g.emit("    inx")
	g.emit("    cpx #8")
	g.emit("    bne peddle_asciifont_patch_underscore")
	g.emit("")
	g.emit("    lda $d018")
	g.emit("    and #$f1")
	g.emit("    ora #$0e")
	g.emit("    sta $d018")
	g.emit("    rts")
	g.emit("")
	g.emit("peddle_asciifont_underscore:")
	g.emit("    .byte 0, 0, 0, 0, 0, 0, 0, 255")
}

func (g *Generator) emitAsciiConvertRuntime() {
	if !g.usedAsciiConvertRuntime {
		return
	}

	g.emit("")
	g.emit("; ASCII/PETSCII conversion runtime")
	g.emit("peddle_toascii:")
	g.emit("    jsr peddle_ascii_convert_prepare")
	g.emit("peddle_toascii_loop:")
	g.emit("    lda peddle_ascii_count_lo")
	g.emit("    ora peddle_ascii_count_hi")
	g.emit("    beq peddle_ascii_convert_done")
	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR1_LO), y")
	g.emit("    cmp #65")
	g.emit("    bcc peddle_toascii_store")
	g.emit("    cmp #91")
	g.emit("    bcs peddle_toascii_store")
	g.emit("    clc")
	g.emit("    adc #32")
	g.emit("peddle_toascii_store:")
	g.emit("    sta (ZP_PTR1_LO), y")
	g.emit("    jsr peddle_ascii_convert_next")
	g.emit("    jmp peddle_toascii_loop")
	g.emit("")
	g.emit("peddle_topetscii:")
	g.emit("    jsr peddle_ascii_convert_prepare")
	g.emit("peddle_topetscii_loop:")
	g.emit("    lda peddle_ascii_count_lo")
	g.emit("    ora peddle_ascii_count_hi")
	g.emit("    beq peddle_ascii_convert_done")
	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR1_LO), y")
	g.emit("    cmp #97")
	g.emit("    bcc peddle_topetscii_check_lf")
	g.emit("    cmp #123")
	g.emit("    bcs peddle_topetscii_check_lf")
	g.emit("    sec")
	g.emit("    sbc #32")
	g.emit("    jmp peddle_topetscii_store")
	g.emit("peddle_topetscii_check_lf:")
	g.emit("    cmp #10")
	g.emit("    bne peddle_topetscii_store")
	g.emit("    lda #13")
	g.emit("peddle_topetscii_store:")
	g.emit("    sta (ZP_PTR1_LO), y")
	g.emit("    jsr peddle_ascii_convert_next")
	g.emit("    jmp peddle_topetscii_loop")
	g.emit("")
	g.emit("peddle_ascii_convert_prepare:")
	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_ascii_count_lo")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_ascii_count_hi")
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR1_HI")
	g.emit("    rts")
	g.emit("")
	g.emit("peddle_ascii_convert_next:")
	g.emit("    inc ZP_PTR1_LO")
	g.emit("    bne peddle_ascii_convert_dec")
	g.emit("    inc ZP_PTR1_HI")
	g.emit("peddle_ascii_convert_dec:")
	g.emit("    lda peddle_ascii_count_lo")
	g.emit("    bne peddle_ascii_convert_dec_low")
	g.emit("    dec peddle_ascii_count_hi")
	g.emit("peddle_ascii_convert_dec_low:")
	g.emit("    dec peddle_ascii_count_lo")
	g.emit("    rts")
	g.emit("")
	g.emit("peddle_ascii_convert_done:")
	g.emit("    rts")
	g.emit("")
	g.emit("peddle_ascii_count_lo:")
	g.emit("    .byte 0")
	g.emit("peddle_ascii_count_hi:")
	g.emit("    .byte 0")
}

func (g *Generator) emitCharToScreenTable() {
	if !g.usedCharToScreenTable && !g.usedPutStrRuntime && !g.usedPutStrColorRuntime {
		return
	}

	table := make([]int, 256)
	for i := 0; i < 256; i++ {
		table[i] = i
	}

	// C64 screen RAM does not use PETSCII/ASCII codes directly.
	//
	// Common useful mappings:
	//   '@'       PETSCII/ASCII 64      -> screen code 0
	//   'A'..'Z' PETSCII/ASCII 65..90   -> screen codes 1..26
	//   'a'..'z' PETSCII/ASCII 97..122  -> screen codes 1..26
	//
	// Other values currently keep identity mapping. This preserves existing
	// behavior for digits, spaces, and common punctuation while giving us one
	// central place to improve the PETSCII-to-screen mapping later.
	table[64] = 0

	for ch := 65; ch <= 90; ch++ {
		table[ch] = ch - 64
	}

	for ch := 97; ch <= 122; ch++ {
		table[ch] = ch - 96
	}

	g.emit("")
	g.emit("; character to C64 screen code lookup table")
	g.emit("peddle_char_to_screen_table:")

	for i := 0; i < 256; i += 16 {
		g.emit(fmt.Sprintf(
			"    .byte %d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
			table[i+0],
			table[i+1],
			table[i+2],
			table[i+3],
			table[i+4],
			table[i+5],
			table[i+6],
			table[i+7],
			table[i+8],
			table[i+9],
			table[i+10],
			table[i+11],
			table[i+12],
			table[i+13],
			table[i+14],
			table[i+15],
		))
	}
}

func (g *Generator) emitScreenBuiltinRuntime() {
	if g.usedGotoXYRuntime {
		g.emit(`
; gotoxy runtime
peddle_gotoxy:
    lda peddle_tmp_int0
    cmp #40
    bcs peddle_gotoxy_done
    lda peddle_tmp_int0+1
    cmp #25
    bcs peddle_gotoxy_done
    clc
    ldx peddle_tmp_int0+1
    ldy peddle_tmp_int0
    jsr $fff0
peddle_gotoxy_done:
    rts
`)
	}

	if !g.usedPutRawRuntime && !g.usedPutCharRuntime && !g.usedPutColorRuntime && !g.usedPutCharColorRuntime {
		return
	}

	g.emit(`
; single-cell screen/color runtime
peddle_screen_addr:
    lda peddle_tmp_int0
    cmp #40
    bcs peddle_screen_addr_clipped
    lda peddle_tmp_int0+1
    cmp #25
    bcs peddle_screen_addr_clipped
    ldx peddle_tmp_int0+1
peddle_screen_addr_row_loop:
    beq peddle_screen_addr_done
    lda ZP_PTR0_LO
    clc
    adc #40
    sta ZP_PTR0_LO
    lda ZP_PTR0_HI
    adc #0
    sta ZP_PTR0_HI
    dex
    jmp peddle_screen_addr_row_loop
peddle_screen_addr_done:
    clc
    rts
peddle_screen_addr_clipped:
    sec
    rts
`)

	if g.usedPutRawRuntime || g.usedPutCharRuntime || g.usedPutColorRuntime {
		g.emit(`
peddle_putscreen_byte:
    jsr peddle_screen_addr
    bcs peddle_putscreen_byte_done
    ldy peddle_tmp_int0
    lda ZP_TMP0
    sta (ZP_PTR0_LO), y
peddle_putscreen_byte_done:
    rts
`)
	}

	if g.usedPutRawRuntime {
		g.emit(`
peddle_putraw:
    lda #<$0400
    sta ZP_PTR0_LO
    lda #>$0400
    sta ZP_PTR0_HI
    jmp peddle_putscreen_byte
`)
	}

	if g.usedPutColorRuntime {
		g.emit(`
peddle_putcolor:
    lda #<$d800
    sta ZP_PTR0_LO
    lda #>$d800
    sta ZP_PTR0_HI
    jmp peddle_putscreen_byte
`)
	}

	if g.usedPutCharRuntime {
		g.emit(`
peddle_putchar:
    lda ZP_TMP0
    tax
    lda peddle_char_to_screen_table, x
    sta ZP_TMP0
    lda #<$0400
    sta ZP_PTR0_LO
    lda #>$0400
    sta ZP_PTR0_HI
    jmp peddle_putscreen_byte
`)
	}

	if g.usedPutCharColorRuntime {
		g.emit(`
peddle_putcharcolor:
    lda ZP_TMP0
    tax
    lda peddle_char_to_screen_table, x
    sta ZP_TMP0
    lda #<$0400
    sta ZP_PTR0_LO
    lda #>$0400
    sta ZP_PTR0_HI
    jsr peddle_screen_addr
    bcs peddle_putcharcolor_done
    ldy peddle_tmp_int0
    lda ZP_TMP0
    sta (ZP_PTR0_LO), y
    lda ZP_PTR0_HI
    clc
    adc #212
    sta ZP_PTR0_HI
    lda ZP_TMP1
    sta (ZP_PTR0_LO), y
peddle_putcharcolor_done:
    rts
`)
	}
}

func (g *Generator) emitReadLineRuntime() {
	if !g.usedReadLineRuntime {
		return
	}

	g.emit("")
	g.emit("; readline runtime")
	g.emit("peddle_readline_echo:")
	g.emit("    .byte 0")
	g.emit("peddle_readline_max:")
	g.emit("    .word 0")
	g.emit("peddle_readline_limit:")
	g.emit("    .word 0")
	g.emit("peddle_readline_len:")
	g.emit("    .word 0")
	g.emit("peddle_readline_char:")
	g.emit("    .byte 0")

	g.emit("")
	g.emit("peddle_readline:")

	// Clear runtime length in the destination array header.
	g.emit("    ldy #2")
	g.emit("    lda #0")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    iny")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    sta peddle_readline_len")
	g.emit("    sta peddle_readline_len+1")

	// Start with limit = array capacity.
	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_readline_limit")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_readline_limit+1")

	// If max < capacity, use max as the effective limit.
	g.emit("    lda peddle_readline_max+1")
	g.emit("    cmp peddle_readline_limit+1")
	g.emit("    bcc peddle_readline_use_max")
	g.emit("    bne peddle_readline_limit_done")
	g.emit("    lda peddle_readline_max")
	g.emit("    cmp peddle_readline_limit")
	g.emit("    bcc peddle_readline_use_max")
	g.emit("    jmp peddle_readline_limit_done")

	g.emit("peddle_readline_use_max:")
	g.emit("    lda peddle_readline_max")
	g.emit("    sta peddle_readline_limit")
	g.emit("    lda peddle_readline_max+1")
	g.emit("    sta peddle_readline_limit+1")

	g.emit("peddle_readline_limit_done:")

	// ZP_PTR1 points at the next destination data byte.
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit("    sta ZP_PTR1_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR1_HI")

	g.emit("peddle_readline_loop:")
	g.emit("    jsr $ffe4")
	g.emit("    cmp #0")
	g.emit("    beq peddle_readline_loop")
	g.emit("    cmp #13")
	g.emit("    beq peddle_readline_done")
	g.emit("    sta peddle_readline_char")

	// Ignore additional characters once the effective limit is reached.
	g.emit("    lda peddle_readline_len+1")
	g.emit("    cmp peddle_readline_limit+1")
	g.emit("    bcc peddle_readline_has_space")
	g.emit("    bne peddle_readline_loop")
	g.emit("    lda peddle_readline_len")
	g.emit("    cmp peddle_readline_limit")
	g.emit("    bcc peddle_readline_has_space")
	g.emit("    jmp peddle_readline_loop")

	g.emit("peddle_readline_has_space:")
	g.emit("    ldy #0")
	g.emit("    lda peddle_readline_char")
	g.emit("    sta (ZP_PTR1_LO), y")

	// Echo accepted characters when requested.
	g.emit("    lda peddle_readline_echo")
	g.emit("    beq peddle_readline_no_echo")
	g.emit("    lda peddle_readline_char")
	g.emit("    jsr $ffd2")
	g.emit("peddle_readline_no_echo:")

	// Advance destination pointer.
	g.emit("    inc ZP_PTR1_LO")
	g.emit("    bne peddle_readline_ptr_no_carry")
	g.emit("    inc ZP_PTR1_HI")
	g.emit("peddle_readline_ptr_no_carry:")

	// Increment current length.
	g.emit("    inc peddle_readline_len")
	g.emit("    bne peddle_readline_loop")
	g.emit("    inc peddle_readline_len+1")
	g.emit("    jmp peddle_readline_loop")

	g.emit("peddle_readline_done:")
	g.emit("    ldy #2")
	g.emit("    lda peddle_readline_len")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP0")
	g.emit("    iny")
	g.emit("    lda peddle_readline_len+1")
	g.emit("    sta (ZP_PTR0_LO), y")
	g.emit("    sta ZP_TMP1")
	g.emit("    lda ZP_TMP0")
	g.emit("    rts")
}

func (g *Generator) emitPutStrRuntime() {
	if !g.usedPutStrRuntime && !g.usedPutStrColorRuntime {
		return
	}

	g.emit("")
	g.emit("; putstr runtime")
	g.emit("peddle_putstr_x:")
	g.emit("    .byte 0")
	g.emit("peddle_putstr_y:")
	g.emit("    .byte 0")
	g.emit("peddle_putstr_start_x:")
	g.emit("    .byte 0")
	g.emit("peddle_putstr_color:")
	g.emit("    .byte 0")
	g.emit("peddle_putstr_write_color:")
	g.emit("    .byte 0")
	g.emit("peddle_putstr_char:")
	g.emit("    .byte 0")

	g.emit("")
	g.emit("peddle_putstr:")
	g.emit("    lda #0")
	g.emit("    sta peddle_putstr_write_color")
	g.emit("    jmp peddle_putstr_common")

	if g.usedPutStrColorRuntime {
		g.emit("")
		g.emit("peddle_putstrcolor:")
		g.emit("    lda #1")
		g.emit("    sta peddle_putstr_write_color")
		g.emit("    jmp peddle_putstr_common")
	}

	g.emit("")
	g.emit("peddle_putstr_common:")
	g.emit("    lda peddle_putstr_x")
	g.emit("    sta peddle_putstr_start_x")
	g.emit("    sta ZP_TMP0")
	g.emit("    lda peddle_putstr_y")
	g.emit("    sta ZP_TMP1")

	// Clip invalid start x.
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emit("    bcc peddle_putstr_start_x_ok")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_start_x_ok:")

	// Clip invalid start y.
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp #25")
	g.emit("    bcc peddle_putstr_start_y_ok")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_start_y_ok:")

	g.emit("")
	g.emit("peddle_putstr_loop:")
	g.emit("    lda peddle_tmp_int0")
	g.emit("    ora peddle_tmp_int0+1")
	g.emit("    bne peddle_putstr_has_chars")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_has_chars:")

	// Load next source character and preserve it before changing counters.
	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR1_LO), y")
	g.emit("    sta peddle_putstr_char")

	// Advance source pointer.
	g.emit("    inc ZP_PTR1_LO")
	g.emit("    bne peddle_putstr_src_no_carry")
	g.emit("    inc ZP_PTR1_HI")
	g.emit("peddle_putstr_src_no_carry:")

	// Decrement remaining length.
	g.emit("    lda peddle_tmp_int0")
	g.emit("    bne peddle_putstr_dec_low")
	g.emit("    dec peddle_tmp_int0+1")
	g.emit("peddle_putstr_dec_low:")
	g.emit("    dec peddle_tmp_int0")

	// Newline / carriage return handling.
	g.emit("    lda peddle_putstr_char")
	g.emit("    cmp #13")
	g.emit("    bne peddle_putstr_not_newline")
	g.emit("    jmp peddle_putstr_newline")
	g.emit("peddle_putstr_not_newline:")

	// Convert character code to C64 screen code through the shared table.
	g.emit("    lda peddle_putstr_char")
	g.emit("    tax")
	g.emit("    lda peddle_char_to_screen_table, x")
	g.emit("    sta peddle_putstr_char")

	// Clip current x.
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emit("    bcc peddle_putstr_current_x_ok")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_current_x_ok:")

	// Clip current y.
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp #25")
	g.emit("    bcc peddle_putstr_current_y_ok")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_current_y_ok:")

	// Compute screen pointer into ZP_PTR0.
	g.emit("    lda #<$0400")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda #>$0400")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldx ZP_TMP1")
	g.emit("peddle_putstr_screen_row_loop:")
	g.emit("    beq peddle_putstr_screen_row_done")
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #40")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    dex")
	g.emit("    jmp peddle_putstr_screen_row_loop")

	g.emit("peddle_putstr_screen_row_done:")
	g.emit("    ldy ZP_TMP0")
	g.emit("    lda peddle_putstr_char")
	g.emit("    sta (ZP_PTR0_LO), y")

	// Optional color write.
	g.emit("    lda peddle_putstr_write_color")
	g.emit("    beq peddle_putstr_advance")

	g.emit("    lda #<$d800")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda #>$d800")
	g.emit("    sta ZP_PTR0_HI")

	g.emit("    ldx ZP_TMP1")
	g.emit("peddle_putstr_color_row_loop:")
	g.emit("    beq peddle_putstr_color_row_done")
	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #40")
	g.emit("    sta ZP_PTR0_LO")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit("    sta ZP_PTR0_HI")
	g.emit("    dex")
	g.emit("    jmp peddle_putstr_color_row_loop")

	g.emit("peddle_putstr_color_row_done:")
	g.emit("    ldy ZP_TMP0")
	g.emit("    lda peddle_putstr_color")
	g.emit("    sta (ZP_PTR0_LO), y")

	// Advance current screen position.
	g.emit("peddle_putstr_advance:")
	g.emit("    inc ZP_TMP0")
	g.emit("    lda ZP_TMP0")
	g.emit("    cmp #40")
	g.emit("    bcs peddle_putstr_wrap_line")
	g.emit("    jmp peddle_putstr_loop")

	g.emit("peddle_putstr_wrap_line:")
	g.emit("    lda #0")
	g.emit("    sta ZP_TMP0")
	g.emit("    inc ZP_TMP1")
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp #25")
	g.emit("    bcc peddle_putstr_continue_after_wrap")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_continue_after_wrap:")
	g.emit("    jmp peddle_putstr_loop")

	// Newline: x = start_x, y++.
	g.emit("peddle_putstr_newline:")
	g.emit("    lda peddle_putstr_start_x")
	g.emit("    sta ZP_TMP0")
	g.emit("    inc ZP_TMP1")
	g.emit("    lda ZP_TMP1")
	g.emit("    cmp #25")
	g.emit("    bcc peddle_putstr_continue_after_newline")
	g.emit("    jmp peddle_putstr_done")
	g.emit("peddle_putstr_continue_after_newline:")
	g.emit("    jmp peddle_putstr_loop")

	g.emit("peddle_putstr_done:")
	g.emit("    rts")
}
