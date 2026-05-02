package codegen

import "peddle/ast"

func (g *Generator) buildFrame(fn *ast.FunctionDecl) *Frame {
	frame := &Frame{
		FunctionName: fn.Name,
		Symbols:      map[string]Symbol{},
	}

	for _, p := range fn.Params {
		frame.Symbols[p.Name] = Symbol{
			SourceName: p.Name,
			Label:      fn.Name + "_" + p.Name,
			Type:       p.Type,
			Size:       sizeof(p.Type),
		}
	}

	for _, l := range fn.Locals {
		frame.Symbols[l.Name] = Symbol{
			SourceName: l.Name,
			Label:      fn.Name + "_" + l.Name,
			Type:       l.Type,
			Size:       sizeof(l.Type),
		}
	}

	if fn.ReturnType.Name != "" {
		frame.Return = &Symbol{
			SourceName: "return",
			Label:      fn.Name + "_return",
			Type:       fn.ReturnType,
			Size:       sizeof(fn.ReturnType),
		}
	}

	return frame
}

func (g *Generator) resolve(name string) (Symbol, bool) {
	if g.currentFrame == nil {
		return Symbol{}, false
	}
	sym, ok := g.currentFrame.Symbols[name]
	return sym, ok
}

func sizeof(t ast.Type) int {
	base := 0

	switch t.Name {
	case "byte", "char", "bool":
		base = 1
	case "int":
		base = 2
	default:
		base = 1
	}

	if t.IsArray {
		return base * t.ArrayLen
	}

	return base
}
