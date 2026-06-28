#!/usr/bin/env bash
# server-bootstrap.sh — dựng WeKnora-Chess TỪ SOURCE trên 1 server Linux sạch (Hetzner).
#
# Idempotent: chạy lại được. Được gọi bởi cloud-init ở first boot, hoặc chạy tay.
# Giả định: Docker đã được cài (cloud-init lo) và repo đã được clone về.
#
# Các bước:
#   1) Build frontend dist qua container node (frontend/Dockerfile copy sẵn dist/,
#      nên phải dựng dist TRƯỚC khi `compose build`). Không cần Node trên host.
#   2) Sinh .env với secret ngẫu nhiên + cấu hình khớp stack cờ vua (vi-VN, qdrant)
#      nếu .env chưa có. Đã có thì GIỮ NGUYÊN (không ghi đè secret -> tránh lệch
#      mật khẩu đã nằm trong volume postgres).
#   3) docker compose (3 file) --profile qdrant up -d --build
#   4) Cài systemd unit weknora-chess.service để stack tự lên lại sau reboot.
#
# Biến môi trường tùy chọn:
#   FRONTEND_NODE_IMAGE   image node để build dist (mặc định node:22-bookworm)
#   NODE_HEAP_MB          giới hạn heap Node khi build frontend, MB (mặc định 4096).
#                         Tăng nếu build frontend bị "Aborted (core dumped)" / hết RAM.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env"
ENV_TEMPLATE="${PROJECT_ROOT}/.env.example"
CRED_FILE="/root/weknora-credentials.txt"
FRONTEND_NODE_IMAGE="${FRONTEND_NODE_IMAGE:-node:22-bookworm}"

log() { printf '[bootstrap] %s\n' "$*"; }

cd "${PROJECT_ROOT}"

# --- 0. kiểm tra docker ---
if ! command -v docker >/dev/null 2>&1; then
  echo "[bootstrap] ERROR: chưa có docker (cloud-init lẽ ra đã cài). Cài Docker rồi chạy lại." >&2
  exit 1
fi

# --- 1. build frontend dist (qua container node, không cần Node trên host) ---
log "1/4 build frontend dist qua container ${FRONTEND_NODE_IMAGE} ..."
COMMIT_ID="$(git -C "${PROJECT_ROOT}" rev-parse --short HEAD 2>/dev/null || echo unknown)"
docker run --rm \
  -e VITE_IS_DOCKER=true \
  -e VITE_FRONTEND_COMMIT="${COMMIT_ID}" \
  -e NODE_OPTIONS="--max-old-space-size=${NODE_HEAP_MB:-4096}" \
  -v "${PROJECT_ROOT}/frontend":/app \
  -w /app \
  "${FRONTEND_NODE_IMAGE}" \
  sh -lc "npm ci && npm run build"

# --- 2. sinh .env nếu chưa có ---
# gen32: chuỗi ngẫu nhiên ĐÚNG 32 byte (AES-256 key). genpw: mật khẩu 24 ký tự
# an toàn cho URL/sed (chỉ [A-Za-z0-9]). Tắt pipefail trong subshell vì head đóng
# stdin sớm khiến tr nhận SIGPIPE (idiom lấy từ scripts/cloud-image/firstboot.sh).
gen32() ( set +o pipefail; LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32 )
genpw() ( set +o pipefail; LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 24 )

replace() {
  local key="$1" val="$2"
  if grep -qE "^${key}=" "${ENV_FILE}"; then
    sed -i "s|^${key}=.*|${key}=${val}|" "${ENV_FILE}"
  else
    printf '%s=%s\n' "${key}" "${val}" >>"${ENV_FILE}"
  fi
}

if [[ -f "${ENV_FILE}" ]]; then
  log "2/4 .env đã tồn tại — giữ nguyên (không sinh lại secret)."
else
  log "2/4 tạo .env từ .env.example + sinh secret ngẫu nhiên ..."
  cp "${ENV_TEMPLATE}" "${ENV_FILE}"

  DB_PWD=$(genpw); REDIS_PWD=$(genpw); JWT=$(genpw)$(genpw)
  SYS_AES=$(gen32); TENANT_AES=$(gen32)
  CRYPTO_KEY=$(gen32); CRYPTO_SALT=$(genpw)

  replace DB_PASSWORD       "${DB_PWD}"
  replace REDIS_PASSWORD    "${REDIS_PWD}"
  replace JWT_SECRET        "${JWT}"
  replace SYSTEM_AES_KEY    "${SYS_AES}"
  replace TENANT_AES_KEY    "${TENANT_AES}"
  replace CRYPTO_MASTER_KEY "${CRYPTO_KEY}"
  replace CRYPTO_SALT       "${CRYPTO_SALT}"

  # cấu hình khớp stack cờ vua đang chạy trên PC
  replace WEKNORA_LANGUAGE     "vi-VN"
  replace RETRIEVE_DRIVER      "qdrant"
  replace STORAGE_TYPE         "local"
  replace GIN_MODE             "release"
  replace DISABLE_REGISTRATION "false"

  # Hetzner ở Đức: bỏ mirror Trung Quốc (apk/apt) cho nhanh & tránh timeout
  replace APK_MIRROR_ARG ""
  replace APT_MIRROR     ""

  umask 077
  PUB_IP=$(curl -fsS --max-time 5 https://ifconfig.me 2>/dev/null \
    || curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null \
    || hostname -I | awk '{print $1}')
  cat >"${CRED_FILE}" <<INFO
========================================
  WeKnora-Chess đã khởi tạo
  Thời điểm: $(date -Iseconds)
========================================

Truy cập : http://${PUB_IP}
NGƯỜI ĐĂNG KÝ ĐẦU TIÊN sẽ thành admin — hãy đăng ký NGAY kẻo bị chiếm chỗ.

Secret ngẫu nhiên (đã ghi vào ${ENV_FILE}, chỉ root đọc được):
  DB_PASSWORD       = ${DB_PWD}
  REDIS_PASSWORD    = ${REDIS_PWD}
  JWT_SECRET        = ${JWT}
  SYSTEM_AES_KEY    = ${SYS_AES}
  TENANT_AES_KEY    = ${TENANT_AES}
  CRYPTO_MASTER_KEY = ${CRYPTO_KEY}
  CRYPTO_SALT       = ${CRYPTO_SALT}

Lưu ý bảo mật:
  - KHÔNG mở 5432/6379/6333/6334/8080 ra Internet. Chỉ mở 80 (+443) qua Hetzner Cloud Firewall.
  - Sau khi đăng ký admin xong, đặt DISABLE_REGISTRATION=true trong ${ENV_FILE} rồi chạy:
        bash ${SCRIPT_DIR}/dc.sh up -d
INFO
  chmod 0600 "${CRED_FILE}"
  log "    secret đã ghi ra ${CRED_FILE}"
fi

# --- 3. build từng image rồi up toàn stack ---
# Build TUẦN TỰ (không song song) để peak RAM thấp — trên box 8GB+swap, build
# đồng thời Go/CGO + Python + C++ dễ OOM. Image hạ tầng (postgres/redis/qdrant)
# do `up` tự pull.
log "3/4 build từng image (tuần tự, lần đầu LÂU: Go/CGO + docreader + Arasan C++) ..."
for svc in app docreader chess-engine frontend; do
  log "    build ${svc} ..."
  bash "${SCRIPT_DIR}/dc.sh" build "${svc}"
done
log "    khởi động toàn stack ..."
bash "${SCRIPT_DIR}/dc.sh" up -d

# --- 4. cài systemd unit để tự chạy lại khi reboot ---
log "4/4 cài systemd unit weknora-chess.service ..."
if command -v systemctl >/dev/null 2>&1; then
  UNIT_SRC="${SCRIPT_DIR}/weknora-chess.service"
  UNIT_DST="/etc/systemd/system/weknora-chess.service"
  sed "s|@PROJECT_ROOT@|${PROJECT_ROOT}|g" "${UNIT_SRC}" >"${UNIT_DST}"
  systemctl daemon-reload
  systemctl enable weknora-chess.service
  log "    đã enable weknora-chess.service (stack tự lên lại sau reboot)."
else
  log "    WARNING: không có systemctl — bỏ qua bước systemd (stack vẫn đang chạy)."
fi

cat <<DONE

[bootstrap] HOÀN TẤT.
  - Kiểm tra:   bash ${SCRIPT_DIR}/dc.sh ps
  - Mở trình duyệt: http://<IP-công-khai>  -> đăng ký admin
  - Vào Settings -> Models nhập API key LLM/embedding/rerank (cloud).
  - Secret xem ở: ${CRED_FILE}
DONE
