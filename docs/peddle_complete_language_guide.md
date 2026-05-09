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
    var x: int

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
var x: int
var mask: byte

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
var x: int
```

---

## Multiple Variables

```peddle
var x, y, z: int
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
    var a: int
    var b: int
    var c: int

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
    var flags: byte
    var mask: byte

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
    var x: int

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
    var a: int
    var b: int

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
    var i: byte

    i = 0

    while i < 10 {
        i = i + 1
    }
}
```

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
fn add(a: int, b: int) -> int {
    return a + b
}
```

---

## Calling Functions

```peddle
fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    var result: int

    result = add(10, 20)
}
```

---

# Return Statements

```peddle
fn double(x: int) -> int {
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
var data: byte[16]
```

## Int Array

```peddle
var nums: int[32]
```

## Char Array

```peddle
var title: char[64]
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
var x: int

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
var n: int

n = len(nums)
```

---

# Array Capacity

Use `size()` to get maximum capacity.

```peddle
var n: int

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

# Strings

Peddle strings are implemented using `char[]`.

Strings are:

- counted strings
- not zero terminated

---

# String Example

```peddle
fn main() {
    var title: char[32]

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

# Structs

Structs group multiple fields together.

---

# Struct Declaration

```peddle
struct Player {
    id: byte
    hp: int
}
```

---

# Struct Variables

```peddle
var p: Player
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
var hp: int

hp = p.hp
```

---

# Structs With Array Fields

```peddle
struct Player {
    id: byte
    name: char[16]
    hp: int
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
var players: Player[8]
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
| `print(x)` | print string |
| `peek(addr)` | read memory |
| `poke(addr, value)` | write memory |
| `len(array)` | runtime length |
| `size(array)` | array capacity |
| `append(array, value)` | append element |
| `copy(dst, src)` | copy arrays/strings |
| `fill(array, value)` | fill array |
| `clear(array)` | clear runtime length |

---

# peek()

Read memory from the C64.

```peddle
var b: byte

b = peek(53280)
```

Constants and hex literals work well for hardware addresses.

```peddle
const BORDER = $d020

fn main() {
    var color: byte

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
    id: byte
    name: char[16]
    hp: int
}

fn main() {
    var players: Player[2]

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

    poke(BORDER, COLOR_OK)
    poke(BG, 0)
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
    var border: byte
    var bg: byte
    var a: byte
    var b: byte
    var s: byte

    var x: int
    var y: int
    var n: int

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
- string literal append/copy
- multiplication
- division and modulo
- variable shifts

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
- functions
- return values
- arrays
- runtime array length
- append/copy/fill/clear
- strings
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
