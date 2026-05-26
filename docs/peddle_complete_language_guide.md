# Peddle Language Guide

# Introduction

Peddle is a compiled systems programming language for the Commodore 64.

It is designed to be:

- simple
- predictable
- easy to learn
- statically typed
- close to the hardware
- optimized for C64 development

Peddle source files use the `.ped` extension.

The compiler translates `.ped` files into 6502 assembly which can then be assembled into `.prg` binaries using `64tass`.

---

# Toolchain

## Compile Source Code

```sh
peddlec -o game.asm game.ped
```

## Assemble PRG

```sh
64tass game.asm -o game.prg
```

## Compile With Size Optimization

```sh
peddlec --opt=size -o game.asm game.ped
```

## Compile With Speed Optimization

```sh
peddlec --opt=speed -o game.asm game.ped
```

## Compile And Run

If your local toolchain is configured with VICE, the compiler can also launch the generated program.

```sh
peddlec --run -o game.asm game.ped
```

---

# Your First Program

```peddle
fn main() {
    print("HELLO WORLD")
}
```

---

# Comments

Peddle supports single-line comments using `#`.

```peddle
# this is a comment

fn main() {
    var x int

    x = 1 # trailing comment
}
```

`//` comments are not supported. The `/` character is the division operator.

---

# Primitive Types

| Type | Size | Description |
|---|---|---|
| byte | 1 byte | unsigned 8-bit |
| bool | 1 byte | boolean |
| char | 1 byte | character |
| int | 2 bytes | signed 16-bit |

---

# Numeric Literals

Peddle supports decimal, hexadecimal, and binary integer literals.

| Form | Example | Value |
|---|---:|---:|
| decimal | `123` | 123 |
| grouped decimal | `1_000` | 1000 |
| C64-style hex | `$d020` | 53280 |
| C-style hex | `0xd020` | 53280 |
| C-style hex uppercase | `0Xff` | 255 |
| binary | `%11110000` | 240 |
| grouped binary | `%1111_0000` | 240 |

Underscores are visual separators and are ignored by the compiler.

```peddle
var x int
var mask byte

x = 1_000
x = $d020
x = 0xd021
mask = %1111_0000
```

Binary literals start with `%`. The same `%` character is also used for modulo when it appears between expressions.

```peddle
mask = %1111_0000
x = 10 % 3
```

---

# Character Literals

Peddle supports single-character literals using single quotes.

```peddle
var ch char

ch = 'A'
ch = ' '
ch = '*'
```

Character literals have type `char`.

Escape sequences are supported:

| Literal | Value | Meaning |
|---|---:|---|
| `'\n'` | 13 | C64 carriage return / newline |
| `'\r'` | 13 | C64 carriage return / newline |
| `'\0'` | 0 | zero byte |
| `'\''` | 39 | single quote |
| `'\\'` | 92 | backslash |

On the C64, the normal KERNAL newline is carriage return byte `13`. Therefore both `'\n'` and `'\r'` compile to `13`.

Example:

```peddle
fn main() {
    putchar(0, 0, 'P')
    putchar(1, 0, 'E')
    putchar(2, 0, 'D')
}
```

---

# Constants

Constants are top-level numeric values.

```peddle
const BORDER = $d020
const BG = 0xd021
const WHITE = 1
const MASK = %1111_0000
const BIG = 1_000
```

Constants can be used in expressions and builtin calls.

```peddle
const BORDER = $d020
const BG = $d021
const BLUE = 6

fn main() {
    poke(BORDER, BLUE)
    poke(BG, 0)
}
```

Constants are currently numeric only.

---

# Variables

## Single Variable

```peddle
var x int
```

---

## Multiple Variables

```peddle
var x, y, z int
```

---

## Assignment

```peddle
x = 10
y = x
```

---

# Arithmetic

Supported arithmetic operators:

| Operator | Meaning |
|---|---|
| `+` | addition |
| `-` | subtraction |
| `*` | multiplication |
| `/` | integer division |
| `%` | integer remainder / modulo |

Example:

```peddle
fn main() {
    var a int
    var b int
    var c int

    a = 100
    b = 50

    c = a + b
    c = c - 25
    c = c * 2
    c = c / 5
    c = c % 7
}
```

Division is integer division.

```peddle
x = 10 / 3 # x becomes 3
y = 10 % 3 # y becomes 1
```

Division by zero is safe and deterministic:

| Expression | Result |
|---|---|
| `x / 0` | `0` |
| `x % 0` | `x` |

---

# Bitwise Operators

Peddle supports bitwise operators for numeric scalar values.

| Operator | Meaning |
|---|---|
| `&` | bitwise AND |
| `|` | bitwise OR |
| `^` | bitwise XOR |
| `<<` | logical shift left |
| `>>` | logical shift right |

Example:

```peddle
fn main() {
    var flags byte
    var mask byte

    flags = %1010_0000
    mask = %1111_0000

    flags = flags & mask
    flags = flags | %0000_0011
    flags = flags ^ %0000_0001

    flags = flags << 1
    flags = flags >> 2
}
```

Right shift is logical. New high bits are filled with zero.

---

# Operator Precedence

From highest to lowest:

| Precedence | Operators |
|---|---|
| field / index / call | `.`, `[]`, `()` |
| unary | `-`, `!` |
| product | `*`, `/`, `%` |
| sum | `+`, `-` |
| shift | `<<`, `>>` |
| bitwise AND | `&` |
| bitwise XOR | `^` |
| bitwise OR | `|` |
| comparisons | `==`, `!=`, `<`, `<=`, `>`, `>=` |

Example:

```peddle
x = a | b & c << d + 1
```

This is parsed as:

```peddle
x = a | (b & (c << (d + 1)))
```

Use parentheses when in doubt.

```peddle
x = (a + b) * c
```

---

# Unary Operators

## Negative Numbers

```peddle
x = -1
```

## Boolean NOT

```peddle
done = !done
```

---

# Comparisons

Supported comparisons:

| Operator |
|---|
| `==` |
| `!=` |
| `<` |
| `<=` |
| `>` |
| `>=` |

Example:

```peddle
if score > 100 {
    print("WIN")
}
```

---

# If Statements

```peddle
fn main() {
    var x int

    x = 10

    if x > 5 {
        print("BIG")
    }
}
```

---

# If / Else

```peddle
fn main() {
    var a int
    var b int

    a = 1
    b = 2

    if a < b {
        print("YES")
    } else {
        print("NO")
    }
}
```

---

# While Loops

```peddle
fn main() {
    var i byte

    i = 0

    while i < 10 {
        i = i + 1
    }
}
```

---

# break and continue

`break` exits the nearest enclosing `while` loop immediately.

```peddle
fn main() {
    var i byte

    i = 0

    while i < 10 {
        if i == 5 {
            break
        }

        i = i + 1
    }
}
```

`continue` skips the rest of the current loop iteration and jumps back to the `while` condition.

```peddle
fn main() {
    var i byte
    var sum int

    i = 0
    sum = 0

    while i < 10 {
        i = i + 1

        if i == 3 {
            continue
        }

        sum = sum + i
    }
}
```

`break` and `continue` are only valid inside `while` loops. Using either outside a loop is a compile-time error.

In nested loops, `break` and `continue` affect the nearest enclosing loop.

---

# Functions

## Basic Function

```peddle
fn hello() {
    print("HELLO")
}
```

---

## Functions With Parameters

```peddle
fn add(a int, b int) int {
    return a + b
}
```

---

## Calling Functions

```peddle
fn add(a int, b int) int {
    return a + b
}

fn main() {
    var result int

    result = add(10, 20)
}
```

---

# Return Statements

```peddle
fn double(x int) int {
    return x + x
}
```

---

# Arrays

Arrays in Peddle have:

- fixed capacity
- runtime length
- embedded storage

Internally arrays store:

1. capacity
2. current runtime length
3. data buffer

---

# Array Declaration

## Byte Array

```peddle
var data byte[16]
```

## Int Array

```peddle
var nums int[32]
```

## Char Array

```peddle
var title char[64]
```

---

# Array Access

## Write

```peddle
nums[0] = 100
nums[1] = 200
```

## Read

```peddle
var x int

x = nums[0]
```

---

# Array Index Expressions

```peddle
nums[i + 1] = 10
```

---

# Runtime Array Length

Use `len()` to get current runtime length.

```peddle
var n int

n = len(nums)
```

---

# Array Capacity

Use `size()` to get maximum capacity.

```peddle
var n int

n = size(nums)
```

---

# append()

Append values to arrays.

## Append Byte

```peddle
append(data, 1)
```

## Append Int

```peddle
append(nums, 1000)
```

## Append String Literal

```peddle
append(title, "!")
```

## Append Char Array

`append()` can append one `char[]` to another `char[]`.

```peddle
var title char[32]
var suffix char[8]

copy(title, "HELLO")
copy(suffix, " C64")

append(title, suffix)
```

This also works with temporary `char[]` values returned by builtins such as `itoa()`.

```peddle
var line char[32]
var score int

score = 123

copy(line, "SCORE ")
append(line, itoa(score))
```

Appending automatically increases runtime length.

---

# copy()

## Copy Arrays

```peddle
copy(dst, src)
```

## Copy String Literal

```peddle
copy(title, "HELLO")
```

## Copy Temporary Char Array

`copy()` can copy a temporary `char[]`, for example the result of `itoa()`.

```peddle
var number char[6]

copy(number, itoa(-123))
```

---

# fill()

Fill an entire array with a value.

```peddle
fill(data, 0)
```

---

# clear()

Reset runtime length to zero.

```peddle
clear(data)
```

The underlying memory is not erased.

Only the runtime length becomes zero.

---

# itoa()

Convert a signed 16-bit `int` to a temporary decimal `char[]`.

```peddle
var line char[32]
var score int

score = -123

copy(line, "SCORE ")
append(line, itoa(score))
putstr(0, 0, line)
```

`itoa(value)` returns a temporary `char[6]`.

The longest signed 16-bit decimal value is:

```text
-32768
```

So the returned temporary buffer can hold every possible `int` value.

Examples:

| Expression | Result |
|---|---|
| `itoa(0)` | `"0"` |
| `itoa(12345)` | `"12345"` |
| `itoa(-123)` | `"-123"` |
| `itoa(-32768)` | `"-32768"` |

The returned array is temporary and is overwritten by the next `itoa()` call.

This is safe because the value is consumed immediately:

```peddle
append(line, itoa(score))
putstr(0, 0, itoa(score))
```

If you need to keep the text, copy it into your own `char[]`.

```peddle
var saved char[6]

copy(saved, itoa(score))
```

---

# Strings

Peddle strings are implemented using `char[]`.

Strings are:

- counted strings
- not zero terminated
- stored in normal `char[]` arrays

Use the normal array/string builtins for string work:

- `copy()`
- `append()`
- `clear()`
- `len()`
- `size()`

Use `itoa()` when you need to convert an `int` into text.

---

## Escape Sequences

String literals support common escape sequences.

| Escape | Value | Meaning |
|---|---:|---|
| `\n` | 13 | C64 carriage return / newline |
| `\r` | 13 | C64 carriage return / newline |
| `\0` | 0 | zero byte |
| `\"` | 34 | double quote |
| `\\` | 92 | backslash |

For C64 KERNAL output, newline means carriage return byte `13`.

```peddle
fn main() {
    print("LINE 1\n")
    print("LINE 2\n")
    print("DONE")
}
```

---

# String Example

```peddle
fn main() {
    var title char[32]

    copy(title, "COMMODORE")
    append(title, " 64")

    print(title)
}
```

---

# Printing

## Print String Literal

```peddle
print("READY")
```

## Print String Variable

```peddle
print(title)
```

---

## Newlines

`print()` uses KERNAL-style output. In string literals, `\n` and `\r` are encoded as C64 carriage return byte `13`.

```peddle
fn main() {
    print("HELLO\n")
    print("WORLD")
}
```

This prints `HELLO`, moves to the next line, then prints `WORLD`.

---

# Structs

Structs group multiple fields together.

---

# Struct Declaration

```peddle
struct Player {
    id byte
    hp int
}
```

---

# Struct Variables

```peddle
var p Player
```

---

# Struct Field Assignment

```peddle
p.id = 1
p.hp = 100
```

---

# Struct Field Read

```peddle
var hp int

hp = p.hp
```

---

# Structs With Array Fields

```peddle
struct Player {
    id byte
    name char[16]
    hp int
}
```

---

# Struct String Operations

```peddle
copy(p.name, "ADA")
append(p.name, "!")
print(p.name)
```

---

# Arrays Of Structs

```peddle
var players Player[8]
```

---

# Struct Array Access

```peddle
players[0].hp = 100
players[1].hp = 120
```

---

# Struct Array String Fields

```peddle
copy(players[0].name, "BOB")
append(players[0].name, "!")
```

---

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
| `netconnect(addr, port)` | connect using the C64 Ultimate modem simulator |
| `netread(buffer, max, timeoutTicks)` | read available network bytes |
| `netreadlf(buffer, max, timeoutTicks)` | read network bytes until CR or LF |
| `netwrite(buffer, len)` | write network bytes |
| `netclose()` | close the current network connection |
| `netconnected()` | check whether the global network connection is open |

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

Currently the conversion handles:

- `A`..`Z` to C64 screen codes `1`..`26`
- `a`..`z` to C64 screen codes `1`..`26`
- all other values unchanged

So space and digits work naturally:

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

Currently the conversion handles:

- `A`..`Z` to C64 screen codes `1`..`26`
- `a`..`z` to C64 screen codes `1`..`26`
- all other values unchanged

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

# itoa() Example

```peddle
fn main() {
    var line char[32]
    var score int
    var neg int

    cls()

    score = 12345
    neg = -123

    copy(line, "SCORE ")
    append(line, itoa(score))
    putstr(0, 0, line)

    clear(line)
    copy(line, "NEG ")
    append(line, itoa(neg))
    putstr(0, 2, line)

    clear(line)
    copy(line, "MIN ")
    append(line, itoa(-32768))
    putstr(0, 4, line)

    clear(line)
    copy(line, "ZERO ")
    append(line, itoa(0))
    putstr(0, 6, line)

    # Leave room for BASIC READY. after program exit.
    gotoxy(0, 22)
}
```

---

# Direct Screen Example

```peddle
fn main() {
    cls()

    border(6)
    background(0)
    textcolor(1)

    putstr(0, 0, "DIRECT SCREEN")
    putstrcolor(0, 1, "COLOR DIRECT", 2)

    gotoxy(0, 5)
    print("KERNAL PRINT\n")

    putchar(0, 8, 'P')
    putchar(1, 8, 'E')
    putchar(2, 8, 'D')
    putchar(3, 8, 'D')
    putchar(4, 8, 'L')
    putchar(5, 8, 'E')

    putcolor(0, 8, 2)
    putcolor(1, 8, 3)
    putcolor(2, 8, 4)
    putcolor(3, 8, 5)
    putcolor(4, 8, 6)
    putcolor(5, 8, 7)

    gotoxy(0, 22)
}
```

---

# Network Builtins

Peddle provides a first network API for the C64 Ultimate modem simulator.

The current network API uses one global connection:

```peddle
netconnect(addr char[], port int) bool
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

# netread()

Read currently available network bytes into a `byte[]` or `char[]` buffer.

```peddle
var rx char[128]
var n int

n = netread(rx, size(rx), 0)
```

`netread()` clears the destination array length before reading, writes up to `min(size(buffer), max)` bytes, updates the runtime length, and returns the number of bytes read.

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

# Border Color Example

```peddle
const BORDER = $d020
const BG = $d021

fn main() {
    poke(BORDER, 2)
    poke(BG, 6)
}
```

---

# Full Example Program

```peddle
const BORDER = $d020
const BG = $d021
const NAME_LEN = 16
const COLOR_OK = %0000_0101

struct Player {
    id byte
    name char[16]
    hp int
}

fn main() {
    var players Player[2]

    players[0].id = 1
    players[0].hp = 100
    copy(players[0].name, "BOB")

    players[1].id = 2
    players[1].hp = 120

    copy(players[1].name, "ADA")
    append(players[1].name, "!")

    print(players[0].name)
    print(players[1].name)

    if players[1].hp > players[0].hp {
        print("ADA LEADS")
    }

    print(itoa(players[1].hp))

    poke(BORDER, COLOR_OK)
    poke(BG, 0)
}
```

---

# Break / Continue Example

```peddle
fn main() {
    var i byte
    var sum int
    var line char[32]

    cls()
    border(6)
    background(0)
    textcolor(1)

    i = 0
    sum = 0

    while i < 10 {
        i = i + 1

        # Skip 3 and 6.
        if i == 3 {
            continue
        }

        if i == 6 {
            continue
        }

        # Stop completely at 8.
        if i == 8 {
            break
        }

        sum = sum + i
    }

    # sum = 1 + 2 + 4 + 5 + 7 = 19

    copy(line, "I ")
    append(line, itoa(i))
    putstr(0, 0, line)

    clear(line)
    copy(line, "SUM ")
    append(line, itoa(sum))
    putstr(0, 2, line)

    if i == 8 {
        putstrcolor(0, 4, "BREAK OK", 5)
    } else {
        putstrcolor(0, 4, "BREAK FAIL", 2)
    }

    if sum == 19 {
        putstrcolor(0, 5, "CONTINUE OK", 5)
    } else {
        putstrcolor(0, 5, "CONTINUE FAIL", 2)
    }

    gotoxy(0, 22)
}
```

---

# Complete Operator Example

```peddle
const BORDER = $d020
const BG = $d021
const MASK = %1111_0000
const BIG = 1_000

fn main() {
    var border byte
    var bg byte
    var a byte
    var b byte
    var s byte

    var x int
    var y int
    var n int

    # byte arithmetic
    a = 20 + 5
    a = a - 3
    a = a * 2
    b = a / 7
    b = a % 7

    # byte bitwise
    border = 10 & 15
    border = border | 32
    border = border ^ 15

    # byte shifts
    s = 2
    border = border << s
    border = border >> 1

    # int arithmetic
    x = BIG + 250
    x = x - 50
    x = x * 3
    y = x / 10
    y = x % 256

    # int bitwise and shifts
    y = y & 1023
    y = y | 4096
    y = y ^ 255

    n = 2
    y = y << n
    y = y >> 3

    bg = y

    poke(BORDER, border)
    poke(BG, bg)

    print("ALL OPS OK")
}
```

---

# Optimization Modes

## Speed Mode

```sh
peddlec --opt=speed -o game.asm game.ped
```

This favors:

- inline code
- fewer subroutine calls
- faster execution

Result:

- larger PRGs
- faster runtime

---

## Size Mode

```sh
peddlec --opt=size -o game.asm game.ped
```

This favors:

- reusable runtime helper functions
- smaller generated code

Result:

- smaller PRGs
- slightly slower runtime

Currently size mode shares runtime helpers for:

- fill()
- copy()
- append()
- multiplication
- division and modulo
- variable shifts

Some larger helpers are shared in both speed and size mode because repeated inline code would quickly waste PRG space:

- string literal append/copy
- `putstr()`
- `putstrcolor()`
- `cls()`
- `itoa()`

Runtime helpers are emitted only when needed. For example, a program that does not call `cls()` does not include the `peddle_cls` helper.

Constant shifts are emitted inline in both optimization modes.

Bitwise operators are small and currently emitted inline in both optimization modes.

---

# Memory Reporting

The compiler supports static memory reporting.

```sh
peddlec --mem-report -o game.asm game.ped
```

Example output:

```text
memory report:
  total static memory: 207 bytes
    literals:          176 bytes
    variables/runtime: 31 bytes
    static symbols:    6
```

You can also enforce a static memory limit.

```sh
peddlec --mem-limit=1024 -o game.asm game.ped
```

If the program exceeds the limit, compilation fails.

---

# Runtime Memory Model

Currently arrays and zero-initialized variables are statically allocated inside the PRG.

This means:

- array storage becomes part of the binary
- large arrays increase PRG size
- no heap allocator exists yet

Future plans include:

- BSS-style zero-init memory
- disk streaming
- runtime memory checks
- heap allocation

---

# Generated Assembly

The generated assembly is intentionally readable.

This makes Peddle useful for:

- learning 6502 assembly
- debugging
- educational projects
- retro game development

---

# Current Language Features

Implemented:

- `#` comments
- constants
- decimal literals
- hexadecimal literals using `$ff` and `0xff`
- binary literals using `%1111_0000`
- underscore numeric separators
- variables
- multiple variable declarations
- arithmetic operators: `+`, `-`, `*`, `/`, `%`
- bitwise operators: `&`, `|`, `^`
- shift operators: `<<`, `>>`
- comparisons
- unary operators
- if/else
- while loops
- break and continue
- functions
- return values
- arrays
- runtime array length
- append/copy/fill/clear
- append/copy with `char[]` sources
- strings
- putchar/putscreen/putcolor
- safe clipping for direct screen writes
- putstr/putstrcolor with string literals and dynamic `char[]`
- signed `itoa()` conversion
- gotoxy
- cls/border/background/textcolor
- string escape sequences including C64 newline/carriage return
- character literals
- structs
- arrays of structs
- struct string fields
- peek/poke
- memory reporting
- optimization modes

---

# Current Limitations

Not implemented yet:

- floating point
- heap allocation
- pointers
- modules/imports
- disk IO
- sprites/sound libraries
- recursion safety
- dynamic arrays
- multiple simultaneous independent temporary string results from `itoa()`
- constant expressions
- typed constants
- local constants

---

# Recommended Workflow

Compile:

```sh
peddlec --opt=size -o game.asm game.ped
```

Assemble:

```sh
64tass game.asm -o game.prg
```

Run in emulator:

```sh
x64sc game.prg
```

Or compile and run directly if configured:

```sh
peddlec --run --opt=size -o game.asm game.ped
```

---

# Final Notes

Peddle is intentionally small and focused.

The goal is to make Commodore 64 development approachable while still generating understandable and efficient 6502 assembly.
