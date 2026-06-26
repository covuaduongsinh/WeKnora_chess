#!/usr/bin/env python3
"""
Điểm vào chính của WeKnora MCP Server

Tệp này cung cấp điểm vào thống nhất để khởi động máy chủ WeKnora MCP.
Có thể chạy theo nhiều cách:
1. python main.py
2. python -m weknora_mcp_server
3. weknora-mcp-server (sau khi cài)
"""

import argparse
import asyncio
import os
import sys
from pathlib import Path


def setup_environment():
    """Thiết lập môi trường và đường dẫn"""
    # Đảm bảo thư mục hiện tại nằm trong đường dẫn Python
    current_dir = Path(__file__).parent.absolute()
    if str(current_dir) not in sys.path:
        sys.path.insert(0, str(current_dir))


def check_dependencies():
    """Kiểm tra phụ thuộc đã cài chưa"""
    try:
        import mcp
        import requests

        return True
    except ImportError as e:
        print(f"Thiếu phụ thuộc: {e}")
        print("Vui lòng chạy: pip install -r requirements.txt")
        return False


def check_environment_variables():
    """Kiểm tra cấu hình biến môi trường"""
    base_url = os.getenv("WEKNORA_BASE_URL")
    api_key = os.getenv("WEKNORA_API_KEY")

    print("=== Kiểm tra môi trường WeKnora MCP Server ===")
    print(f"Base URL: {base_url or 'http://localhost:8080/api/v1 (mặc định)'}")
    print(f"API Key: {'đã đặt' if api_key else 'chưa đặt (cảnh báo)'}")

    if not base_url:
        print("Gợi ý: có thể đặt biến môi trường WEKNORA_BASE_URL")

    if not api_key:
        print("Cảnh báo: nên đặt biến môi trường WEKNORA_API_KEY")

    print("=" * 40)
    return True


def parse_arguments():
    """Phân tích tham số dòng lệnh"""
    parser = argparse.ArgumentParser(
        description="WeKnora MCP Server - Model Context Protocol server for WeKnora API",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Ví dụ:
  python main.py                    # Khởi động với cấu hình mặc định
  python main.py --check-only       # Chỉ kiểm tra môi trường, không khởi động máy chủ
  python main.py --verbose          # Bật log chi tiết
  
Biến môi trường:
  WEKNORA_BASE_URL    URL gốc API WeKnora (mặc định: http://localhost:8080/api/v1)
  WEKNORA_API_KEY     Khóa API WeKnora
        """,
    )

    parser.add_argument(
        "--check-only", action="store_true", help="Chỉ kiểm tra cấu hình môi trường, không khởi động máy chủ"
    )

    parser.add_argument("--verbose", "-v", action="store_true", help="Bật xuất log chi tiết")

    parser.add_argument(
        "--version", action="version", version="WeKnora MCP Server 1.0.0"
    )

    parser.add_argument(
        "--transport",
        choices=["stdio", "sse", "http"],
        default=os.getenv("MCP_TRANSPORT", "stdio"),
        help="Transport type: stdio (default), sse, or http",
    )
    parser.add_argument(
        "--host",
        default=os.getenv("MCP_HOST", "0.0.0.0"),
        help="Bind host for network transports (default: 0.0.0.0)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.getenv("MCP_PORT", "8000")),
        help="Bind port for network transports (default: 8000)",
    )

    return parser.parse_args()


async def main():
    """Hàm chính"""
    args = parse_arguments()

    # Thiết lập môi trường
    setup_environment()

    # Kiểm tra phụ thuộc
    if not check_dependencies():
        sys.exit(1)

    # Kiểm tra biến môi trường
    check_environment_variables()

    # Nếu chỉ kiểm tra môi trường thì thoát
    if args.check_only:
        print("Kiểm tra môi trường hoàn tất.")
        return

    # Đặt mức log
    if args.verbose:
        import logging

        logging.basicConfig(level=logging.DEBUG)
        print("Đã bật chế độ log chi tiết")

    try:
        print(f"Đang khởi động WeKnora MCP Server (transport={args.transport})...")

        from weknora_mcp_server import run_stdio, run_sse, run_http

        # Select transport mode based on CLI argument or MCP_TRANSPORT env var
        # - stdio: Default, used by VS Code Copilot for local integration
        # - sse: Server-Sent Events over HTTP, suitable for cloud/remote deployments
        # - http: Streamable HTTP sessions (MCP 2025-03-26 spec), compatible with REST clients
        if args.transport == "stdio":
            # Stdio mode: communication via stdin/stdout pipes (typical for CLI integrations)
            await run_stdio()
        elif args.transport == "sse":
            # SSE mode: HTTP server with Server-Sent Events for bidirectional streaming
            await run_sse(args.host, args.port)
        elif args.transport == "http":
            # HTTP mode: HTTP REST server with request/response model
            await run_http(args.host, args.port)

    except ImportError as e:
        print(f"Lỗi import: {e}")
        print("Vui lòng đảm bảo mọi tệp ở đúng vị trí")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nMáy chủ đã dừng")
    except Exception as e:
        print(f"Lỗi chạy máy chủ: {e}")
        if args.verbose:
            import traceback

            traceback.print_exc()
        sys.exit(1)


def sync_main():
    """Phiên bản đồng bộ của hàm chính, dùng cho entry_points"""
    asyncio.run(main())


if __name__ == "__main__":
    asyncio.run(main())
