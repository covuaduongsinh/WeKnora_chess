# Backup / Restore & Di trú dữ liệu WeKnora-Chess (local → VPS)

Hướng dẫn sao lưu **toàn bộ** dữ liệu và khôi phục/di trú sang máy khác. Dùng để:
- Chuyển hết dữ liệu từ PC local lên VPS (lần đầu).
- Backup định kỳ trên VPS để phòng hỏng hóc.

## Gói backup gồm những gì

Một file duy nhất `weknora-backup-<ngày>.tar.gz` chứa **mọi loại dữ liệu**:

| Thành phần | Nội dung |
|---|---|
| `db.sql.gz` | PostgreSQL: agent, **cấu hình model**, **khóa học + bài học**, **kho ván đấu**, **ngân hàng bài tập**, tri thức (metadata), trợ lý AI, user/tenant, hội thoại, wiki… |
| `qdrant.tgz` | Vector tri thức (để tìm kiếm/hỏi đáp hoạt động) |
| `data-files.tgz` | File tài liệu đã upload |
| `keys.env` | 4 khóa mã hóa (`SYSTEM_AES_KEY`…) — **bắt buộc** để key model giải mã được sau khi chuyển |

> ⚠️ Gói này chứa **toàn bộ dữ liệu + khóa mã hóa** → giữ bí mật, không đưa lên GitHub.
> (`.gitignore` đã chặn `weknora-backup-*.tar.gz`.)

---

## A. Sao lưu trên PC local

1. **Bật lại Docker** trên PC (nếu đã tắt) và đảm bảo stack cờ vua đang chạy.
   Kiểm tra nhanh trong Terminal (VS Code):
   ```bash
   docker ps --format '{{.Names}}'
   ```
   Phải thấy `WeKnora-postgres`, `WeKnora-qdrant`, `WeKnora-app`.

2. Chạy backup (ở thư mục gốc dự án):
   ```bash
   bash scripts/deploy/backup.sh
   ```
   Xong sẽ in ra đường dẫn gói, ví dụ `weknora-backup-20260628-101500.tar.gz`.

> Muốn lưu ra chỗ khác: `OUT_DIR=D:/backups bash scripts/deploy/backup.sh`

## B. Chuyển gói lên VPS

Trong Terminal trên PC (thay tên gói cho đúng):
```bash
scp weknora-backup-20260628-101500.tar.gz root@65.109.129.242:/opt/WeKnora_chess/
```

## C. Khôi phục trên VPS (GHI ĐÈ)

1. SSH vào VPS:
   ```bash
   ssh root@65.109.129.242
   cd /opt/WeKnora_chess
   ```
2. Chạy restore (thay tên gói cho đúng):
   ```bash
   bash scripts/deploy/restore.sh weknora-backup-20260628-101500.tar.gz --yes
   ```
   Script sẽ: chụp nhanh DB cũ (để rollback) → đồng bộ khóa mã hóa vào `.env` →
   ghi đè DB + Qdrant + data-files → bật lại toàn bộ stack.

> ⚠️ **GHI ĐÈ**: dữ liệu/tài khoản đang có trên VPS sẽ bị thay bằng dữ liệu local.
> Sau khi xong, **đăng nhập VPS bằng tài khoản LOCAL** (email/mật khẩu bạn dùng ở máy local).

## D. Kiểm tra sau khi khôi phục

Mở `http://65.109.129.242` và xác nhận:
- [ ] Đăng nhập được bằng **tài khoản local** → DB đã chuyển.
- [ ] **Kho tri thức**: thấy đủ; mở 1 tài liệu cũ xem được → data-files OK.
- [ ] Hỏi đáp trên tri thức cũ ra kết quả → Qdrant OK.
- [ ] **Settings → Models**: chat thử có phản hồi → **giải mã key model OK** (khóa đồng bộ đúng).
- [ ] **Quản lý cờ vua**: khóa học / kho ván đấu / ngân hàng bài tập hiện đủ.

Lệnh kiểm tra container: `bash scripts/deploy/dc.sh ps` (7 service `Up`).

---

## Backup định kỳ trên VPS (khuyến nghị)

Chạy ngay trên VPS để có bản sao an toàn:
```bash
cd /opt/WeKnora_chess
OUT_DIR=/opt/weknora-backups bash scripts/deploy/backup.sh
```
Muốn tự động hằng ngày (vd 2h sáng), thêm cron:
```bash
mkdir -p /opt/weknora-backups
( crontab -l 2>/dev/null; echo "0 2 * * * cd /opt/WeKnora_chess && OUT_DIR=/opt/weknora-backups bash scripts/deploy/backup.sh >> /var/log/weknora-backup.log 2>&1" ) | crontab -
```
> Nhớ dọn bớt gói cũ kẻo đầy đĩa (mỗi gói gồm cả tri thức + file).

## Khôi phục khi gặp sự cố (disaster recovery)

Trên server bất kỳ đã cài stack (xem [hetzner.md](hetzner.md)), copy gói backup về rồi:
```bash
bash scripts/deploy/restore.sh <goi.tar.gz> --yes
```

## Sự cố thường gặp

- **`container ... chưa chạy`**: bật stack trước — `bash scripts/deploy/dc.sh up -d`.
- **Sau restore, Settings → Models báo lỗi key / chat không chạy**: khóa mã hóa chưa đồng bộ.
  Kiểm tra `keys.env` trong gói có giá trị, và `.env` trên VPS đã được cập nhật 4 khóa
  (`SYSTEM_AES_KEY`, `TENANT_AES_KEY`, `CRYPTO_MASTER_KEY`, `CRYPTO_SALT`), rồi
  `bash scripts/deploy/dc.sh up -d`.
- **Lỡ restore nhầm**: dùng bản chụp nhanh `pre-restore-*.sql.gz` mà `restore.sh` đã tạo trong
  `/opt/WeKnora_chess/` để nạp lại DB cũ.
