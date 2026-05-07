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

func main() {
	outPath := flag.String("o", "out.asm", "output ASM file")
	optMode := flag.String("opt", "speed", "optimization mode: speed or size")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: peddlec -o output.asm [--opt=speed|--opt=size] input.ped")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *optMode != "speed" && *optMode != "size" {
		fmt.Fprintf(os.Stderr, "invalid optimization mode %q: expected speed or size\n", *optMode)
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
		OptMode: codegen.OptMode(*optMode),
	})

	asm, err := g.Generate(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen error:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outPath, []byte(asm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
}
