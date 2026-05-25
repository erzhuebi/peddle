package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genNetConnect(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("netconnect expects two arguments")
	}

	addrType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if !(addrType.IsArray && addrType.Name == "char") {
		return ast.Type{}, fmt.Errorf("netconnect address must be char array")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_net_addr_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_net_addr_hi")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_net_port_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_net_port_hi")

	g.emit("    jsr peddle_netconnect")

	g.usedNetRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "bool"}, nil
}

func (g *Generator) genNetRead(args []ast.Expr) (ast.Type, error) {
	if len(args) != 3 {
		return ast.Type{}, fmt.Errorf("netread expects three arguments")
	}

	bufferType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
		return ast.Type{}, fmt.Errorf("netread buffer must be byte array or char array")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_net_buf_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_net_buf_hi")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_net_max_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_net_max_hi")

	if err := g.genExprTo(args[2], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_net_timeout_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_net_timeout_hi")

	g.emit("    jsr peddle_netread")

	g.usedNetRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genNetWrite(args []ast.Expr) (ast.Type, error) {
	if len(args) != 2 {
		return ast.Type{}, fmt.Errorf("netwrite expects two arguments")
	}

	bufferType, err := g.arrayExprType(args[0])
	if err != nil {
		return ast.Type{}, err
	}
	if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
		return ast.Type{}, fmt.Errorf("netwrite buffer must be byte array or char array")
	}

	if err := g.genArrayAddress(args[0]); err != nil {
		return ast.Type{}, err
	}

	g.emit("    lda ZP_PTR0_LO")
	g.emit("    sta peddle_net_buf_lo")
	g.emit("    lda ZP_PTR0_HI")
	g.emit("    sta peddle_net_buf_hi")

	if err := g.genExprTo(args[1], ast.Type{Name: "int"}); err != nil {
		return ast.Type{}, err
	}
	g.emit("    lda ZP_TMP0")
	g.emit("    sta peddle_net_max_lo")
	g.emit("    lda ZP_TMP1")
	g.emit("    sta peddle_net_max_hi")

	g.emit("    jsr peddle_netwrite")

	g.usedNetRuntime = true
	g.usedTmp16 = true

	return ast.Type{Name: "int"}, nil
}

func (g *Generator) genNetClose(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("netclose expects no arguments")
	}

	g.emit("    jsr peddle_netclose")
	g.usedNetRuntime = true

	return ast.Type{}, nil
}

func (g *Generator) genNetConnected(args []ast.Expr) (ast.Type, error) {
	if len(args) != 0 {
		return ast.Type{}, fmt.Errorf("netconnected expects no arguments")
	}

	g.emit("    lda peddle_net_connected")
	g.usedNetRuntime = true

	return ast.Type{Name: "bool"}, nil
}
