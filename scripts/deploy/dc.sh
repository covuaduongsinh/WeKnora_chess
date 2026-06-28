#!/usr/bin/env bash
# dc.sh — wrapper gói lệnh `docker compose` cho stack WeKnora-Chess.
#
# Stack cờ vua = chồng file compose + bật profile `qdrant`:
#   - docker-compose.yml            (core: frontend/app/docreader/postgres/redis ...)
#   - docker-compose.override.yml   (thay ParadeDB -> pgvector, không cần AVX2)
#   - docker-compose.chess.yml      (thêm sidecar chess-engine + biến WEKNORA_CHESS_*)
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
if [[ -z "${DOMAIN:-}" && -f .env ]]; then
  DOMAIN=$(grep -E '^DOMAIN=' .env | cut -d= -f2- | tr -d '"'"'" | head -1)
fi

# Tự động thêm Caddy overlay khi DOMAIN được đặt
CADDY_FLAG=()
if [[ -n "${DOMAIN:-}" && -f "${PROJECT_ROOT}/docker-compose.caddy.yml" ]]; then
  CADDY_FLAG=(-f docker-compose.caddy.yml)
fi

exec docker compose \
  -f docker-compose.yml \
  -f docker-compose.override.yml \
  -f docker-compose.chess.yml \
  "${CADDY_FLAG[@]}" \
  --profile qdrant \
  "$@"
