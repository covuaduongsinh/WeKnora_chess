#!/usr/bin/env python3
"""
Kiểm thử import MCP
"""

try:
    import mcp.types as types

    print("✓ mcp.types import thành công")
except ImportError as e:
    print(f"✗ mcp.types import thất bại: {e}")

try:
    from mcp.server import NotificationOptions, Server

    print("✓ mcp.server import thành công")
except ImportError as e:
    print(f"✗ mcp.server import thất bại: {e}")

try:
    import mcp.server.stdio

    print("✓ mcp.server.stdio import thành công")
except ImportError as e:
    print(f"✗ mcp.server.stdio import thất bại: {e}")

try:
    from mcp.server.models import InitializationOptions

    print("✓ InitializationOptions từ mcp.server.models import thành công")
except ImportError:
    try:
        from mcp import InitializationOptions

        print("✓ InitializationOptions từ mcp import thành công")
    except ImportError as e:
        print(f"✗ InitializationOptions import thất bại: {e}")

# Kiểm tra cấu trúc gói MCP
import mcp

print(f"\nPhiên bản gói MCP: {getattr(mcp, '__version__', 'Không rõ')}")
print(f"Đường dẫn gói MCP: {mcp.__file__}")
print(f"Nội dung gói MCP: {dir(mcp)}")
