package sema

import "testing"

func TestSemaAllowsGotoXY(t *testing.T) {
	err := checkSource(t, `
fn main() {
    gotoxy(0, 24)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsGotoXYWithVariables(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var x byte
    var y byte

    x = 10
    y = 8

    gotoxy(x, y)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsGotoXYWrongArgCount(t *testing.T) {
	err := checkSource(t, `
fn main() {
    gotoxy(0)
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsGotoXYStringArgument(t *testing.T) {
	err := checkSource(t, `
fn main() {
    gotoxy("X", 0)
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}
