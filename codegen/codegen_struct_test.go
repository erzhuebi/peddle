package codegen

import "testing"

func TestCodegenStructFieldByteReadWrite(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    x: byte
    y: byte
}

fn main() {
    var p: Player
    var x: byte

    p.x = 10
    p.y = 20
    x = p.x
}
`)

	requireASM(t, asm,
		"main_p:",
		".fill 1, 0",
		"sta main_p+0",
		"sta main_p+1",
		"lda main_p+0",
		"sta main_x",
	)
}

func TestCodegenStructFieldIntReadWrite(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    x: byte
    hp: int
}

fn main() {
    var p: Player
    var hp: int

    p.x = 10
    p.hp = 1000
    hp = p.hp
}
`)

	requireASM(t, asm,
		"main_p:",
		".fill 1, 0",
		".fill 2, 0",
		"sta main_p+0",
		"sta main_p+1",
		"sta main_p+2",
		"lda main_p+1",
		"lda main_p+2",
		"sta main_hp",
		"sta main_hp+1",
	)
}

func TestCodegenStructFieldUsedInExpression(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    hp: int
}

fn main() {
    var p: Player
    var x: int

    p.hp = 100
    x = p.hp + 23
}
`)

	requireASM(t, asm,
		"lda main_p+0",
		"sta ZP_TMP0",
		"lda main_p+1",
		"sta ZP_TMP1",
		"adc peddle_tmp_int0",
		"sta main_x",
		"sta main_x+1",
	)
}

func TestCodegenStructArrayFieldByteReadWrite(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    x: byte
    y: byte
    hp: int
}

fn main() {
    var players: Player[10]
    var i: byte
    var x: byte

    i = 3
    players[i].x = 42
    x = players[i].x
}
`)

	requireASM(t, asm,
		"main_players:",
		".word 10",
		".word 10",
		".fill 1, 0",
		".fill 2, 0",
		"lda #<main_players+4",
		"sta ZP_PTR0_LO",
		"lda #>main_players+4",
		"sta ZP_PTR0_HI",
		"ldy #0",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
		"sta main_x",
	)
}

func TestCodegenStructArrayFieldIntWithOffsetReadWrite(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    x: byte
    hp: int
}

fn main() {
    var players: Player[10]
    var i: byte
    var hp: int

    i = 2
    players[i].hp = 1000
    hp = players[i].hp
}
`)

	requireASM(t, asm,
		"main_players:",
		".word 10",
		".word 10",
		".fill 1, 0",
		".fill 2, 0",
		"adc #1",
		"sta ZP_PTR0_LO",
		"sta (ZP_PTR0_LO), y",
		"iny",
		"sta (ZP_PTR0_LO), y",
		"lda (ZP_PTR0_LO), y",
		"sta ZP_TMP0",
		"sta main_hp",
		"sta main_hp+1",
	)
}

func TestCodegenStructCharArrayFieldStringAssignment(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    name: char[16]
}

fn main() {
    var p: Player

    p.name = "player00"
}
`)

	requireASM(t, asm,
		"main_p:",
		".word 16",
		".word 0",
		".fill 16, 0",
		"literal_0:",
		"ldy #2",
		"ldy #3",
		"ldy #0",
		"lda literal_0, y",
		"sta (ZP_PTR0_LO), y",
	)
}

func TestCodegenStructArrayCharArrayFieldStringAssignment(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var players: Player[100]
    var i: byte

    i = 7
    players[i].name = "player07"
}
`)

	requireASM(t, asm,
		"main_players:",
		".word 100",
		".word 100",
		".word 16",
		".word 0",
		".fill 16, 0",
		"lda #<main_players+4",
		"sta ZP_PTR0_LO",
		"lda #>main_players+4",
		"sta ZP_PTR0_HI",
		"adc #1",
		"sta ZP_PTR0_LO",
		"ldy #2",
		"ldy #3",
		"ldy #0",
		"lda literal_0, y",
		"sta (ZP_PTR0_LO), y",
	)
}
