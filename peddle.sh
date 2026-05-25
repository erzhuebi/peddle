#!/bin/sh
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

RUN=0
DEPLOY=0
DEPLOY_HOST=""
OPT="${OPT:-size}"

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
            echo "usage:"
            echo "  peddle.sh file.ped"
            echo "  peddle.sh --run file.ped"
            echo "  peddle.sh --deploy --host HOST file.ped"
            echo ""
            echo "environment:"
            echo "  OPT              optimization mode, default: size"
            echo "  C64U_HOST        C64 Ultimate hostname or IP, used for --deploy"
            echo "  C64U_REMOTE_DIR  remote FTP directory, default: /USB1"
            echo ""
            echo "examples:"
            echo "  ./peddle.sh examples/hello.ped"
            echo "  ./peddle.sh --run examples/hello.ped"
            echo "  ./peddle.sh --deploy --host 192.168.1.64 examples/net_probe.ped"
            echo "  C64U_HOST=192.168.1.64 ./peddle.sh --deploy examples/net_probe.ped"
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
    echo "usage:"
    echo "  peddle.sh file.ped"
    echo "  peddle.sh --run file.ped"
    echo "  peddle.sh --deploy --host HOST file.ped"
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
    if command -v x64sc >/dev/null 2>&1; then
        x64sc -autostart "$BASE.prg"
    elif command -v x64 >/dev/null 2>&1; then
        x64 -autostart "$BASE.prg"
    else
        echo "VICE not found (x64sc/x64)"
        exit 1
    fi
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