package sema

import "testing"

func TestSemaAllowsPutStrWithStringLiteral(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstr(0, 0, "HELLO")
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsPutStrColorWithStringLiteral(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstrcolor(0, 1, "HELLO", 2)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsPutStrWithNumericVariables(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var x byte
    var y byte

    x = 10
    y = 5

    putstr(x, y, "HELLO")
    putstrcolor(x, y, "COLOR", 7)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsPutStrCharArray(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var title char[16]

    copy(title, "HELLO")
    putstr(0, 0, title)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsPutStrColorCharArray(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var title char[16]

    copy(title, "HELLO")
    putstrcolor(0, 0, title, 2)
}
`)

	if err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsPutStrWrongArgCount(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstr(0, 0)
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPutStrColorWrongArgCount(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstrcolor(0, 0, "HELLO")
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPutStrNonNumericCoordinate(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstr("X", 0, "HELLO")
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPutStrColorNonNumericColor(t *testing.T) {
	err := checkSource(t, `
fn main() {
    putstrcolor(0, 0, "HELLO", "RED")
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPutStrNonCharArray(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var data byte[16]

    putstr(0, 0, data)
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsPutStrColorNonCharArray(t *testing.T) {
	err := checkSource(t, `
fn main() {
    var data byte[16]

    putstrcolor(0, 0, data, 2)
}
`)

	if err == nil {
		t.Fatalf("expected sema error")
	}
}
