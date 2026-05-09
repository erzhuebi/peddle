package main

import (
	"flag"
	"fmt"
	"os"

	"peddle/codegen"
	"peddle/lexer"
	"peddle/parser"
	"peddle/sema"
)

const Version = "0.1.0"

func main() {
	outPath := flag.String("o", "out.asm", "output ASM file")
	optMode := flag.String("opt", "speed", "optimization mode: speed or size")
	memReport := flag.Bool("mem-report", false, "print compiler memory usage report")
	memLimit := flag.Int("mem-limit", 0, "maximum static data memory in bytes, 0 disables the limit")

	showVersion := flag.Bool("version", false, "print compiler version")
	showVersionShort := flag.Bool("v", false, "print compiler version")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: peddlec -o output.asm [--opt=speed|--opt=size] [--mem-report] [--mem-limit=N] input.ped")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("peddlec %s\n", Version)
		fmt.Println("target: c64/6502")
		fmt.Println("assembler: 64tass")
		return
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *optMode != "speed" && *optMode != "size" {
		fmt.Fprintf(os.Stderr, "invalid optimization mode %q: expected speed or size\n", *optMode)
		os.Exit(1)
	}

	if *memLimit < 0 {
		fmt.Fprintf(os.Stderr, "invalid memory limit %d: expected 0 or greater\n", *memLimit)
		os.Exit(1)
	}

	inputPath := flag.Arg(0)

	src, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintln(os.Stderr, "parse error:", err)
		}
		os.Exit(1)
	}

	checker := sema.New()
	if err := checker.Check(program); err != nil {
		fmt.Fprintln(os.Stderr, "semantic error:", err)
		os.Exit(1)
	}

	g := codegen.NewWithOptions(codegen.Options{
		OptMode:           codegen.OptMode(*optMode),
		StaticMemoryLimit: *memLimit,
	})

	asm, err := g.Generate(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen error:", err)
		os.Exit(1)
	}

	if *memReport {
		report := g.MemoryReport()
		variablesAndRuntime := report.StaticDataBytes - report.LiteralBytes

		fmt.Fprintln(os.Stderr, "memory report:")
		fmt.Fprintf(os.Stderr, "  total static memory: %d bytes\n", report.StaticDataBytes)
		fmt.Fprintf(os.Stderr, "    literals:          %d bytes\n", report.LiteralBytes)
		fmt.Fprintf(os.Stderr, "    variables/runtime: %d bytes\n", variablesAndRuntime)
		fmt.Fprintf(os.Stderr, "    static symbols:    %d\n", report.StaticSymbolCount)
	}

	if err := os.WriteFile(*outPath, []byte(asm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
}
