package sema

import "testing"

func TestSemaAllowsArrayParameterReferenceMutation(t *testing.T) {
	src := `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums byte[4]
    push(nums, 7)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsArrayParameterForwarding(t *testing.T) {
	src := `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn pushTwo(nums byte[4]) {
    push(nums, 1)
    push(nums, 2)
}

fn main() {
    var nums byte[4]
    pushTwo(nums)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsCharArrayParameterBuiltins(t *testing.T) {
	src := `
fn setTitle(title char[8]) {
    copy(title, "OK")
    append(title, "!")
}

fn main() {
    var title char[8]
    setTitle(title)
    print(title)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaAllowsStructArrayParameterMutation(t *testing.T) {
	src := `
struct Player {
    hp byte
}

fn damage(players Player[2], idx byte) {
    players[idx].hp = players[idx].hp - 1
}

fn main() {
    var players Player[2]
    damage(players, 1)
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsArrayParameterCapacityMismatch(t *testing.T) {
	src := `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums byte[2]
    push(nums, 7)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsArrayParameterElementTypeMismatch(t *testing.T) {
	src := `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums int[4]
    push(nums, 7)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsArrayElementAsArrayParameter(t *testing.T) {
	src := `
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums byte[4]
    push(nums[0], 7)
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}

func TestSemaRejectsStringLiteralAsArrayParameter(t *testing.T) {
	src := `
fn show(line char[5]) {
    print(line)
}

fn main() {
    show("HELLO")
}
`

	if err := checkSource(t, src); err == nil {
		t.Fatalf("expected sema error")
	}
}
