# Syntax, Types, and Control Flow

# Primitive Types

| Type | Size | Description |
|---|---|---|
| byte | 1 byte | unsigned 8-bit |
| bool | 1 byte | boolean |
| char | 1 byte | character |
| int | 2 bytes | signed 16-bit |

---


# Boolean Literals

Peddle supports the boolean literals `true` and `false`.

```peddle
var done bool
var enabled bool

done = false
enabled = true
```

Boolean values are stored as one byte:

| Literal | Value | Meaning |
|---|---:|---|
| `false` | 0 | false |
| `true` | 1 | true |

Using `0` and `1` for boolean values is still allowed. This is useful for low-level C64 code and keeps existing programs compatible.

```peddle
var done bool

done = 0
done = 1
```

Boolean literals can be used in assignments, comparisons, `if` statements, and `while` loops.

```peddle
fn main() {
    var done bool

    done = false

    while done == false {
        if key() == 'q' {
            done = true
        }
    }
}
```

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

# For Loops

Peddle supports three `for` loop forms.

```peddle
for {
    # infinite loop
}
```

```peddle
for condition {
    # loop while condition is non-zero / true
}
```

```peddle
for i = 0 to 9 {
    # counted loop
}
```

Counted `for` loops are inclusive: `for i = 0 to 9` runs with `i` set to every value from `0` through `9`.

The loop variable must already exist and must be `byte` or `int`. The end expression is evaluated once before the loop body runs.

---

# break and continue

`break` exits the nearest enclosing loop immediately.

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

`continue` skips the rest of the current loop iteration and starts the next iteration.

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

`break` and `continue` are only valid inside loops. Using either outside a loop is a compile-time error.

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


---

[Back to documentation index](README.md)
