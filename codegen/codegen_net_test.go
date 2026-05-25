package codegen

import "testing"

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
