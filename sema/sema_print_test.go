package sema

import "testing"

func TestSemaAllowsPrintTemporaryCharArrays(t *testing.T) {
	src := `
fn main() {
    var score int
    var b byte

    score = -123
    b = 27
    print(itoa(score))
    print(itox(b))
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}
