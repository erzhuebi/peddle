package codegen

import (
	"strings"
	"testing"
)

func TestCodegenAsciiFontUsesRuntimeHelper(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    asciifont()
}
`)

	requireASM(t, asm,
		"jsr peddle_asciifont",
		"peddle_asciifont:",
		"peddle_asciifont_copy:",
		"sta $3800, x",
		"sta $3f00, x",
		"peddle_asciifont_copy_lowercase:",
		"lda $d808, x",
		"sta $3a08, x",
		"peddle_asciifont_patch_underscore:",
		"sta $38f8, x",
		"and #$f1",
		"ora #$0e",
		"sta $d018",
		"peddle_asciifont_underscore:",
		".byte 0, 0, 0, 0, 0, 0, 0, 255",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenAsciiConversionBuiltins(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var text char[32]

    copy(text, "HELLO")
    toascii(text)
    topetscii(text)
}
`)

	requireASM(t, asm,
		"jsr peddle_toascii",
		"jsr peddle_topetscii",
		"peddle_toascii:",
		"adc #32",
		"peddle_topetscii:",
		"sbc #32",
		"cmp #10",
		"lda #13",
		"peddle_ascii_convert_prepare:",
		"peddle_ascii_convert_next:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenAsciiFontRuntimeEmittedOnlyWhenUsed(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    cls()
}
`)

	if strings.Contains(asm, "peddle_asciifont:") {
		t.Fatalf("unexpected asciifont runtime in ASM:\n%s", asm)
	}

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
