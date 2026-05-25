#!/usr/bin/env python3
import argparse
import signal
import socket
import sys
import time


DEFAULT_HOST = "0.0.0.0"
DEFAULT_PORT = 6764
DEFAULT_RECV_SIZE = 256

# Important for C64 / ACIA polling:
# Send echoed bytes slowly so the C64 program can poll and read them.
DEFAULT_BYTE_DELAY = 0.03


running = True
server_socket = None
client_conn = None


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


def format_bytes(data):
    printable = []

    for b in data:
        if 32 <= b <= 126:
            printable.append(chr(b))
        elif b == 13:
            printable.append("\\r")
        elif b == 10:
            printable.append("\\n")
        elif b == 9:
            printable.append("\\t")
        else:
            printable.append(f"\\x{b:02x}")

    return "".join(printable)


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


def send_slowly(conn, data, byte_delay):
    sent = 0

    for b in data:
        if not running:
            break

        conn.sendall(bytes([b]))
        sent += 1

        if byte_delay > 0:
            time.sleep(byte_delay)

    return sent


def handle_client(conn, addr, recv_size, byte_delay):
    global client_conn

    client_conn = conn

    print(f"Client connected from {addr[0]}:{addr[1]}")

    try:
        while running:
            data = conn.recv(recv_size)

            if not data:
                break

            print(f"received {len(data)} bytes: {format_bytes(data)}")

            sent = send_slowly(conn, data, byte_delay)
            print(f"echoed   {sent} bytes with {byte_delay:.3f}s byte delay")

    except OSError as e:
        print(f"client error: {e}")

    finally:
        print("Client disconnected")

        try:
            conn.close()
        except OSError:
            pass

        client_conn = None


def parse_args():
    parser = argparse.ArgumentParser(
        description="Simple slow TCP echo server for C64 Ultimate / Peddle network tests."
    )

    parser.add_argument(
        "--host",
        default=DEFAULT_HOST,
        help=f"host/interface to bind, default: {DEFAULT_HOST}",
    )

    parser.add_argument(
        "--port",
        type=int,
        default=DEFAULT_PORT,
        help=f"TCP port to listen on, default: {DEFAULT_PORT}",
    )

    parser.add_argument(
        "--recv-size",
        type=int,
        default=DEFAULT_RECV_SIZE,
        help=f"receive buffer size, default: {DEFAULT_RECV_SIZE}",
    )

    parser.add_argument(
        "--byte-delay",
        type=float,
        default=DEFAULT_BYTE_DELAY,
        help=f"delay between echoed bytes in seconds, default: {DEFAULT_BYTE_DELAY}",
    )

    return parser.parse_args()


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

    print("C64 echo server started")
    print(f"Listening on port {args.port}")
    print(f"Byte delay: {args.byte_delay:.3f}s")

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

            handle_client(conn, addr, args.recv_size, args.byte_delay)

    finally:
        print("Server stopped")
        shutdown_all()


if __name__ == "__main__":
    main()