package codegen

import "testing"

func TestCodegenStructPointerParameterFieldMutation(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var player Player

    player.hp = 5
    damage(&player)
}
`)

	requireASM(t, asm,
		"damage_p:",
		".fill 2, 0",
		"lda #<main_player",
		"sta damage_p",
		"sta damage_p+1",
		"lda damage_p",
		"sta ZP_PTR0_LO",
		"lda damage_p+1",
		"sta ZP_PTR0_HI",
		"lda (ZP_PTR0_LO), y",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStructArrayElementPointerArgument(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    hp int
}

fn heal(p *Player) {
    p.hp = 100
}

fn main() {
    var players Player[4]
    var i byte

    i = 2
    heal(&players[i])
}
`)

	requireASM(t, asm,
		"heal_p:",
		".fill 2, 0",
		"lda #<main_players+4",
		"sta ZP_PTR0_LO",
		"lda #>main_players+4",
		"sta ZP_PTR0_HI",
		"lda ZP_PTR0_LO",
		"sta ZP_TMP0",
		"lda ZP_PTR0_HI",
		"sta ZP_TMP1",
		"sta heal_p+1",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenScalarPointerParameterByteMutation(t *testing.T) {
	asm := compileSource(t, `
fn bump(x *byte) {
    x = x + 1
}

fn main() {
    var value byte

    value = 5
    bump(&value)
}
`)

	requireASM(t, asm,
		"bump_x:",
		".fill 2, 0",
		"lda #<main_value",
		"sta bump_x",
		"lda #>main_value",
		"sta bump_x+1",
		"lda bump_x",
		"sta ZP_PTR0_LO",
		"lda bump_x+1",
		"sta ZP_PTR0_HI",
		"lda (ZP_PTR0_LO), y",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenScalarPointerParameterUintMutation(t *testing.T) {
	asm := compileSource(t, `
fn bump(x *uint) {
    x = x + 1
}

fn main() {
    var value uint

    value = 65534
    bump(&value)
}
`)

	requireASM(t, asm,
		"bump_x:",
		".fill 2, 0",
		"lda #<main_value",
		"sta bump_x",
		"lda #>main_value",
		"sta bump_x+1",
		"lda (ZP_PTR0_LO), y",
		"sta ZP_TMP0",
		"iny",
		"lda (ZP_PTR0_LO), y",
		"sta ZP_TMP1",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenScalarArrayElementPointerArgument(t *testing.T) {
	asm := compileSource(t, `
fn bump(x *uint) {
    x = x + 1
}

fn main() {
    var values uint[4]
    var i byte

    i = 2
    values[i] = 10
    bump(&values[i])
}
`)

	requireASM(t, asm,
		"bump_x:",
		".fill 2, 0",
		"lda #<main_values+4",
		"sta ZP_PTR0_LO",
		"lda #>main_values+4",
		"sta ZP_PTR0_HI",
		"lda ZP_PTR0_LO",
		"sta ZP_TMP0",
		"lda ZP_PTR0_HI",
		"sta ZP_TMP1",
		"sta bump_x+1",
		"sta (ZP_PTR0_LO), y",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
