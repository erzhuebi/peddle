package codegen

import "testing"

func TestCodegenBreakAndContinueAssembles(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var i byte
    var values byte[10]

    i = 0

    while i < 10 {
        i = i + 1

        if i == 3 {
            continue
        }

        if i == 8 {
            break
        }

        append(values, i)
    }
}
`)

	requireASM(t, asm,
		"jmp L",
		"main_values:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNestedBreakTargetsInnerLoop(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x byte
    var y byte

    while x < 3 {
        y = 0

        while y < 3 {
            if y == 1 {
                break
            }

            y = y + 1
        }

        x = x + 1
    }
}
`)

	requireASM(t, asm,
		"jmp L",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
