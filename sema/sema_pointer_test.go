package sema

import "testing"

func TestSemaAllowsStructPointerParameterMutation(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var player Player
    damage(&player)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsStructArrayElementPointerArgument(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var players Player[4]
    var i byte

    i = 2
    damage(&players[i])
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsPointerArgumentWithoutAddressOf(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var player Player
    damage(player)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsWrongStructPointerArgument(t *testing.T) {
	src := `
struct Player {
    hp byte
}

struct Enemy {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var enemy Enemy
    damage(&enemy)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaAllowsScalarPointerParameterAliases(t *testing.T) {
	src := `
fn bumpByte(x *byte) {
    x = x + 1
}

fn setBool(x *bool) {
    x = true
}

fn setChar(x *char) {
    x = 'Z'
}

fn bumpInt(x *int) {
    x = x + 1
}

fn bumpUint(x *uint) {
    x = x + 1
}

fn main() {
    var b byte
    var flag bool
    var ch char
    var i int
    var u uint

    bumpByte(&b)
    setBool(&flag)
    setChar(&ch)
    bumpInt(&i)
    bumpUint(&u)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsScalarArrayElementPointerArgument(t *testing.T) {
	src := `
fn bump(x *uint) {
    x = x + 1
}

fn main() {
    var values uint[4]
    var i byte

    i = 2
    bump(&values[i])
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsAddressOfScalarForPointerArgument(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn damage(p *Player) {
    p.hp = p.hp - 1
}

fn main() {
    var x byte
    damage(&x)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsWrongScalarPointerArgument(t *testing.T) {
	src := `
fn bump(x *uint) {
    x = x + 1
}

fn main() {
    var i int
    bump(&i)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsScalarPointerArgumentWithoutAddressOf(t *testing.T) {
	src := `
fn bump(x *uint) {
    x = x + 1
}

fn main() {
    var u uint
    bump(u)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsArrayPointerParameter(t *testing.T) {
	src := `
fn bad(values *byte[4]) {
    return
}

fn main() {
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPointerParamForwarding(t *testing.T) {
	src := `
fn inner(x *uint) {
    x = x + 1
}

fn outer(x *uint) {
    inner(&x)
}

fn main() {
    var u uint
    outer(&u)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
