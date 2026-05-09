package codegen

import (
	"fmt"
	"strings"

	"peddle/ast"
)

type OptMode string

const (
	OptModeSpeed OptMode = "speed"
	OptModeSize  OptMode = "size"
)

type Options struct {
	OptMode           OptMode
	StaticMemoryLimit int
}

type MemoryReport struct {
	StaticDataBytes   int
	LiteralBytes      int
	StaticSymbolCount int
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
	usedMulByteRuntime      bool
	usedMulIntRuntime       bool
	usedShlByteRuntime      bool
	usedShrByteRuntime      bool
	usedShlIntRuntime       bool
	usedShrIntRuntime       bool
	usedStringCopyRuntime   bool
	usedStringAppendRuntime bool

	memoryReport MemoryReport

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

	g.memoryReport = g.computeMemoryReport()

	if g.options.StaticMemoryLimit > 0 && g.memoryReport.StaticDataBytes > g.options.StaticMemoryLimit {
		return "", fmt.Errorf("static memory usage %d bytes exceeds limit %d bytes", g.memoryReport.StaticDataBytes, g.options.StaticMemoryLimit)
	}

	g.emitRuntime()
	g.emitLiterals()
	g.emitStaticData()

	return g.out.String(), nil
}

func (g *Generator) MemoryReport() MemoryReport {
	return g.memoryReport
}

func (g *Generator) computeMemoryReport() MemoryReport {
	report := MemoryReport{}

	for _, frame := range g.frames {
		for _, sym := range frame.Symbols {
			report.StaticDataBytes += sym.Size
			report.StaticSymbolCount++
		}

		if frame.Return != nil {
			report.StaticDataBytes += frame.Return.Size
			report.StaticSymbolCount++
		}
	}

	for _, lit := range g.literals {
		report.LiteralBytes += len(lit)
	}

	return report
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
