#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

OPT="${OPT:-size}"
VICE_CYCLES="${VICE_CYCLES:-80000000}"
VICE_PAUSE="${VICE_PAUSE:-1}"
NO_COMPILE=0

usage() {
    cat <<EOF
usage:
  examples/smoke/capture_screenshots.sh [options] [program ...]

options:
  --no-compile      use existing .prg files
  --list            list smoke entry programs and exit
  --help, -h        show this help

program:
  A .ped path, a path relative to examples/smoke, or a name without .ped.
  With no programs, all smoke entry programs are captured except */lib/*.

environment:
  OPT=size|speed        compile optimization mode, default: size
  VICE=/path/to/x64sc   VICE executable override
  VICE_DIRECTORY=DIR    VICE data directory override
  VICE_CYCLES=N         cycle limit before screenshot, default: 80000000
  VICE_PAUSE=N          seconds between captures, default: 1

output:
  Screenshots are written beside each program as:
    <program>-vice.ppm.png
EOF
}

list_entries() {
    find "$SCRIPT_DIR" -type f -name '*.ped' ! -path '*/lib/*' | sort
}

for arg in "$@"; do
    case "$arg" in
        --help|-h)
            usage
            exit 0
            ;;
        --list)
            list_entries
            exit 0
            ;;
    esac
done

if [ -n "${VICE:-}" ]; then
    VICE_BIN="$VICE"
elif command -v x64sc >/dev/null 2>&1; then
    VICE_BIN=$(command -v x64sc)
elif command -v x64 >/dev/null 2>&1; then
    VICE_BIN=$(command -v x64)
else
    echo "error: VICE not found (x64sc/x64)" >&2
    exit 1
fi

if [ -n "${VICE_DIRECTORY:-}" ]; then
    VICE_DIR="$VICE_DIRECTORY"
elif [ -d /opt/homebrew/Cellar/vice/3.10/share/vice ]; then
    VICE_DIR=/opt/homebrew/Cellar/vice/3.10/share/vice
elif [ -d /usr/local/share/vice ]; then
    VICE_DIR=/usr/local/share/vice
else
    VICE_DIR=
fi

resolve_program() {
    input=$1

    case "$input" in
        /*)
            src=$input
            ;;
        *.ped)
            if [ -f "$input" ]; then
                src=$(CDPATH= cd -- "$(dirname -- "$input")" && pwd)/$(basename -- "$input")
            elif [ -f "$SCRIPT_DIR/$input" ]; then
                src=$SCRIPT_DIR/$input
            else
                src=
            fi
            ;;
        *)
            if [ -f "$SCRIPT_DIR/$input.ped" ]; then
                src=$SCRIPT_DIR/$input.ped
            elif [ -f "$SCRIPT_DIR/$input" ]; then
                src=$SCRIPT_DIR/$input
            else
                src=
            fi
            ;;
    esac

    if [ -z "$src" ] || [ ! -f "$src" ]; then
        echo "error: smoke program not found: $input" >&2
        return 1
    fi

    printf '%s\n' "$src"
}

capture_one() {
    src=$1
    base=${src%.ped}
    prg=$base.prg
    shot=$base-vice.ppm

    echo "==> $(basename -- "$src")"

    if [ "$NO_COMPILE" -eq 0 ]; then
        (cd "$REPO_ROOT" && OPT="$OPT" ./peddle.sh "$src")
    fi

    if [ ! -f "$prg" ]; then
        echo "error: missing PRG after compile: $prg" >&2
        return 1
    fi

    rm -f "$shot" "$shot.png"

    if [ -n "$VICE_DIR" ]; then
        "$VICE_BIN" \
            -directory "$VICE_DIR" \
            -silent \
            +sound \
            -warp \
            -quicksaveformat ppm \
            -exitscreenshot "$shot" \
            -limitcycles "$VICE_CYCLES" \
            -autostart "$prg"
    else
        "$VICE_BIN" \
            -silent \
            +sound \
            -warp \
            -quicksaveformat ppm \
            -exitscreenshot "$shot" \
            -limitcycles "$VICE_CYCLES" \
            -autostart "$prg"
    fi
    rc=$?

    if [ -f "$shot.png" ]; then
        echo "wrote $shot.png"
        return 0
    fi

    if [ -f "$shot" ]; then
        echo "wrote $shot"
        return 0
    fi

    echo "error: VICE did not write screenshot for $src (exit $rc)" >&2
    return 1
}

entry_file=${TMPDIR:-/tmp}/peddle-smoke-screens.$$
trap 'rm -f "$entry_file"' EXIT HUP INT TERM

: > "$entry_file"

while [ "$#" -gt 0 ]; do
    case "$1" in
        --no-compile)
            NO_COMPILE=1
            shift
            ;;
        --list)
            list_entries
            exit 0
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        --*)
            echo "error: unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
        *)
            resolve_program "$1" >> "$entry_file" || exit 1
            shift
            ;;
    esac
done

failures=0

if [ ! -s "$entry_file" ]; then
    list_entries | while IFS= read -r src; do
        printf '%s\n' "$src" >> "$entry_file"
    done
fi

while IFS= read -r src; do
    if ! capture_one "$src"; then
        failures=$((failures + 1))
    fi
    sleep "$VICE_PAUSE"
done < "$entry_file"

if [ "$failures" -ne 0 ]; then
    echo "failed captures: $failures" >&2
    exit 1
fi
