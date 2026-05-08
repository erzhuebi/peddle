package codegen

import (
	"strings"
	"testing"

	"peddle/lexer"
	"peddle/parser"
	"peddle/sema"
)

func generateProgramWithOptions(t *testing.T, src string, options Options) (*Generator, string, error) {
	t.Helper()

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if err := sema.New().Check(prog); err != nil {
		t.Fatalf("sema error: %v", err)
	}

	g := NewWithOptions(options)
	asm, err := g.Generate(prog)
	return g, asm, err
}

func TestCodegenMemoryReportTracksStaticDataBytes(t *testing.T) {
	g, _, err := generateProgramWithOptions(t, `
fn main() {
    var a: byte[10]
    var b: int[4]
    var s: char[16]
}
`, Options{OptMode: OptModeSpeed})

	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	report := g.MemoryReport()

	if report.StaticDataBytes != 46 {
		t.Fatalf("got static data bytes %d, want 46", report.StaticDataBytes)
	}

	if report.StaticSymbolCount != 3 {
		t.Fatalf("got static symbol count %d, want 3", report.StaticSymbolCount)
	}
}

func TestCodegenMemoryReportTracksStructStaticDataBytes(t *testing.T) {
	g, _, err := generateProgramWithOptions(t, `
struct Player {
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var p: Player
    var players: Player[2]
}
`, Options{OptMode: OptModeSpeed})

	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	report := g.MemoryReport()

	if report.StaticDataBytes != 73 {
		t.Fatalf("got static data bytes %d, want 73", report.StaticDataBytes)
	}

	if report.StaticSymbolCount != 2 {
		t.Fatalf("got static symbol count %d, want 2", report.StaticSymbolCount)
	}
}

func TestCodegenMemoryLimitAllowsProgramWithinLimit(t *testing.T) {
	_, asm, err := generateProgramWithOptions(t, `
fn main() {
    var a: byte[10]
}
`, Options{
		OptMode:           OptModeSpeed,
		StaticMemoryLimit: 14,
	})

	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(asm, "main_a:") {
		t.Fatalf("expected generated ASM to contain main_a label")
	}
}

func TestCodegenMemoryLimitRejectsProgramOverLimit(t *testing.T) {
	_, _, err := generateProgramWithOptions(t, `
fn main() {
    var a: byte[10]
}
`, Options{
		OptMode:           OptModeSpeed,
		StaticMemoryLimit: 13,
	})

	if err == nil {
		t.Fatalf("expected memory limit error")
	}

	if !strings.Contains(err.Error(), "static memory usage 14 bytes exceeds limit 13 bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCodegenMemoryReportTracksLiterals(t *testing.T) {
	g, _, err := generateProgramWithOptions(t, `
fn main() {
    print("ABC")
    print("DE")
}
`, Options{OptMode: OptModeSpeed})

	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	report := g.MemoryReport()

	if report.LiteralBytes != 5 {
		t.Fatalf("got literal bytes %d, want 5", report.LiteralBytes)
	}
}
