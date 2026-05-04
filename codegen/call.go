package codegen

import (
	"fmt"

	"peddle/ast"
)

func (g *Generator) genCall(name string, args []ast.Expr) (ast.Type, error) {
	switch name {
	case "print":
		return g.genPrint(args)

	case "poke":
		return g.genPoke(args)

	case "peek":
		return g.genPeek(args)

	case "strlen":
		return g.genStrlen(args)

	case "strcpy":
		return g.genStrcpy(args)

	case "stradd":
		return g.genStradd(args)
	}

	fn, ok := g.functions[name]
	if !ok {
		return ast.Type{}, fmt.Errorf("unknown function %q", name)
	}

	if len(args) != len(fn.Params) {
		return ast.Type{}, fmt.Errorf("function %s expects %d args, got %d", name, len(fn.Params), len(args))
	}

	for i, arg := range args {
		if err := g.genExprTo(arg, fn.Params[i].Type); err != nil {
			return ast.Type{}, err
		}

		sym := g.frames[name].Symbols[fn.Params[i].Name]
		g.storeAIntoSymbol(sym)
	}

	g.emit(fmt.Sprintf("    jsr %s", name))
	return fn.ReturnType, nil
}
