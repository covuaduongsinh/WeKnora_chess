#!/usr/bin/env python3
"""
Script khởi động WeKnora MCP Server
"""

import asyncio
import os
import sys


def check_environment():
    """Kiểm tra cấu hình môi trường"""
    base_url = os.getenv("WEKNORA_BASE_URL")
    api_key = os.getenv("WEKNORA_API_KEY")

    if not base_url:
        print(
            "Cảnh báo: biến môi trường WEKNORA_BASE_URL chưa đặt, dùng giá trị mặc định: http://localhost:8080/api/v1"
        )

    if not api_key:
        print("Cảnh báo: biến môi trường WEKNORA_API_KEY chưa đặt")

    print(f"WeKnora Base URL: {base_url or 'http://localhost:8080/api/v1'}")
    print(f"API Key: {'đã đặt' if api_key else 'chưa đặt'}")


def main():
    """Hàm chính"""
    print("Đang khởi động WeKnora MCP Server...")
    check_environment()

    try:
        from weknora_mcp_server import run

        asyncio.run(run())
    except ImportError as e:
        print(f"Lỗi import: {e}")
        print("Vui lòng đảm bảo đã cài tất cả phụ thuộc: pip install -r requirements.txt")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nMáy chủ đã dừng")
    except Exception as e:
        print(f"Lỗi chạy máy chủ: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
