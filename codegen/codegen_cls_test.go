package codegen

import "testing"

func TestCodegenClsUsesRuntimeHelper(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()
}
`)

	requireASM(t, asm,
		"jsr peddle_cls",
		"peddle_cls:",
		"peddle_cls_loop_full:",
		"sta $0400, x",
		"sta $0500, x",
		"sta $0600, x",
		"peddle_cls_loop_last:",
		"sta $0700, x",
		"jsr $fff0",
		"rts",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenClsRuntimeEmittedOnlyOnce(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()
    cls()
}
`)

	if countOccurrences(asm, "peddle_cls:") != 1 {
		t.Fatalf("expected peddle_cls runtime label once\n\nASM:\n%s", asm)
	}

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
