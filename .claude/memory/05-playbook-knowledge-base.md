# 05 — Playbook vận hành & mở rộng lớp cờ

Hướng dẫn dùng/mở rộng những gì ĐÃ CÓ (không dựng lại từ đầu).

## 5.1. Chạy với engine cờ
```bash
# Core + chess-engine (Arasan)
docker compose -f docker-compose.yml -f docker-compose.chess.yml up -d
docker compose ps          # chess-engine phải Up

# Engine mạnh/nhanh hơn nếu CPU hỗ trợ AVX2:
#   sửa docker-compose.chess.yml → ARASAN_BUILD: avx2 → rebuild
```
Overlay đã mount `config/builtin_agents.yaml` từ host → chỉnh agent HLV **không cần rebuild image** (chỉ restart `app`).

## 5.2. Agent "HLV Cờ vua" (đã có)
Định nghĩa: `config/builtin_agents.yaml` → `builtin-chess-coach`.
- System prompt tiếng Việt, hướng dạy học; `kb_selection_mode: none`.
- 6 tool: `chess_analyze_position` (đánh giá thế cờ/FEN), `chess_best_move` (nước tốt nhất), `chess_evaluate_game` (chấm ván/PGN), `chess_explain_move` (giải thích 1 nước), `chess_lookup_opening` (tra khai cuộc), `chess_generate_puzzle` (sinh bài tập) + `thinking`.
- Tool cờ tự render **bàn cờ tương tác**; agent chỉ diễn giải kết quả bằng lời. Có thể chèn ` ```chess ` chứa FEN/PGN để hiển thị bàn cờ không cần gọi tool.

**Chỉnh agent:** sửa block `builtin-chess-coach` trong YAML (prompt, temperature, allowed_tools…) → `docker compose ... restart app`. Mọi sửa đổi file dùng chung này → ghi `04-nhat-ky-tuy-bien.md` (mục C).

## 5.3. Chess Wikilink (đã có — dùng khi soạn nội dung)
Cú pháp: `[[game/<slug>|Nhãn]]`, nhúng `![[…]]`. Loại: `game` / `position` / `lesson` / `course` / puzzle.
- Gõ `[[` trong ô soạn bài giảng/khóa học (hoặc trang wiki thủ công) → **autocomplete** gợi ý, chèn đúng `slug|nhãn`.
- Slug bất biến, sinh tự động — **đừng tự gõ tay**, dùng autocomplete/picker (có xem trước bàn cờ).
- Link sai nhẹ vẫn mở đúng nhờ fuzzy; link gãy hẳn có nút "Tạo mới".
- Chi tiết & demo: `docs/chess-wikilink-demo.md` (`node scripts/seed_chess_wikilink_demo.mjs` để tạo dữ liệu demo — cần `API_KEY` tenant lấy ở UI Cài đặt → API key).

## 5.4. Bật RAG cờ (tùy chọn — mặc định TẮT)
Để agent HLV **trích dẫn lý thuyết/sách/ván mẫu** từ kho tri thức:
1. Đảm bảo tenant có ≥1 KB đã cấu hình embedding (KB cờ sẽ sao chép cấu hình đó).
2. Đặt `CHESS_KB_INDEX=true` cho service `app` (cần embedding + vector store + worker chạy ổn).
3. Tạo/sửa ván/thế/bài giảng → KB **"Tri thức cờ vua"** tự sinh bản ghi (best-effort, không chặn thao tác; import PGN hàng loạt KHÔNG trigger).
4. Trong `builtin-chess-coach`: đổi `kb_selection_mode` sang `selected`/`all` và thêm `knowledge_search` vào `allowed_tools` → restart `app`.
5. Hỏi thử về ván vừa tạo → kiểm tra agent truy hồi được nội dung. Bật Langfuse (`--profile langfuse`) nếu cần soi pipeline.

## 5.5. Dựng knowledge base nội dung (cho RAG / Wiki)
Nếu muốn KB tri thức cờ tổng quát (ngoài KB tự sinh), gợi ý tách theo nhóm `02-mien-co-vua.md` §2.1:

| KB | Quyền | Ưu tiên |
|---|---|---|
| `luat-co-vua`, `khai-cuoc`, `chien-thuat` | Mở (Viewer) | Cao |
| `tan-cuoc`, `van-hoa-lich-su` | Mở | Trung bình |
| `giao-trinh-6-cap` | Nội bộ HLV (Contributor+) | Cao |
| `van-hanh-duongsinh` | Hạn chế (nội bộ) | Tùy nhu cầu |

Khi nạp: gắn tag **cấp độ** (Tốt…Vua) + **nhóm nội dung**; khai cuộc thêm **ECO**; giữ FEN/PGN raw; nạp xong chạy vài câu hỏi mẫu để kiểm recall.

CLI nhanh:
```bash
weknora auth login --host http://localhost:8080
weknora kb list
weknora link --kb khai-cuoc
weknora doc upload ./khai-cuoc/sicilian.md
weknora chat "Y tuong chinh cua Sicilian la gi?"
```

## 5.6. Thêm tool cờ mới (mẫu mở rộng)
1. Tạo `internal/agent/tools/chess_<ten>.go` (theo mẫu các tool sẵn + `chess_common.go`).
2. Đăng ký tool vào registry (file dùng chung → **ghi mục C của nhật ký**).
3. Thêm vào `allowed_tools` của agent trong `builtin_agents.yaml`.
4. Nếu cần dữ liệu mới → thêm migration `000070+` (`.up`/`.down`).
5. `gofmt`, test, commit `feat(chess): ...`.

## 5.7. Lưu ý production
- Production thật tại `weknora.covuaduongsinh.com` (systemd `weknora-chess.service`). **Backup DB trước** mọi thao tác có rủi ro; không drop bảng/xóa volume khi chưa backup.
- Deploy code Go chạy trong container → phải **rebuild image** thì code mới mới vào (giống lưu ý của Arkon, nhưng đây là dự án khác — đừng nhầm hạ tầng).
