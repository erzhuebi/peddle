package codegen

import "testing"

func TestCodegenForLoopFormsAssemble(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var i byte
    var total int

    for {
        break
    }

    for i < 3 {
        i = i + 1
    }

    for i = 0 to 3 {
        if i == 2 {
            continue
        }

        total = total + i
    }
}
`)

	requireASM(t, asm,
		"main_for_end_",
		"inc main_i",
		"jmp L",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenCountedForIntAssemble(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var i int
    var total int

    for i = -2 to 2 {
        total = total + i
    }
}
`)

	requireASM(t, asm,
		"main_for_end_",
		"inc main_i",
		"inc main_i+1",
		"peddle_tmp_int0",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
