package sema

import "testing"

func TestSemaAllowsMemWindowReadsWritesAndAddresses(t *testing.T) {
	src := `
fn set(x *byte) {
    x = 99
}

fn main() {
    var screen mem[1000] at $0400
    var b byte
    var addr uint
    var n int

    screen[0] = 65
    screen[999] = screen[0]
    b = screen[1]
    n = len(screen)
    n = size(screen)
    addr = &screen
    addr = &screen[2]
    set(&screen[3])
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsMemParameterPassingAndForwarding(t *testing.T) {
	src := `
fn put(buf mem[1000], index byte, value byte) {
    buf[index] = value
}

fn forward(buf mem[1000]) {
    put(buf, 2, 7)
}

fn main() {
    var screen mem[1000] at $0400

    put(screen, 1, 32)
    forward(screen)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsOverlappingMemWindows(t *testing.T) {
	src := `
fn main() {
    var screen mem[1000] at $0400
    var alias mem[40] at $0400

    screen[0] = 1
    alias[0] = 2
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsInvalidMemDeclarations(t *testing.T) {
	tests := []string{
		`
fn main() {
    var screen mem[1000]
}
`,
		`
fn main() {
    var screen mem[0] at $0400
}
`,
		`
fn main() {
    var screen mem[4] at $ffff
}
`,
		`
struct View {
    screen mem[1000]
}

fn main() {
}
`,
		`
fn bad() mem[4] {
    var screen mem[4] at $c000
    return screen
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

func TestSemaRejectsInvalidMemUse(t *testing.T) {
	tests := []string{
		`
fn put(buf mem[1000]) {
}

fn main() {
    var small mem[40] at $0400
    put(small)
}
`,
		`
fn main() {
    var screen mem[1000] at $0400
    screen[1000] = 1
}
`,
		`
fn main() {
    var screen mem[1000] at $0400
    screen[true] = 1
}
`,
		`
fn main() {
    var screen mem[1000] at $0400
    append(screen, 1)
}
`,
		`
fn main() {
    var a mem[10] at $0400
    var b mem[10] at $0500
    copy(a, b)
}
`,
		`
fn main() {
    var screen mem[1000] at $0400
    fill(screen, 0)
}
`,
		`
fn main() {
    var screen mem[1000] at $0400
    var addr uint
    addr = &screen[-1]
}
`,
		`
fn set(x *uint) {
    x = 1
}

fn main() {
    var screen mem[1000] at $0400
    set(&screen[0])
}
`,
	}

	for _, src := range tests {
		if err := checkSource(t, src); err == nil {
			t.Fatalf("expected sema error")
		}
	}
}
