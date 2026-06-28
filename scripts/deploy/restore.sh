#!/usr/bin/env bash
# restore.sh — Khôi phục / di trú TOÀN BỘ dữ liệu từ 1 gói backup do backup.sh tạo.
#
# ⚠ GHI ĐÈ: xóa sạch DB / Qdrant / data-files hiện tại trên máy này rồi nạp dữ liệu
# từ gói. Chạy trên máy ĐÍCH (vd VPS) nơi stack WeKnora-Chess đang chạy.
#
# Dùng:
#   bash scripts/deploy/restore.sh <goi.tar.gz> --yes
#
# Tùy chọn:
#   --yes | -y      bỏ qua hỏi xác nhận (BẮT BUỘC khi chạy qua SSH không tương tác)
#   --no-snapshot   không tự backup DB hiện tại trước khi ghi đè (mặc định CÓ snapshot)
#
# Biến tùy chọn: PG_CONTAINER, QDRANT_CONTAINER, APP_CONTAINER
set -euo pipefail

export MSYS_NO_PATHCONV=1

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env"

PG_CONTAINER="${PG_CONTAINER:-WeKnora-postgres}"
QDRANT_CONTAINER="${QDRANT_CONTAINER:-WeKnora-qdrant}"
APP_CONTAINER="${APP_CONTAINER:-WeKnora-app}"

BUNDLE=""; ASSUME_YES=0; SNAPSHOT=1
for a in "$@"; do
  case "$a" in
    --yes|-y)      ASSUME_YES=1 ;;
    --no-snapshot) SNAPSHOT=0 ;;
    *.tar.gz)      BUNDLE="$a" ;;
    *) echo "tham số lạ: $a" >&2; exit 1 ;;
  esac
done

log() { printf '[restore] %s\n' "$*"; }
die() { printf '[restore] ERROR: %s\n' "$*" >&2; exit 1; }

[ -n "$BUNDLE" ] || die "thiếu gói backup. Vd: restore.sh weknora-backup-YYYYmmdd-HHMMSS.tar.gz --yes"
[ -f "$BUNDLE" ] || die "không thấy gói: $BUNDLE"
command -v docker >/dev/null 2>&1 || die "không thấy docker"
[ -f "$ENV_FILE" ] || die "không thấy .env ở $ENV_FILE"

get_env() { grep -E "^$1=" "$ENV_FILE" | head -1 | cut -d= -f2- ; }
DB_USER="$(get_env DB_USER)"; DB_USER="${DB_USER:-postgres}"
DB_NAME="$(get_env DB_NAME)"; DB_NAME="${DB_NAME:-WeKnora}"

running() { [ "$(docker inspect -f '{{.State.Running}}' "$1" 2>/dev/null || echo false)" = "true" ]; }
running "$PG_CONTAINER" || die "container $PG_CONTAINER chưa chạy — bật stack trước (dc.sh up -d)"

if [ "$ASSUME_YES" != "1" ]; then
  echo "⚠ CẢNH BÁO: sẽ GHI ĐÈ toàn bộ DB/Qdrant/file hiện tại trên máy này bằng dữ liệu trong gói."
  printf "Gõ 'yes' để tiếp tục: "
  read -r ans; [ "$ans" = "yes" ] || die "đã hủy"
fi

WORK="$(mktemp -d)"; trap 'rm -rf "$WORK"' EXIT
log "giải nén gói ..."
tar xzf "$BUNDLE" -C "$WORK"
for f in db.sql.gz qdrant.tgz data-files.tgz keys.env; do
  [ -f "$WORK/$f" ] || die "gói thiếu $f (không phải gói backup hợp lệ?)"
done

# 0) snapshot DB hiện tại để lỡ sai còn quay lại
if [ "$SNAPSHOT" = "1" ]; then
  SNAP="${PROJECT_ROOT}/pre-restore-$(date +%Y%m%d-%H%M%S).sql.gz"
  log "0/6 chụp nhanh DB hiện tại -> $SNAP"
  docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" --clean --if-exists \
    | gzip > "$SNAP" 2>/dev/null || log "   (cảnh báo: snapshot thất bại — vẫn tiếp tục)"
fi

# 1) đồng bộ 4 khóa mã hóa vào .env (nếu không, key model migrate sẽ KHÔNG giải mã được)
log "1/6 đồng bộ khóa mã hóa vào .env ..."
set_env() {
  local key="$1" val="$2"
  if grep -qE "^${key}=" "$ENV_FILE"; then
    sed -i "s|^${key}=.*|${key}=${val}|" "$ENV_FILE"
  else
    printf '%s=%s\n' "$key" "$val" >> "$ENV_FILE"
  fi
}
while IFS='=' read -r k v; do
  [ -n "$k" ] || continue
  set_env "$k" "$v"
done < "$WORK/keys.env"

# 2) dừng các service ghi vào DB để giải phóng kết nối
log "2/6 dừng app/docreader/frontend/mcp ..."
docker stop "$APP_CONTAINER" WeKnora-docreader WeKnora-frontend WeKnora-mcp >/dev/null 2>&1 || true

# 3) restore DB: drop & tạo lại DB cho sạch rồi nạp dump
log "3/6 restore PostgreSQL (drop & tạo lại $DB_NAME) ..."
docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${DB_NAME}' AND pid<>pg_backend_pid();" >/dev/null
docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS \"${DB_NAME}\";" >/dev/null
docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"${DB_NAME}\";" >/dev/null
gunzip -c "$WORK/db.sql.gz" | docker exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=0 >/dev/null

# 4) restore Qdrant — PHẢI dừng qdrant trước khi ghi đè storage (tránh hỏng dữ liệu)
log "4/6 restore Qdrant ..."
docker stop "$QDRANT_CONTAINER" >/dev/null 2>&1 || true
docker run --rm -i --volumes-from "$QDRANT_CONTAINER" alpine \
  sh -c 'find /qdrant/storage -mindepth 1 -delete 2>/dev/null; tar xzf - -C /qdrant/storage' < "$WORK/qdrant.tgz"
docker start "$QDRANT_CONTAINER" >/dev/null 2>&1 || true

# 5) restore data-files
log "5/6 restore data-files ..."
docker run --rm -i --volumes-from "$APP_CONTAINER" alpine \
  sh -c 'find /data/files -mindepth 1 -delete 2>/dev/null; tar xzf - -C /data/files' < "$WORK/data-files.tgz"

# 6) bật lại toàn stack (app tái tạo container -> nạp .env khóa mới)
log "6/6 bật lại toàn bộ stack ..."
bash "${SCRIPT_DIR}/dc.sh" up -d

log "XONG. Mở web và ĐĂNG NHẬP BẰNG TÀI KHOẢN LOCAL để kiểm tra."
[ "$SNAPSHOT" = "1" ] && log "Bản DB cũ đã lưu ở: ${SNAP:-(none)} (xóa khi chắc chắn ổn)."
