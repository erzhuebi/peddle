# Peddle

**Peddle** is a small programming language for the Commodore 64 / C64 Ultimate.

It is designed as a compact alternative to BASIC: simple enough to keep the compiler understandable, but expressive enough to write real C64 programs with native code, hardware access, arrays, functions, and eventually C64 Ultimate specific features such as networking/modem support.

The compiler is called **`peddlec`**.

The name **Peddle** is a homage to **Chuck Peddle**, one of the creators of the MOS 6502 CPU used in the Commodore 64.

---

## Goals

Peddle aims to be:

- small and predictable
- easy to cross-compile on macOS/Linux
- efficient on real C64 hardware
- simple enough to keep generated programs small
- close to the machine when needed

The current compiler pipeline is:

```text
source -> lexer -> parser -> AST -> sema -> codegen -> 6502 ASM -> 64tass -> PRG
```

---

## Toolchain

Required tools:

- Go
- 64tass
- VICE emulator, usually `x64sc` or `x64`

### macOS

```sh
brew install go 64tass vice
```

### Linux

```sh
sudo apt install golang 64tass vice
```

---

## Building

Build the compiler:

```sh
make build
```

This creates:

```text
build/peddlec
```

Run the Hello World example:

```sh
make hello
```

Run tests:

```sh
make test
```

Show available targets:

```sh
make
```

---

## Project Structure

```text
peddle/
  ast/             AST definitions
  lexer/           tokenizer
  parser/          parser
  sema/            semantic/type checker
  codegen/         6502 assembly generator
  cmd/peddlec/     compiler entry point
  examples/        example Peddle programs
```

---

## Language Overview

A Peddle program consists of functions. The entry point is:

```peddle
fn main() {
    print("HELLO WORLD")
}
```

Variables must currently be declared at the beginning of a block.

---

## Types

Available primitive types:

```peddle
byte    // 8-bit unsigned value
char    // 8-bit character value
bool    // 8-bit boolean: 0 = false, non-zero = true
int     // 16-bit integer
```

Examples:

```peddle
fn main() {
    var b: byte
    var c: char
    var ok: bool
    var x: int

    b = 42
    c = 65
    ok = 1
    x = 1000
}
```

---

## Variables

Variables are declared with `var`:

```peddle
fn main() {
    var x: byte
    var y: int

    x = 10
    y = 1000
}
```

Current restriction:

```peddle
fn main() {
    var x: byte

    x = 1

    // var y: byte  // not allowed here yet
}
```

---

## Numeric Conversion

Peddle currently allows implicit conversion between small numeric types and `int`.

```peddle
fn main() {
    var b: byte
    var i: int

    b = 10
    i = b      // byte -> int
    b = i      // int -> byte, high byte is truncated
}
```

On C64 this behavior is useful for hardware registers and low-byte operations, but it should be used deliberately.

---

## Arithmetic

Supported arithmetic operators:

```peddle
+
-
```

Example:

```peddle
fn main() {
    var a: int
    var b: int
    var c: int

    a = 1000
    b = 2000

    c = a + b
    c = c - a
}
```

Mixed byte/int expressions are supported:

```peddle
fn main() {
    var b: byte
    var i: int
    var x: int

    b = 10
    i = 1000

    x = b + i
    x = i - b
}
```

Not implemented yet:

```peddle
*
/
%
```

---

## Unary Operators

Supported unary operators:

```peddle
-x
!x
```

Example:

```peddle
fn main() {
    var x: int
    var done: bool

    x = 10
    x = -x

    done = 0
    done = !done
}
```

`!x` returns `1` when `x` is zero, otherwise `0`.

---

## Comparisons

Supported comparison operators:

```peddle
==
!=
<
<=
>
>=
```

Example:

```peddle
fn main() {
    var a: int
    var b: int

    a = -1
    b = 1

    if a < b {
        print("OK")
    }
}
```

`int` comparisons are signed 16-bit comparisons.

---

## If / Else

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

A numeric value can be used directly as a condition:

```peddle
fn main() {
    var x: int

    x = 1

    if x {
        print("NON ZERO")
    }
}
```

---

## While Loops

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

## Functions

Functions are declared with `fn`.

```peddle
fn add(a: int, b: int): int {
    return a + b
}

fn main() {
    var x: int

    x = add(1, 2)
}
```

Void-style functions are also supported:

```peddle
fn sayHello() {
    print("HELLO")
    return
}

fn main() {
    sayHello()
}
```

Current calling convention:

- parameters are stored in static function-local memory
- return values are stored in a static return slot
- recursion is not supported

Conceptually:

```asm
sta add_a
sta add_b
jsr add
lda add_return
```

---

## Arrays

Fixed-size arrays are supported:

```peddle
byte[n]
char[n]
int[n]
```

Example:

```peddle
fn main() {
    var a: byte[4]
    var x: byte

    a[0] = 7
    x = a[0]
}
```

`int[]` uses two bytes per element:

```peddle
fn main() {
    var values: int[4]
    var x: int

    values[0] = 1234
    x = values[0]
}
```

Array indexes may be expressions:

```peddle
fn main() {
    var a: byte[8]
    var i: byte
    var x: byte

    i = 2

    a[i + 1] = 7
    x = a[i + 1]
}
```

There is currently no runtime bounds checking.

---

## Strings

String literals are emitted as zero-terminated `char[]` data.

```peddle
fn main() {
    print("HELLO WORLD")
}
```

Strings can be assigned to `char[]` variables if the target array is large enough:

```peddle
fn main() {
    var s: char[6]

    s = "HELLO"
    print(s)
}
```

The terminating zero byte is included, so `"HELLO"` needs `char[6]`.

---

## Built-ins

### `print`

Prints a string literal or `char[]`.

```peddle
fn main() {
    print("HELLO")
}
```

```peddle
fn main() {
    var s: char[6]

    s = "HELLO"
    print(s)
}
```

Currently `print` is intended for strings, not numeric formatting.

---

### `poke`

Writes one byte to a fixed C64 memory address.

```peddle
fn main() {
    poke(53280, 0)   // border color
    poke(53281, 6)   // background color
}
```

The address is currently required to be a numeric literal.

---

### `peek`

Reads one byte from a fixed C64 memory address.

```peddle
fn main() {
    var border: byte

    border = peek(53280)
    poke(53281, border)
}
```

The address is currently required to be a numeric literal.

---

## Structs

Structs are part of the language syntax and semantic checker, but code generation for structs is not implemented yet.

Example syntax:

```peddle
struct Player {
    x: byte
    y: byte
    hp: int
}
```

Planned usage:

```peddle
fn main() {
    var p: Player

    p.x = 10
    p.y = 20
    p.hp = 100
}
```

Planned implementation work:

- static memory layout for structs
- field offsets
- field read/write codegen
- arrays of structs later

---

## Memory Model

Peddle currently uses a static memory model.

Important properties:

- no heap
- no garbage collector
- no dynamic allocation
- globals/function locals are emitted as static data
- arrays have fixed size
- function parameters and return values use fixed memory slots

This keeps code generation simple and predictable for the C64.

---

## Zero Page Usage

Current generated code uses these zero page locations:

```asm
ZP_TMP0     = $fb
ZP_TMP1     = $fc
ZP_PTR0_LO  = $fd
ZP_PTR0_HI  = $fe
```

These are used for temporary values and indirect memory access.

---

## Generated Program Layout

Generated assembly includes:

- BASIC stub
- entry point calling `main`
- generated functions
- optional runtime helpers
- string literals
- static data

The BASIC stub starts the native program automatically with `SYS 2064`.

---

## Example: Hello World

```peddle
fn main() {
    print("HELLO WORLD")
}
```

Build and run:

```sh
make hello
```

---

## Example: Screen Colors

```peddle
fn main() {
    poke(53280, 0)
    poke(53281, 6)

    print("PEDDLE")
}
```

---

## Example: Function + Arrays

```peddle
fn add(a: int, b: int): int {
    return a + b
}

fn main() {
    var values: int[4]
    var x: int

    values[0] = add(1000, 234)
    x = values[0]

    if x > 1000 {
        print("OK")
    }
}
```

---

## Current Limitations

Not implemented yet:

- multiplication, division, modulo
- struct code generation
- arrays of structs
- dynamic memory
- recursion
- numeric formatting for `print`
- variable addresses for `peek` / `poke`
- runtime bounds checks
- explicit casts such as `byte(x)`

---

## Design Principles

Peddle favors:

- small compiler implementation
- small generated binaries
- predictable memory usage
- C64 hardware friendliness
- directness over abstraction
- incremental feature growth

The guiding idea is: enough language to build real C64 software, without turning the compiler into a large modern language implementation.

---

## License

MIT License.
