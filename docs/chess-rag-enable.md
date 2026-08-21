# Bật RAG cờ cho agent "HLV Cờ vua" (runbook)

> Mục tiêu: để agent HLV **trích dẫn** nội dung trong KB **"Tri thức cờ vua"** thay vì chỉ "tính cờ" bằng engine.
> **An toàn:** indexer là best-effort, không chặn CRUD; có thể tắt lại bất cứ lúc nào (mục Rollback).

## Loại nội dung được index

| Loại | Điều kiện index |
|---|---|
| Ván cờ · Bài tập · Thế cờ · Bài giảng | Luôn (khi `CHESS_KB_INDEX` bật) |
| **Sách + Chương** | Chỉ khi sách `status="published"` — bản thảo KHÔNG vào kho |
| **Bài viết** (Ngân hàng bài viết) | Chỉ khi `status="published"` — bản thảo KHÔNG vào kho |

Hạ `published → draft` sẽ **tự động GỠ** khỏi kho, không để lại tri thức cũ.

> ⚠️ **Giới hạn 500:** `ListArticles`/`ListBooks`/`ListPositions` đều chặn ở 500 bản ghi.
> Trên 500 mục `published` cùng loại, reindex sẽ **cắt âm thầm** — không có cảnh báo.

## Cách nhanh nhất: panel "Kho tri thức" trong giao diện

Vào **Quản lý cờ vua → nút "Kho tri thức"** (góc trên bên phải). Panel hiện:
- 4 điều kiện hoạt động (bật index / kho tồn tại / có embedding model / agent nhìn thấy được kho), **kèm câu xử lý ngay tại chỗ** khi điều kiện nào đỏ;
- số mục **đã đẩy vào kho tách theo loại** (bài viết / sách / chương / ván / bài tập / thế cờ);
- tiến độ embedding (Xong / Đang chờ / Lỗi / Tổng);
- nút **"Đẩy lại index"** và **"Làm mới"**.

Panel cần quyền **Contributor** trở lên. Các mục `curl` bên dưới giữ lại làm phương án dự phòng (khi không vào được UI, hoặc thao tác từ server).

> **"Đã đẩy vào kho" KHÁC "đã embed xong".** Đẩy đi xong mà `Đang chờ` vẫn > 0 thì agent
> chưa tìm thấy — phải đợi embedding nền chạy hết. Đây đúng là chỗ từng bị hiểu nhầm
> thành "thành công giả" ở đợt bật RAG đầu tiên.

## 0b. BẮT BUỘC sau khi chạy migration 000073/000074

Hai migration này thêm **hệ thẻ thống nhất** và cột **`search_text`** (khử dấu).
Cả hai đều để **rỗng** cho dữ liệu cũ — phải nạp một lần:

> **Quản lý cờ vua → tab "Thẻ" → nút "Nạp thẻ từ dữ liệu cũ"**
> (tương đương `POST /api/v1/chess/tags/backfill`, cần quyền Contributor)

Một nút làm **hai** việc: tách các thẻ CSV cũ (thế cờ/sách/bài viết) thành thẻ
thật, và tính `search_text` cho toàn bộ bản ghi.

⚠️ **Không chạy bước này thì ô tìm sẽ không ra gì với dữ liệu cũ** — cột
`search_text` rỗng nên mọi truy vấn `LIKE` đều trượt. Idempotent, chạy lại vô hại.

Từ đây trở đi mọi lần tạo/sửa tự điền, không phải làm lại.

**Nhãn tài liệu trong KB cờ:** từ đợt này, tài liệu cờ được gắn nhãn theo LOẠI
("Ván cờ", "Sách", "Bài viết"…) để mở KB lên còn lọc được. Nhãn chỉ gắn lúc
index, nên tài liệu index từ TRƯỚC đó chưa có — chạy "Đẩy lại index" một lần là
có. *Lưu ý:* tool `knowledge_search` hiện chưa nhận tham số lọc theo nhãn, nên
đây là tiện ích cho NGƯỜI xem KB, chưa phải bộ lọc cho agent.

---

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
Từ đây, **tạo/sửa** ván·thế cờ·bài giảng·bài viết·sách (đã published) sẽ tự sinh bản ghi trong KB "Tri thức cờ vua".

## 2. Index dữ liệu cũ (backfill)
Import ván hàng loạt (`POST /chess/games/import`) **không** tự index. Sau khi bật ở bước 1, bấm
**"Đẩy lại index"** trong panel Kho tri thức, hoặc gọi **một lần**:
```bash
curl -X POST http://localhost:8080/api/v1/chess/library/reindex \
  -H "X-API-Key: <API_KEY_TENANT>"
# → {"success":true,"data":{
#      "games_total":N, "puzzles_total":N, "positions_total":N,
#      "books_total":N, "chapters_total":N, "articles_total":N,
#      "enqueued":N, "failed":0, "errors":[],
#      "note":"đã enqueue để index; embedding chạy nền — kiểm tra index-status sau ~1 phút"
#   }}
```
> Nếu trả lỗi `CHESS_KB_INDEX chưa bật` → kiểm tra lại bước 1 (env + restart). Endpoint cần quyền **Contributor**.
>
> `*_total` = số mục **đủ điều kiện** được đẩy đi (sách/bài viết chỉ đếm bản `published`).
> `enqueued` = số đã đưa vào hàng đợi embedding — **chưa phải** đã embed xong.

## 3. Đấu nối retrieval cho agent HLV
> **Trạng thái hiện tại: block `builtin-chess-coach` trong repo ĐÃ cấu hình xong** —
> `allowed_tools` đã có `knowledge_search` + `grep_chunks`, `kb_selection_mode: "all"`,
> `retrieve_kb_only_when_mentioned: false`, threshold đã đặt. **Không cần sửa gì thêm**,
> chỉ cần đảm bảo VPS đang chạy đúng bản `config/builtin_agents.yaml` này rồi restart `app`.

Cấu hình đang dùng (để đối chiếu, không phải việc phải làm):
```yaml
      allowed_tools:
        - "thinking"
        - "knowledge_search"     # bắt buộc để nạp KB
        - "grep_chunks"          # tra từ khóa
        - "chess_analyze_position"
        # … 6 tool cờ còn lại …
      kb_selection_mode: "all"
      retrieve_kb_only_when_mentioned: false
```
**Vì sao `all`?** KB "Tri thức cờ vua" được tạo tự động theo từng tenant nên không có ID cố định để dùng `selected` trong YAML builtin. `all` nạp mọi KB của tenant (đã lọc theo tool). Phù hợp khi tenant Dương Sinh chủ yếu là nội dung cờ. *(Nếu tenant có nhiều KB ngoài cờ gây nhiễu → chuyển `selected` + liệt kê ID KB cờ.)*

Overlay đã mount YAML → chỉ **restart `app`**, không cần rebuild:
```bash
docker compose -f docker-compose.yml -f docker-compose.chess.yml restart app
```

### ⚠️ Cạm bẫy: bản ghi đè trong DB làm YAML vô tác dụng
`GetAgentByID` **ưu tiên bản ghi trong bảng `custom_agents`, YAML chỉ là fallback**.
Nếu ai đó từng bấm **Lưu** trên agent "HLV Cờ vua" trong giao diện, một hàng đã được tạo
cho tenant đó — và từ lúc ấy **mọi sửa đổi YAML đều KHÔNG có tác dụng**, agent vẫn chạy
cấu hình cũ (có thể thiếu `knowledge_search`, `kb_selection_mode` cũ).

Triệu chứng: log hiện `tool not found: knowledge_search` dù YAML có đủ. Kiểm tra:
```sql
SELECT id, tenant_id, config FROM custom_agents WHERE id='builtin-chess-coach';
```
Có hàng → sửa qua **UI** (thêm `knowledge_search`, đặt phạm vi KB = tất cả), hoặc **xóa hàng đó**
để rơi về YAML.

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
> Cách nhanh: mở panel **Kho tri thức** (đầu file) — nó hiện đúng những thông tin dưới đây
> kèm câu xử lý. Phần `curl` này dành cho lúc thao tác từ server.
>
> Embedding chạy **NỀN**: `reindex` trả `enqueued` ≠ "đã embed xong". Sau ~1 phút, gọi:
```bash
curl -s https://weknora.covuaduongsinh.com/api/v1/chess/library/index-status \
  -H "X-API-Key: <API_KEY_TENANT>" | jq .data
# {
#   "enabled": true, "kb_exists": true, "kb_id": "...",
#   "embedding_model_id": "...", "embedding_configured": true,
#   "vector_enabled": true, "keyword_enabled": true, "searchable": true,
#   "total": 7, "completed": 7, "pending": 0, "failed": 0,
#   "enabled_docs": 7, "disabled_docs": 0, "sample_error": "",
#   "by_type": {"article": 12, "book": 3, "chapter": 40, "game": 7}
# }
```
**`by_type` trả lời câu "bài viết/sách đã vào kho chưa"** — đếm từ bảng ánh xạ
`chess_kb_index`. Cần con số riêng này vì bản ghi Knowledge trong KB **không mang
`chess_type`**, nên nhìn danh sách tài liệu không phân biệt được loại nào.

> ⚠️ `by_type` = **đã đẩy vào kho**; `completed` = **đã embed xong** (gộp mọi loại).
> `by_type.article=12` mà `pending=12` nghĩa là 12 bài đã đẩy nhưng chưa embed → agent chưa thấy.

Đọc kết quả theo bảng nhánh (kiểm tra **theo thứ tự**):

| Triệu chứng | Nghĩa | Cách xử lý |
|---|---|---|
| `enabled:false` | `CHESS_KB_INDEX` chưa bật | Làm lại Bước 1 (env + restart `app`). |
| `kb_exists:false` | KB cờ chưa tạo (chưa index lần nào / **chưa có KB embedding mẫu**) | Đảm bảo tenant có ≥1 KB cấu hình embedding → reindex lại. |
| `searchable:false` (vector+keyword đều tắt) | **NGUYÊN NHÂN GỐC HAY GẶP** — KB cờ bị tắt index → knowledge_search KHÔNG "nhìn thấy" KB (capability filter) + embedding bị skip lúc index. KB cũ tạo trước bản vá thường dính lỗi này. | **XÓA KB "Tri thức cờ vua"** trong UI (Kho tri thức → … → Xóa) rồi gọi lại `reindex`. Code mới tạo KB với vector+keyword BẬT tường minh → fix dứt điểm. |
| `embedding_configured:false` | KB cờ không có embedding model → chunk không lên vector store | Cấu hình embedding cho 1 KB; xóa KB cờ rồi reindex để tạo lại với model đúng. |
| `failed > 0`, có `sample_error` | Embedding lỗi (model hỏng / rate-limit / hết quota) | Đọc `sample_error`; sửa cấu hình model rồi reindex lại. |
| `disabled_docs > 0` (mà `searchable:true`) | Tài liệu chưa được bật sau xử lý | Reindex lại để xử lý lại; nếu vẫn → kiểm tra log worker. |
| `pending` lâu không về 0 | Worker embedding chưa chạy/kẹt | Kiểm tra container worker (asynq) + log `app`. |
| `by_type` thiếu hẳn loại đang tìm (vd không có `article`) | Loại đó chưa được đẩy vào kho — thường vì nội dung còn ở `draft` | Đổi sang `published` (sách/bài viết), rồi "Đẩy lại index". |
| `searchable:true`, `completed==total`, `enabled_docs==total` | OK — RAG truy hồi được | Nếu agent vẫn không ra nội dung → **kiểm bản ghi đè DB ở mục 3** trước, rồi mới soi `kb_selection_mode`/threshold trong YAML. |

> SQL fallback (khi không gọi được endpoint): `SELECT name, embedding_model_id FROM knowledge_bases WHERE name='Tri thức cờ vua';` và `SELECT parse_status, count(*) FROM knowledges WHERE knowledge_base_id='<kb_id>' GROUP BY parse_status;`

## 5. Rollback (tắt lại)
- `.env`: `CHESS_KB_INDEX=false` → restart `app` (ngừng index; dữ liệu đã index vẫn còn).
- `builtin_agents.yaml`: trả `kb_selection_mode: "none"` + bỏ `knowledge_search`/`grep_chunks` → restart `app`.

## 6. Sửa nội dung sau khi đã index

- **Đổi slug** (ván/bài tập/thế cờ/sách/chương/bài viết): tài liệu của slug CŨ được **tự động gỡ**
  khỏi kho trước khi index slug mới → không còn bản trùng. Wikilink cũ vẫn mở đúng nhờ bảng alias.
- **Hạ `published → draft`**: sách/bài viết bị **gỡ** khỏi kho ngay.
- **Xóa hẳn**: gỡ tài liệu + ánh xạ + mọi backlink 2 chiều.

Sau mỗi thao tác trên, mở panel Kho tri thức bấm "Làm mới" để đối chiếu `by_type`.

## Ghi chú merge
Sửa file dùng chung khi bật: `config/builtin_agents.yaml` (C1), `.env`. Đã ghi `04-nhat-ky-tuy-bien.md`. Code backfill (`ReindexAll` + route `/chess/library/reindex`) thuộc lớp cờ; router.go (C1) đã có 1 dòng đăng ký route mới.
