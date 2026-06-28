# CI/CD: push → GitHub Actions build → GHCR → VPS tự cập nhật

Thay vì build trên VPS (chậm 15–20'), giờ mỗi lần `git push` lên `main`:
1. GitHub Actions build 4 image cờ vua trên runner mạnh của GitHub.
2. Đẩy lên **GHCR** (`ghcr.io/covuaduongsinh/weknora-chess-*`).
3. SSH vào VPS → `git pull` (lấy compose/scripts mới) → **pull image + restart** (vài giây, KHÔNG biên dịch).

Workflow: [.github/workflows/cicd-deploy.yml](../../.github/workflows/cicd-deploy.yml).
Build-on-VPS cũ (`redeploy.sh`) vẫn giữ làm phương án dự phòng.

> ⚠ **Trước khi push, hãy test ở LOCAL** (`git push` = deploy VPS/production). Xem
> quy trình local-first + `make local-deploy` / `make local-status` ở
> [dev-workflow.md](dev-workflow.md) để tránh "VPS mới hơn local".

---

## Cấu hình MỘT LẦN (làm trước khi push)

### 1. Tạo SSH key riêng cho deploy + nạp vào GitHub Secrets
Để Actions SSH được vào VPS. Trên PC:
```bash
ssh-keygen -t ed25519 -f deploy_key -N "" -C "github-actions-deploy"
```
- Thêm **public key** vào VPS:
  ```bash
  ssh root@65.109.129.242 "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys" < deploy_key.pub
  ```
- Vào **GitHub repo → Settings → Secrets and variables → Actions → New repository secret**, tạo 3 secret:
  | Secret | Giá trị |
  |---|---|
  | `VPS_HOST` | `65.109.129.242` |
  | `VPS_USER` | `root` |
  | `VPS_SSH_KEY` | private key **mã hóa base64 1 dòng**: trên VPS chạy `base64 -w0 ~/deploy_key` rồi copy dòng kết quả. (Workflow tự `base64 -d` lại — tránh lỗi `error in libcrypto` do xuống dòng khi copy.) |
- Xóa `deploy_key` / `deploy_key.pub` trên PC sau khi đã nạp xong (giữ bí mật).

### 2. Cho VPS quyền kéo image private từ GHCR
Image GHCR mặc định **private** → VPS phải đăng nhập 1 lần.
- Tạo GitHub **PAT** (Settings → Developer settings → Tokens) có quyền **`read:packages`**.
- Trên VPS, đăng nhập GHCR (lưu vào `/root/.docker/config.json`, dùng mãi):
  ```bash
  ssh root@65.109.129.242
  echo '<PAT_read_packages>' | docker login ghcr.io -u covuaduongsinh --password-stdin
  ```
> Cách khác (đỡ phải login): vào GitHub → trang **Packages** của mỗi image `weknora-chess-*` → **Package settings → Change visibility → Public**. Khi đó VPS pull khỏi cần đăng nhập. (Đổi lại: ai cũng tải được image đã build — không lộ source, nhưng lộ binary.)

### 3. Quyền ghi package cho Actions
Workflow đã khai báo `permissions: packages: write` nên thường chạy được ngay. Nếu lần build báo lỗi push GHCR `denied`, vào **Settings → Actions → General → Workflow permissions** chọn **Read and write permissions**.

---

## Dùng hằng ngày

Chỉ cần:
```bash
git add -A && git commit -m "..." && git push origin main
```
→ Mở tab **Actions** trên GitHub xem tiến trình. Khi job **deploy** xanh là VPS đã chạy bản mới (`http://65.109.129.242`). Lần build đầu lâu (~10–15' do build từ đầu); các lần sau nhanh hơn nhờ cache.

## Rollback (quay về bản cũ)
Mỗi bản được gắn tag = commit SHA. Trên VPS, chạy lại với SHA cũ:
```bash
ssh root@65.109.129.242 "cd /opt/WeKnora_chess && IMAGE_TAG=<sha_cu> bash scripts/deploy/pull-deploy.sh"
```
(Xem SHA các bản trong lịch sử commit hoặc trang Packages của GHCR.)

## Cập nhật thủ công (không cần push)
Khi image `latest` đã có sẵn trên GHCR mà muốn ép VPS lấy lại:
```bash
ssh root@65.109.129.242 "cd /opt/WeKnora_chess && bash scripts/deploy/pull-deploy.sh"
```

## Xử lý sự cố
- **Job deploy lỗi `Permission denied (publickey)`**: `VPS_SSH_KEY` sai, hoặc public key chưa nằm trong `~/.ssh/authorized_keys` trên VPS.
- **VPS pull lỗi `denied`/`unauthorized`**: chưa `docker login ghcr.io` trên VPS (bước 2), hoặc PAT thiếu `read:packages`.
- **Job build lỗi push `denied`**: bật Read/write permissions (bước 3).
- **Muốn quay lại build-on-VPS**: `ssh root@65.109.129.242 "cd /opt/WeKnora_chess && bash scripts/deploy/redeploy.sh"`.

---

## Thứ tự lần đầu (quan trọng)
1. Làm xong **bước 1, 2, 3** ở trên (secrets + VPS login GHCR).
2. `git push` toàn bộ (gồm các file CI/CD này).
3. Actions chạy: build → push GHCR → SSH vào VPS `git pull` (VPS có `docker-compose.ghcr.yml` + `pull-deploy.sh`) → pull image + up.
