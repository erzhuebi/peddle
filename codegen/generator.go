package codegen

import (
	"strings"

	"peddle/ast"
)

type Generator struct {
	out strings.Builder

	functions map[string]*ast.FunctionDecl
	frames    map[string]*Frame

	currentFn    *ast.FunctionDecl
	currentFrame *Frame

	literals []string

	usedPrint bool
	usedTmp16 bool

	labelCounter int
}

type Frame struct {
	FunctionName string
	Symbols      map[string]Symbol
	Return       *Symbol
}

type Symbol struct {
	SourceName string
	Label      string
	Type       ast.Type
	Size       int
}

func New() *Generator {
	return &Generator{
		functions: map[string]*ast.FunctionDecl{},
		frames:    map[string]*Frame{},
	}
}

func (g *Generator) Generate(p *ast.Program) (string, error) {
	for _, fn := range p.Functions {
		g.functions[fn.Name] = fn
		g.frames[fn.Name] = g.buildFrame(fn)
	}

	g.emitHeader()

	for _, fn := range p.Functions {
		if err := g.genFunction(fn); err != nil {
			return "", err
		}
	}

	g.emitRuntime()
	g.emitLiterals()
	g.emitStaticData()

	return g.out.String(), nil
}

func (g *Generator) genFunction(fn *ast.FunctionDecl) error {
	g.currentFn = fn
	g.currentFrame = g.frames[fn.Name]

	g.emit(fn.Name + ":")

	for _, stmt := range fn.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}

	g.emit("    rts")
	g.emit("")
	return nil
}
