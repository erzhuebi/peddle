#!/bin/sh
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

RUN=0
OPT="${OPT:-size}"

if [ "$1" = "--run" ]; then
    RUN=1
    shift
fi

if [ $# -ne 1 ]; then
    echo "usage:"
    echo "  peddle.sh file.ped"
    echo "  peddle.sh --run file.ped"
    exit 1
fi

INPUT="$1"
BASE="${INPUT%.ped}"

"$SCRIPT_DIR/build/peddlec" --opt="$OPT" -o "$BASE.asm" "$INPUT"

64tass "$BASE.asm" -o "$BASE.prg"

if [ -f "$SCRIPT_DIR/scripts/prg_memory_map.sh" ]; then
    sh "$SCRIPT_DIR/scripts/prg_memory_map.sh" "$BASE.prg" || true
fi

echo ""
echo "wrote $BASE.prg"

if [ "$RUN" = "1" ]; then
    if command -v x64sc >/dev/null 2>&1; then
        x64sc -autostart "$BASE.prg"
    elif command -v x64 >/dev/null 2>&1; then
        x64 -autostart "$BASE.prg"
    else
        echo "VICE not found (x64sc/x64)"
        exit 1
    fi
fi