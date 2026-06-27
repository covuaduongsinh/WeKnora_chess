#!/usr/bin/env python3
"""Cầu nối UCI <-> HTTP cho engine cờ vua (mặc định: Arasan, MIT).

Sidecar này bọc một engine UCI và expose một HTTP API tối giản khớp với
internal/chess/http_engine.go của WeKnora:

    POST /analyze   {"fen": "...", "depth": 14}
    -> 200 {"best_move","eval_cp","is_mate","mate_in","depth","pv","side_to_move"}

    GET  /health    -> 200 "ok"

Chỉ dùng thư viện chuẩn Python. Engine giao tiếp qua stdin/stdout. Engine là
tiến trình riêng (ranh giới UCI) nên không ràng buộc giấy phép code gọi nó.

Biến môi trường:
    CHESS_ENGINE_PATH   đường dẫn binary engine UCI (mặc định: arasanx-64)
    CHESS_BRIDGE_PORT   cổng HTTP (mặc định: 8080)
    CHESS_DEFAULT_DEPTH độ sâu mặc định khi request không chỉ định (mặc định 14)
"""
import json
import os
import subprocess
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

ENGINE_PATH = os.environ.get("CHESS_ENGINE_PATH", "arasanx-64")
PORT = int(os.environ.get("CHESS_BRIDGE_PORT", "8080"))
DEFAULT_DEPTH = int(os.environ.get("CHESS_DEFAULT_DEPTH", "14"))


class Engine:
    """Bọc một tiến trình engine UCI, tuần tự hóa truy cập bằng khóa."""

    def __init__(self, path):
        self.proc = subprocess.Popen(
            [path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1,
        )
        self.lock = threading.Lock()
        self._send("uci")
        self._wait("uciok")
        self._send("isready")
        self._wait("readyok")

    def _send(self, line):
        self.proc.stdin.write(line + "\n")
        self.proc.stdin.flush()

    def _wait(self, token):
        for line in self.proc.stdout:
            if line.strip().startswith(token):
                return

    def analyze(self, fen, depth):
        with self.lock:
            self._send("ucinewgame")
            self._send("position fen " + fen)
            self._send("go depth %d" % depth)
            result = {
                "best_move": "",
                "eval_cp": 0,
                "is_mate": False,
                "mate_in": 0,
                "depth": depth,
                "pv": [],
            }
            for line in self.proc.stdout:
                line = line.strip()
                if line.startswith("info "):
                    self._parse_info(line, result)
                elif line.startswith("bestmove"):
                    parts = line.split()
                    if len(parts) >= 2 and parts[1] != "(none)":
                        result["best_move"] = parts[1]
                    return result
            return result

    @staticmethod
    def _parse_info(line, result):
        fields = line.split()
        i = 0
        while i < len(fields):
            tok = fields[i]
            if tok == "depth" and i + 1 < len(fields):
                try:
                    result["depth"] = int(fields[i + 1])
                except ValueError:
                    pass
            elif tok == "score" and i + 2 < len(fields):
                kind, val = fields[i + 1], fields[i + 2]
                try:
                    num = int(val)
                except ValueError:
                    num = 0
                if kind == "cp":
                    result["is_mate"] = False
                    result["eval_cp"] = num
                elif kind == "mate":
                    result["is_mate"] = True
                    result["mate_in"] = num
            elif tok == "pv":
                result["pv"] = fields[i + 1:]
                break
            i += 1


engine = Engine(ENGINE_PATH)


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass  # giữ log sạch

    def _json(self, code, payload):
        body = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/health":
            self.send_response(200)
            self.send_header("Content-Length", "2")
            self.end_headers()
            self.wfile.write(b"ok")
        else:
            self._json(404, {"error": "not found"})

    def do_POST(self):
        if self.path != "/analyze":
            self._json(404, {"error": "not found"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
            data = json.loads(self.rfile.read(length) or b"{}")
            fen = data.get("fen", "").strip()
            if not fen:
                self._json(400, {"error": "thiếu fen"})
                return
            depth = int(data.get("depth") or DEFAULT_DEPTH)
            result = engine.analyze(fen, depth)
            self._json(200, result)
        except Exception as exc:  # noqa: BLE001 - sidecar phải bền bỉ
            self._json(500, {"error": str(exc)})


if __name__ == "__main__":
    server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
    print("chess-engine bridge nghe ở cổng %d (engine: %s)" % (PORT, ENGINE_PATH), flush=True)
    server.serve_forever()
