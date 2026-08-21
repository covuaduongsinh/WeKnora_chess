#!/usr/bin/env bash
# dc.sh — wrapper gói lệnh `docker compose` cho stack WeKnora-Chess.
#
# Stack cờ vua = chồng file compose + bật profile `qdrant`:
#   - docker-compose.yml            (core: frontend/app/docreader/postgres/redis ...)
#   - docker-compose.override.yml   (thay ParadeDB -> pgvector, không cần AVX2)
#   - docker-compose.chess.yml      (thêm sidecar chess-engine + biến WEKNORA_CHESS_*)
#   - docker-compose.llm.yml        (tự động thêm khi LLM_GATEWAY_ENABLED=true trong .env)
#   - docker-compose.llm-claude.yml (tự động thêm khi CLAUDE_BRIDGE_ENABLED=true trong .env)
#   - docker-compose.caddy.yml      (tự động thêm khi DOMAIN được đặt trong .env)
#   - --profile qdrant              (bật container qdrant; RETRIEVE_DRIVER=qdrant cần nó)
#
# Gói vào một chỗ để systemd / redeploy.sh / thao tác tay đều dùng chung, tránh gõ sai.
#
# Ví dụ:
#   scripts/deploy/dc.sh up -d --build
#   scripts/deploy/dc.sh ps
#   scripts/deploy/dc.sh logs -f app
#   scripts/deploy/dc.sh down
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"

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
if [[ -n "${DOMAIN:-}" && -f "${PROJECT_ROOT}/docker-compose.caddy.yml" ]]; then
  CADDY_FLAG=(-f docker-compose.caddy.yml)
fi

# Đọc LLM_GATEWAY_ENABLED từ .env nếu chưa có trong môi trường (`|| true`: xem
# ghi chú ở khối DOMAIN phía trên — cùng một lỗi, .env chưa có dòng này là
# tình huống MẶC ĐỊNH cho biến mới này nên bắt buộc phải chịu được).
if [[ -z "${LLM_GATEWAY_ENABLED:-}" && -f .env ]]; then
  LLM_GATEWAY_ENABLED=$(grep -E '^LLM_GATEWAY_ENABLED=' .env | cut -d= -f2- | tr -d '"'"'" | head -1 || true)
fi

# Cổng LLM là OPT-IN: không bật thì stack chạy y hệt như trước, không đụng gì.
# Xem docs/llm-gateway.md.
LLM_FLAG=()
if [[ "${LLM_GATEWAY_ENABLED:-}" == "true" && -f "${PROJECT_ROOT}/docker-compose.llm.yml" ]]; then
  LLM_FLAG=(-f docker-compose.llm.yml)
fi

# Đọc CLAUDE_BRIDGE_ENABLED từ .env nếu chưa có trong môi trường (cùng khuôn
# `|| true` như DOMAIN/LLM_GATEWAY_ENABLED phía trên).
if [[ -z "${CLAUDE_BRIDGE_ENABLED:-}" && -f .env ]]; then
  CLAUDE_BRIDGE_ENABLED=$(grep -E '^CLAUDE_BRIDGE_ENABLED=' .env | cut -d= -f2- | tr -d '"'"'" | head -1 || true)
fi

# Cầu nối Claude Agent SDK (Giai đoạn 2, tùy chọn) — cũng OPT-IN, độc lập với
# LLM_GATEWAY_ENABLED. Xem docs/llm-gateway.md mục "Giai đoạn 2".
CLAUDE_FLAG=()
if [[ "${CLAUDE_BRIDGE_ENABLED:-}" == "true" && -f "${PROJECT_ROOT}/docker-compose.llm-claude.yml" ]]; then
  CLAUDE_FLAG=(-f docker-compose.llm-claude.yml)
fi

exec docker compose \
  -f docker-compose.yml \
  -f docker-compose.override.yml \
  -f docker-compose.chess.yml \
  "${LLM_FLAG[@]}" \
  "${CLAUDE_FLAG[@]}" \
  "${CADDY_FLAG[@]}" \
  --profile qdrant \
  "$@"
