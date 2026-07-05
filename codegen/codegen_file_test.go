package codegen

import "testing"

func TestCodegenFileBuiltins(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var data char[64]
    var rx byte[64]
    var f byte
    var n int

    copy(data, "HELLO FILE")

    f = fileopen("PEDDLEFILE", "w", 8)
    n = filewrite(f, data, len(data))
    fileclose(f)

    f = fileopen("PEDDLEFILE", "r", 8)
    n = fileread(f, rx, size(rx))
    fileclose(f)

    n = filesave("PEDDLEFILE", data, len(data), 8)
    n = fileload("PEDDLEFILE", rx, 8)
}
`)

	requireASM(t, asm,
		"jsr peddle_fileopen",
		"jsr peddle_filewrite",
		"jsr peddle_fileclose",
		"jsr peddle_fileread",
		"jsr peddle_filesave",
		"jsr peddle_fileload",
		"peddle_fileopen:",
		"peddle_fileload:",
		"peddle_filesave:",
		"KERNAL_SETLFS = $ffba",
		"KERNAL_SETNAM = $ffbd",
		"KERNAL_OPEN   = $ffc0",
		"KERNAL_CHRIN  = $ffcf",
		"KERNAL_CHROUT = $ffd2",
		"PEDDLE_FILE_LFN       = 1",
		"peddle_file_build_read_suffix:",
		"peddle_file_build_write_suffix:",
	)

	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenFileReadClearsAndStoresArrayLength(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var rx byte[64]
    var f byte
    var n int

    f = fileopen("PEDDLEFILE", "r", 8)
    n = fileread(f, rx, size(rx))
}
`)

	fileread := netRuntimeBlock(t, asm, "peddle_fileread:", "peddle_filewrite:")
	requireContains(t, fileread,
		"jsr peddle_file_prepare_read_buffer",
		"jsr KERNAL_CHKIN",
		"jsr KERNAL_CHRIN",
		"jsr peddle_file_store_buffer_length",
	)

	prepareRead := netRuntimeBlock(t, asm, "peddle_file_prepare_read_buffer:", "peddle_file_prepare_write_buffer:")
	requireContains(t, prepareRead,
		"; Clear destination array length before reading.",
		"ldy #2\n    lda #0\n    sta (ZP_PTR0_LO), y\n    iny\n    sta (ZP_PTR0_LO), y",
	)

	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenFileBuiltinsAcceptIndexedStructFieldArrays(t *testing.T) {
	asm := compileSource(t, `
struct FileSlot {
    name char[32]
    text char[64]
    bytes byte[64]
}

fn main() {
    var slots FileSlot[2]
    var i byte = 1
    var f byte
    var n int

    copy(slots[i].name, "PEDDLEFILE")
    copy(slots[i].text, "HELLO")

    f = fileopen(slots[i].name, "w", 8)
    n = filewrite(f, slots[i].text, len(slots[i].text))
    fileclose(f)

    n = filesave(slots[i].name, slots[i].text, len(slots[i].text), 8)
    n = fileload(slots[i].name, slots[i].bytes, 8)
}
`)

	requireASM(t, asm,
		"jsr peddle_fileopen",
		"jsr peddle_filewrite",
		"jsr peddle_fileclose",
		"jsr peddle_filesave",
		"jsr peddle_fileload",
		"sta peddle_file_name_lo",
		"sta peddle_file_buf_lo",
		"sta peddle_file_max_lo",
	)

	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenFileRuntimeNotEmittedWithoutFileBuiltins(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int

    x = 1
}
`)

	requireNotContains(t, asm,
		"peddle_fileopen:",
		"peddle_file_name_buffer:",
		"KERNAL_SETLFS = $ffba",
	)
}
