package sema

import "testing"

func TestSemaAllowsUintLiteralsAndAddressOfVariables(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn take(addr uint) {
    return
}

fn main() {
    var addr uint
    var max uint
    var border uint
    var x byte
    var player Player
    var data byte[4]

    max = 65535
    border = 0xd020
    addr = &x
    addr = &player
    addr = &data
    take(&x)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsUintUnsignedComparisons(t *testing.T) {
	src := `
fn main() {
    var hi uint
    var mid uint
    var ok bool

    hi = 65535
    mid = 32768
    ok = hi > mid
    ok = mid < hi
    ok = hi >= 65535
    ok = mid <= hi
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsUintNegativeLiteral(t *testing.T) {
	src := `
fn main() {
    var u uint
    u = -1
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsUintTooLargeLiteral(t *testing.T) {
	src := `
fn main() {
    var u uint
    u = 65536
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsSignedIntVariableToUint(t *testing.T) {
	src := `
fn main() {
    var i int
    var u uint

    i = -1
    u = i
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsSignedIntArgumentToUintParam(t *testing.T) {
	src := `
fn take(addr uint) {
    return
}

fn main() {
    var i int
    i = -1
    take(i)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
