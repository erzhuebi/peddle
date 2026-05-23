# Examples

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

# Boolean Literals Example

```peddle
fn main() {
    var done bool
    var enabled bool
    var count byte

    cls()
    textcolor(1)

    done = false
    enabled = true
    count = 0

    if enabled == true {
        putstrcolor(0, 0, "ENABLED TRUE OK", 5)
    } else {
        putstrcolor(0, 0, "ENABLED TRUE FAIL", 2)
    }

    if done == false {
        putstrcolor(0, 1, "DONE FALSE OK", 5)
    } else {
        putstrcolor(0, 1, "DONE FALSE FAIL", 2)
    }

    while done == false {
        count = count + 1

        if count == 5 {
            done = true
        }
    }

    if done == true {
        putstrcolor(0, 3, "DONE TRUE OK", 5)
    } else {
        putstrcolor(0, 3, "DONE TRUE FAIL", 2)
    }

    if count == 5 {
        putstrcolor(0, 4, "WHILE BOOL OK", 5)
    } else {
        putstrcolor(0, 4, "WHILE BOOL FAIL", 2)
    }

    gotoxy(0, 22)
}
```

Run it with:

```bash
./peddle.sh --run examples/bool_literals.ped
```

Expected screen output:

```text
ENABLED TRUE OK
DONE FALSE OK

DONE TRUE OK
WHILE BOOL OK
```

---

# Keyboard input

The `keyboard_key.ped` example shows how to read keys without blocking the program.

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

Run it with:

```bash
./peddle.sh --run examples/keyboard_key.ped
```

`key()` returns `0` when no key is waiting. When a key is pressed, it returns the C64 KERNAL/PETSCII character code.

---

## Target file: `examples/keyboard_key.ped`

If you already created this example and it works in VICE, no change is needed.

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

---

[Back to documentation index](README.md)
