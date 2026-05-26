#!/usr/bin/env python3
import argparse
import fcntl
import os
import pty
import select
import signal
import socket
import struct
import subprocess
import sys
import termios
import time


DEFAULT_HOST = "0.0.0.0"
DEFAULT_PORT = 6764
DEFAULT_COLS = 40
DEFAULT_ROWS = 24
DEFAULT_RECV_SIZE = 256
DEFAULT_TERM = "vt100"
DEFAULT_C64_CHUNK = 16
DEFAULT_C64_PAUSE = 0.015
DEFAULT_C64_SLOW_COMMAND_PAUSE = 0.08
DEFAULT_TRACE_WIDTH = 16

C64_ESC = 0x1B

running = True
server_socket = None
client_conn = None


def c64_cmd(command, *args):
    return bytes([C64_ESC, ord(command), *args, C64_ESC])


def byte_token(b):
    if b == C64_ESC:
        return "<ESC>"
    if b == 13:
        return "<CR>"
    if b == 10:
        return "<LF>"
    if b == 9:
        return "<TAB>"
    if b == 8:
        return "<BS>"
    if b == 127:
        return "<DEL>"
    if 32 <= b <= 126:
        return chr(b)
    return f"<{b:02X}>"


def trace_bytes(args, label, data):
    if args.trace_off or not data:
        return

    width = args.trace_width

    for offset in range(0, len(data), width):
        chunk = data[offset : offset + width]
        hex_text = " ".join(f"{b:02X}" for b in chunk)
        ascii_text = "".join(chr(b) if 32 <= b <= 126 else "." for b in chunk)
        token_text = " ".join(byte_token(b) for b in chunk)
        print(f"{label} {offset:04X}  {hex_text:<{width * 3}} |{ascii_text:<{width}}|  {token_text}", flush=True)


def send_c64(conn, data, args, label="S->C64"):
    i = 0

    while i < len(data):
        if data[i] == C64_ESC:
            end = data.find(bytes([C64_ESC]), i + 1)

            if end < 0:
                packet = data[i:]
                trace_bytes(args, label, packet)
                conn.sendall(packet)
                return

            packet = data[i : end + 1]
            trace_bytes(args, label, packet)
            conn.sendall(packet)

            if len(packet) > 1 and chr(packet[1]) in "CJLS":
                time.sleep(DEFAULT_C64_SLOW_COMMAND_PAUSE)
            else:
                time.sleep(DEFAULT_C64_PAUSE)

            i = end + 1
        else:
            end = data.find(bytes([C64_ESC]), i)
            if end < 0:
                end = len(data)

            while i < end:
                chunk_end = i + DEFAULT_C64_CHUNK
                if chunk_end > end:
                    chunk_end = end

                chunk = data[i:chunk_end]
                trace_bytes(args, label, chunk)
                conn.sendall(chunk)
                time.sleep(DEFAULT_C64_PAUSE)
                i = chunk_end


class C64TerminalTranslator:
    def __init__(self, cols, rows):
        self.cols = cols
        self.rows = rows
        self.x = 0
        self.y = 0
        self.saved_x = 0
        self.saved_y = 0
        self.pending = b""

    def feed(self, data):
        data = self.pending + data
        self.pending = b""

        out = bytearray()
        i = 0

        while i < len(data):
            b = data[i]

            if b == C64_ESC:
                next_i = self._parse_escape(data, i, out)
                if next_i is None:
                    self.pending = data[i:]
                    break
                i = next_i
                continue

            self._emit_byte(out, b)
            i += 1

        return bytes(out)

    def _parse_escape(self, data, i, out):
        if i + 1 >= len(data):
            return None

        introducer = data[i + 1]

        if introducer == ord("["):
            end = i + 2
            while end < len(data):
                if 0x40 <= data[end] <= 0x7E:
                    self._handle_csi(data[i + 2 : end + 1], out)
                    return end + 1
                end += 1
            return None

        if introducer == ord("]"):
            end = i + 2
            while end < len(data):
                if data[end] == 0x07:
                    return end + 1
                if data[end] == C64_ESC and end + 1 < len(data) and data[end + 1] == ord("\\"):
                    return end + 2
                end += 1
            return None

        if introducer in (ord("("), ord(")"), ord("*"), ord("+")):
            if i + 2 >= len(data):
                return None
            return i + 3

        self._handle_single_escape(introducer, out)
        return i + 2

    def _handle_single_escape(self, command, out):
        if command == ord("c"):
            self._clear(out)
        elif command == ord("7"):
            self.saved_x = self.x
            self.saved_y = self.y
        elif command == ord("8"):
            self._move(out, self.saved_x, self.saved_y)
        elif command == ord("D"):
            self._line_feed(out)
        elif command == ord("E"):
            self._carriage_return(out)
            self._line_feed(out)
        elif command == ord("M"):
            self.y -= 1
            if self.y < 0:
                self.y = 0
            self._move(out, self.x, self.y)
        elif command == ord("H"):
            pass

    def _handle_csi(self, seq, out):
        if not seq:
            return

        final = chr(seq[-1])
        body = seq[:-1].decode("ascii", errors="ignore")
        body = "".join(ch for ch in body if ch not in " ")

        private = ""
        if body and body[0] in "?<=>":
            private = body[0]
            body = body[1:]

        params = self._parse_params(body)

        if final in ("H", "f"):
            row = self._param(params, 0, 1) - 1
            col = self._param(params, 1, 1) - 1
            self._move(out, col, row)
        elif final == "A":
            self._move(out, self.x, self.y - self._param(params, 0, 1))
        elif final == "B":
            self._move(out, self.x, self.y + self._param(params, 0, 1))
        elif final == "C":
            self._move(out, self.x + self._param(params, 0, 1), self.y)
        elif final == "D":
            self._move(out, self.x - self._param(params, 0, 1), self.y)
        elif final == "G":
            self._move(out, self._param(params, 0, 1) - 1, self.y)
        elif final == "d":
            self._move(out, self.x, self._param(params, 0, 1) - 1)
        elif final == "J":
            mode = self._param(params, 0, 0)
            if mode in (2, 3):
                self._clear(out)
            else:
                out.extend(c64_cmd("J"))
        elif final == "K":
            mode = self._param(params, 0, 0)
            if mode == 2:
                out.extend(c64_cmd("L"))
            else:
                out.extend(c64_cmd("K"))
        elif final == "X":
            count = self._param(params, 0, 1)
            old_x = self.x
            old_y = self.y
            for _ in range(count):
                self._printable(out, ord(" "))
            self._move(out, old_x, old_y)
        elif final == "s":
            self.saved_x = self.x
            self.saved_y = self.y
        elif final == "u":
            self._move(out, self.saved_x, self.saved_y)
        elif final in ("h", "l") and private == "?":
            for value in params:
                if value == 25:
                    out.extend(c64_cmd("V" if final == "h" else "v"))
                elif value in (47, 1047, 1048, 1049):
                    self._clear(out)

    def _parse_params(self, body):
        if body == "":
            return []

        params = []
        for part in body.replace(":", ";").split(";"):
            digits = "".join(ch for ch in part if ch.isdigit())
            if digits == "":
                params.append(0)
            else:
                params.append(int(digits))

        return params

    def _param(self, params, index, default):
        if index >= len(params) or params[index] == 0:
            return default
        return params[index]

    def _emit_byte(self, out, b):
        if b == 13:
            self._carriage_return(out)
        elif b == 10:
            self._line_feed(out)
        elif b in (8, 127):
            self._backspace(out)
        elif b == 9:
            self._tab(out)
        elif 32 <= b <= 126:
            self._printable(out, b)
        elif b == 7:
            pass
        elif b >= 128:
            self._printable(out, ord("?"))

    def _printable(self, out, b):
        out.append(b)
        self.x += 1

        if self.x >= self.cols:
            self.x = 0
            self.y += 1

            if self.y >= self.rows:
                self.y = self.rows - 1

    def _carriage_return(self, out):
        out.append(13)
        self.x = 0

    def _line_feed(self, out):
        out.append(10)
        self.y += 1

        if self.y >= self.rows:
            self.y = self.rows - 1

    def _backspace(self, out):
        out.append(8)

        if self.x > 0:
            self.x -= 1
        elif self.y > 0:
            self.y -= 1
            self.x = self.cols - 1

    def _tab(self, out):
        while True:
            self._printable(out, ord(" "))
            if self.x % 4 == 0:
                break

    def _clear(self, out):
        out.extend(c64_cmd("C"))
        self.x = 0
        self.y = 0

    def _move(self, out, x, y):
        if x < 0:
            x = 0
        if y < 0:
            y = 0
        if x >= self.cols:
            x = self.cols - 1
        if y >= self.rows:
            y = self.rows - 1

        self.x = x
        self.y = y
        out.extend(c64_cmd("M", x, y))


def get_local_ips():
    ips = []

    try:
        hostname = socket.gethostname()
        for info in socket.getaddrinfo(hostname, None, socket.AF_INET):
            ip = info[4][0]
            if ip not in ips and not ip.startswith("127."):
                ips.append(ip)
    except OSError:
        pass

    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()

        if ip not in ips:
            ips.append(ip)
    except OSError:
        pass

    return ips


def set_winsize(fd, rows, cols):
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))


def shell_path(requested):
    if requested:
        return requested

    for candidate in (os.environ.get("SHELL"), "/bin/bash", "/bin/sh"):
        if candidate and os.path.exists(candidate):
            return candidate

    return "/bin/sh"


def shell_command(shell):
    base = os.path.basename(shell)

    if base == "zsh":
        return [shell, "-f"]
    if base == "bash":
        return [shell, "--noprofile", "--norc"]

    return [shell]


def translate_input(data):
    out = bytearray()

    for b in data:
        if b == 0:
            continue
        if b == 20:
            out.append(127)
        elif b == 13:
            out.append(13)
        else:
            out.append(b)

    return bytes(out)


def spawn_shell(args):
    master_fd, slave_fd = pty.openpty()
    set_winsize(slave_fd, args.rows, args.cols)

    env = os.environ.copy()
    env["TERM"] = args.term
    env["COLUMNS"] = str(args.cols)
    env["LINES"] = str(args.rows)
    env["CLICOLOR"] = "0"
    env["NO_COLOR"] = "1"
    env["LS_COLORS"] = ""
    env["BASH_SILENCE_DEPRECATION_WARNING"] = "1"
    env.setdefault("PS1", "$ ")
    env.setdefault("PROMPT", "%~ %# ")

    proc = subprocess.Popen(
        shell_command(args.shell),
        cwd=args.cwd,
        env=env,
        stdin=slave_fd,
        stdout=slave_fd,
        stderr=slave_fd,
        close_fds=True,
        preexec_fn=os.setsid,
    )

    os.close(slave_fd)

    return master_fd, proc


def terminate_shell(master_fd, proc):
    try:
        os.close(master_fd)
    except OSError:
        pass

    if proc.poll() is None:
        try:
            os.killpg(proc.pid, signal.SIGHUP)
        except OSError:
            proc.terminate()

        try:
            proc.wait(timeout=1.0)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()


def shutdown_all():
    global running, server_socket, client_conn

    running = False

    if client_conn is not None:
        try:
            client_conn.close()
        except OSError:
            pass
        client_conn = None

    if server_socket is not None:
        try:
            server_socket.close()
        except OSError:
            pass
        server_socket = None


def signal_handler(signum, frame):
    print("\nShutting down gracefully...")
    shutdown_all()
    sys.exit(0)


def send_start_screen(conn, args):
    message = (
        f"PEDDLE TERM LIVE {args.cols}X{args.rows}\r\n"
        "TYPE EXIT TO QUIT\r\n"
    ).encode("ascii")

    send_c64(conn, c64_cmd("C") + message, args)


def handle_client(conn, addr, args):
    global client_conn

    client_conn = conn
    translator = C64TerminalTranslator(args.cols, args.rows)
    master_fd, proc = spawn_shell(args)

    print(f"Client connected from {addr[0]}:{addr[1]}")

    try:
        conn.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
    except OSError:
        pass

    try:
        send_start_screen(conn, args)

        while running:
            if proc.poll() is not None:
                try:
                    send_c64(conn, c64_cmd("Q"), args)
                    time.sleep(0.05)
                except OSError:
                    pass
                break

            try:
                readable, _, _ = select.select([conn, master_fd], [], [], 0.05)
            except OSError:
                break

            if conn in readable:
                data = conn.recv(args.recv_size)

                if not data:
                    break

                trace_bytes(args, "C64->S", data)
                pty_input = translate_input(data)
                trace_bytes(args, "S->PTY", pty_input)
                os.write(master_fd, pty_input)

            if master_fd in readable:
                try:
                    data = os.read(master_fd, 4096)
                except OSError:
                    data = b""

                if not data:
                    break

                trace_bytes(args, "PTY->S", data)
                out = translator.feed(data)
                if out:
                    send_c64(conn, out, args)

    except OSError as e:
        print(f"client error: {e}")

    finally:
        print("Client disconnected")
        terminate_shell(master_fd, proc)

        try:
            conn.close()
        except OSError:
            pass

        client_conn = None


def parse_args():
    parser = argparse.ArgumentParser(
        description="Interactive PTY terminal service for Peddle/C64 Ultimate."
    )

    parser.add_argument("--host", default=DEFAULT_HOST, help=f"host/interface to bind, default: {DEFAULT_HOST}")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help=f"TCP port, default: {DEFAULT_PORT}")
    parser.add_argument("--cols", type=int, default=DEFAULT_COLS, help=f"terminal columns, default: {DEFAULT_COLS}")
    parser.add_argument("--rows", type=int, default=DEFAULT_ROWS, help=f"terminal rows, default: {DEFAULT_ROWS}")
    parser.add_argument("--recv-size", type=int, default=DEFAULT_RECV_SIZE, help=f"receive size, default: {DEFAULT_RECV_SIZE}")
    parser.add_argument("--term", default=DEFAULT_TERM, help=f"TERM value for the PTY, default: {DEFAULT_TERM}")
    parser.add_argument("--cwd", default=os.getcwd(), help="initial working directory")
    parser.add_argument("--shell", default="", help="shell executable, default: $SHELL or /bin/sh")
    parser.add_argument("--trace-off", action="store_true", help="disable the default hex/ASCII trace output")
    parser.add_argument("--trace-width", type=int, default=DEFAULT_TRACE_WIDTH, help=f"bytes per trace line, default: {DEFAULT_TRACE_WIDTH}")

    args = parser.parse_args()
    args.shell = shell_path(args.shell)
    args.cwd = os.path.abspath(os.path.expanduser(args.cwd))

    if args.trace_width < 1:
        args.trace_width = DEFAULT_TRACE_WIDTH

    return args


def main():
    global server_socket

    args = parse_args()

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind((args.host, args.port))
    server_socket.listen(1)
    server_socket.settimeout(0.5)

    print("C64 terminal server started")
    print(f"Listening on port {args.port}")
    print(f"Terminal size: {args.cols}x{args.rows}")
    print(f"TERM: {args.term}")
    print(f"Shell: {args.shell}")
    print(f"Initial cwd: {args.cwd}")
    print("Protocol: VT100/ANSI PTY -> TEP tiny escape protocol")
    print(f"I/O trace: {'off' if args.trace_off else 'on'}")

    ips = get_local_ips()
    if ips:
        print("Connect from C64 to one of these IPs:")
        for ip in ips:
            print(f"  {ip}:{args.port}")
    else:
        print(f"Listening on {args.host}:{args.port}")

    print("Press Ctrl-C to stop.\n")

    try:
        while running:
            try:
                conn, addr = server_socket.accept()
            except socket.timeout:
                continue
            except OSError:
                break

            handle_client(conn, addr, args)

    finally:
        print("Server stopped")
        shutdown_all()


if __name__ == "__main__":
    main()
