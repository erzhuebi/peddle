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
| `putscreen(x, y, code)` | write raw C64 screen code to screen RAM |
| `putcolor(x, y, color)` | write color RAM value at position |
| `putstr(x, y, text)` | write a string literal or `char[]` directly to screen RAM |
| `putstrcolor(x, y, text, color)` | write a string literal or `char[]` to screen RAM and color RAM |
| `key()` | non-blocking keyboard read; returns `0` if no key is waiting |
| `waitkey()` | blocking keyboard read; waits until a key is available |
| `readline(buffer, echo, max)` | read a line into a `char[]` buffer |

---

# C64 Screen Builtins

Peddle provides direct C64 screen and color helpers for text-mode screen programming.

Important distinction:

- `print()` uses the KERNAL cursor
- `gotoxy()` moves the KERNAL cursor
- `textcolor()` affects KERNAL `print()`
- `putchar()`, `putscreen()`, `putstr()`, and `putcolor()` write directly to screen/color RAM

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

`textcolor()` affects KERNAL-style `print()` output. It does not automatically affect direct screen RAM writes such as `putchar()` or `putscreen()`.

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
- `putscreen()`
- `putstr()`
- `putcolor()`

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

# putscreen()

Write a raw C64 screen code at a fixed screen position.

```peddle
putscreen(0, 0, 16)
putscreen(1, 0, 5)
putscreen(2, 0, 4)
```

`putscreen()` does no character conversion.

Use it when you already know the C64 screen code you want to write.

`putscreen()` is clipped to the visible screen. If `x >= 40` or `y >= 25`, it does nothing.

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
putscreen() writes raw screen codes directly to screen RAM
putchar()   converts a character to a screen code
putstr()    converts string characters to screen codes
```

For example, the space character is simple because it is `32` in both common text handling and screen memory. Letters are different: PETSCII/KERNAL character codes and raw screen codes are not the same thing.

Therefore, do not pass a value returned by `key()` directly to `putscreen()` unless you intentionally want to use it as a raw screen code.

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
