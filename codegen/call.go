package codegen

import (
	"fmt"
	"strconv"

	"peddle/ast"
)

func (g *Generator) genCall(name string, args []ast.Expr) (ast.Type, error) {
	switch name {
	case "print":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("print expects one argument")
		}
		g.usedPrint = true

		switch arg := args[0].(type) {
		case *ast.StringExpr:
			label := g.addLiteral(arg.Value)
			g.emit(fmt.Sprintf("    lda #<%s", label))
			g.emit("    sta ZP_PTR0_LO")
			g.emit(fmt.Sprintf("    lda #>%s", label))
			g.emit("    sta ZP_PTR0_HI")
			g.emit("    jsr peddle_print_string")
			return ast.Type{}, nil

		case *ast.IdentExpr:
			sym, ok := g.resolve(arg.Name)
			if !ok {
				return ast.Type{}, fmt.Errorf("unknown variable %q", arg.Name)
			}
			g.emit(fmt.Sprintf("    lda #<%s", sym.Label))
			g.emit("    sta ZP_PTR0_LO")
			g.emit(fmt.Sprintf("    lda #>%s", sym.Label))
			g.emit("    sta ZP_PTR0_HI")
			g.emit("    jsr peddle_print_string")
			return ast.Type{}, nil

		default:
			return ast.Type{}, fmt.Errorf("unsupported print argument")
		}

	case "poke":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("poke expects two arguments")
		}

		addrNum, ok := args[0].(*ast.NumberExpr)
		if !ok {
			return ast.Type{}, fmt.Errorf("poke currently requires constant address")
		}

		value := args[1]
		if err := g.genExprTo(value, ast.Type{Name: "byte"}); err != nil {
			return ast.Type{}, err
		}

		addr, err := strconv.Atoi(addrNum.Value)
		if err != nil {
			return ast.Type{}, err
		}

		g.emit(fmt.Sprintf("    sta $%04x", addr))
		return ast.Type{}, nil
	}

	fn, ok := g.functions[name]
	if !ok {
		return ast.Type{}, fmt.Errorf("unknown function %q", name)
	}

	frame := g.frames[name]

	for i, arg := range args {
		param := fn.Params[i]
		sym := frame.Symbols[param.Name]

		if err := g.genExprTo(arg, param.Type); err != nil {
			return ast.Type{}, err
		}

		g.storeAIntoSymbol(sym)
	}

	g.emit(fmt.Sprintf("    jsr %s", name))
	return fn.ReturnType, nil
}
