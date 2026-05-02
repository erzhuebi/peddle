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
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: peddlec input.ped -o output.asm")
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

	g := codegen.New()
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
