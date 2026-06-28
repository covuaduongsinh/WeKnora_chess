# Quy trình phát triển: local-first, tránh "VPS mới hơn local"

## Vấn đề muốn tránh

Có **2 đường deploy độc lập**:

- `git push` lên `main` → CI/CD tự build image + tự deploy **VPS** (xem [cicd.md](cicd.md)). 0 thao tác tay.
- Cập nhật **stack LOCAL** là **bước riêng** — không tự chạy khi bạn sửa code.

Nếu bạn `git push` mà **quên build lại local**, container local sẽ đứng ở bản cũ trong
khi VPS nhảy lên bản mới → "VPS mới hơn local", và tệ hơn: **bạn deploy lên production
(VPS) code chưa từng chạy thử ở local**.

**Bất biến cần giữ:** *local ≥ VPS khi đang phát triển* — test ở local trước, rồi mới push.

## Vòng lặp chuẩn

```
sửa code  →  make local-deploy  →  test ở http://localhost  →  git push  →  (CI deploy VPS)
```

1. **`make local-deploy`** — build app+frontend **từ nguồn** (working tree), recreate
   `app`+`frontend`, restart frontend (chống nginx stale-IP), rồi in `schema_migrations`,
   `/healthz`, và commit đang chạy. (OOM build đã giảm thiểu bằng `.wslconfig` 12GB+swap.)
2. Mở `http://localhost` (Ctrl+Shift+R nếu vừa đổi frontend) và **kiểm thử thật**.
3. `git push` → CI build + deploy VPS. (Tuỳ chọn) verify `https://weknora.covuaduongsinh.com`.

### Mode nhanh (đồng bộ với VPS, không build)
Khi chỉ muốn local **giống hệt bản đã push** (không cần test code chưa commit):
```
make local-deploy ARGS=--ghcr          # pull ghcr …:latest
make local-deploy ARGS="--ghcr <sha>"  # pull đúng tag/sha CI đã build
```
(GHCR private → `docker login ghcr.io` một lần; mật khẩu dùng `gh auth token`.)

## Tăng tốc build (quan trọng)

Build app lần ĐẦU lâu (~10–15') vì final stage cài nhiều gói apt + biên dịch Go.
Các lần SAU phải nhanh nhờ cache — đảm bảo:

- **Chỉ build phần đã đổi:**
  ```
  make local-deploy ARGS=--app-only       # chỉ sửa backend (Go)
  make local-deploy ARGS=--frontend-only  # chỉ sửa UI (Vue)
  ```
  Tránh tốn ~3–4' build frontend khi không động tới nó (và ngược lại).
- **Layer apt nặng được cache qua các commit:** nhãn commit
  (`LABEL org.opencontainers.image.revision`) đặt ở **CUỐI** `docker/Dockerfile.app`,
  KHÔNG đặt trước `apt-get install`. Nếu đặt trước, mỗi commit (đổi `COMMIT_ID_ARG`)
  sẽ **bust cache** toàn bộ layer apt → build nào cũng cài lại 12'. (Đã sửa; đừng
  chuyển nhãn lên trên.) Lợi cho cả CI (cache GHA layer).
- **Go build tăng tiến:** builder stage đã có `--mount=type=cache` cho
  `/go/pkg/mod` + `/root/.cache/go-build` → chỉ biên dịch package đã đổi.
- **Đừng `docker system prune` bừa** giữa các phiên — sẽ xoá cache layer/Go, build
  sau lại chậm như lần đầu.

## Báo "lệch" & cảnh báo khi push

- **Kiểm thủ công bất cứ lúc nào:**
  ```
  make local-status
  ```
  In `IN SYNC ✅` hoặc `LOCAL BEHIND ⚠ (running=<x> source=<y>)`. (So `COMMIT_ID` bake
  trong image app đang chạy với `git rev-parse --short HEAD`.)

- **Cảnh báo mềm khi `git push`** — bật hook tracked trong repo (one-time):
  ```
  git config core.hooksPath scripts/dev/githooks
  ```
  Từ đó, nếu bạn push một bản mà stack local chưa build/test, hook in cảnh báo nhưng
  **vẫn cho push**. Muốn **chặn cứng**: đặt `WEKNORA_STRICT_PREPUSH=1`.

## Ghi chú
- `make local-deploy` chỉ recreate `app`+`frontend` (`--no-deps`) — không động tới
  engine/docreader/redis/postgres/qdrant đang chạy.
- Migration nằm trong image app, tự chạy lúc boot; lỗi migration chỉ WARN → kiểm
  `schema_migrations` SAU khi app healthy (local-deploy đã in sẵn).
