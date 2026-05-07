package codegen

import "testing"

func TestCodegenNonArrayStructFieldReadWrite(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id: byte
    hp: int
}

fn main() {
    var p: Player
    var id: byte
    var hp: int

    p.id = 7
    p.hp = 120

    id = p.id
    hp = p.hp
}
`)

	requireASM(t, asm,
		"main_p:",
		"sta main_p",
		"sta main_p+1",
		"sta main_p+2",
		"lda main_p",
		"lda main_p+1",
		"lda main_p+2",
		"sta main_id",
		"sta main_hp",
		"sta main_hp+1",
	)
}

func TestCodegenStructArrayFieldInsideNonArrayStruct(t *testing.T) {
	asm := compileSource(t, `
struct Message {
    text: char[16]
    count: int
}

fn main() {
    var msg: Message
    var n: int

    copy(msg.text, "HELLO")
    append(msg.text, "!")
    print(msg.text)

    n = len(msg.text)

    clear(msg.text)

    if len(msg.text) == 0 {
        print("CLEARED")
    }

    msg.count = n
}
`)

	requireASM(t, asm,
		"main_msg:",
		".word 16",
		".word 0",
		"literal_0:",
		"literal_1:",
		"jsr peddle_print_counted_string",
		"ldy #2",
		"lda #0",
		"sta main_msg+20",
		"sta main_msg+21",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenStructArrayFieldAccessAssembles(t *testing.T) {
	asm := compileSource(t, `
struct Player {
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var p: Player

    p.id = 1
    copy(p.name, "ADA")
    append(p.name, "!")
    p.hp = 120

    print(p.name)
}
`)

	requireASM(t, asm,
		"main_p:",
		".word 16",
		".word 0",
		"literal_0:",
		"literal_1:",
		"jsr peddle_print_counted_string",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}
