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
