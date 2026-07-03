package codegen

import (
	"fmt"

	"peddle/ast"
)

func isSoundBuiltin(name string) bool {
	switch name {
	case "sound_init", "sound_reset", "sound_load", "sound_play", "sound_stop", "sound_stop_voices", "sound_num", "sound_memfree":
		return true
	default:
		return false
	}
}

func (g *Generator) builtinReturnSymbols(name string) ([]Symbol, bool) {
	switch name {
	case "sound_load":
		return []Symbol{
			{
				SourceName: "sound_load_id",
				Label:      "peddle_sound_load_return_id",
				Type:       ast.Type{Name: "uint"},
				Size:       2,
			},
			{
				SourceName: "sound_load_err",
				Label:      "peddle_sound_load_return_err",
				Type:       ast.Type{Name: "int"},
				Size:       2,
			},
		}, true
	default:
		return nil, false
	}
}

func (g *Generator) genSoundInit(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("sound_init expects one argument")
	}

	t, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if !(t.IsArray && t.Name == "byte") {
		return ast.Type{}, fmt.Errorf("sound_init pool must be byte array")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_sound_pool_header_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_sound_pool_header_hi")
	g.emit("    jsr peddle_sound_init")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{}, nil
}

func (g *Generator) genSoundReset(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("sound_reset expects no arguments")
	}

	g.emit("    jsr peddle_sound_reset")

	g.usedSoundRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genSoundLoad(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("sound_load expects two arguments")
	}

	t, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if !(t.IsArray && t.Name == "byte") {
		return ast.Type{}, fmt.Errorf("sound_load data must be byte array")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_sound_data_header_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_sound_data_header_hi")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_kind_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_kind_hi")

	g.emit("    jsr peddle_sound_load")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "uint"}, nil
}

func (g *Generator) genSoundPlay(args []ast.Expr) (ast.Type, error) {
	if len(args) != 4 {
		return ast.Type{}, fmt.Errorf("sound_play expects four arguments")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_handle_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_handle_hi")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_play_voices_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_play_voices_hi")

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_play_priority_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_play_priority_hi")

	if err := g.genExprTo(args[3], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_play_flags_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_play_flags_hi")
	g.emit("    jsr peddle_sound_play")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genSoundStop(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("sound_stop expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_handle_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_handle_hi")
	g.emit("    jsr peddle_sound_stop")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{}, nil
}

func (g *Generator) genSoundStopVoices(args []ast.Expr) (ast.Type, error) {
	if len(args) != 1 {
		return ast.Type{}, fmt.Errorf("sound_stop_voices expects one argument")
	}

	if err := g.genExprTo(args[0], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_sound_stop_voices_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_sound_stop_voices_hi")
	g.emit("    jsr peddle_sound_stop_voices")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{}, nil
}

func (g *Generator) genSoundNum(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("sound_num expects no arguments")
	}

	g.emit("    jsr peddle_sound_num")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genSoundMemFree(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("sound_memfree expects no arguments")
	}

	g.emit("    jsr peddle_sound_memfree")

	g.usedSoundRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}
