# Deploy WeKnora-Chess lên Hetzner (tự động, build-on-server)

Hướng dẫn này giúp bạn đưa stack cờ vua từ Docker trên PC lên 1 server Hetzner.
Cơ chế: tạo server → dán 1 đoạn cloud-init → server tự cài Docker, clone fork, **build từ
source** và chạy. Update về sau chỉ cần 1 lệnh.

> **Vì sao phải build từ source?** Đây là fork private có code cờ vua tùy biến. Các image
> công khai `wechatopenai/weknora-*` trên Docker Hub KHÔNG có tính năng cờ vua, nên không
> dùng `scripts/cloud-image/prepare.sh` (kéo image vanilla) được.

## Kiến trúc

```
GitHub (fork private)  --PAT read-only-->  Hetzner CX32/CX33 (Ubuntu 24.04 LTS)
                                            ├─ cloud-init: swap 4GB + Docker + clone
                                            └─ server-bootstrap.sh:
                                                 ├─ build frontend dist (container node)
                                                 ├─ sinh .env (secret ngẫu nhiên, vi-VN, qdrant)
                                                 ├─ docker compose (3 file) --profile qdrant up --build
                                                 └─ cài systemd (tự chạy lại khi reboot)

Container chạy: frontend(80) · app(8080) · docreader · postgres(pgvector) · redis · qdrant · chess-engine
```

## Yêu cầu server

- **Ubuntu 24.04 LTS** (ổn định với `get.docker.com`; tránh 26.04 nếu không cần).
- **~8GB RAM** (CX32/CX33). cloud-init tự thêm **swap 4GB** → đủ ~12GB để build app (Go/CGO).
- ~40GB disk trở lên (image + volume).

---

## Bước 1 — Tạo GitHub PAT (read-only)

1. GitHub → **Settings → Developer settings → Personal access tokens → Fine-grained tokens → Generate new token**.
2. **Repository access**: Only select repositories → chọn `WeKnora_chess`.
3. **Permissions → Repository permissions → Contents: Read-only**.
4. Generate → **copy token** (chỉ hiện 1 lần).

> Token này chỉ để clone/pull. Có thể thu hồi/đổi bất cứ lúc nào trên GitHub.

## Bước 2 — Thêm SSH key vào Hetzner

Trong Hetzner Console → server creation, mục **SSH keys** → thêm public key của bạn
(`~/.ssh/id_ed25519.pub`). Cần có để `ssh root@<IP>` về sau (xem log, redeploy).

## Bước 3 — Tạo Cloud Firewall (BẮT BUỘC)

compose publish ra host các cổng `80` (frontend), `8080` (app), `6333/6334` (qdrant).
Trên IP public, **8080 và qdrant sẽ lộ ra Internet** nếu không chặn.

Hetzner Console → **Firewalls → Create Firewall**:
- Inbound rules — chỉ cho phép:
  - TCP **22** (SSH)
  - TCP **80** (HTTP)  *(thêm TCP **443** khi bạn làm HTTPS)*
- Apply to → server bạn sắp tạo (hoặc gắn sau khi tạo).

> postgres/redis vốn không publish nên đã an toàn; firewall lo phần 8080/qdrant.

## Bước 4 — Tạo server + dán cloud-init

1. Mở [`scripts/deploy/hetzner-cloud-init.yaml`](../../scripts/deploy/hetzner-cloud-init.yaml),
   copy toàn bộ nội dung.
2. Sửa đúng 1 dòng: thay `__GITHUB_PAT__` bằng token ở Bước 1.
   *(GITHUB_USER / REPO_URL đã điền sẵn cho `covuaduongsinh/WeKnora_chess`.)*
3. Hetzner Console → **Create server**:
   - Location: tùy (vd Falkenstein/Nuremberg).
   - Image: **Ubuntu 24.04**.
   - Type: **CX32/CX33 (~8GB)**.
   - **Cloud config** (cuộn xuống phần "Cloud config" / user data): dán nội dung đã sửa.
   - SSH keys: chọn key Bước 2. Firewall: chọn firewall Bước 3.
   - **Create & Buy now**.

## Bước 5 — Chờ & theo dõi

Lần đầu build mất ~20–40 phút (Go/CGO + docreader + Arasan C++).

```bash
ssh root@<IP> tail -f /var/log/weknora-bootstrap.log
```

Thấy dòng `[bootstrap] HOÀN TẤT.` là xong.

---

## Kiểm tra (verification)

```bash
ssh root@<IP>
cd /opt/WeKnora_chess
bash scripts/deploy/dc.sh ps         # 7 service: frontend/app/docreader/postgres/redis/qdrant/chess-engine
curl -f http://localhost:8080/health # app sống
cat /root/weknora-credentials.txt    # secret đã sinh
```

Trên trình duyệt:
1. Mở `http://<IP>` → **đăng ký admin** (người đầu tiên = admin, làm ngay).
2. **Settings → Models**: nhập API key LLM / embedding / rerank (OpenAI/DeepSeek/Gemini…).
3. Test cờ vua: tạo hội thoại với **agent HLV Cờ vua**, hỏi 1 thế cờ → engine `chess-engine`
   phải phản hồi (xác nhận overlay `docker-compose.chess.yml` + `WEKNORA_CHESS_*` hoạt động).
4. (Khuyến nghị) Khóa đăng ký: sửa `DISABLE_REGISTRATION=true` trong `/opt/WeKnora_chess/.env`
   rồi `bash scripts/deploy/dc.sh up -d`.

Reboot thử (`reboot`) → `weknora-chess.service` tự đưa stack lên lại.

---

## Cập nhật khi có code mới

Trên PC: `git push origin main`. Trên server:

```bash
ssh root@<IP> "bash /opt/WeKnora_chess/scripts/deploy/redeploy.sh"
```

`redeploy.sh` = `git pull` → build lại frontend dist → `compose up -d --build` → dọn image cũ.
**Không** đụng `.env` và volume (postgres/qdrant/data-files giữ nguyên dữ liệu).

## Lệnh vận hành thường dùng

Tất cả qua wrapper [`scripts/deploy/dc.sh`](../../scripts/deploy/dc.sh) (đã gói sẵn 3 file
compose + `--profile qdrant`):

```bash
cd /opt/WeKnora_chess
bash scripts/deploy/dc.sh ps          # trạng thái
bash scripts/deploy/dc.sh logs -f app # log backend
bash scripts/deploy/dc.sh restart app # khởi động lại 1 service
bash scripts/deploy/dc.sh down        # tắt stack
bash scripts/deploy/dc.sh up -d       # bật stack
```

---

## Xử lý sự cố

- **Build app chết vì hết RAM (OOM / "signal: killed")**: tăng swap lên 8GB —
  `swapoff /swapfile; fallocate -l 8G /swapfile; chmod 600 /swapfile; mkswap /swapfile; swapon /swapfile`
  rồi `bash scripts/deploy/redeploy.sh`. Hoặc tạm resize server lên 16GB để build, sau đó resize
  xuống (volume giữ nguyên).
- **`git pull`/clone đòi mật khẩu**: PAT sai hoặc hết hạn. Sửa `/root/.git-credentials`
  (dòng `https://covuaduongsinh:<PAT>@github.com`), `chmod 600`.
- **Không vào được `http://<IP>`**: kiểm tra firewall đã mở 80; `bash scripts/deploy/dc.sh ps`
  xem `frontend` healthy chưa.
- **cloud-init không chạy**: `cat /var/log/cloud-init-output.log` và `/var/log/weknora-bootstrap.log`.

## Tùy chọn nâng cao (không thuộc MVP)

- **HTTPS + domain**: trỏ A record về IP, mở 443 trên firewall, đặt Caddy (tự xin Let's Encrypt)
  reverse-proxy trước `frontend:80`.
- **Engine cờ mạnh/nhanh hơn**: CPU Hetzner CX (Intel) có AVX2 → đổi `ARASAN_BUILD: avx2`
  trong [`docker-compose.chess.yml`](../../docker-compose.chess.yml) rồi rebuild `chess-engine`.
- **CI/CD push-to-deploy**: build image cờ vua trong GitHub Actions đẩy lên GHCR, server chỉ
  `pull` → mỗi lần push tự deploy, server có thể nhỏ hơn (chỉ chạy, không build).
