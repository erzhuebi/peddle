# Peddle Language

Peddle is a minimal, C-like language targeting the Commodore 64 (6502).
It is designed for small, efficient programs compiled to native C64 PRG
files.

## Features

### Types

``` peddle
byte    // 8-bit
char    // 8-bit character
bool    // 8-bit (0=false, non-zero=true)
int     // 16-bit
```

### Variables

``` peddle
var x: byte
var y: int
```

### Arrays

``` peddle
var a: byte[4]
var b: int[10]

a[0] = 1
b[2] = 1000
```

### Strings

``` peddle
var s: char[6]
s = "HELLO"
print(s)
```

### Arithmetic

``` peddle
var a: int
var b: int
var c: int

c = a + b
c = a - b
```

### Unary Operators

``` peddle
x = -x
b = !b
```

### Comparisons

``` peddle
if a < b { }
if a >= b { }
```

### Control Flow

#### If / Else

``` peddle
if a < b {
    print("YES")
} else {
    print("NO")
}
```

#### While

``` peddle
var i: byte

i = 0
while i < 10 {
    i = i + 1
}
```

### Functions

``` peddle
fn add(a: int, b: int): int {
    return a + b
}

fn main() {
    var x: int
    x = add(1, 2)
}
```

### Built-in Functions

#### print

``` peddle
print("HELLO")
```

#### poke

``` peddle
poke(53280, 0)
```

#### peek

``` peddle
var x: byte
x = peek(53280)
```

### Type Conversion (implicit)

``` peddle
var b: byte
var i: int

i = b      // byte -> int
b = i      // int -> byte (truncated)
```

## Notes

-   No heap or dynamic memory
-   Static memory layout
-   Optimized for small binaries
-   Designed for C64 hardware access

## Example Program

``` peddle
fn main() {
    var s: char[6]

    s = "HELLO"
    print(s)
}
```
