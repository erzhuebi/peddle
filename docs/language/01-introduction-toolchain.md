# Introduction and Toolchain

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


---

[Back to documentation index](README.md)
