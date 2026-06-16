package sema

import "testing"

func TestSemaAllowsSoundBuiltins(t *testing.T) {
	src := `
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1

fn main() {
    var pool byte[64]
    var data byte[16]
    var id uint
    var err int
    var n int

    data[0] = 0

    sound_init(pool)
    sound_reset()
    id, err = sound_load(data, SOUND_STREAM)
    err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    sound_stop(id)
    n = sound_num()
    n = sound_memfree()
}
`

	if err := checkSource(t, src); err != nil {
		t.Fatalf("unexpected sema error: %v", err)
	}
}

func TestSemaRejectsInvalidSoundCalls(t *testing.T) {
	tests := []string{
		`
fn main() {
    var data byte[16]
    sound_init(data, data)
}
`,
		`
fn main() {
    var pool int[16]
    sound_init(pool)
}
`,
		`
fn main() {
    var data char[16]
    var id uint
    var err int
    id, err = sound_load(data, 1)
}
`,
		`
fn main() {
    var data byte[16]
    var id uint
    var err int
    id, err = sound_load(data, "BAD")
}
`,
		`
fn main() {
    var data byte[16]
    sound_load(data, 1)
}
`,
		`
fn main() {
    var data byte[16]
    var id uint
    id = sound_load(data, 1)
}
`,
		`
fn main() {
    var data byte[16]
    var id uint
    id = sound_play(data)
}
`,
		`
fn main() {
    var data byte[16]
    var id uint
    var err int
    var extra int
    id, err, extra = sound_load(data, 1)
}
`,
		`
fn main() {
    var data byte[16]
    var id int
    var err uint
    id, err = sound_load(data, 1)
}
`,
	}

	for _, src := range tests {
		if err := checkSource(t, src); err == nil {
			t.Fatalf("expected sema error for:\n%s", src)
		}
	}
}
