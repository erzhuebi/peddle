package sema

import "testing"

func TestSemaAllowsMixedNumericExpressions(t *testing.T) {
	src := `
fn main() {
    var b: byte
    var i: int
    var x: int

    b = 10
    i = 1000
    x = b + i
    x = i + b
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsDirectBoolCondition(t *testing.T) {
	src := `
fn main() {
    var done: bool

    done = 1

    if done {
        print("OK")
    }
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsDirectIntCondition(t *testing.T) {
	src := `
fn main() {
    var x: int

    x = 1

    if x {
        print("OK")
    }
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsBinaryStringMath(t *testing.T) {
	src := `
fn main() {
    var x: int
    x = "A" + 1
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsUnaryMinusString(t *testing.T) {
	src := `
fn main() {
    var x: int
    x = -"A"
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsUnaryBangString(t *testing.T) {
	src := `
fn main() {
    var b: bool
    b = !"A"
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsInvalidArrayIndexType(t *testing.T) {
	src := `
fn main() {
    var a: byte[4]
    var s: char[4]

    s = "HEY"
    a[s] = 1
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaAllowsMultipleReturns(t *testing.T) {
	src := `
fn choose(x: int) -> int {
    if x {
        return 1
    }

    return 2
}

fn main() {
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}
