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
    var a byte[10]
    var i byte

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
    id byte
    name char[16]
    hp int
}

fn main() {
    var players Player[4]
    var i byte

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
    id byte
    name char[16]
    hp int
}

fn main() {
    var nums int[10]
    var copyNums int[10]

    var players Player[4]
    var i byte
    var n int

    var title char[32]

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

func TestCodegenAppendToIndexedStructCharArrayPreservesSource(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    name char[16]
    hp int
}

fn main() {
    var players Player[10]
    var i byte = 0

    while i < 10 {
        copy(players[i].name, "PLAYER0")
        append(players[i].name, itoa(i))
        append(players[i].name, "\n")
        i = i + 1
    }

    print(players[9].name)
}
`)

	if got := strings.Count(asm, "lda ZP_PTR1_LO\n    pha"); got < 2 {
		t.Fatalf("expected both indexed struct-field appends to preserve source pointer/length, got %d\n\nASM:\n%s", got, asm)
	}

	requireASM(t, asm,
		"jsr peddle_string_append_literal",
		"pla\n    sta peddle_tmp_int0+1",
		"pla\n    sta ZP_PTR1_LO",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenAppendToIndexedStructFieldArraysPreservesValue(t *testing.T) {
	src := `
struct Bucket {
    values byte[8]
    totals int[8]
    marker byte
}

fn main() {
    var buckets Bucket[4]
    var i byte = 1
    var b byte = 7
    var n int = 1024

    append(buckets[i].values, b)
    append(buckets[i].totals, n)
}
`

	for _, mode := range []OptMode{OptModeSpeed, OptModeSize} {
		t.Run(string(mode), func(t *testing.T) {
			asm := compileSourceWithOptions(t, src, Options{OptMode: mode})

			if got := strings.Count(asm, "lda peddle_tmp_int0\n    pha\n    lda peddle_tmp_int0+1\n    pha"); got < 2 {
				t.Fatalf("expected both indexed field appends to preserve pending value, got %d\n\nASM:\n%s", got, asm)
			}

			if mode == OptModeSize {
				requireASM(t, asm,
					"jsr peddle_append_byte",
					"jsr peddle_append_int",
				)
			}

			requireReferencedLabelsDefined(t, asm)
			requireASMAssemblesWith64tass(t, asm)
		})
	}
}

func TestCodegenFillIndexedStructFieldArraysPreservesFillValue(t *testing.T) {
	src := `
struct Bucket {
    values byte[8]
    totals int[8]
    marker byte
}

fn main() {
    var buckets Bucket[4]
    var i byte = 1
    var b byte = 3
    var n int = 777

    fill(buckets[i].values, b)
    fill(buckets[i].totals, n)
}
`

	for _, mode := range []OptMode{OptModeSpeed, OptModeSize} {
		t.Run(string(mode), func(t *testing.T) {
			asm := compileSourceWithOptions(t, src, Options{OptMode: mode})

			if got := strings.Count(asm, "lda ZP_TMP0\n    pha\n    lda ZP_TMP1\n    pha"); got < 2 {
				t.Fatalf("expected both indexed field fills to preserve fill value, got %d\n\nASM:\n%s", got, asm)
			}

			if mode == OptModeSize {
				requireASM(t, asm,
					"jsr peddle_fill_byte",
					"jsr peddle_fill_int",
				)
			}

			requireReferencedLabelsDefined(t, asm)
			requireASMAssemblesWith64tass(t, asm)
		})
	}
}

func TestCodegenPutStrColorIndexedStructCharArrayPreservesTextArg(t *testing.T) {
	asm := compileSource(t, `
struct Row {
    name char[16]
    color byte
}

fn main() {
    var rows Row[4]
    var i byte = 1

    copy(rows[i].name, "ROW")
    rows[i].color = 2
    putstrcolor(0, 0, rows[i].name, rows[i].color + 1)
}
`)

	requireASM(t, asm,
		"lda peddle_tmp_int0\n    pha\n    lda peddle_tmp_int0+1\n    pha\n    lda ZP_PTR1_LO\n    pha\n    lda ZP_PTR1_HI\n    pha",
		"jsr peddle_putstrcolor",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenIndexedStructFieldArrayCommandsAssemble(t *testing.T) {
	src := `
struct Bucket {
    name char[16]
    values byte[8]
    totals int[8]
    marker byte
}

fn main() {
    var buckets Bucket[4]
    var i byte = 1
    var j byte = 2
    var l int
    var s int

    copy(buckets[i].name, buckets[j].name)
    copy(buckets[i].values, buckets[j].values)
    copy(buckets[i].totals, buckets[j].totals)

    clear(buckets[i].name)
    clear(buckets[i].values)
    clear(buckets[i].totals)

    l = len(buckets[i].name)
    s = size(buckets[i].totals)

    if l == s {
        print("MATCH")
    }
}
`

	for _, mode := range []OptMode{OptModeSpeed, OptModeSize} {
		t.Run(string(mode), func(t *testing.T) {
			asm := compileSourceWithOptions(t, src, Options{OptMode: mode})

			if mode == OptModeSize {
				requireASM(t, asm, "jsr peddle_array_copy")
			}

			requireNoASM(t, asm, "_skip_broken")
			requireReferencedLabelsDefined(t, asm)
			requireASMAssemblesWith64tass(t, asm)
		})
	}
}

func TestCodegenCountedStringLiteralsAssemble(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    print("ABC")
    print("")
    print("DONE")
}
`)

	requireNoASM(t, asm,
		"peddle_print_string:",
		"jsr peddle_print_string",
		".byte 65,66,67,0",
	)

	requireASM(t, asm,
		"jsr peddle_print_counted_string",
		"literal_0:",
		"literal_1:",
		"literal_2:",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenEmptyStringLiteralLengthIsZero(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var s char[10]

    copy(s, "")
    append(s, "")
    print("")
}
`)

	requireASM(t, asm,
		"lda #<0",
		"lda #>0",
		"jsr peddle_print_counted_string",
		"literal_0:",
		"literal_1:",
		"literal_2:",
	)

	requireNoASM(t, asm,
		".byte 0,0",
		"peddle_print_string:",
		"jsr peddle_print_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenClearArraysAssembles(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id byte
    name char[16]
    hp int
}

fn main() {
    var nums int[10]
    var title char[16]
    var players Player[4]

    append(nums, 1)
    clear(nums)

    copy(title, "HELLO")
    clear(title)

    players[0].id = 1
    clear(players)
}
`)

	requireNoASM(t, asm,
		"_skip_broken",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenOptSizeArrayHelpersAssemble(t *testing.T) {
	asm := compileSourceWithOptions(t, `
fn main() {
    var a byte[10]
    var b byte[10]
    var i int[10]
    var j int[10]
    var s char[16]

    append(a, 1)
    fill(a, 2)
    copy(b, a)

    append(i, 1000)
    fill(i, 7)
    copy(j, i)

    copy(s, "HELLO")
    append(s, "!")
}
`, Options{OptMode: OptModeSize})

	requireASM(t, asm,
		"jsr peddle_append_byte",
		"jsr peddle_fill_byte",
		"jsr peddle_array_copy",
		"jsr peddle_append_int",
		"jsr peddle_fill_int",
		"jsr peddle_string_copy_literal",
		"jsr peddle_string_append_literal",
	)

	requireNoASM(t, asm,
		"_skip_broken",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
