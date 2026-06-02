# Optimization, Memory, and Workflow

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
- array parameters store only a reference to the caller's array
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
- sound runtime API
- memory reporting
- optimization modes
- source-level imports

---

# Current Limitations

Not implemented yet:

- floating point
- heap allocation
- generic pointer values
- disk IO
- sprite libraries
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


---

[Back to documentation index](README.md)
