# Arrays, Strings, and Structs

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

# Arrays As Function Parameters

Array parameters are passed by reference. The callee receives the caller's array
header and data buffer, not a copy.

```peddle
fn push(nums byte[4], value byte) {
    append(nums, value)
}

fn main() {
    var nums byte[4]

    push(nums, 10)
    push(nums, 20)

    # len(nums) is 2
    # nums[0] is 10
    # nums[1] is 20
}
```

The parameter type includes the array capacity, so `byte[4]` and `byte[8]` are
different parameter types.

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

# Memory Windows

`mem[N] at ADDRESS` declares a fixed byte-addressable memory window. It is for
C64 memory such as screen RAM, color RAM, hardware registers, or explicit
buffers.

```peddle
var screen mem[1000] at $0400
var color  mem[1000] at $D800

screen[0] = 65
color[0] = 1
```

`mem` is not an array. It has no hidden capacity/length header and does not
claim storage in the program image. Element `0` is exactly the declared address.

```peddle
var vic mem[47] at $D000

vic[$20] = 0
```

`len(mem)` and `size(mem)` both return the declared fixed size.

```peddle
var n int

n = len(screen)
n = size(screen)
```

`mem` parameters are passed by reference as a 16-bit base address.

```peddle
fn clear(buf mem[1000], value byte) {
    var i int

    for i = 0 to size(buf) - 1 {
        buf[i] = value
    }
}

fn main() {
    var screen mem[1000] at $0400

    clear(screen, 32)
}
```

Use `&mem` to get the base address as `uint`, and `&mem[i]` to get an element
address. `&mem[i]` can also be passed to `*byte` parameters.

```peddle
fn set(x *byte) {
    x = 99
}

fn main() {
    var screen mem[1000] at $0400
    var addr uint

    addr = &screen
    addr = &screen[10]
    set(&screen[0])
}
```

Array-only operations such as `append()`, `copy()`, `fill()`, file reads, and
network reads do not accept `mem` in this first pass. Use arrays when you need
runtime length metadata.

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


---

[Back to documentation index](README.md)
