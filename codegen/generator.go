package codegen

import (
	"strings"

	"peddle/ast"
)

type OptMode string

const (
	OptModeSpeed OptMode = "speed"
	OptModeSize  OptMode = "size"
)

type Options struct {
	OptMode OptMode
}

type Generator struct {
	out strings.Builder

	options Options

	functions map[string]*ast.FunctionDecl
	structs   map[string]*ast.StructDecl
	frames    map[string]*Frame

	currentFn    *ast.FunctionDecl
	currentFrame *Frame

	literals []string

	usedPrint bool
	usedTmp16 bool

	usedArrayCopyRuntime    bool
	usedFillByteRuntime     bool
	usedFillIntRuntime      bool
	usedAppendByteRuntime   bool
	usedAppendIntRuntime    bool
	usedStringCopyRuntime   bool
	usedStringAppendRuntime bool

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
	return NewWithOptions(Options{
		OptMode: OptModeSpeed,
	})
}

func NewWithOptions(options Options) *Generator {
	if options.OptMode == "" {
		options.OptMode = OptModeSpeed
	}

	return &Generator{
		options:   options,
		functions: map[string]*ast.FunctionDecl{},
		structs:   map[string]*ast.StructDecl{},
		frames:    map[string]*Frame{},
	}
}

func (g *Generator) Generate(p *ast.Program) (string, error) {
	for _, s := range p.Structs {
		g.structs[s.Name] = s
	}

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
