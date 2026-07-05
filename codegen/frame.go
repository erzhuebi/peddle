package codegen

import (
	"fmt"
	"strings"

	"peddle/ast"
)

func (g *Generator) buildGlobalSymbol(v *ast.VarDecl) Symbol {
	return Symbol{
		SourceName:   v.Name,
		Label:        "peddle_global_" + sanitizeLabel(v.Name),
		Type:         v.Type,
		Size:         g.sizeof(v.Type),
		HasAtAddress: v.HasAtAddress,
		AtAddress:    v.AtAddress,
	}
}

func (g *Generator) buildFrame(fn *ast.FunctionDecl) *Frame {
	frame := &Frame{
		FunctionName: fn.Name,
		Symbols:      map[string]Symbol{},
	}

	for _, p := range fn.Params {
		size := g.sizeof(p.Type)
		isReference := p.Type.IsArray || p.Type.IsMem
		if isReference {
			size = 2
		}

		frame.Symbols[p.Name] = Symbol{
			SourceName:  p.Name,
			Label:       fn.Name + "_" + p.Name,
			Type:        p.Type,
			Size:        size,
			IsReference: isReference,
		}
	}

	for _, l := range fn.Locals {
		frame.Symbols[l.Name] = Symbol{
			SourceName:   l.Name,
			Label:        fn.Name + "_" + l.Name,
			Type:         l.Type,
			Size:         g.sizeof(l.Type),
			HasAtAddress: l.HasAtAddress,
			AtAddress:    l.AtAddress,
		}
	}

	returnTypes := functionReturnTypes(fn)
	for i, returnType := range returnTypes {
		label := fn.Name + "_return"
		if len(returnTypes) > 1 {
			label = fmt.Sprintf("%s_return_%d", fn.Name, i)
		}

		frame.Returns = append(frame.Returns, Symbol{
			SourceName: fmt.Sprintf("return_%d", i),
			Label:      label,
			Type:       returnType,
			Size:       g.sizeof(returnType),
		})
	}
	if len(frame.Returns) > 0 {
		frame.Return = &frame.Returns[0]
	}

	return frame
}

func (g *Generator) resolve(name string) (Symbol, bool) {
	if g.currentFrame != nil {
		if sym, ok := g.currentFrame.Symbols[name]; ok {
			return sym, true
		}
	}

	sym, ok := g.globals[name]
	return sym, ok
}

func (g *Generator) sizeof(t ast.Type) int {
	if t.IsMem {
		return 0
	}

	if t.IsPointer {
		return 2
	}

	base := 0

	switch t.Name {
	case "byte", "char", "bool":
		base = 1
	case "int", "uint":
		base = 2
	default:
		if s, ok := g.structs[t.Name]; ok {
			for _, f := range s.Fields {
				base += g.sizeof(f.Type)
			}
		} else {
			base = 1
		}
	}

	if t.IsArray {
		return 4 + base*t.ArrayLen
	}

	return base
}

func (g *Generator) fieldInfo(base ast.Type, field string) (ast.Type, int, error) {
	if base.IsPointer {
		base = ast.Type{Name: base.Name, IsArray: base.IsArray, ArrayLen: base.ArrayLen}
	}

	if base.IsMem {
		return ast.Type{}, 0, fmt.Errorf("cannot access field %q on mem type %s", field, base.String())
	}

	if base.IsArray {
		return ast.Type{}, 0, fmt.Errorf("cannot access field %q on array type %s", field, base.String())
	}

	s, ok := g.structs[base.Name]
	if !ok {
		return ast.Type{}, 0, fmt.Errorf("type %s has no fields", base.Name)
	}

	offset := 0

	for _, f := range s.Fields {
		if f.Name == field {
			return f.Type, offset, nil
		}

		offset += g.sizeof(f.Type)
	}

	return ast.Type{}, 0, fmt.Errorf("type %s has no field %q", base.Name, field)
}

func sanitizeLabel(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}
