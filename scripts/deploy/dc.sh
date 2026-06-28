#!/usr/bin/env bash
# dc.sh — wrapper gói lệnh `docker compose` cho stack WeKnora-Chess.
#
# Stack cờ vua = chồng 3 file compose + bật profile `qdrant`:
#   - docker-compose.yml            (core: frontend/app/docreader/postgres/redis ...)
#   - docker-compose.override.yml   (thay ParadeDB -> pgvector, không cần AVX2)
#   - docker-compose.chess.yml      (thêm sidecar chess-engine + biến WEKNORA_CHESS_*)
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
exec docker compose \
  -f docker-compose.yml \
  -f docker-compose.override.yml \
  -f docker-compose.chess.yml \
  --profile qdrant \
  "$@"
