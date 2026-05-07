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

	case "len":
		return g.genLen(args)

	case "size":
		return g.genSize(args)

	case "append":
		return g.genAppend(args)

	case "copy":
		return g.genCopy(args)

	case "fill":
		return g.genFill(args)

	case "clear":
		return g.genClear(args)
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
