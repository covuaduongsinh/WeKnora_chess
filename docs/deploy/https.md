# HTTPS + tên miền (Caddy + Let's Encrypt)

Đặt **Caddy** trước frontend làm reverse-proxy; Caddy **tự xin & tự gia hạn** chứng chỉ
Let's Encrypt → hết cảnh báo "Not secure", truy cập bằng `https://ten-mien` thay cho IP.

File: [docker-compose.caddy.yml](../../docker-compose.caddy.yml) + [docker/caddy/Caddyfile](../../docker/caddy/Caddyfile).

## Yêu cầu (làm trước)

1. **Có tên miền / subdomain** (vd `chess.tencongty.com`).
2. **DNS**: tạo **A record** trỏ tên miền về **65.109.129.242**
   *(tùy chọn: AAAA record → `2a01:4f9:c012:237e::1` cho IPv6)*. Chờ vài phút cho DNS lan.
3. **Hetzner Cloud Firewall**: mở thêm **443 (TCP)** *(UDP 443 tùy chọn cho HTTP/3)* — hiện chỉ mở 22 + 80.
   Console → Firewalls → `weknora-fw` → Inbound → Add rule: TCP 443, source Any IPv4/IPv6.

## Bật HTTPS

1. Đặt tên miền vào `.env` trên VPS:
   ```bash
   echo "DOMAIN=ten-mien.com" >> /opt/WeKnora_chess/.env
   ```
   *(hoặc sửa dòng `DOMAIN=` nếu đã có)*
2. Khởi động kèm overlay Caddy (Caddy chiếm 80/443, frontend lùi vào trong):
   ```bash
   cd /opt/WeKnora_chess
   docker compose -f docker-compose.yml -f docker-compose.override.yml \
     -f docker-compose.chess.yml -f docker-compose.ghcr.yml \
     -f docker-compose.caddy.yml --profile qdrant up -d
   ```
3. Mở `https://ten-mien.com` — lần đầu Caddy xin chứng chỉ mất ~10–30 giây. Xem log nếu cần:
   ```bash
   docker logs -f WeKnora-caddy
   ```

## Giữ HTTPS sau mỗi lần CI deploy

Để Caddy luôn nằm trong stack (kể cả khi CI tự deploy), thêm `-f docker-compose.caddy.yml` vào
danh sách file trong [scripts/deploy/dc.sh](../../scripts/deploy/dc.sh) và
[scripts/deploy/pull-deploy.sh](../../scripts/deploy/pull-deploy.sh). (Chỉ làm SAU khi HTTPS đã chạy ổn,
và `DOMAIN` đã có trong `.env` — nếu không Caddy sẽ không xin được chứng chỉ.)

## Sự cố thường gặp
- **Không ra HTTPS / log Caddy báo lỗi ACME**: DNS chưa trỏ đúng IP, hoặc firewall chưa mở 443/80.
  Kiểm tra: `dig +short ten-mien.com` phải ra `65.109.129.242`.
- **`bind: address already in use` cổng 80**: còn container khác giữ 80 — đảm bảo dùng overlay caddy
  (đã `!reset` cổng 80 của frontend) và không chạy đồng thời stack không-caddy.
- **Đổi tên miền**: sửa `DOMAIN` trong `.env` → `... up -d` lại; Caddy xin chứng chỉ mới tự động.
