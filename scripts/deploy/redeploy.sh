#!/usr/bin/env bash
# redeploy.sh — cập nhật WeKnora-Chess trên server sau khi bạn push code mới lên GitHub.
#
#   git pull  ->  build lại frontend dist  ->  docker compose up -d --build  ->  dọn image cũ
#
# Chạy từ PC của bạn:
#   ssh root@<IP> "bash /opt/WeKnora_chess/scripts/deploy/redeploy.sh"
# Hoặc SSH vào server rồi chạy trực tiếp.
#
# .env và các docker volume (postgres/qdrant/data-files) KHÔNG bị động tới.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
FRONTEND_NODE_IMAGE="${FRONTEND_NODE_IMAGE:-node:22-bookworm}"

cd "${PROJECT_ROOT}"

echo "[redeploy] git pull ..."
git pull --ff-only

echo "[redeploy] build lại frontend dist ..."
COMMIT_ID="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
docker run --rm \
  -e VITE_IS_DOCKER=true \
  -e VITE_FRONTEND_COMMIT="${COMMIT_ID}" \
  -e NODE_OPTIONS="--max-old-space-size=${NODE_HEAP_MB:-4096}" \
  -v "${PROJECT_ROOT}/frontend":/app \
  -w /app \
  "${FRONTEND_NODE_IMAGE}" \
  sh -lc "npm ci && npm run build"

echo "[redeploy] build từng image (tuần tự, đỡ ngốn RAM) rồi up ..."
for svc in app docreader chess-engine frontend; do
  echo "[redeploy]   build ${svc} ..."
  bash "${SCRIPT_DIR}/dc.sh" build "${svc}"
done
bash "${SCRIPT_DIR}/dc.sh" up -d

echo "[redeploy] dọn image dangling (an toàn, không đụng image đang dùng) ..."
docker image prune -f >/dev/null 2>&1 || true

echo "[redeploy] xong. Trạng thái:"
bash "${SCRIPT_DIR}/dc.sh" ps
