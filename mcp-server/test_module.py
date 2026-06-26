#!/usr/bin/env python3
"""
Script kiểm thử mô-đun WeKnora MCP Server

Kiểm thử các cách khởi động và chức năng của mô-đun
"""

import os
import subprocess
import sys
from pathlib import Path


def test_imports():
    """Kiểm thử import mô-đun"""
    print("=== Kiểm thử import mô-đun ===")

    try:
        # Kiểm thử phụ thuộc cơ bản
        import mcp

        print("✓ mcp import mô-đun thành công")

        import requests

        print("✓ requests import mô-đun thành công")

        # Kiểm thử mô-đun chính
        import weknora_mcp_server

        print("✓ weknora_mcp_server import mô-đun thành công")

        # Kiểm thử import gói
        from weknora_mcp_server import WeKnoraClient, run

        print("✓ WeKnoraClient và hàm run import thành công")

        # Kiểm thử điểm vào chính
        import main

        print("✓ main import mô-đun thành công")

        return True

    except ImportError as e:
        print(f"✗ import thất bại: {e}")
        return False


def test_environment():
    """Kiểm thử cấu hình môi trường"""
    print("\n=== Kiểm thử cấu hình môi trường ===")

    base_url = os.getenv("WEKNORA_BASE_URL")
    api_key = os.getenv("WEKNORA_API_KEY")

    print(f"WEKNORA_BASE_URL: {base_url or 'chưa đặt (sẽ dùng giá trị mặc định)'}")
    print(f"WEKNORA_API_KEY: {'đã đặt' if api_key else 'chưa đặt'}")

    if not base_url:
        print("Gợi ý: có thể đặt biến môi trường WEKNORA_BASE_URL")

    if not api_key:
        print("Gợi ý: nên đặt biến môi trường WEKNORA_API_KEY")

    return True


def test_client_creation():
    """Kiểm thử tạo client"""
    print("\n=== Kiểm thử tạo client ===")

    try:
        from weknora_mcp_server import WeKnoraClient

        base_url = os.getenv("WEKNORA_BASE_URL", "http://localhost:8080/api/v1")
        api_key = os.getenv("WEKNORA_API_KEY", "test_key")

        client = WeKnoraClient(base_url, api_key)
        print("✓ WeKnoraClient tạo thành công")

        # Kiểm tra thuộc tính client
        assert client.base_url == base_url
        assert client.api_key == api_key
        print("✓ Cấu hình client đúng")

        return True

    except Exception as e:
        print(f"✗ Tạo client thất bại: {e}")
        return False


def test_file_structure():
    """Kiểm thử cấu trúc tệp"""
    print("\n=== Kiểm thử cấu trúc tệp ===")

    required_files = [
        "__init__.py",
        "main.py",
        "run_server.py",
        "weknora_mcp_server.py",
        "requirements.txt",
        "setup.py",
        "pyproject.toml",
        "README.md",
        "INSTALL.md",
        "LICENSE",
        "MANIFEST.in",
    ]

    missing_files = []
    for file in required_files:
        if Path(file).exists():
            print(f"✓ {file}")
        else:
            print(f"✗ {file} (thiếu)")
            missing_files.append(file)

    if missing_files:
        print(f"Tệp thiếu: {missing_files}")
        return False

    print("✓ Mọi tệp bắt buộc đều tồn tại")
    return True


def test_entry_points():
    """Kiểm thử điểm vào"""
    print("\n=== Kiểm thử điểm vào ===")

    # Kiểm thử tùy chọn trợ giúp của main.py
    try:
        result = subprocess.run(
            [sys.executable, "main.py", "--help"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode == 0:
            print("✓ main.py --help hoạt động bình thường")
        else:
            print(f"✗ main.py --help thất bại: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("✗ main.py --help quá thời gian")
        return False
    except Exception as e:
        print(f"✗ main.py --help Lỗi: {e}")
        return False

    # Kiểm thử kiểm tra môi trường
    try:
        result = subprocess.run(
            [sys.executable, "main.py", "--check-only"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode == 0:
            print("✓ main.py --check-only hoạt động bình thường")
        else:
            print(f"✗ main.py --check-only thất bại: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("✗ main.py --check-only quá thời gian")
        return False
    except Exception as e:
        print(f"✗ main.py --check-only Lỗi: {e}")
        return False

    return True


def test_wiki_tools():
    """Kiểm thử đăng ký công cụ Wiki và phương thức Client"""
    print("\n=== Kiểm thử công cụ Wiki ===")

    try:
        import weknora_mcp_server

        # Xác minh phương thức Client tồn tại
        client = weknora_mcp_server.WeKnoraClient("http://localhost:8080/api/v1", "test")
        for method in ["wiki_search", "wiki_read_page", "wiki_index_view"]:
            assert hasattr(client, method), f"WeKnoraClient thiếu phương thức: {method}"
            assert callable(getattr(client, method)), f"{method} không gọi được"
            print(f"✓ WeKnoraClient.{method} tồn tại")

        return True

    except Exception as e:
        print(f"✗ Kiểm thử công cụ Wiki thất bại: {e}")
        return False


def test_package_installation():
    """Kiểm thử cài gói (chế độ phát triển)"""
    print("\n=== Kiểm thử cài gói ===")

    try:
        # Kiểm tra có thể cài ở chế độ phát triển không
        result = subprocess.run(
            [sys.executable, "setup.py", "check"],
            capture_output=True,
            text=True,
            timeout=30,
        )

        if result.returncode == 0:
            print("✓ setup.py kiểm tra đạt")
        else:
            print(f"✗ setup.py kiểm tra thất bại: {result.stderr}")
            return False

    except subprocess.TimeoutExpired:
        print("✗ setup.py kiểm tra quá thời gian")
        return False
    except Exception as e:
        print(f"✗ setup.py kiểm tra lỗi: {e}")
        return False

    return True


def main():
    """Chạy tất cả kiểm thử"""
    print("Kiểm thử mô-đun WeKnora MCP Server")
    print("=" * 50)

    tests = [
        ("Import mô-đun", test_imports),
        ("Cấu hình môi trường", test_environment),
        ("Tạo client", test_client_creation),
        ("Cấu trúc tệp", test_file_structure),
        ("Điểm vào", test_entry_points),
        ("Công cụ Wiki", test_wiki_tools),
        ("Cài gói", test_package_installation),
    ]

    passed = 0
    total = len(tests)

    for test_name, test_func in tests:
        try:
            if test_func():
                passed += 1
            else:
                print(f"Kiểm thử thất bại: {test_name}")
        except Exception as e:
            print(f"Kiểm thử ngoại lệ: {test_name} - {e}")

    print("\n" + "=" * 50)
    print(f"Kết quả kiểm thử: {passed}/{total} đạt")

    if passed == total:
        print("✓ Tất cả kiểm thử đạt! Mô-đun có thể dùng bình thường.")
        return True
    else:
        print("✗ Một số kiểm thử thất bại, vui lòng kiểm tra các lỗi trên.")
        return False


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
