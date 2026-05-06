package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func requireNoASM(t *testing.T, asm string, parts ...string) {
	t.Helper()

	for _, part := range parts {
		if strings.Contains(asm, part) {
			t.Fatalf("ASM unexpectedly contains %q\n\nASM:\n%s", part, asm)
		}
	}
}

func requireReferencedLabelsDefined(t *testing.T, asm string) {
	t.Helper()

	defined := map[string]bool{}
	referenced := map[string]bool{}

	labelDef := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*):$`)
	labelRef := regexp.MustCompile(`^\s*(?:beq|bne|bcc|bcs|bmi|bpl|bvc|bvs|jmp|jsr)\s+([A-Za-z_][A-Za-z0-9_]*)`)

	for _, line := range strings.Split(asm, "\n") {
		line = strings.TrimSpace(line)

		if m := labelDef.FindStringSubmatch(line); m != nil {
			defined[m[1]] = true
			continue
		}

		if m := labelRef.FindStringSubmatch(line); m != nil {
			referenced[m[1]] = true
		}
	}

	for label := range referenced {
		if !defined[label] {
			t.Fatalf("ASM references undefined label %q\n\nASM:\n%s", label, asm)
		}
	}
}

func requireASMAssemblesWith64tass(t *testing.T, asm string) {
	t.Helper()

	if _, err := exec.LookPath("64tass"); err != nil {
		t.Skip("64tass not found")
	}

	dir := t.TempDir()
	asmPath := filepath.Join(dir, "test.asm")
	prgPath := filepath.Join(dir, "test.prg")

	if err := os.WriteFile(asmPath, []byte(asm), 0644); err != nil {
		t.Fatalf("write asm: %v", err)
	}

	cmd := exec.Command("64tass", asmPath, "-o", prgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("64tass failed: %v\n\n%s\n\nASM:\n%s", err, string(out), asm)
	}
}

func TestCodegenArrayIndexWriteDoesNotEmitBrokenLabels(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var a: byte[10]
    var i: byte

    i = 5
    a[i] = 1
}
`)

	requireNoASM(t, asm,
		"_skip_broken",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStructArrayFieldAssignmentPreservesValueAcrossIndexCalculation(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var players: Player[4]
    var i: byte

    i = 1
    players[i].hp = 120
}
`)

	requireASM(t, asm,
		"sta ZP_PTR1_LO",
		"sta ZP_PTR1_HI",
		"lda #<120",
		"sta ZP_TMP0",
		"lda #>120",
		"sta ZP_TMP1",
		"lda ZP_PTR1_LO",
		"sta ZP_PTR0_LO",
		"lda ZP_PTR1_HI",
		"sta ZP_PTR0_HI",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenArraysDemoAssembles(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var nums: int[10]
    var copyNums: int[10]

    var players: Player[4]
    var i: byte
    var n: int

    var title: char[32]

    print("PEDDLE ARRAY DEMO ")

    copy(title, "TEAM ")
    append(title, "ALPHA ")
    print(title)

    append(nums, 10)
    append(nums, 20)
    append(nums, 30)

    nums[5] = 99

    n = len(nums)
    if n == 6 {
        print("LEN UPDATED ")
    }

    n = size(nums)
    if n == 10 {
        print("SIZE OK ")
    }

    copy(copyNums, nums)

    fill(nums, 1)

    i = 0
    players[i].id = 1
    copy(players[i].name, "BOB ")
    players[i].hp = 100

    i = 1
    players[i].id = 2
    copy(players[i].name, "ADA")
    append(players[i].name, "! ")
    players[i].hp = 120

    i = 0
    print(players[i].name)

    i = 1
    print(players[i].name)

    if players[1].hp > players[0].hp {
        print("ADA LEADS ")
    }

    i = 0
    while i < 3 {
        append(nums, i)
        i = i + 1
    }

    if len(nums) > 6 {
        print("APPEND OK ")
    }

    poke(53280, 0)
    poke(53281, 6)

    print("DONE")
}
`)

	requireNoASM(t, asm,
		"_skip_broken",
	)

	requireASM(t, asm,
		"jsr peddle_print_counted_string",
		"sta ZP_PTR1_LO",
		"sta ZP_PTR1_HI",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
