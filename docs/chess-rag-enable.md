# Bật RAG cờ cho agent "HLV Cờ vua" (runbook)

> Mục tiêu: để agent HLV **trích dẫn lý thuyết/sách/ván mẫu** từ KB **"Tri thức cờ vua"** thay vì chỉ "tính cờ" bằng engine.
> Trạng thái mặc định: **TẮT** (`CHESS_KB_INDEX` off, `kb_selection_mode: none`). Runbook này là phần **chuẩn bị sẵn** của WS2 — code backfill đã có, chỉ còn các bước bật dưới đây.
> **An toàn:** indexer là best-effort, không chặn CRUD; có thể tắt lại bất cứ lúc nào (mục Rollback).

## 0. Tiền điều kiện
- Tenant có **≥ 1 KB đã cấu hình model embedding** (KB cờ sẽ SAO CHÉP cấu hình embedding đó khi tự tạo). *(Thầy đã xác nhận local có embedding.)*
- Stack chạy đầy đủ (embedding + vector store + worker).

## 1. Bật ingest (đánh chỉ mục)
Trong `.env` của service `app`:
```env
CHESS_KB_INDEX=true
```
Rồi restart `app`:
```bash
docker compose -f docker-compose.yml -f docker-compose.chess.yml up -d app
# hoặc: make dev-app
```
Từ đây, **tạo/sửa** ván·thế cờ·bài giảng sẽ tự sinh bản ghi trong KB "Tri thức cờ vua".

## 2. Index dữ liệu cũ (backfill)
Import hàng loạt (`POST /chess/games/import`) **không** tự index. Sau khi bật ở bước 1, gọi **một lần**:
```bash
curl -X POST http://localhost:8080/api/v1/chess/library/reindex \
  -H "X-API-Key: <API_KEY_TENANT>"
# → {"success":true,"data":{"games_indexed":N,"puzzles_indexed":M}}
```
> Nếu trả lỗi `CHESS_KB_INDEX chưa bật` → kiểm tra lại bước 1 (env + restart). Endpoint cần quyền **Contributor**.

## 3. Đấu nối retrieval cho agent HLV
Sửa block `builtin-chess-coach` trong `config/builtin_agents.yaml`:
```yaml
      # 1) Cho phép công cụ tìm tri thức:
      allowed_tools:
        - "thinking"
        - "knowledge_search"     # ← THÊM (bắt buộc để nạp KB)
        - "grep_chunks"          # ← THÊM (tùy chọn, tra từ khóa)
        - "chess_analyze_position"
        - "chess_best_move"
        - "chess_evaluate_game"
        - "chess_explain_move"
        - "chess_lookup_opening"
        - "chess_generate_puzzle"
      # 2) Cho phép truy hồi KB:
      kb_selection_mode: "all"   # ← ĐỔI từ "none"
```
**Vì sao `all`?** KB "Tri thức cờ vua" được tạo tự động theo từng tenant nên không có ID cố định để dùng `selected` trong YAML builtin. `all` nạp mọi KB của tenant (đã lọc theo tool). Phù hợp khi tenant Dương Sinh chủ yếu là nội dung cờ. *(Nếu tenant có nhiều KB ngoài cờ gây nhiễu → chuyển `selected` + liệt kê ID KB cờ.)*

Overlay đã mount YAML → chỉ **restart `app`**, không cần rebuild:
```bash
docker compose -f docker-compose.yml -f docker-compose.chess.yml restart app
```

## 4. Nghiệm thu
1. Nạp ≥ vài tài liệu lý thuyết vào KB cờ (hoặc dựa vào ván/bài tập đã reindex).
2. Hỏi HLV một câu lý thuyết về nội dung vừa nạp.
3. Kỳ vọng: câu trả lời **có trích nguồn** từ "Tri thức cờ vua".
4. Soi pipeline (tùy chọn): bật Langfuse `--profile langfuse`.

## 5. Rollback (tắt lại)
- `.env`: `CHESS_KB_INDEX=false` → restart `app` (ngừng index; dữ liệu đã index vẫn còn).
- `builtin_agents.yaml`: trả `kb_selection_mode: "none"` + bỏ `knowledge_search`/`grep_chunks` → restart `app`.

## Ghi chú merge
Sửa file dùng chung khi bật: `config/builtin_agents.yaml` (C1), `.env`. Đã ghi `04-nhat-ky-tuy-bien.md`. Code backfill (`ReindexAll` + route `/chess/library/reindex`) thuộc lớp cờ; router.go (C1) đã có 1 dòng đăng ký route mới.
