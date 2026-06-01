package sema

import "testing"

func TestSemaAllowsMultipleReturnAssignment(t *testing.T) {
	src := `
fn values() (byte, bool, char, int, uint) {
    var b byte
    var ok bool
    var c char
    var i int
    var u uint

    b = 7
    ok = true
    c = 'A'
    i = -3
    u = 65535

    return b, ok, c, i, u
}

fn main() {
    var b byte
    var ok bool
    var c char
    var i int
    var u uint

    b, ok, c, i, u = values()
    _, ok, _, i, _ = values()
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsWrongMultipleReturnValueCount(t *testing.T) {
	tests := []string{
		`
fn pair() (int, uint) {
    var a int
    return a
}

fn main() {
}
`,
		`
fn pair() (int, uint) {
    var a int
    var b uint
    var c byte
    return a, b, c
}

fn main() {
}
`,
	}

	for _, src := range tests {
		if err := checkSource(t, src); err == nil {
			t.Fatalf("expected sema error")
		}
	}
}

func TestSemaRejectsWrongMultipleAssignmentTargetCount(t *testing.T) {
	tests := []string{
		`
fn triple() (int, int, int) {
    var a int
    return a, a, a
}

fn main() {
    var a int
    var b int
    a, b = triple()
}
`,
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn main() {
    var a int
    var b int
    var c int
    a, b, c = pair()
}
`,
	}

	for _, src := range tests {
		if err := checkSource(t, src); err == nil {
			t.Fatalf("expected sema error")
		}
	}
}

func TestSemaRejectsMultipleAssignmentTypeMismatch(t *testing.T) {
	src := `
fn pair() (uint, int) {
    var u uint
    var i int
    return u, i
}

fn main() {
    var i int
    var u uint
    i, u = pair()
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsMultipleReturnCallInSingleValueContexts(t *testing.T) {
	tests := []string{
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn main() {
    var a int
    a = pair()
}
`,
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn take(a int) {
}

fn main() {
    take(pair())
}
`,
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn main() {
    if pair() {
    }
}
`,
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn main() {
    var a int
    a = pair() + 1
}
`,
		`
fn pair() (int, int) {
    var a int
    return a, a
}

fn main() {
    pair()
}
`,
	}

	for _, src := range tests {
		if err := checkSource(t, src); err == nil {
			t.Fatalf("expected sema error")
		}
	}
}
