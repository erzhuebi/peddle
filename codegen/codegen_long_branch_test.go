package codegen

import (
	"fmt"
	"strings"
	"testing"
)

func longBranchBody(repetitions int) string {
	var b strings.Builder

	for i := 0; i < repetitions; i++ {
		b.WriteString("        total = total + i\n")
	}

	return b.String()
}

func longBranchSource(control string, body string) string {
	return fmt.Sprintf(`
fn main() {
    var flag bool
    var i byte
    var total int

    flag = true
    i = 0
    total = 0

%s {
%s    }
}
`, control, body)
}

func requireASMContainsLongBranchPattern(t *testing.T, asm string) {
	t.Helper()

	lines := strings.Split(asm, "\n")
	for i := 0; i+3 < len(lines); i++ {
		if lines[i] != "    cmp #0" {
			continue
		}

		branchParts := strings.Fields(lines[i+1])
		jumpParts := strings.Fields(lines[i+2])
		if len(branchParts) != 2 || len(jumpParts) != 2 {
			continue
		}
		if branchParts[0] != "bne" || jumpParts[0] != "jmp" {
			continue
		}
		if lines[i+3] == branchParts[1]+":" {
			return
		}
	}

	t.Fatalf("expected inverted branch plus jmp long-branch pattern\n\nASM:\n%s", asm)
}

func TestCodegenLongBranchPatternForNonComparisonCondition(t *testing.T) {
	asm := compileSource(t, longBranchSource("if flag", longBranchBody(80)))

	requireASMContainsLongBranchPattern(t, asm)
	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenLongBranchesAssembleForLargeIfAndLoopBodies(t *testing.T) {
	tests := []struct {
		name    string
		control string
		body    string
	}{
		{
			name:    "if bool",
			control: "if flag",
			body:    longBranchBody(80),
		},
		{
			name:    "if comparison",
			control: "if i < 100",
			body:    longBranchBody(80),
		},
		{
			name:    "while bool",
			control: "while flag",
			body:    longBranchBody(80),
		},
		{
			name:    "while comparison",
			control: "while i < 100",
			body:    longBranchBody(80),
		},
		{
			name:    "for bool",
			control: "for flag",
			body:    longBranchBody(80),
		},
		{
			name:    "for comparison",
			control: "for i < 100",
			body:    longBranchBody(80),
		},
		{
			name:    "counted for",
			control: "for i = 0 to 3",
			body:    longBranchBody(80),
		},
		{
			name:    "counted for continue",
			control: "for i = 0 to 3",
			body: longBranchBody(40) + `        if i == 2 {
            continue
        }
` + longBranchBody(40),
		},
		{
			name:    "counted for break",
			control: "for i = 0 to 3",
			body: longBranchBody(40) + `        if i == 2 {
            break
        }
` + longBranchBody(40),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := compileSource(t, longBranchSource(tt.control, tt.body))

			requireReferencedLabelsDefined(t, asm)
			requireASMAssemblesWith64tass(t, asm)
		})
	}
}
