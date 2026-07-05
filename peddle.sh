#!/bin/sh
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

RUN=0
DEPLOY=0
DEPLOY_HOST=""
OPT="${OPT:-size}"
VICE_NET_HOST="${VICE_NET_HOST:-127.0.0.1}"
VICE_NET_PORT="${VICE_NET_PORT:-25232}"
VICE_NET_BAUD="${VICE_NET_BAUD:-50}"
VICE_NET_SHIM_PID=""

usage() {
    echo "usage:"
    echo "  peddle.sh file.ped"
    echo "  peddle.sh --run file.ped"
    echo "  peddle.sh --deploy --host HOST file.ped"
    echo ""
    echo "environment:"
    echo "  OPT              optimization mode, default: size"
    echo "  VICE             VICE executable, default: x64sc or x64"
    echo "  VICE_NET_HOST    modem shim listen host, default: 127.0.0.1"
    echo "  VICE_NET_PORT    modem shim listen port, default: 25232"
    echo "  VICE_NET_BAUD    VICE RS232 baud setting, default: 50"
    echo "  C64U_HOST        C64 Ultimate hostname or IP, used for --deploy"
    echo "  C64U_REMOTE_DIR  remote FTP directory, default: /USB1"
    echo ""
    echo "examples:"
    echo "  ./peddle.sh examples/smoke/hello.ped"
    echo "  ./peddle.sh --run examples/smoke/net_probe.ped"
    echo "  ./peddle.sh examples/demos/pong.ped"
    echo "  ./peddle.sh --deploy --host 192.168.1.64 examples/smoke/net_probe.ped"
    echo "  C64U_HOST=192.168.1.64 ./peddle.sh --deploy examples/smoke/net_probe.ped"
}

cleanup_background() {
    if [ -n "$VICE_NET_SHIM_PID" ]; then
        echo "stopping VICE modem shim"
        kill "$VICE_NET_SHIM_PID" 2>/dev/null || true
        wait "$VICE_NET_SHIM_PID" 2>/dev/null || true
        VICE_NET_SHIM_PID=""
    fi
}

trap cleanup_background EXIT
trap 'cleanup_background; exit 130' INT
trap 'cleanup_background; exit 143' TERM

require_python3() {
    if ! command -v python3 >/dev/null 2>&1; then
        echo "error: python3 is required for --run"
        exit 1
    fi
}

find_vice() {
    if [ -n "$VICE" ]; then
        if command -v "$VICE" >/dev/null 2>&1 || [ -x "$VICE" ]; then
            echo "$VICE"
            return
        fi

        echo "error: VICE executable not found: $VICE" >&2
        exit 1
    fi

    if command -v x64sc >/dev/null 2>&1; then
        echo "x64sc"
        return
    fi

    if command -v x64 >/dev/null 2>&1; then
        echo "x64"
        return
    fi

    echo "error: VICE not found (x64sc/x64)" >&2
    exit 1
}

start_vice_modem_shim() {
    require_python3

    SHIM_SCRIPT="$SCRIPT_DIR/scripts/vice_modem_shim.py"
    if [ ! -f "$SHIM_SCRIPT" ]; then
        echo "error: missing VICE modem shim: $SHIM_SCRIPT"
        exit 1
    fi

    echo "starting VICE modem shim on $VICE_NET_HOST:$VICE_NET_PORT"
    PYTHONUNBUFFERED=1 python3 "$SHIM_SCRIPT" --host "$VICE_NET_HOST" --port "$VICE_NET_PORT" --serial-baud "$VICE_NET_BAUD" &
    VICE_NET_SHIM_PID=$!

    sleep 0.3
    if ! kill -0 "$VICE_NET_SHIM_PID" 2>/dev/null; then
        wait "$VICE_NET_SHIM_PID" 2>/dev/null || true
        VICE_NET_SHIM_PID=""
        echo "error: VICE modem shim failed to start"
        exit 1
    fi
}

while [ $# -gt 0 ]; do
    case "$1" in
        --run)
            if [ "$DEPLOY" = "1" ]; then
                echo "error: --run and --deploy cannot be used together"
                exit 1
            fi
            RUN=1
            shift
            ;;
        --deploy)
            if [ "$RUN" = "1" ]; then
                echo "error: --run and --deploy cannot be used together"
                exit 1
            fi
            DEPLOY=1
            shift
            ;;
        --host)
            if [ $# -lt 2 ]; then
                echo "error: --host requires a value"
                exit 1
            fi
            DEPLOY_HOST="$2"
            shift 2
            ;;
        --host=*)
            DEPLOY_HOST="${1#--host=}"
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        --*)
            echo "error: unknown option: $1"
            exit 1
            ;;
        *)
            if [ -n "$INPUT" ]; then
                echo "error: only one input file can be specified"
                exit 1
            fi
            INPUT="$1"
            shift
            ;;
    esac
done

if [ -z "$INPUT" ]; then
    usage
    exit 1
fi

if [ "$RUN" = "0" ] && [ "$DEPLOY" = "0" ] && [ -n "$DEPLOY_HOST" ]; then
    echo "error: --host can only be used together with --deploy"
    exit 1
fi

BASE="${INPUT%.ped}"

"$SCRIPT_DIR/build/peddlec" --opt="$OPT" -o "$BASE.asm" "$INPUT"

64tass "$BASE.asm" -o "$BASE.prg"

if [ -f "$SCRIPT_DIR/scripts/prg_memory_map.sh" ]; then
    sh "$SCRIPT_DIR/scripts/prg_memory_map.sh" "$BASE.prg" || true
fi

echo ""
echo "wrote $BASE.prg"

if [ "$RUN" = "1" ]; then
    VICE_BIN=$(find_vice)

    start_vice_modem_shim

    "$VICE_BIN" \
        -acia1 \
        -acia1base 0xDE00 \
        -acia1mode 1 \
        -acia1irq 0 \
        -myaciadev 0 \
        -rsdev1 "$VICE_NET_HOST:$VICE_NET_PORT" \
        -rsdev1baud "$VICE_NET_BAUD" \
        -autostart "$BASE.prg"
fi

if [ "$DEPLOY" = "1" ]; then
    DEPLOY_SCRIPT="$SCRIPT_DIR/scripts/deploy_to_c64u.sh"

    if [ ! -f "$DEPLOY_SCRIPT" ]; then
        echo "deploy error: missing $DEPLOY_SCRIPT"
        exit 1
    fi

    if [ -n "$DEPLOY_HOST" ]; then
        sh "$DEPLOY_SCRIPT" --host "$DEPLOY_HOST" "$BASE.prg"
    else
        sh "$DEPLOY_SCRIPT" "$BASE.prg"
    fi
fi
