# Peddle Language Guide (Beginner-Friendly)

Welcome to **Peddle**, a small systems-oriented language for the Commodore 64.

This guide walks you through all currently supported language features with explanations and examples.

---

# 1. Your First Program

Every program starts with a `main` function:

```peddle
fn main() {
    print("HELLO WORLD")
}
```

---

# 2. Types

Peddle has a small set of built-in types.

## Primitive Types

- `byte` – 8-bit value
- `char` – 8-bit character
- `bool` – 0 = false, non-zero = true
- `int` – 16-bit integer

## Example

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

# 3. Variables

Variables must be declared at the beginning of a block.

```peddle
fn main() {
    var x: int
    var y: byte

    x = 10
    y = 5
}
```

---

# 4. Numeric Conversion

Implicit conversion between numeric types is allowed.

```peddle
fn main() {
    var b: byte
    var i: int

    b = 10
    i = b
    b = i
}
```

---

# 5. Arithmetic

Supported operators:

- `+`
- `-`

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

---

# 6. Unary Operators

```peddle
-x
!x
```

```peddle
fn main() {
    var x: int
    var done: bool

    x = -10
    done = !done
}
```

---

# 7. Comparisons

Supported:

```
== != < <= > >=
```

```peddle
fn main() {
    var a: int
    var b: int

    if a < b {
        print("OK")
    }
}
```

---

# 8. Control Flow

## If / Else

```peddle
fn main() {
    var x: int

    if x {
        print("NON ZERO")
    } else {
        print("ZERO")
    }
}
```

## While Loop

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

# 9. Functions

## Definition

```peddle
fn add(a: int, b: int) -> int {
    return a + b
}
```

## Call

```peddle
fn main() {
    var x: int
    x = add(1, 2)
}
```

---

# 10. Arrays

Arrays are fixed size.

```peddle
fn main() {
    var a: byte[4]
    a[0] = 7
}
```

---

# 11. Strings

Strings are special `char[]` arrays with:

- capacity (2 bytes)
- length (2 bytes)
- data

```peddle
fn main() {
    var s: char[6]

    s = "HELLO"
    print(s)
}
```

---

# 12. Structs

Structs group fields.

```peddle
struct Player {
    x: byte
    hp: int
}
```

## Usage

```peddle
fn main() {
    var p: Player

    p.x = 10
    p.hp = 100
}
```

---

# 13. Arrays of Structs

```peddle
fn main() {
    var players: Player[10]
    players[0].hp = 100
}
```

---

# 14. Field Access

```peddle
p.x = 10
x = p.x
```

---

# 15. Array Indexing

```peddle
a[i] = 5
x = a[i]
```

---

# 16. Struct Array Field Access

```peddle
players[i].hp = 100
hp = players[i].hp
```

---

# 17. Built-in Functions

## print

```peddle
print("HELLO")
```

## poke

```peddle
poke(53280, 0)
```

## peek

```peddle
var b: byte
b = peek(53280)
```

---

# 18. Memory Model

- static memory only
- no heap
- no recursion
- fixed-size arrays

---

# 19. Current Limitations

Not supported yet:

- multiplication / division
- dynamic memory
- recursion
- bounds checking

---

# 20. Summary

Peddle is:

- small
- predictable
- close to hardware

Perfect for writing real C64 programs with minimal abstraction.

