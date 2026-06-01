package codegen

import (
	"strings"
	"testing"
)

func TestCodegenMultipleReturnSlotsAndAssignmentCopies(t *testing.T) {
	asm := compileSource(t, `
fn pair() (byte, uint) {
    var b byte
    var u uint

    b = 7
    u = 65535
    return b, u
}

fn flags() (bool, int) {
    var ok bool
    var value int

    ok = true
    value = -3
    return ok, value
}

fn main() {
    var b byte
    var u uint
    var value int

    b, u = pair()
    _, value = flags()
}
`)

	requireASM(t, asm,
		"pair_return_0:",
		"pair_return_1:",
		"flags_return_0:",
		"flags_return_1:",
		"sta pair_return_0",
		"sta pair_return_1",
		"sta pair_return_1+1",
		"jsr pair",
		"lda pair_return_0",
		"sta main_b",
		"lda pair_return_1",
		"sta main_u",
		"sta main_u+1",
		"jsr flags",
		"lda flags_return_1",
		"sta main_value",
		"sta main_value+1",
	)

	afterFlagsCall := asm[strings.Index(asm, "    jsr flags"):]
	if strings.Contains(afterFlagsCall, "lda flags_return_0") {
		t.Fatalf("ignored return value should not be copied after flags call\n\nASM:\n%s", asm)
	}

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenMultipleReturnWidenAndTruncateCopies(t *testing.T) {
	asm := compileSource(t, `
fn values() (byte, int) {
    var b byte
    var i int

    b = 3
    i = 260
    return b, i
}

fn main() {
    var wide int
    var low byte

    wide, low = values()
}
`)

	requireASM(t, asm,
		"values_return_0:",
		"values_return_1:",
		"jsr values",
		"lda values_return_0",
		"lda #0",
		"sta main_wide+1",
		"lda values_return_1",
		"sta main_low",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
