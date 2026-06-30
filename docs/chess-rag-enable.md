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

## 3b. Bật trên PRODUCTION (`weknora.covuaduongsinh.com`)
> Các lệnh ở trên dùng compose local. Trên VPS (GHCR/Caddy) làm tuần tự, **AN TOÀN TRƯỚC**:
1. **Backup DB trước tiên** (bắt buộc — thao tác có rủi ro):
   ```bash
   bash scripts/deploy/backup.sh    # hoặc theo docs/deploy/backup-restore.md
   ```
2. Đặt `CHESS_KB_INDEX=true` trong `.env` trên VPS (mục service `app`).
3. Áp lại & khởi động lại `app` (không cần rebuild nếu chỉ đổi env):
   ```bash
   bash scripts/deploy/redeploy.sh   # hoặc: dc up -d app  (xem scripts/deploy/dc.sh)
   ```
4. Backfill dữ liệu cũ — gọi reindex vào URL production (API key tenant ở UI → Cài đặt → API key):
   ```bash
   curl -X POST https://weknora.covuaduongsinh.com/api/v1/chess/library/reindex \
     -H "X-API-Key: <API_KEY_TENANT>"
   ```
5. Sửa block `builtin-chess-coach` (mục 3) trong `config/builtin_agents.yaml` trên VPS → restart `app`.
6. Nghiệm thu (mục 4). Nếu trục trặc → Rollback (mục 5) + restore backup nếu cần.

> **Khuyến nghị mạnh:** chạy thử **local 1 lượt** (mục 1–4) trước khi làm production để bắt sớm lỗi pipeline embedding/worker.

## 4. Nghiệm thu
1. Nạp ≥ vài tài liệu lý thuyết vào KB cờ (hoặc dựa vào ván/bài tập đã reindex).
2. Hỏi HLV một câu lý thuyết về nội dung vừa nạp.
3. Kỳ vọng: câu trả lời **có trích nguồn** từ "Tri thức cờ vua".
4. Soi pipeline (tùy chọn): bật Langfuse `--profile langfuse`.

## 4b. Chẩn đoán bằng endpoint trạng thái (khi RAG "rỗng")
> Embedding chạy **NỀN**: `reindex` trả `enqueued` ≠ "đã embed xong". Sau ~1 phút, gọi:
```bash
curl -s https://weknora.covuaduongsinh.com/api/v1/chess/library/index-status \
  -H "X-API-Key: <API_KEY_TENANT>" | jq .data
# {
#   "enabled": true, "kb_exists": true, "kb_id": "...",
#   "embedding_model_id": "...", "embedding_configured": true,
#   "total": 7, "completed": 7, "pending": 0, "failed": 0, "sample_error": ""
# }
```
Đọc kết quả theo bảng nhánh:

| Triệu chứng | Nghĩa | Cách xử lý |
|---|---|---|
| `enabled:false` | `CHESS_KB_INDEX` chưa bật | Làm lại Bước 1 (env + restart `app`). |
| `kb_exists:false` | KB cờ chưa tạo (chưa index lần nào / **chưa có KB embedding mẫu**) | Đảm bảo tenant có ≥1 KB cấu hình embedding → reindex lại. |
| `embedding_configured:false` | **NGUYÊN NHÂN GỐC** — KB cờ không có embedding model → chunk không lên vector store | Cấu hình embedding cho 1 KB; xóa KB cờ rỗng rồi reindex để tạo lại với model đúng. |
| `failed > 0`, có `sample_error` | Embedding lỗi (model hỏng / rate-limit / hết quota) | Đọc `sample_error`; sửa cấu hình model rồi reindex lại. |
| `pending` lâu không về 0 | Worker embedding chưa chạy/kẹt | Kiểm tra container worker (asynq) + log `app`. |
| `completed == total > 0` | OK — RAG truy hồi được | Nếu agent vẫn không ra nội dung → soi `kb_selection_mode`/threshold trong YAML. |

> SQL fallback (khi không gọi được endpoint): `SELECT name, embedding_model_id FROM knowledge_bases WHERE name='Tri thức cờ vua';` và `SELECT parse_status, count(*) FROM knowledges WHERE knowledge_base_id='<kb_id>' GROUP BY parse_status;`

## 5. Rollback (tắt lại)
- `.env`: `CHESS_KB_INDEX=false` → restart `app` (ngừng index; dữ liệu đã index vẫn còn).
- `builtin_agents.yaml`: trả `kb_selection_mode: "none"` + bỏ `knowledge_search`/`grep_chunks` → restart `app`.

## Ghi chú merge
Sửa file dùng chung khi bật: `config/builtin_agents.yaml` (C1), `.env`. Đã ghi `04-nhat-ky-tuy-bien.md`. Code backfill (`ReindexAll` + route `/chess/library/reindex`) thuộc lớp cờ; router.go (C1) đã có 1 dòng đăng ký route mới.
