package codegen

import (
	"fmt"
	"strconv"
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

	constants map[string]int
	functions map[string]*ast.FunctionDecl
	structs   map[string]*ast.StructDecl
	frames    map[string]*Frame

	currentFn    *ast.FunctionDecl
	currentFrame *Frame

	literals []string

	forLoopTemps []Symbol

	usedPrint bool
	usedTmp16 bool

	usedNetRuntime          bool
	usedFileRuntime         bool
	usedSoundRuntime        bool
	usedClsRuntime          bool
	usedAsciiFontRuntime    bool
	usedAsciiConvertRuntime bool
	usedArrayCopyRuntime    bool
	usedFillByteRuntime     bool
	usedFillIntRuntime      bool
	usedAppendByteRuntime   bool
	usedAppendIntRuntime    bool
	usedMulByteRuntime      bool
	usedMulIntRuntime       bool
	usedDivModByteRuntime   bool
	usedDivModIntRuntime    bool
	usedShlByteRuntime      bool
	usedShrByteRuntime      bool
	usedShlIntRuntime       bool
	usedShrIntRuntime       bool
	usedStringCopyRuntime   bool
	usedStringAppendRuntime bool
	usedPutStrRuntime       bool
	usedPutStrColorRuntime  bool
	usedCharToScreenTable   bool
	usedItoaRuntime         bool
	usedItoxRuntime         bool
	usedReadLineRuntime     bool

	memoryReport MemoryReport

	labelCounter int
}

type Frame struct {
	FunctionName string
	Symbols      map[string]Symbol
	Return       *Symbol
	Returns      []Symbol
}

type Symbol struct {
	SourceName   string
	Label        string
	Type         ast.Type
	Size         int
	IsReference  bool
	HasAtAddress bool
	AtAddress    int
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
		constants: map[string]int{},
		functions: map[string]*ast.FunctionDecl{},
		structs:   map[string]*ast.StructDecl{},
		frames:    map[string]*Frame{},
	}
}

func (g *Generator) Generate(p *ast.Program) (string, error) {
	for _, c := range p.Consts {
		n, err := strconv.Atoi(c.Value)
		if err != nil {
			return "", fmt.Errorf("invalid const %q value %q", c.Name, c.Value)
		}
		g.constants[c.Name] = n
	}

	for _, s := range p.Structs {
		g.structs[s.Name] = s
	}

	for _, fn := range p.Functions {
		g.functions[fn.Name] = fn
		g.frames[fn.Name] = g.buildFrame(fn)
	}

	g.scanSoundRuntimeUse(p)

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
			if sym.Size == 0 {
				continue
			}
			report.StaticDataBytes += sym.Size
			report.StaticSymbolCount++
		}

		for _, sym := range frame.Returns {
			if sym.Size == 0 {
				continue
			}
			report.StaticDataBytes += sym.Size
			report.StaticSymbolCount++
		}
	}

	for _, lit := range g.literals {
		report.LiteralBytes += len(lit)
	}

	for _, sym := range g.forLoopTemps {
		report.StaticDataBytes += sym.Size
		report.StaticSymbolCount++
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

func (g *Generator) scanSoundRuntimeUse(p *ast.Program) {
	for _, fn := range p.Functions {
		for _, stmt := range fn.Body {
			g.scanSoundStmt(stmt)
		}
	}
}

func (g *Generator) scanSoundStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		g.scanSoundExpr(s.Value)
		if target, ok := s.Target.(*ast.IndexLValue); ok {
			g.scanSoundExpr(target.Index)
		}
	case *ast.CallStmt:
		if isSoundBuiltin(s.Name) {
			g.usedSoundRuntime = true
		}
		for _, arg := range s.Args {
			g.scanSoundExpr(arg)
		}
	case *ast.WhileStmt:
		g.scanSoundExpr(s.Cond)
		for _, inner := range s.Body {
			g.scanSoundStmt(inner)
		}
	case *ast.ForStmt:
		g.scanSoundExpr(s.Cond)
		g.scanSoundExpr(s.Start)
		g.scanSoundExpr(s.End)
		for _, inner := range s.Body {
			g.scanSoundStmt(inner)
		}
	case *ast.IfStmt:
		g.scanSoundExpr(s.Cond)
		for _, inner := range s.Then {
			g.scanSoundStmt(inner)
		}
		for _, inner := range s.Else {
			g.scanSoundStmt(inner)
		}
	case *ast.ReturnStmt:
		for _, value := range returnValues(s) {
			g.scanSoundExpr(value)
		}
	}
}

func (g *Generator) scanSoundExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		if isSoundBuiltin(e.Name) {
			g.usedSoundRuntime = true
		}
		for _, arg := range e.Args {
			g.scanSoundExpr(arg)
		}
	case *ast.UnaryExpr:
		g.scanSoundExpr(e.Expr)
	case *ast.BinaryExpr:
		g.scanSoundExpr(e.Left)
		g.scanSoundExpr(e.Right)
	case *ast.IndexExpr:
		g.scanSoundExpr(e.Index)
	case *ast.IndexFieldExpr:
		g.scanSoundExpr(e.Index)
	}
}
