package sema

import (
	"strings"
	"testing"
)

func TestSemaAllowsItoxForNumericScalars(t *testing.T) {
	src := `
fn main() {
    var b byte
    var ch char
    var ok bool
    var i int
    var h2 char[2]
    var h4 char[4]

    h2 = itox(b)
    h2 = itox(ch)
    h2 = itox(ok)
    h4 = itox(i)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaItoxIntReturnsFourDigitString(t *testing.T) {
	src := `
fn main() {
    var i int
    var h2 char[2]

    h2 = itox(i)
}
`

	err := checkSource(t, src)
	if err == nil {
		t.Fatalf("expected itox(int) assignment to char[2] to fail")
	}
}

func TestSemaRejectsItoxArrayArgument(t *testing.T) {
	src := `
fn main() {
    var data byte[4]
    var h char[4]

    h = itox(data)
}
`

	err := checkSource(t, src)
	if err == nil {
		t.Fatalf("expected itox(array) to fail")
	}

	if !strings.Contains(err.Error(), "itox argument must be numeric") {
		t.Fatalf("expected itox numeric error, got: %v", err)
	}
}

func TestSemaRejectsItoxWrongArgCount(t *testing.T) {
	src := `
fn main() {
    var h char[4]

    h = itox(1, 2)
}
`

	err := checkSource(t, src)
	if err == nil {
		t.Fatalf("expected itox wrong arg count to fail")
	}

	if !strings.Contains(err.Error(), "itox expects one argument") {
		t.Fatalf("expected itox arg count error, got: %v", err)
	}
}
