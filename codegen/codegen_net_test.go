package codegen

import (
	"strings"
	"testing"
)

func TestCodegenNetBuiltins(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var addr char[128]
    var tx char[64]
    var rx byte[128]
    var ok bool
    var n int

    copy(addr, "example.com")
    copy(tx, "GET / HTTP/1.0")

    ok = netconnect(addr, 80)

    if ok {
        n = netwrite(tx, len(tx))
        n = netread(rx, size(rx), 50)
        netclose()
    }
}
`)

	requireASM(t, asm,
		"jsr peddle_netconnect",
		"jsr peddle_netwrite",
		"jsr peddle_netread",
		"jsr peddle_netclose",
		"peddle_netconnect:",
		"ACIA_DATA    = $de00",
		".byte 67, 79, 78, 78, 69, 67, 84, 13",
	)

	netconnect := netRuntimeBlock(t, asm, "peddle_netconnect:", "peddle_netconnect_fail:")

	requireContains(t, netconnect,
		"; Wait for the complete modem result line terminator. The first byte\n    ; after this match belongs to TCP payload and must not be flushed.",
		"lda #8\n    sta peddle_net_pattern_len",
		"lda #1\n    sta peddle_net_connected\n    lda #0\n    sta peddle_net_skip_lf\n    lda #1\n    rts",
	)

	requireNotContains(t, netconnect,
		"; The modem often leaves CR/LF after CONNECT.",
		"jsr peddle_net_guard_delay\n    jsr peddle_acia_flush\n\n    lda #1\n    sta peddle_net_connected",
	)

	netread := netRuntimeBlock(t, asm, "peddle_netread:", "peddle_netwrite:")

	requireContains(t, netread,
		"; If at least one byte was read, return immediately with the available\n    ; chunk. Timeout only waits for the first byte.\n    lda peddle_net_count_lo\n    ora peddle_net_count_hi\n    bne peddle_netread_done",
		"ldy #2\n    lda peddle_net_count_lo\n    sta (ZP_PTR0_LO), y\n    iny\n    lda peddle_net_count_hi\n    sta (ZP_PTR0_LO), y",
		"lda peddle_net_timeout_lo\n    ora peddle_net_timeout_hi\n    beq peddle_netread_done",
	)

	requireNotContains(t, netread,
		"Reset idle timeout after every received byte",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetConnected(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var ok bool

    ok = netconnected()
}
`)

	requireASM(t, asm,
		"lda peddle_net_connected",
	)
}

func TestCodegenNetAvailable(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var n int

    n = netavailable()
}
`)

	requireASM(t, asm,
		"jsr peddle_netavailable",
		"peddle_netavailable:",
		"lda peddle_net_ring_count_lo",
		"sta ZP_TMP0",
		"lda peddle_net_ring_count_hi",
		"sta ZP_TMP1",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetReadLF(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var line char[128]
    var found bool

    found = netreadlf(line, size(line), 0)
}
`)

	requireASM(t, asm,
		"jsr peddle_netreadlf",
		"peddle_netreadlf:",
		"peddle_net_skip_lf:",
	)

	netreadlf := netRuntimeBlock(t, asm, "peddle_netreadlf:", "peddle_netclose:")

	requireContains(t, netreadlf,
		"; Start with the existing destination length. netreadlf() appends into",
		"lda peddle_net_skip_lf\n    beq peddle_netreadlf_check_terminator",
		"cmp #13\n    beq peddle_netreadlf_found_cr",
		"cmp #10\n    beq peddle_netreadlf_found_lf",
		"sta peddle_net_skip_lf\n    sta peddle_net_line_found",
		"sta (ZP_PTR1_LO), y",
		"lda peddle_net_had_byte\n    bne peddle_netreadlf_done",
		"lda peddle_net_line_found\n    rts",
	)

	requireNotContains(t, netreadlf,
		"Clear destination array length",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetBufferSetupAndRuntime(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var backlog byte[1024]
    var rx byte[64]
    var n int

    netbuffer(backlog)
    n = netread(rx, size(rx), 0)
}
`)

	requireASM(t, asm,
		"jsr peddle_netbuffer",
		"peddle_net_ring_data_lo:",
		"peddle_net_ring_cap_lo:",
		"peddle_net_ring_read_lo:",
		"peddle_net_ring_count_lo:",
		"peddle_netbuffer:",
		"peddle_net_drain_ring_to_buffer:",
		"peddle_net_drain_acia_to_ring:",
	)

	netread := netRuntimeBlock(t, asm, "peddle_netread:", "peddle_netwrite:")
	requireASMOrder(t, netread,
		"    jsr peddle_net_drain_ring_to_buffer",
		"    jsr peddle_acia_can_read",
	)
	requireContains(t, netread,
		"peddle_netread_buffer_full:\n    ; The caller buffer is full. Preserve any immediately available extra\n    ; bytes in the user-provided backlog, stopping before ACIA_DATA if the\n    ; backlog is also full.\n    jsr peddle_net_drain_acia_to_ring",
	)

	drain := netRuntimeBlock(t, asm, "peddle_net_drain_acia_to_ring:", "peddle_net_drain_acia_to_ring_done:")
	requireContains(t, drain,
		"jsr peddle_net_ring_has_space\n    beq peddle_net_drain_acia_to_ring_done\n\n    jsr peddle_acia_can_read",
	)
	requireASMOrder(t, drain,
		"    jsr peddle_net_ring_has_space",
		"    jsr peddle_acia_can_read",
		"    jsr peddle_acia_read",
		"    jsr peddle_net_ring_push",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetReadLFUsesRingBufferSource(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var backlog byte[1024]
    var line char[80]
    var found bool

    netbuffer(backlog)
    found = netreadlf(line, size(line), 0)
}
`)

	netreadlf := netRuntimeBlock(t, asm, "peddle_netreadlf:", "peddle_netclose:")

	requireContains(t, netreadlf,
		"jsr peddle_net_ring_has_bytes",
		"jsr peddle_net_ring_pop",
		"jsr peddle_net_drain_acia_to_ring",
	)
	requireASMOrder(t, netreadlf,
		"    jsr peddle_net_ring_has_bytes",
		"    jsr peddle_acia_can_read",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetBuiltinsAcceptIndexedStructFieldArrays(t *testing.T) {
	asm := compileSource(t, `
struct Conn {
    addr char[32]
    tx char[64]
    rx byte[64]
    line char[64]
    backlog byte[128]
}

fn main() {
    var conns Conn[2]
    var i byte = 1
    var ok bool
    var found bool
    var n int

    copy(conns[i].addr, "example.com")
    copy(conns[i].tx, "PING")

    netbuffer(conns[i].backlog)
    ok = netconnect(conns[i].addr, 80)

    if ok {
        n = netwrite(conns[i].tx, len(conns[i].tx))
        n = netread(conns[i].rx, size(conns[i].rx), 0)
        found = netreadlf(conns[i].line, size(conns[i].line), 0)
        netclose()
    }
}
`)

	requireASM(t, asm,
		"jsr peddle_netbuffer",
		"jsr peddle_netconnect",
		"jsr peddle_netwrite",
		"jsr peddle_netread",
		"jsr peddle_netreadlf",
		"jsr peddle_netclose",
		"sta peddle_net_addr_lo",
		"sta peddle_net_buf_lo",
		"sta peddle_net_max_lo",
	)

	requireReferencedLabelsDefined(t, asm)
	requireASMAssemblesWith64tass(t, asm)
}

func TestCodegenNetRuntimeNotEmittedWithoutNetworkBuiltins(t *testing.T) {
	asm := compileSource(t, `
fn main() {
    var x int

    x = 1
}
`)

	requireNotContains(t, asm,
		"peddle_netbuffer:",
		"peddle_netavailable:",
		"peddle_net_ring_data_lo:",
		"peddle_net_drain_ring_to_buffer:",
	)
}

func netRuntimeBlock(t *testing.T, asm string, start string, end string) string {
	t.Helper()

	startIdx := strings.Index(asm, start)
	if startIdx < 0 {
		t.Fatalf("ASM does not contain start marker %q\n\nASM:\n%s", start, asm)
	}

	endIdx := strings.Index(asm[startIdx:], end)
	if endIdx < 0 {
		t.Fatalf("ASM does not contain end marker %q after %q\n\nASM:\n%s", end, start, asm)
	}

	return asm[startIdx : startIdx+endIdx]
}

func requireContains(t *testing.T, text string, parts ...string) {
	t.Helper()

	for _, part := range parts {
		if !strings.Contains(text, part) {
			t.Fatalf("text does not contain %q\n\nTEXT:\n%s", part, text)
		}
	}
}

func requireNotContains(t *testing.T, text string, parts ...string) {
	t.Helper()

	for _, part := range parts {
		if strings.Contains(text, part) {
			t.Fatalf("text unexpectedly contains %q\n\nTEXT:\n%s", part, text)
		}
	}
}
