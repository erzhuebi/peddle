package codegen

import "testing"

func TestCodegenArrayParameterPassedByReference(t *testing.T) {
	asm := compileSource(t, `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums byte[4]
    var n int
    var first byte

    push(nums, 7)
    n = len(nums)
    first = nums[0]
}
`)

	requireASM(t, asm,
		"push_nums:",
		".fill 2, 0",
		"lda #<main_nums",
		"sta push_nums",
		"sta push_nums+1",
		"lda push_nums",
		"sta ZP_PTR0_LO",
		"lda push_nums+1",
		"sta ZP_PTR0_HI",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenArrayParameterIndexReadWrite(t *testing.T) {
	asm := compileSource(t, `
fn setSecond(nums byte[4], value byte) {
    nums[1] = value
}

fn getSecond(nums byte[4]) byte {
    return nums[1]
}

fn main() {
    var nums byte[4]
    var second byte

    setSecond(nums, 9)
    second = getSecond(nums)
}
`)

	requireASM(t, asm,
		"setSecond_nums:",
		".fill 2, 0",
		"getSecond_nums:",
		".fill 2, 0",
		"lda setSecond_nums",
		"sta ZP_PTR0_LO",
		"lda setSecond_nums+1",
		"sta ZP_PTR0_HI",
		"sta (ZP_PTR0_LO), y",
		"lda getSecond_nums",
		"sta ZP_PTR0_LO",
		"lda getSecond_nums+1",
		"sta ZP_PTR0_HI",
		"lda (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenArrayParameterForwarding(t *testing.T) {
	asm := compileSource(t, `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn pushTwo(nums byte[4]) {
    push(nums, 1)
    push(nums, 2)
}

fn main() {
    var nums byte[4]

    pushTwo(nums)
}
`)

	requireASM(t, asm,
		"push_nums:",
		".fill 2, 0",
		"pushTwo_nums:",
		".fill 2, 0",
		"lda #<main_nums",
		"sta pushTwo_nums",
		"lda pushTwo_nums",
		"sta ZP_PTR0_LO",
		"lda pushTwo_nums+1",
		"sta ZP_PTR0_HI",
		"sta push_nums+1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenCharArrayParameterBuiltins(t *testing.T) {
	asm := compileSourceWithOptions(t, `
fn setTitle(title char[8]) {
    copy(title, "OK")
    append(title, "!")
    print(title)
}

fn main() {
    var title char[8]

    setTitle(title)
}
`, Options{OptMode: OptModeSize})

	requireASM(t, asm,
		"setTitle_title:",
		".fill 2, 0",
		"lda #<main_title",
		"sta setTitle_title",
		"lda setTitle_title",
		"sta ZP_PTR0_LO",
		"lda setTitle_title+1",
		"sta ZP_PTR0_HI",
		"jsr peddle_string_copy_literal",
		"jsr peddle_string_append_literal",
		"jsr peddle_print_counted_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStructArrayParameterPassedByReference(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    hp byte
}

fn damage(players Player[2], idx byte) {
    players[idx].hp = players[idx].hp - 1
}

fn main() {
    var players Player[2]
    var hp byte

    players[1].hp = 9
    damage(players, 1)
    hp = players[1].hp
}
`)

	requireASM(t, asm,
		"damage_players:",
		".fill 2, 0",
		"lda #<main_players",
		"sta damage_players",
		"sta damage_players+1",
		"lda damage_players",
		"sta ZP_PTR0_LO",
		"lda damage_players+1",
		"sta ZP_PTR0_HI",
		"lda (ZP_PTR0_LO), y",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenIndexedStructFieldArraysPassedAsParameters(t *testing.T) {
	asm := compileSource(t, `
struct Bucket {
    name char[16]
    bytes byte[8]
    totals int[8]
    marker byte
}

fn tag(name char[16]) {
    append(name, "!")
}

fn pushByte(values byte[8], value byte) {
    append(values, value)
}

fn pushInt(values int[8], value int) {
    append(values, value)
}

fn main() {
    var buckets Bucket[4]
    var i byte = 1

    copy(buckets[i].name, "B")
    tag(buckets[i].name)
    pushByte(buckets[i].bytes, 7)
    pushInt(buckets[i].totals, 1024)

    if len(buckets[i].name) == 2 {
        print("PARAM OK")
    }
}
`)

	requireASM(t, asm,
		"tag_name:",
		".fill 2, 0",
		"pushByte_values:",
		".fill 2, 0",
		"pushInt_values:",
		".fill 2, 0",
		"jsr tag",
		"jsr pushByte",
		"jsr pushInt",
	)

	requireNoASM(t, asm, "_skip_broken")
	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
