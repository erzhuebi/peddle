package sema

import "testing"

func TestSemaFileBuiltinsAcceptValidCalls(t *testing.T) {
	src := `
fn main() {
    var name char[32]
    var data char[64]
    var bytes byte[64]
    var f byte
    var n int

    copy(name, "PEDDLEFILE")
    copy(data, "HELLO FILE")

    f = fileopen(name, "w", 8)
    n = filewrite(f, data, len(data))
    fileclose(f)

    f = fileopen(name, "r", 8)
    n = fileread(f, bytes, size(bytes))
    fileclose(f)

    n = filesave(name, data, len(data), 8)
    n = fileload("PEDDLEFILE", data, 8)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaFileLoadRejectsScalarBuffer(t *testing.T) {
	src := `
fn main() {
    var n int

    n = fileload("PEDDLEFILE", n, 8)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaFileSaveRejectsNonNumericLen(t *testing.T) {
	src := `
fn main() {
    var name char[32]
    var data byte[64]

    filesave(name, data, data, 8)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaFileOpenRejectsNonCharName(t *testing.T) {
	src := `
fn main() {
    var name byte[32]
    var f byte

    f = fileopen(name, "r", 8)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaFileCloseRejectsArrayHandle(t *testing.T) {
	src := `
fn main() {
    var handle byte[1]

    fileclose(handle)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaFileBuiltinsRejectWrongArity(t *testing.T) {
	src := `
fn main() {
    var n int

    n = fileload("PEDDLEFILE")
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
