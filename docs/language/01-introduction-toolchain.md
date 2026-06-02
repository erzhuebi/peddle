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
# This is a comment
fn main() {
    print("HELLO")
}
```

---

# Imports

Use `import` to split a program across multiple `.ped` files. Imports are
merged by the compiler before parsing, so imported files share the same global
namespace as the main file.

```peddle
import "game/player"
import "/generics/file"
import "./math/helpers"
import "game/../generics/file"
```

Import paths are always relative to the project root, which is the directory
containing the main `.ped` file passed to `peddlec`. A leading `/` means that
project root, not the computer's filesystem root.

The compiler appends `.ped` when the import path does not already end with
`.ped`. It also imports each physical file only once, even if nested imports
refer to it more than once.

`.` and `..` are allowed in import paths, but the final resolved file must stay
inside the project root. Symlinks are resolved before this check.

```peddle
# main.ped
import "/game/player"

fn main() {
    player_init()
}

# game/player.ped
fn player_init() {
    print("READY")
}
```


---

[Back to documentation index](README.md)
