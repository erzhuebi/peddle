#!/bin/sh
set -e

if [ $# -ne 1 ]; then
    echo "usage: scripts/prg-memory-map.sh file.prg"
    exit 1
fi

PRG="$1"

if [ ! -f "$PRG" ]; then
    echo "memory map error: file not found: $PRG"
    exit 1
fi

file_size=$(wc -c < "$PRG" | tr -d ' ')

if [ "$file_size" -lt 3 ]; then
    echo "memory map error: PRG file is too small: $PRG"
    exit 1
fi

set -- $(od -An -N2 -tu1 "$PRG")

if [ $# -lt 2 ]; then
    echo "memory map error: unable to read PRG load address: $PRG"
    exit 1
fi

load_lo=$1
load_hi=$2

prg_start=$((load_lo + load_hi * 256))
prg_size=$((file_size - 2))
prg_end=$((prg_start + prg_size - 1))

expected_start=2049      # $0801
allowed_end=53247        # $CFFF
forbidden_start=53248    # $D000
max_address=65535        # $FFFF

fmt_addr() {
    printf '$%04X' "$1"
}

range_size() {
    echo $(($2 - $1 + 1))
}

print_range() {
    start=$1
    end=$2
    label=$3
    bytes=$(range_size "$start" "$end")

    printf '  %s-%s   %-28s %6d bytes\n' "$(fmt_addr "$start")" "$(fmt_addr "$end")" "$label" "$bytes"
}

status="OK"

if [ "$prg_start" -ne "$expected_start" ]; then
    status="ERROR"
fi

if [ "$prg_end" -gt "$max_address" ]; then
    status="ERROR"
fi

if [ "$prg_end" -ge "$forbidden_start" ]; then
    status="ERROR"
fi

allowed_total=$(range_size "$prg_start" "$allowed_end")

if [ "$prg_end" -lt "$allowed_end" ]; then
    free_bytes=$((allowed_end - prg_end))
else
    free_bytes=0
fi

echo ""
echo "Memory map"
echo "----------"
printf 'PRG load start:     %s\n' "$(fmt_addr "$prg_start")"
printf 'PRG end:            %s\n' "$(fmt_addr "$prg_end")"
printf 'PRG size:           %d bytes\n' "$prg_size"
echo ""
echo "Allowed PRG range:"
printf '  %-13s   %-28s %6d bytes\n' "$(fmt_addr "$prg_start")-$(fmt_addr "$allowed_end")" "PRG may grow here" "$allowed_total"
printf '  %-13s   %-28s %6d bytes\n' "" "free until forbidden memory" "$free_bytes"
echo ""
echo "Used ranges:"
print_range "$prg_start" "$prg_end" "PRG image"
echo ""
echo "Reserved / forbidden ranges:"
print_range 0 255 "zero page"
print_range 256 511 "CPU stack"
print_range 512 1023 "KERNAL/BASIC workspace"
print_range 1024 2023 "screen RAM"
print_range 53248 57343 "I/O / color RAM"
print_range 57344 65535 "KERNAL ROM"
echo ""
echo "Banking:"
echo "  BASIC ROM:        disabled at startup"
echo "  KERNAL ROM:       enabled"
echo "  I/O:              enabled"
echo ""
echo "Status:"
echo "  $status"

if [ "$prg_start" -ne "$expected_start" ]; then
    echo ""
    echo "memory map error: unexpected PRG load address $(fmt_addr "$prg_start"), expected $(fmt_addr "$expected_start")"
    exit 1
fi

if [ "$prg_end" -gt "$max_address" ]; then
    echo ""
    echo "memory map error: PRG exceeds 64 KB address space"
    exit 1
fi

if [ "$prg_end" -ge "$forbidden_start" ]; then
    echo ""
    echo "memory map error: PRG reaches forbidden memory at $(fmt_addr "$forbidden_start")"
    echo "PRG end is $(fmt_addr "$prg_end")"
    exit 1
fi