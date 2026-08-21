#!/usr/bin/env bash
# pull-deploy.sh — cập nhật VPS bằng image DỰNG SẴN từ GHCR (KHÔNG build trên VPS).
#
# Dùng bởi CI (GitHub Actions SSH vào chạy sau khi build xong), hoặc chạy tay khi
# muốn cập nhật/rollback:
#   IMAGE_TAG=<github-sha|latest> bash scripts/deploy/pull-deploy.sh
#
# Yêu cầu: VPS đã `docker login ghcr.io` 1 lần để kéo được image private
# (xem docs/deploy/cicd.md). .env và volume dữ liệu KHÔNG bị động tới.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
export IMAGE_TAG="${IMAGE_TAG:-latest}"

cd "${PROJECT_ROOT}"


# Đọc DOMAIN từ .env nếu chưa có trong môi trường
# (`|| true`: nếu .env không có dòng DOMAIN= nào thì grep thoát 1 — dưới
#  `set -euo pipefail` điều đó làm cả script dừng ngay, im lặng, không in gì.
#  Đã xác nhận đây là lỗi có sẵn, tái hiện được ngay trên .env thật.)
if [[ -z "${DOMAIN:-}" && -f .env ]]; then
  DOMAIN=$(grep -E '^DOMAIN=' .env | cut -d= -f2- | tr -d '"'"'" | head -1 || true)
fi

# Tự động thêm Caddy overlay khi DOMAIN được đặt
CADDY_FLAG=()
if [[ -n "${DOMAIN:-}" && -f docker-compose.caddy.yml ]]; then
  CADDY_FLAG=(-f docker-compose.caddy.yml)
fi

# Đọc LLM_GATEWAY_ENABLED từ .env nếu chưa có trong môi trường (`|| true`: xem
# ghi chú ở khối DOMAIN phía trên — cùng một lỗi, .env chưa có dòng này là
# tình huống MẶC ĐỊNH cho biến mới này nên bắt buộc phải chịu được).
if [[ -z "${LLM_GATEWAY_ENABLED:-}" && -f .env ]]; then
  LLM_GATEWAY_ENABLED=$(grep -E '^LLM_GATEWAY_ENABLED=' .env | cut -d= -f2- | tr -d '"'"'" | head -1 || true)
fi

# Cổng LLM là OPT-IN: không bật thì deploy chạy y hệt như trước.
LLM_FLAG=()
if [[ "${LLM_GATEWAY_ENABLED:-}" == "true" && -f docker-compose.llm.yml ]]; then
  LLM_FLAG=(-f docker-compose.llm.yml)
fi

# Đọc CLAUDE_BRIDGE_ENABLED từ .env nếu chưa có trong môi trường (cùng khuôn
# `|| true` như DOMAIN/LLM_GATEWAY_ENABLED phía trên).
if [[ -z "${CLAUDE_BRIDGE_ENABLED:-}" && -f .env ]]; then
  CLAUDE_BRIDGE_ENABLED=$(grep -E '^CLAUDE_BRIDGE_ENABLED=' .env | cut -d= -f2- | tr -d '"'"'" | head -1 || true)
fi

# Cầu nối Claude Agent SDK (Giai đoạn 2, tùy chọn) — cũng OPT-IN.
CLAUDE_FLAG=()
if [[ "${CLAUDE_BRIDGE_ENABLED:-}" == "true" && -f docker-compose.llm-claude.yml ]]; then
  CLAUDE_FLAG=(-f docker-compose.llm-claude.yml)
fi

dc() {
  docker compose \
    -f docker-compose.yml \
    -f docker-compose.override.yml \
    -f docker-compose.chess.yml \
    -f docker-compose.ghcr.yml \
    "${LLM_FLAG[@]}" \
    "${CLAUDE_FLAG[@]}" \
    "${CADDY_FLAG[@]}" \
    --profile qdrant "$@"
}

# ⚠️ claude-bridge KHÔNG có image dựng sẵn trên GHCR — nó build thẳng từ repo
# nguồn Meridian (xem docker-compose.llm-claude.yml). Triết lý của script này
# là "không build trên VPS" (`--no-build` ở dưới), nên nếu bật cầu nối mà chưa
# từng build, `dc up -d --no-build` sẽ báo lỗi "no such image" khó hiểu.
# Chặn sớm với thông báo rõ ràng thay vì để lỗi mù mờ đó xảy ra.
if [[ -n "${CLAUDE_FLAG[*]:-}" ]]; then
  if [[ -z "$(dc images -q claude-bridge 2>/dev/null)" ]]; then
    echo "[pull-deploy] LOI: CLAUDE_BRIDGE_ENABLED=true nhung image claude-bridge chua tung build." >&2
    echo "[pull-deploy] Chay MOT LAN truoc: scripts/deploy/dc.sh build claude-bridge" >&2
    echo "[pull-deploy] Chi tiet: docs/llm-gateway.md muc \"Giai doan 2\"." >&2
    exit 1
  fi
fi

echo "[pull-deploy] IMAGE_TAG=${IMAGE_TAG} — kéo image cờ vua từ GHCR ..."
# CỐ Ý không kéo llm-gateway ở đây: image đó ghim tag `main-stable` của bên thứ
# ba, không đổi theo commit của ta. Để `up -d` tự kéo lần đầu; nâng phiên bản là
# thao tác có chủ đích (`dc.sh pull llm-gateway`), không lẫn vào deploy thường.
dc pull app frontend docreader chess-engine

echo "[pull-deploy] khởi động lại (không build trên VPS) ..."
dc up -d --no-build

echo "[pull-deploy] dọn image cũ (dangling) ..."
docker image prune -f >/dev/null 2>&1 || true

echo "[pull-deploy] xong. Trạng thái:"
dc ps
