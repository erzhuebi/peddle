#!/usr/bin/env python3
import argparse
import select
import signal
import socket
import sys
import time


DEFAULT_HOST = "127.0.0.1"
DEFAULT_PORT = 25232
DEFAULT_SERIAL_BAUD = 50
BUFFER_SIZE = 4096


running = True
listen_socket = None
vice_socket = None
target_socket = None


def close_socket(sock):
    if sock is None:
        return

    try:
        sock.close()
    except OSError:
        pass


def tune_tcp_socket(sock):
    try:
        sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
    except OSError:
        pass


def shutdown_all():
    global running, listen_socket, vice_socket, target_socket

    running = False
    close_socket(target_socket)
    close_socket(vice_socket)
    close_socket(listen_socket)
    target_socket = None
    vice_socket = None
    listen_socket = None


def signal_handler(signum, frame):
    print("\nShutting down gracefully...")
    shutdown_all()
    sys.exit(0)


def printable(data):
    parts = []

    for b in data:
        if 32 <= b <= 126:
            parts.append(chr(b))
        elif b == 13:
            parts.append("\\r")
        elif b == 10:
            parts.append("\\n")
        else:
            parts.append(f"\\x{b:02x}")

    return "".join(parts)


def trace(prefix, data, enabled):
    if enabled and data:
        print(f"{prefix} {len(data):4d}: {printable(data)}")


def send_modem_line(conn, text, trace_enabled):
    data = text.encode("ascii") + b"\r"
    conn.sendall(data)
    trace("SHIM->VICE", data, trace_enabled)


def parse_dial_target(line):
    upper = line.upper()

    if not upper.startswith("ATDT"):
        return None

    target = line[4:].strip()
    if ":" not in target:
        return None

    host, port_text = target.rsplit(":", 1)
    if not host or not port_text:
        return None

    return host, int(port_text)


def connect_target(host, port, timeout):
    sock = socket.create_connection((host, port), timeout=timeout)
    tune_tcp_socket(sock)
    sock.setblocking(False)
    return sock


def send_serial(conn, data, baud):
    if baud <= 0:
        conn.sendall(data)
        return

    delay = 10.0 / baud

    for b in data:
        conn.sendall(bytes((b,)))
        time.sleep(delay)


def bridge(vice_conn, target_conn, serial_baud, trace_enabled):
    vice_conn.setblocking(True)
    sockets = [vice_conn, target_conn]
    in_escape = False
    pending_plus = 0
    pending_plus_time = 0.0
    command = bytearray()

    while running:
        try:
            readable, _, errored = select.select(sockets, [], sockets, 0.1)
        except OSError:
            break

        if errored:
            break

        if not readable and not in_escape and pending_plus:
            if time.monotonic() - pending_plus_time >= 0.2:
                try:
                    target_conn.sendall(b"+" * pending_plus)
                except OSError:
                    return
                pending_plus = 0

        for sock in readable:
            try:
                data = sock.recv(BUFFER_SIZE)
            except OSError:
                return

            if not data:
                return

            if sock is vice_conn:
                trace("VICE->TCP ", data, trace_enabled)
                outgoing = bytearray()

                for b in data:
                    if in_escape:
                        if b in (13, 10):
                            if not command:
                                continue

                            line = command.decode("ascii", errors="ignore").strip()
                            command.clear()
                            print(f"modem command: {line}")

                            if line.upper() == "ATH":
                                send_modem_line(vice_conn, "OK", trace_enabled)
                                return

                            send_modem_line(vice_conn, "ERROR", trace_enabled)
                            in_escape = False
                            continue

                        command.append(b)
                        continue

                    if b == ord("+"):
                        pending_plus = pending_plus + 1
                        pending_plus_time = time.monotonic()

                        if pending_plus == 3:
                            if outgoing:
                                try:
                                    target_conn.sendall(outgoing)
                                except OSError:
                                    return
                                outgoing.clear()

                            print("modem escape: +++")
                            in_escape = True
                            pending_plus = 0
                            command.clear()

                        continue

                    if pending_plus:
                        outgoing.extend(b"+" * pending_plus)
                        pending_plus = 0

                    outgoing.append(b)

                if outgoing:
                    try:
                        target_conn.sendall(outgoing)
                    except OSError:
                        return
            else:
                trace("TCP->VICE ", data, trace_enabled)
                try:
                    send_serial(vice_conn, data, serial_baud)
                except OSError:
                    return


def handle_vice_connection(conn, addr, args):
    global target_socket

    print(f"VICE connected from {addr[0]}:{addr[1]}")
    conn.settimeout(0.1)

    command = bytearray()

    try:
        while running:
            try:
                data = conn.recv(BUFFER_SIZE)
            except socket.timeout:
                continue
            except OSError as e:
                print(f"VICE connection error: {e}")
                break

            if not data:
                break

            trace("VICE->SHIM", data, args.trace)

            for b in data:
                if b == ord("+"):
                    command.append(b)
                    if command.endswith(b"+++"):
                        print("modem escape: +++")
                        command.clear()
                    continue

                if b in (13, 10):
                    if not command:
                        continue

                    line = command.decode("ascii", errors="ignore").strip()
                    command.clear()
                    upper = line.upper()
                    print(f"modem command: {line}")

                    if upper == "AT" or upper == "ATH":
                        send_modem_line(conn, "OK", args.trace)
                        close_socket(target_socket)
                        target_socket = None
                        continue

                    target = parse_dial_target(line)
                    if target is None:
                        send_modem_line(conn, "ERROR", args.trace)
                        continue

                    host, port = target
                    try:
                        print(f"dialing TCP target {host}:{port}")
                        target_socket = connect_target(host, port, args.connect_timeout)
                    except OSError as e:
                        print(f"dial failed: {e}")
                        send_modem_line(conn, "NO CARRIER", args.trace)
                        continue

                    send_modem_line(conn, "CONNECT", args.trace)
                    bridge(conn, target_socket, args.serial_baud, args.trace)
                    close_socket(target_socket)
                    target_socket = None
                    return

                command.append(b)

    finally:
        print("VICE disconnected")
        close_socket(target_socket)
        target_socket = None


def parse_args():
    parser = argparse.ArgumentParser(
        description="C64 Ultimate-style AT modem shim for VICE ACIA/RS232 tests."
    )
    parser.add_argument("--host", default=DEFAULT_HOST, help=f"host to bind, default: {DEFAULT_HOST}")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help=f"port to bind, default: {DEFAULT_PORT}")
    parser.add_argument(
        "--connect-timeout",
        type=float,
        default=5.0,
        help="TCP dial timeout in seconds, default: 5.0",
    )
    parser.add_argument(
        "--serial-baud",
        type=int,
        default=DEFAULT_SERIAL_BAUD,
        help=f"pace TCP-to-ACIA bytes at this baud, default: {DEFAULT_SERIAL_BAUD}",
    )
    parser.add_argument("--trace", action="store_true", help="print bridged bytes")
    return parser.parse_args()


def main():
    global listen_socket, vice_socket

    args = parse_args()

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    listen_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    listen_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    listen_socket.bind((args.host, args.port))
    listen_socket.listen(1)
    listen_socket.settimeout(0.5)

    print("VICE modem shim started")
    print(f"Listening for VICE RS232 on {args.host}:{args.port}")
    print(f"Serial pacing: {args.serial_baud} baud")
    print("Point VICE at this with:")
    print(f"  -rsdev1 {args.host}:{args.port}")
    print("Press Ctrl-C to stop.\n")

    try:
        while running:
            try:
                vice_socket, addr = listen_socket.accept()
            except socket.timeout:
                continue
            except OSError:
                break

            tune_tcp_socket(vice_socket)
            handle_vice_connection(vice_socket, addr, args)
            close_socket(vice_socket)
            vice_socket = None

            time.sleep(0.1)

    finally:
        print("Shim stopped")
        shutdown_all()


if __name__ == "__main__":
    main()
