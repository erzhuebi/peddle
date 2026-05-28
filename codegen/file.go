package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genFileOpen(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("fileopen expects three arguments")
	}

	if err := g.genFileStringArg(args[0], "peddle_file_name"); err != nil {
		return ast.Type{}, err
	}
	if err := g.genFileStringArg(args[1], "peddle_file_mode"); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_device")

	g.emit("    jsr peddle_fileopen")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "byte"}, nil
}

func (g *Generator) genFileClose(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("fileclose expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_handle")

	g.emit("    jsr peddle_fileclose")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{}, nil
}

func (g *Generator) genFileRead(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("fileread expects three arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_handle")

	if err := g.genFileBufferArg(args[1]); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_max_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_file_max_hi")

	g.emit("    jsr peddle_fileread")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genFileWrite(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("filewrite expects three arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_handle")

	if err := g.genFileBufferArg(args[1]); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_max_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_file_max_hi")

	g.emit("    jsr peddle_filewrite")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genFileLoad(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("fileload expects three arguments")
	}

	if err := g.genFileStringArg(args[0], "peddle_file_name"); err != nil {
		return ast.Type{}, err
	}
	if err := g.genFileBufferArg(args[1]); err != nil {
		return ast.Type{}, err
	}
	if err := g.genFileBufferCapacityToMax(args[1]); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_device")

	g.emit("    jsr peddle_fileload")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genFileSave(args []ast.Expr) (ast.Type, error) {
	if len(args) != 4 {
		return ast.Type{}, fmt.Errorf("filesave expects four arguments")
	}

	if err := g.genFileStringArg(args[0], "peddle_file_name"); err != nil {
		return ast.Type{}, err
	}
	if err := g.genFileBufferArg(args[1]); err != nil {
		return ast.Type{}, err
	}

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_max_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_file_max_hi")

	if err := g.genExprTo(args[3], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_file_device")

	g.emit("    jsr peddle_filesave")

	g.usedFileRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genFileStringArg(arg ast.Expr, prefix string) error {
	if str, ok := arg.(*ast.StringExpr); ok {
		label := g.addLiteral(str.Value)

		g.emit(fmt.Sprintf("    lda #<%s", label))
		g.emit(fmt.Sprintf("    sta %s_lo", prefix))
		g.emit(fmt.Sprintf("    lda #>%s", label))
		g.emit(fmt.Sprintf("    sta %s_hi", prefix))
		g.emit(fmt.Sprintf("    lda #<%d", len(str.Value)))
		g.emit(fmt.Sprintf("    sta %s_len_lo", prefix))
		g.emit(fmt.Sprintf("    lda #>%d", len(str.Value)))
		g.emit(fmt.Sprintf("    sta %s_len_hi", prefix))

		g.usedTmp16 = true
		return nil
	}

	t, err := g.arrayExprType(arg)
	if err != nil {
		return err
	}
	if !(t.IsArray && t.Name == "char") {
		return fmt.Errorf("file string argument must be char array")
	}

	if err := g.genArrayAddress(arg); err != nil {
		return err
	}

	g.emit("    ldy #2")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit(fmt.Sprintf("    sta %s_len_lo", prefix))
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit(fmt.Sprintf("    sta %s_len_hi", prefix))

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    clc")
	g.emit("    adc #4")
	g.emit(fmt.Sprintf("    sta %s_lo", prefix))
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    adc #0")
	g.emit(fmt.Sprintf("    sta %s_hi", prefix))

	g.usedTmp16 = true
	return nil
}

func (g *Generator) genFileBufferArg(arg ast.Expr) error {
	t, err := g.arrayExprType(arg)
	if err != nil {
		return err
	}
	if !(t.IsArray && (t.Name == "byte" || t.Name == "char")) {
		return fmt.Errorf("file buffer must be byte array or char array")
	}

	if err := g.genArrayAddress(arg); err != nil {
		return err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_file_buf_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_file_buf_hi")

	return nil
}

func (g *Generator) genFileBufferCapacityToMax(arg ast.Expr) error {
	if err := g.genArrayAddress(arg); err != nil {
		return err
	}

	g.emit("    ldy #0")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_file_max_lo")
	g.emit("    iny")
	g.emit("    lda (ZP_PTR0_LO), y")
	g.emit("    sta peddle_file_max_hi")
	g.usedTmp16 = true
	return nil
}
