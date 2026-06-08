# Networking and Terminal

Peddle includes a first network API for the C64 Ultimate modem simulator.

This chapter documents the public Peddle network builtins, the reference Python terminal server, the `term.ped` example, and TEP, the Tiny Escape Protocol used by the terminal example.

---

# Network Builtins

```peddle
netconnect(addr char[], port int) bool
netbuffer(backlog byte[])
netavailable() int
netread(buffer byte[]|char[], max int, timeoutTicks int) int
netreadlf(buffer byte[]|char[], max int, timeoutTicks int) bool
netwrite(buffer byte[]|char[], len int) int
netclose()
netconnected() bool
```

Networking currently supports one global connection. There are no socket handles in this version.

The C64 Ultimate modem simulator is accessed through a SwiftLink/6551 ACIA-compatible interface at `$de00`.

`netconnect(addr, port)` opens the connection and returns `true` when the modem reports `CONNECT`.

`netbuffer(backlog)` reserves a caller-owned `byte[]` as exclusive network receive backlog storage.

`netavailable()` returns how many bytes are currently queued in that Peddle runtime backlog.

`netwrite(buffer, len)` writes bytes from a `byte[]` or `char[]` buffer.

`netclose()` closes the current connection and clears the runtime backlog.

`netconnected()` returns whether the global connection is currently open.

---

# netavailable()

```peddle
n = netavailable()
```

`netavailable()` returns the number of bytes currently queued in the Peddle runtime backlog configured with `netbuffer(backlog)`.

It does not include bytes still waiting inside the C64 Ultimate modem/TCP layer. Those bytes become visible to Peddle only when `netread()` or `netreadlf()` services the ACIA.

If `netbuffer(backlog)` has not been called, `netavailable()` returns `0`.

`netclose()` clears the runtime backlog, so read wanted bytes before closing the connection.

---

# netread()

```peddle
n = netread(rx, size(rx), 0)
```

`netread(buffer, max, timeoutTicks)` reads available bytes into `buffer` and returns the number of bytes stored.

Optional backlog storage can be configured once:

```peddle
var backlog byte[2048]

netbuffer(backlog)
```

After `netbuffer(backlog)`, the supplied `byte[]` belongs to the network runtime. Do not use it for normal application data.

The timeout is measured in C64 ticks/jiffies, not milliseconds.

- `0` means non-blocking
- around `50` ticks is about one second on PAL
- around `60` ticks is about one second on NTSC

Current behavior is intended for games and interactive programs:

- if backlog bytes already exist, copy those oldest bytes into the destination first
- if at least one byte is already available, read the available burst and return as soon as no more byte is immediately available
- if the destination fills and a backlog buffer was configured, continue draining immediately available modem bytes into the backlog
- if both destination and backlog are full, stop reading from the modem so lower network layers can apply backpressure
- if no byte is available at all, wait up to `timeoutTicks`
- if `timeoutTicks` is `0`, return immediately
- the destination array length is updated to the number of bytes read

For a game loop, use timeout `0` and call `netread()` every frame or update slot.

---

# netreadlf()

```peddle
found = netreadlf(line, size(line), 0)
```

`netreadlf(buffer, max, timeoutTicks)` appends bytes into `buffer` until CR or LF is found, the buffer is full, or the read returns no more bytes.

If `netbuffer(backlog)` is configured, `netreadlf()` also reads from and preserves extra bytes in the same backlog.

It returns:

- `true` when a line terminator was found
- `false` when no complete line is available yet

The CR/LF terminator is not stored in the buffer. This lets a program keep the same buffer across multiple game-loop ticks until a full line has arrived.

---

# Reference Terminal Files

The reference terminal service is:

```text
scripts/c64_term_server.py
```

The reference Peddle terminal client is:

```text
examples/term.ped
```

Run the server:

```sh
python3 scripts/c64_term_server.py
```

Deploy the client:

```sh
./peddle.sh --deploy --host 192.168.106.29 examples/term.ped
```

The server starts a real pseudo-terminal with a `40x24` terminal size. It receives VT100/ANSI output from the Mac/Linux shell and translates that output into TEP commands for the C64 client.

The Peddle client does not parse full VT100/ANSI. It only parses TEP and draws to C64 screen RAM.

---

# TEP: Tiny Escape Protocol

TEP is the small command protocol used between `scripts/c64_term_server.py` and `examples/term.ped`.

The ESC byte is decimal `27`, hexadecimal `$1b`.

Each TEP command is:

```text
<ESC><command><payload><ESC>
```

There are no spaces in the protocol. Spaces in documentation examples are only separators for readability.

For example, clear screen is exactly:

```text
<ESC>C<ESC>
```

The transmitted bytes are:

```text
27, 67, 27
```

---

# TEP Commands

| Command | Bytes | Meaning |
|---|---|---|
| `C` | `<ESC>C<ESC>` | clear screen and home cursor |
| `H` | `<ESC>H<ESC>` | home cursor |
| `M` | `<ESC>M<x><y><ESC>` | move cursor to zero-based column `x`, row `y` |
| `K` | `<ESC>K<ESC>` | erase from cursor to end of line |
| `J` | `<ESC>J<ESC>` | erase from cursor to end of screen |
| `L` | `<ESC>L<ESC>` | erase current line |
| `S` | `<ESC>S<ESC>` | scroll up one line |
| `Q` | `<ESC>Q<ESC>` | end terminal session |

The `M` command uses two raw byte values for the cursor position.

Move to column `4`, row `9` is:

```text
<ESC>M<x=4><y=9><ESC>
```

The transmitted bytes are:

```text
27, 77, 4, 9, 27
```

Those coordinate bytes are not ASCII characters. ASCII `"4"` and `"9"` would be bytes `52` and `57`, and must not be used for the `M` payload.

---

# Terminal Client Notes

`examples/term.ped` uses:

- `asciifont()` for ASCII-style lowercase and underscore display
- `toascii()` before sending typed keys to the server
- `netread(rx, size(rx), 0)` for non-blocking input from the server
- direct screen RAM drawing with `putraw()` and `putcolor()`

Interactive programs such as `top` need a live terminal. They do not fit a command/response model because they constantly redraw the screen using terminal control sequences.

The reference server handles the modern terminal side. The C64 only handles the compact TEP command stream.
