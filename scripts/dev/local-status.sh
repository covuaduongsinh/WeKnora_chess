#!/usr/bin/env bash
# local-status.sh — Báo "lệch" giữa stack LOCAL đang chạy và mã nguồn (git HEAD).
#
# Mục đích: phát hiện sớm tình trạng "app đang chạy ở local cũ hơn working tree"
# (nguyên nhân gây cảm giác "VPS mới hơn local"). So COMMIT_ID đã bake trong image
# app đang chạy với `git rev-parse --short HEAD`.
#
# Dùng:
#   scripts/dev/local-status.sh              # in báo cáo cho người đọc (luôn exit 0)
#   scripts/dev/local-status.sh --porcelain  # in "STATE running source"; exit 0=SYNC 3=BEHIND 2=khác
#
# Biến: WEKNORA_APP_CONTAINER (mặc định WeKnora-app).
set -uo pipefail

PORCELAIN=0
[ "${1:-}" = "--porcelain" ] && PORCELAIN=1

APP_CTR="${WEKNORA_APP_CONTAINER:-WeKnora-app}"

# 7 ký tự đầu cho cả hai phía để khớp dù build local (git --short) hay CI (sha[:7]).
head7="$(git rev-parse --short=7 HEAD 2>/dev/null || echo unknown)"

emit() { # state running source
  if [ "$PORCELAIN" = 1 ]; then echo "$1 ${2:-} ${3:-}"; fi
}

if ! docker inspect "$APP_CTR" >/dev/null 2>&1; then
  emit DOWN "" "$head7"
  [ "$PORCELAIN" = 1 ] && exit 2
  echo "ℹ Stack local chưa chạy (không thấy container $APP_CTR)."
  echo "  → Chạy: make local-deploy"
  exit 0
fi

# Commit nằm ở NHÃN ảnh org.opencontainers.image.revision (gắn ở final stage của
# docker/Dockerfile.app). Đọc nhãn của ĐÚNG image mà container đang chạy.
img_id="$(docker inspect "$APP_CTR" --format '{{.Image}}' 2>/dev/null)"
running_raw="$(docker image inspect "$img_id" --format '{{ index .Config.Labels "org.opencontainers.image.revision" }}' 2>/dev/null | tr -d '\r\n')"
running7="${running_raw:0:7}"

if [ -z "$running_raw" ] || [ "$running_raw" = "unknown" ] || [ "$running_raw" = "<no value>" ]; then
  emit UNKNOWN "$running_raw" "$head7"
  [ "$PORCELAIN" = 1 ] && exit 2
  echo "⚠ Image app đang chạy CHƯA gắn nhãn commit (build trước thay đổi này)."
  echo "  git HEAD (short) : $head7"
  echo "  → Chạy 'make local-deploy' (build lại) để bật báo lệch."
  exit 0
fi

if [ "$running7" = "$head7" ]; then
  emit SYNC "$running7" "$head7"
  [ "$PORCELAIN" = 1 ] && exit 0
  echo "IN SYNC ✅ — app local đang chạy đúng bản nguồn ($running7)."
  exit 0
fi

emit BEHIND "$running7" "$head7"
[ "$PORCELAIN" = 1 ] && exit 3
echo "LOCAL BEHIND ⚠ — app local KHÁC mã nguồn."
echo "  running app COMMIT_ID : $running_raw"
echo "  git HEAD (short)      : $head7"
echo "  → Chạy: make local-deploy   (build từ nguồn + recreate)"
exit 0
