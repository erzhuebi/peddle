#!/bin/sh
set -e

SCRIPT_NAME=$(basename "$0")
C64U_REMOTE_DIR="${C64U_REMOTE_DIR:-/USB1}"

while [ $# -gt 0 ]; do
    case "$1" in
        --host)
            if [ $# -lt 2 ]; then
                echo "deploy error: --host requires a value"
                exit 1
            fi
            C64U_HOST="$2"
            shift 2
            ;;
        --host=*)
            C64U_HOST="${1#--host=}"
            shift
            ;;
        --help|-h)
            echo "usage:"
            echo "  ./$SCRIPT_NAME --host HOST path/to/program.prg"
            echo ""
            echo "environment:"
            echo "  C64U_HOST        C64 Ultimate hostname or IP"
            echo "  C64U_REMOTE_DIR  remote FTP directory, default: /USB1"
            echo ""
            echo "examples:"
            echo "  ./$SCRIPT_NAME --host 192.168.1.64 examples/move_joy.prg"
            echo "  C64U_HOST=192.168.1.64 ./$SCRIPT_NAME examples/move_joy.prg"
            exit 0
            ;;
        --*)
            echo "deploy error: unknown option: $1"
            exit 1
            ;;
        *)
            if [ -n "$LOCAL_PRG" ]; then
                echo "deploy error: only one .prg file can be specified"
                exit 1
            fi
            LOCAL_PRG="$1"
            shift
            ;;
    esac
done

if [ -z "$LOCAL_PRG" ]; then
    echo "usage:"
    echo "  ./$SCRIPT_NAME --host HOST path/to/program.prg"
    echo ""
    echo "examples:"
    echo "  ./$SCRIPT_NAME --host 192.168.1.64 examples/move_joy.prg"
    echo "  C64U_HOST=192.168.1.64 ./$SCRIPT_NAME examples/move_joy.prg"
    exit 1
fi

if [ -z "$C64U_HOST" ]; then
    echo "deploy error: C64U_HOST is not set"
    echo "use one of:"
    echo "  ./$SCRIPT_NAME --host 192.168.1.64 $LOCAL_PRG"
    echo "  C64U_HOST=192.168.1.64 ./$SCRIPT_NAME $LOCAL_PRG"
    exit 1
fi

if [ ! -f "$LOCAL_PRG" ]; then
    echo "deploy error: file not found: $LOCAL_PRG"
    exit 1
fi

case "$LOCAL_PRG" in
    *.prg|*.PRG)
        ;;
    *)
        echo "deploy error: expected a .prg file: $LOCAL_PRG"
        exit 1
        ;;
esac

if ! command -v lftp >/dev/null 2>&1; then
    echo "deploy error: lftp not found"
    echo "install with:"
    echo "  brew install lftp"
    exit 1
fi

REMOTE_NAME=$(basename "$LOCAL_PRG")

lftp "$C64U_HOST" <<EOF
set cmd:fail-exit yes
cd "$C64U_REMOTE_DIR"
put "$LOCAL_PRG" -o "$REMOTE_NAME"
bye
EOF

echo "deployed $LOCAL_PRG to $C64U_HOST:$C64U_REMOTE_DIR/$REMOTE_NAME"