#!/usr/bin/env bash
# backup.sh — Sao lưu TOÀN BỘ dữ liệu WeKnora-Chess thành 1 gói .tar.gz duy nhất.
#
# Chạy trên máy đang chạy stack (PC local để di trú lên VPS, hoặc trên chính VPS
# để backup định kỳ). Cần Docker đang chạy và các container WeKnora đang lên.
#
# Gói gồm:
#   - db.sql.gz       : toàn bộ PostgreSQL (agent, model, khóa học, ván đấu, bài tập,
#                       tri thức metadata, user/tenant, session, wiki...) qua pg_dump
#   - qdrant.tgz      : vector tri thức (Qdrant /qdrant/storage)
#   - data-files.tgz  : file tài liệu đã upload (/data/files, kèm .crypto_state.json)
#   - keys.env        : 4 khóa mã hóa trong .env (để restore giải mã được key model)
#   - manifest.txt    : metadata (ngày, version) để cảnh báo lệch version khi restore
#
# Dùng:
#   bash scripts/deploy/backup.sh                 # xuất ra thư mục hiện tại
#   OUT_DIR=/duong/dan bash scripts/deploy/backup.sh
#
# Biến tùy chọn: OUT_DIR, PG_CONTAINER, QDRANT_CONTAINER, APP_CONTAINER
set -euo pipefail

# Ngăn Git Bash (Windows) tự đổi các đường dẫn NỘI BỘ container (vd /qdrant/storage)
# thành đường dẫn Windows. Vô hại trên Linux. Script không bind-mount host path
# (đều dùng stdin/stdout redirect) nên bật cờ này an toàn.
export MSYS_NO_PATHCONV=1

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env"
OUT_DIR="${OUT_DIR:-$(pwd)}"

PG_CONTAINER="${PG_CONTAINER:-WeKnora-postgres}"
QDRANT_CONTAINER="${QDRANT_CONTAINER:-WeKnora-qdrant}"
APP_CONTAINER="${APP_CONTAINER:-WeKnora-app}"

log() { printf '[backup] %s\n' "$*"; }
die() { printf '[backup] ERROR: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null 2>&1 || die "không thấy docker (Docker chưa chạy?)"
[ -f "$ENV_FILE" ] || die "không thấy .env ở $ENV_FILE"

get_env() { grep -E "^$1=" "$ENV_FILE" | head -1 | cut -d= -f2- ; }
DB_USER="$(get_env DB_USER)"; DB_USER="${DB_USER:-postgres}"
DB_NAME="$(get_env DB_NAME)"; DB_NAME="${DB_NAME:-WeKnora}"

running() { [ "$(docker inspect -f '{{.State.Running}}' "$1" 2>/dev/null || echo false)" = "true" ]; }
running "$PG_CONTAINER" || die "container $PG_CONTAINER chưa chạy — hãy bật stack trước (dc.sh up -d)"

STAMP="$(date +%Y%m%d-%H%M%S)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# 1) PostgreSQL (pg_dump qua stdout -> file host; không bind-mount)
log "1/5 dump PostgreSQL (db=$DB_NAME, user=$DB_USER) ..."
docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" --clean --if-exists \
  | gzip > "$WORK/db.sql.gz" || die "pg_dump thất bại"

# 2) Qdrant — dừng để nhất quán, tar qua stdout, rồi bật lại nếu trước đó đang chạy
log "2/5 sao lưu Qdrant ..."
if running "$QDRANT_CONTAINER"; then QWAS=1; else QWAS=0; fi
[ "$QWAS" = "1" ] && docker stop "$QDRANT_CONTAINER" >/dev/null
docker run --rm --volumes-from "$QDRANT_CONTAINER" alpine \
  tar czf - -C /qdrant/storage . > "$WORK/qdrant.tgz" || die "tar Qdrant thất bại"
[ "$QWAS" = "1" ] && docker start "$QDRANT_CONTAINER" >/dev/null

# 3) data-files (file tĩnh — backup khi đang chạy được)
log "3/5 sao lưu data-files ..."
docker run --rm --volumes-from "$APP_CONTAINER" alpine \
  tar czf - -C /data/files . > "$WORK/data-files.tgz" || die "tar data-files thất bại"

# 4) khóa mã hóa (BẮT BUỘC để key model giải mã được sau restore)
log "4/5 trích khóa mã hóa từ .env ..."
: > "$WORK/keys.env"
for k in SYSTEM_AES_KEY TENANT_AES_KEY CRYPTO_MASTER_KEY CRYPTO_SALT; do
  printf '%s=%s\n' "$k" "$(get_env "$k")" >> "$WORK/keys.env"
done

# 5) manifest
log "5/5 ghi manifest + đóng gói ..."
{
  echo "created=$(date -Iseconds)"
  echo "db_name=$DB_NAME"
  echo "db_user=$DB_USER"
  echo "weknora_version=$(get_env WEKNORA_VERSION)"
  echo "pg_image=$(docker inspect -f '{{.Config.Image}}' "$PG_CONTAINER" 2>/dev/null || echo '?')"
  echo "qdrant_image=$(docker inspect -f '{{.Config.Image}}' "$QDRANT_CONTAINER" 2>/dev/null || echo '?')"
} > "$WORK/manifest.txt"

BUNDLE="$OUT_DIR/weknora-backup-$STAMP.tar.gz"
tar czf "$BUNDLE" -C "$WORK" db.sql.gz qdrant.tgz data-files.tgz keys.env manifest.txt

log "XONG -> $BUNDLE ($(du -h "$BUNDLE" 2>/dev/null | cut -f1))"
log "⚠ GIỮ BÍ MẬT: gói chứa toàn bộ dữ liệu + khóa mã hóa. Đừng commit lên Git."
