#!/usr/bin/env python3
"""
Script khởi động nhanh WeKnora MCP Server

Đây là script khởi động đơn giản, cung cấp các chức năng cơ bản nhất.
Để có thêm tùy chọn, vui lòng dùng main.py
"""

import os
import sys
from pathlib import Path


def main():
    """Hàm khởi động đơn giản"""
    # Thêm thư mục hiện tại vào đường dẫn Python
    current_dir = Path(__file__).parent.absolute()
    if str(current_dir) not in sys.path:
        sys.path.insert(0, str(current_dir))

    # Kiểm tra biến môi trường
    base_url = os.getenv("WEKNORA_BASE_URL", "http://localhost:8080/api/v1")
    api_key = os.getenv("WEKNORA_API_KEY", "")

    print("WeKnora MCP Server")
    print(f"Base URL: {base_url}")
    print(f"API Key: {'đã đặt' if api_key else 'chưa đặt'}")
    print("-" * 40)

    try:
        # Import và chạy
        from main import sync_main

        sync_main()
    except ImportError:
        print("Lỗi: không thể import mô-đun cần thiết")
        print("Vui lòng đảm bảo chạy: pip install -r requirements.txt")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nMáy chủ đã dừng")
    except Exception as e:
        print(f"Lỗi: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
