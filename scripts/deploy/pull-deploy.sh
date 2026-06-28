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

dc() {
  docker compose \
    -f docker-compose.yml \
    -f docker-compose.override.yml \
    -f docker-compose.chess.yml \
    -f docker-compose.ghcr.yml \
    --profile qdrant "$@"
}

echo "[pull-deploy] IMAGE_TAG=${IMAGE_TAG} — kéo image cờ vua từ GHCR ..."
dc pull app frontend docreader chess-engine

echo "[pull-deploy] khởi động lại (không build trên VPS) ..."
dc up -d --no-build

echo "[pull-deploy] dọn image cũ (dangling) ..."
docker image prune -f >/dev/null 2>&1 || true

echo "[pull-deploy] xong. Trạng thái:"
dc ps
