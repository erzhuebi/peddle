# Builtins and C64 Screen

# Builtin Functions

| Function | Description |
|---|---|
| `print(x)` | print string using KERNAL-style output |
| `peek(addr)` | read memory |
| `poke(addr, value)` | write memory |
| `len(array)` | runtime length |
| `size(array)` | array capacity |
| `append(array, value)` | append element |
| `copy(dst, src)` | copy arrays/strings |
| `fill(array, value)` | fill array |
| `clear(array)` | clear runtime length |
| `itoa(value)` | convert signed `int` to temporary decimal `char[]` |
| `cls()` | clear screen RAM and reset KERNAL text cursor |
| `border(color)` | set border color using `$d020` |
| `background(color)` | set background color using `$d021` |
| `textcolor(color)` | set KERNAL text color using `$0286` |
| `gotoxy(x, y)` | set KERNAL text cursor position |
| `putchar(x, y, ch)` | write a character to screen RAM at position |
| `putraw(x, y, code)` | write raw C64 screen code to screen RAM |
| `putcolor(x, y, color)` | write color RAM value at position |
| `putcharcolor(x, y, ch, color)` | write a character and color RAM value at position |
| `putstr(x, y, text)` | write a string literal or `char[]` directly to screen RAM |
| `putstrcolor(x, y, text, color)` | write a string literal or `char[]` to screen RAM and color RAM |
| `key()` | non-blocking keyboard read; returns `0` if no key is waiting |
| `waitkey()` | blocking keyboard read; waits until a key is available |
| `readline(buffer, echo, max)` | read a line into a `char[]` buffer |
| `joy(port)` | read joystick port state |
| `ticks()` | read 16-bit system tick counter |
| `elapsed(last)` | wrap-safe elapsed ticks since `last` |
| `tickdue(last, interval)` | wrap-safe check if an interval has passed |
| `asciifont()` | install a RAM font for ASCII-style terminal output |
| `toascii(buffer)` | convert a `char[]` from C64 keyboard/PETSCII-style text to ASCII-style text |
| `topetscii(buffer)` | convert a `char[]` from ASCII-style text to C64/PETSCII-style text |
| `netconnect(addr, port)` | connect using the C64 Ultimate modem simulator |
| `netbuffer(backlog)` | reserve a byte array as network receive backlog |
| `netavailable()` | return bytes currently queued in the network receive backlog |
| `netread(buffer, max, timeoutTicks)` | read available network bytes |
| `netreadlf(buffer, max, timeoutTicks)` | read network bytes until CR or LF |
| `netwrite(buffer, len)` | write network bytes |
| `netclose()` | close the current network connection |
| `netconnected()` | check whether the global network connection is open |
| `fileload(name, buffer, device)` | load a whole file into an array |
| `filesave(name, buffer, len, device)` | save array bytes to a file |
| `fileopen(name, mode, device)` | open a file stream |
| `fileread(handle, buffer, max)` | read bytes from a file stream |
| `filewrite(handle, buffer, len)` | write bytes to a file stream |
| `fileclose(handle)` | close a file stream |
| `sound_init(pool)` | initialize sound runtime with a user-provided byte pool |
| `sound_reset()` | stop playback and clear loaded sounds |
| `sound_load(data, kind)` | load sound data and return `(uint, int)` |
| `sound_play(id, voices, priority, flags)` | play a loaded sound and return `int` |
| `sound_stop(id)` | stop active players using the handle |
| `sound_stop_voices(voices)` | stop the active players' selected SID voices |
| `sound_num()` | return number of loaded sounds |
| `sound_memfree()` | return remaining sound pool bytes |

---

# C64 Screen Builtins

Peddle provides direct C64 screen and color helpers for text-mode screen programming.

Important distinction:

- `print()` uses the KERNAL cursor
- `gotoxy()` moves the KERNAL cursor
- `textcolor()` affects KERNAL `print()`
- `putchar()`, `putraw()`, `putstr()`, `putcolor()`, and `putcharcolor()` write directly to screen/color RAM

---

# cls()

Clear the visible screen and reset the KERNAL text cursor to the top-left position.

```peddle
fn main() {
    cls()
    print("READY")
}
```

`cls()` clears screen RAM from `$0400` to `$07e7` with space characters and resets the KERNAL cursor using the KERNAL `PLOT` routine.

This means `print()` starts at the top-left after `cls()`.

---

# border()

Set the C64 border color.

```peddle
border(6)
```

This writes to `$d020`.

---

# background()

Set the C64 background color.

```peddle
background(0)
```

This writes to `$d021`.

---

# textcolor()

Set the KERNAL text color used by `print()`.

```peddle
textcolor(1)
print("WHITE TEXT")
```

This writes to `$0286`.

`textcolor()` affects KERNAL-style `print()` output. It does not automatically affect direct screen RAM writes such as `putchar()` or `putraw()`.

---

# gotoxy()

Set the KERNAL text cursor position.

```peddle
gotoxy(10, 8)
print("HELLO")
```

`gotoxy(x, y)` moves the KERNAL cursor to column `x`, row `y`.

It affects `print()` output.

It does not affect direct screen RAM functions such as:

- `putchar()`
- `putraw()`
- `putstr()`
- `putcolor()`
- `putcharcolor()`

Coordinates outside the visible screen are clipped. If `x >= 40` or `y >= 25`, `gotoxy()` does nothing.

Example:

```peddle
fn main() {
    cls()

    gotoxy(0, 5)
    print("ROW 5")

    gotoxy(10, 8)
    print("COL 10 ROW 8")

    gotoxy(0, 22)
}
```

The final `gotoxy()` is useful when running from BASIC, because BASIC prints `READY.` at the current KERNAL cursor position after the program exits.

Avoid parking the cursor on the very last row unless you intentionally want BASIC output to scroll the screen after the program exits. `gotoxy(0, 22)` or `gotoxy(0, 23)` is usually safer than `gotoxy(0, 24)`.

---

# putchar()

Write a character at a fixed screen position.

```peddle
putchar(0, 0, 'P')
putchar(1, 0, 'E')
putchar(2, 0, 'D')
```

`putchar(x, y, ch)` writes to screen RAM at:

```text
$0400 + y * 40 + x
```

`putchar()` accepts normal character values and converts letters to C64 screen codes before writing them.

For example:

```peddle
putchar(0, 0, 'P')
putchar(1, 0, 80)
```

Both write the letter `P`.

The conversion uses Peddle's character-to-screen-code table.

Currently the important mappings include:

- `@` to C64 screen code `0`
- `A`..`Z` to C64 screen codes `1`..`26`
- `a`..`z` to C64 screen codes `1`..`26`
- space and digits keep their usual values

So `@`, space, letters, and digits work naturally:

```peddle
putchar(0, 0, ' ')
putchar(1, 0, '1')
```

`putchar()` is clipped to the visible screen. If `x >= 40` or `y >= 25`, it does nothing.

---

# putraw()

Write a raw C64 screen code at a fixed screen position.

```peddle
putraw(0, 0, 16)
putraw(1, 0, 5)
putraw(2, 0, 4)
```

`putraw()` does no character conversion.

Use it when you already know the C64 screen code you want to write.

`putraw()` is clipped to the visible screen. If `x >= 40` or `y >= 25`, it does nothing.

---

# putcolor()

Write a color value to color RAM at a fixed screen position.

```peddle
putcolor(0, 0, 2)
putcolor(1, 0, 3)
```

`putcolor(x, y, color)` writes to color RAM at:

```text
$d800 + y * 40 + x
```

`putcolor()` is clipped to the visible screen. If `x >= 40` or `y >= 25`, it does nothing.

---

# putcharcolor()

Write one character and its color at a fixed screen position.

```peddle
putcharcolor(0, 0, 'P', 2)
putcharcolor(1, 0, 'E', 3)
```

`putcharcolor(x, y, ch, color)` combines `putchar()` and `putcolor()` for one
cell. It converts `ch` through the same character-to-screen-code table as
`putchar()`, writes the converted character to screen RAM, then writes `color`
to the matching color RAM position.

`putcharcolor()` is clipped to the visible screen. If `x >= 40` or `y >= 25`, it
does nothing.

---

# putstr()

Write a string directly to screen RAM.

```peddle
putstr(0, 0, "HELLO")
putstr(0, 1, "PEDDLE")
```

`putstr(x, y, text)` starts at column `x`, row `y` and writes converted C64 screen codes to screen RAM.

It does not use the KERNAL cursor and does not affect where `print()` will write next.

`putstr()` does not change color RAM. Use `putstrcolor()` when you want to write text and color together.

`putstr()` supports:

- string literals
- `char[]` variables
- `char[]` struct fields
- temporary `char[]` results such as `itoa(...)`

```peddle
var title char[32]
var score int

copy(title, "SCORE ")
append(title, itoa(score))

putstr(0, 0, "READY")
putstr(0, 1, title)
putstr(0, 2, itoa(score))
```

---

## putstr() Character Conversion

`putstr()` uses the same character-to-screen-code conversion as `putchar()`.

The conversion uses Peddle's character-to-screen-code table.

Currently the important mappings include:

- `@` to C64 screen code `0`
- `A`..`Z` to C64 screen codes `1`..`26`
- `a`..`z` to C64 screen codes `1`..`26`
- space and digits keep their usual values

---

## putstr() Newlines

Inside `putstr()`, `\n` and `\r` move to the next row and return to the original start column.

```peddle
putstr(5, 3, "A\nB")
```

This writes:

```text
A at column 5, row 3
B at column 5, row 4
```

---

## putstr() Clipping

`putstr()` never writes past the visible screen area.

If the start position is outside the screen, it does nothing.

If the string reaches the end of the visible screen, it stops safely.

```peddle
putstr(0, 24, "THIS TEXT IS CLIPPED AT THE END OF SCREEN")
```

Horizontal overflow continues onto the next row while still inside the visible screen.

```peddle
putstr(38, 0, "ABCD")
```

This writes `A` and `B` at columns 38 and 39, then continues with `C` and `D` at the start of the next row.

---

# putstrcolor()

Write a string directly to screen RAM and also write color RAM for each character.

```peddle
putstrcolor(0, 0, "HELLO", 2)
```

`putstrcolor(x, y, text, color)` behaves like `putstr()`, but also writes `color` to color RAM for every visible character written.

Newline characters only move the current position. They do not write a screen byte or a color byte.

```peddle
putstrcolor(0, 0, "RED", 2)
putstrcolor(0, 1, "GREEN", 5)
```

`putstrcolor()` supports the same text argument forms as `putstr()`:

- string literals
- `char[]` variables
- `char[]` struct fields
- temporary `char[]` results such as `itoa(...)`

```peddle
var line char[32]

copy(line, "GREEN")

putstrcolor(0, 0, "RED", 2)
putstrcolor(0, 1, line, 5)
putstrcolor(0, 2, itoa(123), 1)
```

---

# key()

Peddle provides a simple non-blocking keyboard builtin for C64 programs:

```peddle
key() char
```

`key()` reads one character from the C64 KERNAL keyboard buffer.

It returns:

- the next C64 KERNAL/PETSCII character code if a key is waiting
- `0` if no key is currently waiting

`key()` does not block. This makes it useful for games and interactive programs where the main loop should continue running even when no key is pressed.

Internally, `key()` uses the C64 KERNAL `GETIN` routine at `$FFE4`.

Example:

```peddle
fn main() {
    var k char
    var line char[32]

    cls()
    putstr(0, 0, "PRESS KEYS")

    while 1 == 1 {
        k = key()

        if k != 0 {
            clear(line)
            copy(line, "KEY ")
            append(line, itoa(k))

            putstr(0, 2, "        ")
            putstr(0, 2, line)
        }
    }
}
```

Common values include:

```text
SPACE   32
RETURN  13
```

Other keys should be treated as C64/PETSCII-style key codes and can be checked in VICE or on real hardware.

### PETSCII and screen codes

`key()` returns C64 KERNAL/PETSCII-style character codes. These are not the same as raw screen codes.

Use this rule:

```text
key()       returns PETSCII/KERNAL character codes
print()     uses KERNAL/PETSCII output
putraw() writes raw screen codes directly to screen RAM
putchar()   converts a character to a screen code
putcharcolor() converts a character to a screen code and writes color RAM
putstr()    converts string characters to screen codes
```

For example, the space character is simple because it is `32` in both common text handling and screen memory. Letters are different: PETSCII/KERNAL character codes and raw screen codes are not the same thing.

Therefore, do not pass a value returned by `key()` directly to `putraw()` unless you intentionally want to use it as a raw screen code.

---

# waitkey()

`waitkey()` is the blocking form of `key()`.

```peddle
waitkey() char
```

It waits until a key is available and then returns the C64 KERNAL/PETSCII character code.

`waitkey()` does not echo the key to the screen.

Internally, `waitkey()` repeatedly calls the C64 KERNAL `GETIN` routine at `$FFE4` until it receives a non-zero value.

Example:

```peddle
fn main() {
    var k char
    var line char[32]

    cls()
    putstr(0, 0, "PRESS ANY KEY")

    k = waitkey()

    clear(line)
    copy(line, "KEY ")
    append(line, itoa(k))
    putstr(0, 2, line)

    gotoxy(0, 22)
}
```

Use `waitkey()` when the program should stop and wait for one key. Use `key()` when the program should continue running even if no key is pressed.

---

# readline()

Read a line of keyboard input into a `char[]` buffer.

```peddle
readline(buffer char[], echo bool, max int) int
```

`readline()` blocks until RETURN is pressed.

Arguments:

| Argument | Meaning |
|---|---|
| `buffer` | target `char[]` that receives the typed text |
| `echo` | `true` echoes accepted characters, `false` stores silently |
| `max` | maximum number of characters to accept |

Return value:

- the number of characters stored in `buffer`

Important behavior:

- `readline()` clears the buffer first
- RETURN ends input and is not stored
- the effective limit is `min(size(buffer), max)`
- if the limit is reached, additional typed characters are ignored until RETURN
- if `echo` is `true`, accepted characters are echoed using KERNAL character output
- if `echo` is `false`, input is stored without printing typed characters

Example with echo:

```peddle
fn main() {
    var name char[32]
    var line char[32]
    var n int

    cls()

    putstr(0, 0, "NAME? ")
    gotoxy(6, 0)

    n = readline(name, true, 16)

    clear(line)
    copy(line, "HELLO ")
    append(line, name)
    putstr(0, 2, line)

    clear(line)
    copy(line, "LEN ")
    append(line, itoa(n))
    putstr(0, 3, line)

    gotoxy(0, 22)
}
```

Example without echo:

```peddle
fn main() {
    var secret char[16]
    var n int

    cls()

    putstr(0, 0, "SECRET? ")
    gotoxy(8, 0)

    n = readline(secret, false, 8)

    putstr(0, 2, "SECRET STORED")
    putstr(0, 3, itoa(n))

    gotoxy(0, 22)
}
```

Because arrays are mutable arguments in Peddle, `readline()` writes into the caller's `char[]` storage and updates the runtime length of that buffer.

---

# joy()

Read the current joystick state.

```peddle
joy(port) byte
```

`joy(1)` reads joystick port 1.

`joy(2)` reads joystick port 2.

Other port values return `255`, which means no direction or fire button is pressed.

C64 joystick bits are active-low:

```text
bit 0 = up
bit 1 = down
bit 2 = left
bit 3 = right
bit 4 = fire

pressed     means bit == 0
not pressed means bit == 1
```

A common pattern is to mask the lower five joystick bits:

```peddle
var j byte

j = joy(2) & 31
```

Common values are:

```text
31 = idle
30 = up
29 = down
27 = left
23 = right
15 = fire
```

For diagonal movement, test individual bits instead of comparing only the full value:

```peddle
if (j & 4) == 0 {
    # left pressed
}

if (j & 8) == 0 {
    # right pressed
}
```

---

# ticks()

Read the current 16-bit system tick counter.

```peddle
ticks() int
```

`ticks()` returns a 16-bit counter based on the C64 KERNAL jiffy clock. It advances roughly once per video frame while the normal KERNAL interrupt is running.

Typical rates:

```text
PAL C64:  about 50 ticks per second
NTSC C64: about 60 ticks per second
```

The returned value wraps from `65535` back to `0`. For timing checks, prefer `elapsed(last)` or `tickdue(last, interval)` instead of manually subtracting two tick values.

Example:

```peddle
var start int

start = ticks()
```

---

# elapsed()

Return how many ticks have passed since a previous tick value.

```peddle
elapsed(last int) int
```

`elapsed(last)` computes the difference between the current tick counter and `last` using wrap-safe 16-bit arithmetic.

Example:

```peddle
var start int
var passed int

start = ticks()
passed = elapsed(start)
```

The result is intended for relatively short intervals. For common game timing, such as a few frames or a few seconds, it is safe and convenient.

---

# tickdue()

Check if a tick interval has passed.

```peddle
tickdue(last int, interval int) bool
```

`tickdue(last, interval)` returns `true` if at least `interval` ticks have passed since `last`. It handles tick counter wraparound internally.

This is the recommended helper for frame-based throttling in games.

Example:

```peddle
var lastMove int
var moveEvery int

moveEvery = 4
lastMove = ticks()

while true {
    if tickdue(lastMove, moveEvery) {
        lastMove = ticks()
        # update movement here
    }
}
```

For long-running loops, update the stored tick value whenever the timed slot is reached, even if no movement happened. This avoids keeping an old timestamp for many minutes.

---

# ASCII/PETSCII Helpers

Terminal-style programs often need normal ASCII text when talking to Mac or Linux services.

```peddle
asciifont()
toascii(buffer)
topetscii(buffer)
```

`asciifont()` installs a RAM character set intended for terminal output. It keeps normal C64 output behavior but patches useful ASCII display cases such as `_` and lowercase `a`..`z`.

`toascii(buffer)` converts a `char[]` in place. It maps C64 keyboard-style `A`..`Z` to ASCII `a`..`z`, which is useful after `readline()` before sending commands to a Unix-like shell.

`topetscii(buffer)` converts a `char[]` in place in the other direction. It maps ASCII `a`..`z` to C64-style `A`..`Z` and maps LF `10` to C64 carriage return `13`.

Example:

```peddle
var cmd char[80]

asciifont()
n = readline(cmd, true, 38)
toascii(cmd)
```

---

# Network Builtins

Peddle provides a first network API for the C64 Ultimate modem simulator.

The current network API uses one global connection:

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

There are no socket handles in this version. A program opens one connection with `netconnect()`, uses the read/write helpers, then closes it with `netclose()`.

The C64 Ultimate modem simulator is expected to provide a SwiftLink-style 6551 ACIA at `$de00`.

Timeouts are measured in C64 ticks, not milliseconds:

```text
PAL C64:  about 50 ticks per second
NTSC C64: about 60 ticks per second
```

`timeoutTicks` controls how long a read waits for the first byte.

- `0` means non-blocking
- if no byte is available and timeout is `0`, the read returns immediately
- after at least one byte has been read, the read returns as soon as no more bytes are immediately available
- if `netbuffer(backlog)` is configured, extra immediately available bytes can be preserved in that backlog

This keeps network reads friendly for games and other programs with a main loop.

---

# netconnect()

Open the global network connection.

```peddle
var addr char[128]
var ok bool

copy(addr, "192.168.0.10")
ok = netconnect(addr, 6764)
```

`netconnect(addr, port)` returns `true` when the modem reports a successful connection.

The address is a `char[]` and the port is an `int`.

---

# netbuffer()

Reserve a `byte[]` as exclusive network receive backlog storage.

```peddle
var backlog byte[2048]

netbuffer(backlog)
```

After this call, the array belongs to the network runtime. Do not use it for application data.

---

# netavailable()

Return the number of bytes currently queued in the Peddle runtime network backlog.

```peddle
var n int

n = netavailable()
```

`netavailable()` reports only bytes already stored in the backlog configured with `netbuffer(backlog)`. It does not include bytes still waiting inside the C64 Ultimate modem/TCP layer.

If `netbuffer(backlog)` has not been called, `netavailable()` returns `0`.

`netclose()` clears the runtime backlog. Read any wanted queued bytes before closing the connection.

---

# netread()

Read currently available network bytes into a `byte[]` or `char[]` buffer.

```peddle
var rx char[128]
var n int

n = netread(rx, size(rx), 0)
```

`netread()` clears the destination array length before reading, writes up to `min(size(buffer), max)` bytes, updates the runtime length, and returns the number of bytes read.

When a backlog is configured, `netread()` copies old backlog bytes first, then reads immediately available modem bytes into the destination. If the destination fills, it continues draining immediately available modem bytes into the backlog until the backlog is full or no more byte is available.

With timeout `0`, `netread()` is non-blocking.

---

# netreadlf()

Read network bytes into an existing buffer until a line terminator is found.

```peddle
var line char[128]
var found bool

found = netreadlf(line, size(line), 0)
```

`netreadlf()` appends into the current buffer. It does not clear the destination first.

If `netbuffer(backlog)` is configured, `netreadlf()` also reads from and preserves extra bytes in the same backlog.

It returns:

- `true` if it reads CR `13` or LF `10`
- `false` if no complete line is available yet

The CR or LF marker is not stored in the buffer. If a CR is followed by LF, the LF is swallowed on the next call so CRLF does not produce an empty line.

This pattern lets a game loop assemble a line over several frames:

```peddle
var line char[128]
var found bool

found = netreadlf(line, size(line), 0)

if found {
    putstr(0, 0, line)
    clear(line)
}
```

---

# netwrite()

Write bytes from a `byte[]` or `char[]` buffer.

```peddle
var tx char[64]
var n int

copy(tx, "HELLO WORLD\r")
n = netwrite(tx, len(tx))
```

`netwrite(buffer, len)` writes up to `min(size(buffer), len)` bytes and returns the number of bytes written.

---

# netclose()

Close the current network connection.

```peddle
netclose()
```

---

# netconnected()

Check whether the global network connection is currently marked open.

```peddle
if netconnected() {
    print("ONLINE")
}
```

---

# Sound

Peddle sound is a runtime API, not language syntax. The program provides a
`byte[]` memory pool explicitly, and loaded sound bytes are copied into that
pool.

```peddle
const SOUND_STREAM = 1
const SOUND_ALL = 7
const SOUND_REPLACE = 1

fn main() {
    var pool byte[4096]
    var data byte[128]
    var id uint
    var err int

    data[0] = 7
    data[1] = 15
    data[2] = 8
    data[3] = 0
    data[4] = 103
    data[5] = 17
    data[6] = 4
    data[7] = 0
    data[8] = 17
    data[9] = 1
    data[10] = 30
    data[11] = 4
    data[12] = 0
    data[13] = 16
    data[14] = 0

    sound_init(pool)
    id, err = sound_load(data, SOUND_STREAM)

    if err == 0 {
        err = sound_play(id, SOUND_ALL, 0, SOUND_REPLACE)
    }
}
```

`sound_init(pool)` can be called again with the same or another `byte[]`. It
stops playback, clears loaded sounds, resets the pool length to zero, and
installs the IRQ player.

`sound_reset()` keeps the current pool, stops playback, clears loaded sounds,
and resets the pool length to zero.

`sound_load(data, kind)` copies `len(data)` bytes into the pool and returns a
`uint` handle plus an `int` error code. There is no per-sound free in this first
version, so memory is reclaimed with `sound_reset()` or another `sound_init()`.

`sound_num()` returns the number of loaded sounds. `sound_memfree()` returns the
remaining bytes in the current pool.

The supported format is `SOUND_STREAM`, a timed event stream:

```text
0                     end
1, frames             wait frames IRQ ticks
2, voice, note        set note frequency for logical voice
3, voice              gate off logical voice
4, voice, waveform    set waveform/control for logical voice
5, voice, ad          set attack/decay for logical voice
6, voice, sr          set sustain/release for logical voice
7, volume             set global volume
8, voice, lo, hi      set raw frequency for logical voice
9, reg, value         raw SID register write
```

Logical voices are `0`, `1`, and `2`. Time advances only when the stream reaches
a `wait` command, so all commands before the next wait happen in the same
logical sound frame. `sound_play(id, voices, priority, flags)` starts a loaded
stream with explicit voice ownership; use `SOUND_REPLACE` for exclusive playback
or `SOUND_OVERLAY` to layer streams on different voices. Use
`sound_stop_voices(voices)` to stop only the selected SID voices.

For a detailed guide with constants, note numbering, helper functions, raw SID
writes, and a three-voice chord example, see [Sound](09-sound.md).

---

# peek()

Read memory from the C64.

```peddle
var b byte

b = peek(53280)
```

Constants and hex literals work well for hardware addresses.

```peddle
const BORDER = $d020

fn main() {
    var color byte

    color = peek(BORDER)
}
```

---

# poke()

Write memory to the C64.

```peddle
poke(53280, 6)
```

Constants and hex literals are recommended for C64 hardware registers.

```peddle
const BORDER = $d020
const BG = 0xd021
const BLUE = $06

fn main() {
    poke(BORDER, BLUE)
    poke(BG, 0)
}
```

When the address is a numeric literal or constant, the compiler can emit direct absolute addressing.

---


---

[Back to documentation index](README.md)
