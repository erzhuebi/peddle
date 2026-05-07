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

---

# Your First Program

```peddle
fn main() {
    print("HELLO WORLD")
}
```

---

# Comments

Peddle supports single-line comments.

```peddle
// this is a comment

fn main() {
    var x: int

    x = 1 // trailing comment
}
```

---

# Primitive Types

| Type | Size | Description |
|---|---|---|
| byte | 1 byte | unsigned 8-bit |
| bool | 1 byte | boolean |
| char | 1 byte | character |
| int | 2 bytes | signed 16-bit |

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

Supported operators:

| Operator | Meaning |
|---|---|
| + | addition |
| - | subtraction |

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
}
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
| == |
| != |
| < |
| <= |
| > |
| >= |

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
| print(x) | print string |
| peek(addr) | read memory |
| poke(addr, value) | write memory |
| len(array) | runtime length |
| size(array) | array capacity |
| append(array, value) | append element |
| copy(dst, src) | copy arrays/strings |
| fill(array, value) | fill array |
| clear(array) | clear runtime length |

---

# peek()

Read memory from the C64.

```peddle
var b: byte

b = peek(53280)
```

---

# poke()

Write memory to the C64.

```peddle
poke(53280, 6)
```

---

# Border Color Example

```peddle
fn main() {
    poke(53280, 2)
    poke(53281, 6)
}
```

---

# Full Example Program

```peddle
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

    poke(53280, 5)
    poke(53281, 0)
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

---

# Runtime Memory Model

Currently arrays are statically allocated inside the PRG.

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

- comments
- variables
- multiple variable declarations
- arithmetic
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

---

# Final Notes

Peddle is intentionally small and focused.

The goal is to make Commodore 64 development approachable while still generating understandable and efficient 6502 assembly.
