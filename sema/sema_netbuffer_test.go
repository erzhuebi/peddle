package sema

import "testing"

func TestSemaNetBufferAcceptsByteArray(t *testing.T) {
	src := `
fn main() {
    var backlog byte[1024]

    netbuffer(backlog)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaNetBufferRejectsCharArray(t *testing.T) {
	src := `
fn main() {
    var backlog char[1024]

    netbuffer(backlog)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaNetBufferRejectsScalar(t *testing.T) {
	src := `
fn main() {
    var backlog int

    netbuffer(backlog)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaNetBufferRejectsWrongArity(t *testing.T) {
	src := `
fn main() {
    netbuffer()
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaNetAvailableReturnsInt(t *testing.T) {
	src := `
fn main() {
    var n int

    n = netavailable()
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaNetAvailableRejectsArguments(t *testing.T) {
	src := `
fn main() {
    var backlog byte[32]

    netavailable(backlog)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
