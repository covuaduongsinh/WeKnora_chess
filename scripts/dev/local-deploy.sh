#!/usr/bin/env bash
# local-deploy.sh — Đưa stack LOCAL về đúng working tree bằng MỘT lệnh.
#
# Quy trình local-first: sửa code → `make local-deploy` (build+recreate+verify ở
# local) → test ở http://localhost → `git push` (CI deploy VPS). Nhờ vậy local
# luôn ≥ VPS khi đang phát triển (tránh "VPS mới hơn local").
#
# Dùng:
#   scripts/dev/local-deploy.sh                 # build app+frontend TỪ NGUỒN
#   scripts/dev/local-deploy.sh --app-only      # CHỈ build+recreate app (sửa backend)
#   scripts/dev/local-deploy.sh --frontend-only # CHỈ build+recreate frontend (sửa UI)
#   scripts/dev/local-deploy.sh --ghcr [tag]    # nhanh: pull image CI từ GHCR
#   (kết hợp được: --ghcr --app-only ...)
#
# Mẹo TỐC ĐỘ: chỉ build phần đã đổi; layer apt nặng được Docker cache qua các lần
# build (LABEL commit đặt cuối Dockerfile.app nên không bust cache). Biến: GHCR_PREFIX.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

MODE=build
TAG=latest
SERVICES="app frontend"
while [ $# -gt 0 ]; do
  case "$1" in
    --ghcr) MODE=ghcr; shift
            if [ $# -gt 0 ] && [ "${1#--}" = "$1" ]; then TAG="$1"; shift; fi ;;
    --app-only) SERVICES="app"; shift ;;
    --frontend-only) SERVICES="frontend"; shift ;;
    -h|--help) sed -n '2,17p' "$0"; exit 0 ;;
    *) echo "Tham số lạ: $1" >&2; exit 2 ;;
  esac
done

has() { case " $SERVICES " in *" $1 "*) return 0 ;; *) return 1 ;; esac; }

COMPOSE=(docker compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.chess.yml)
GHCR_PREFIX="${GHCR_PREFIX:-ghcr.io/covuaduongsinh/weknora-chess}"
APP_IMG=wechatopenai/weknora-app:latest
UI_IMG=wechatopenai/weknora-ui:latest

if [ "$MODE" = build ]; then
  echo "==> [1/4] Build TỪ NGUỒN: $SERVICES ..."
  has app      && "$ROOT/scripts/build_images.sh" --app
  has frontend && "$ROOT/scripts/build_images.sh" --frontend
else
  echo "==> [1/4] Pull GHCR ($TAG) + retag: $SERVICES ..."
  if has app; then docker pull "$GHCR_PREFIX-app:$TAG"; docker tag "$GHCR_PREFIX-app:$TAG" "$APP_IMG"; fi
  if has frontend; then docker pull "$GHCR_PREFIX-ui:$TAG"; docker tag "$GHCR_PREFIX-ui:$TAG" "$UI_IMG"; fi
fi

echo "==> [2/4] Recreate ($SERVICES, --no-deps)..."
# shellcheck disable=SC2086
"${COMPOSE[@]}" up -d --no-deps --force-recreate $SERVICES

echo "==> [3/4] Restart frontend (chống nginx stale-IP sau khi app đổi IP)..."
docker restart WeKnora-frontend >/dev/null

if has app; then
  echo "==> [4/4] Chờ app healthy..."
  health=unknown
  for _ in $(seq 1 60); do
    health="$(docker inspect -f '{{ if .State.Health }}{{ .State.Health.Status }}{{ else }}no-healthcheck{{ end }}' WeKnora-app 2>/dev/null || echo none)"
    case "$health" in healthy|no-healthcheck) break ;; esac
    sleep 3
  done
  echo "    app health: $health"
fi

echo
echo "==== Kết quả ===="
mig="$(docker exec WeKnora-postgres sh -lc 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "select version,dirty from schema_migrations;"' 2>/dev/null | tr -d '\r')"
echo "  schema_migrations : ${mig:-<không đọc được>}"
curl -sS -o /dev/null -w "  localhost/healthz : HTTP %{http_code}\n" http://localhost/healthz --max-time 15 || true
"$ROOT/scripts/dev/local-status.sh" || true
echo
echo "✅ Xong. Mở http://localhost (Ctrl+Shift+R nếu vừa đổi frontend)."
