package sema

import "testing"

func TestSemaAllowsStructFieldAccess(t *testing.T) {
	src := `
struct Player {
    x: byte
    hp: int
}

fn main() {
    var p: Player
    var hp: int

    p.x = 10
    p.hp = 100
    hp = p.hp
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsUnknownStructField(t *testing.T) {
	src := `
struct Player {
    x: byte
}

fn main() {
    var p: Player

    p.y = 10
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsFieldAccessOnNonStruct(t *testing.T) {
	src := `
fn main() {
    var x: int

    x.y = 10
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
