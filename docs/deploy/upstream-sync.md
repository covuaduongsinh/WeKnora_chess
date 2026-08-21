# Đồng bộ với upstream Tencent/WeKnora

Quy trình merge upstream cho fork `WeKnora_chess`. Đọc trước mỗi lần `git merge upstream`.

## 1. Vì sao migration cờ nằm ở dải `000900+`

Tới 22/8/2026 lớp cờ chiếm `000062`–`000074`. Upstream **cũng** dùng đúng dải đó
(`000062_mcp_oauth`, `000063_knowledge_multi_tags`, … `000074_mcp_oauth_refresh_lease`)
— **đụng khít 13/13**.

`golang-migrate` chỉ lưu **một** con số trong `schema_migrations(version, dirty)`.
Production khi đó ở version `74`. Nếu merge mà không xử lý:

- hai file cùng số → runner báo **duplicate migration version**, app không lên; hoặc
- đổi tên đại khái → **13 migration upstream bị bỏ qua im lặng** (DB đã ở 74) →
  thiếu bảng `principals`, `storage_backends`, `mcp_oauth*`… → lỗi runtime rải rác,
  rất khó truy.

Nên migration cờ đã dời sang `000900`–`000912`, trả dải `000062`–`000899` cho upstream.
**Migration cờ mới đánh tiếp từ `000913`.**

## 2. Điều kiện bắt buộc: migration cờ phải idempotent

Quy trình bên dưới cho migration cờ **chạy lại mỗi lần deploy**, nên mọi file trong
dải `000900+` phải là no-op khi chạy lần hai:

- `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`, `DROP INDEX IF EXISTS`
- `ALTER TABLE … ADD COLUMN IF NOT EXISTS`
- Không `ADD CONSTRAINT`, không `CREATE TYPE`, không `CREATE TABLE x` trần

Kiểm nhanh trước khi commit migration mới:

```bash
grep -nE 'CREATE TABLE [^I]|CREATE INDEX [^I]|ADD CONSTRAINT|CREATE TYPE' migrations/versioned/0009*.sql
# kỳ vọng: không ra gì
```

> ⚠️ **SQLite là ngoại lệ.** SQLite **không có** `ADD COLUMN IF NOT EXISTS`. Ba migration
> lite của fork (`000901`–`000903`) vì thế **không** idempotent. Bản lite là DB dùng thử
> local — khi đổi số hoặc đặt lại version, **xoá file `.db` rồi để nó dựng lại từ đầu**.

## 3. Quy trình merge (mỗi chặng một release)

Merge **từng release một**, đừng nhảy thẳng tới tag mới nhất — conflict cùng một vùng
code sẽ lặp lại và `rerere` sẽ tự áp lại cách giải của chặng trước.

```bash
git fetch upstream --tags
git config rerere.enabled true          # nhớ cách resolve, tự áp lại chặng sau
git config merge.conflictStyle zdiff3   # hiện cả bản gốc chung

# đo trước, chưa đụng working tree:
git merge-tree --write-tree --name-only HEAD v0.7.0

git switch -c chore/upstream-sync
git merge v0.6.3      # rồi v0.7.0 → v0.7.1 → v0.7.2
```

### Bốn bước kiểm BẮT BUỘC mỗi chặng (rút từ lần merge 0.6.2→0.7.2)

Bốn lỗi dưới đây đều **không hiện trong danh sách conflict của git** — chỉ lộ khi build/test:

1. **`grep -rn "locales/zh-CN" frontend/src`** — upstream liên tục thêm test/helper đọc `zh-CN.ts`,
   file mà fork đã xoá. Đã gặp **3 lần** trong 4 chặng (`TagEditDialog.test.ts`,
   `workspaceTerminology.test.ts`, `BatchTagDialog.test.ts` + `localeKeyAudit.ts`).
   Triệu chứng: `ENOENT` / `ERR_MODULE_NOT_FOUND`. Chuyển tham chiếu sang `vi-VN`.
2. **Constructor dùng chung mà fork đã chèn tham số** — `NewWikiPageService`,
   `NewAgentService`, `NewChessLibraryService`… Khi upstream chèn thêm tham số của họ,
   **test của upstream vẫn truyền số cũ** → lỗi biên dịch test. Đã gặp 2 lần.
3. **Chạy `npm run build` sau MỌI sửa file i18n** — `npm test` và `vue-tsc` **không parse**
   file locale, nên thiếu dấu phẩy hoặc nháy lồng sai chỉ esbuild mới bắt.
   Khi chèn khoá vào cuối một container, luôn thêm `,` vào dòng cuối trước đó.
4. **`go test ./internal/agent/tools/`** — 0.7.2 đặt ra hợp đồng "mọi tool lộ ra UI phải khai
   model-handle policy". Tool cờ khai ở `internal/modelcontext/tool_policy_chess.go` (file riêng,
   dùng `init()`). Thêm tool cờ mới thì **phải thêm tên vào đó**, nếu không test đỏ.

### File luôn phải kiểm bằng tay sau merge

| File | Vì sao |
|---|---|
| `internal/application/service/session_agent_qa.go` | Fork nới ràng buộc **rerank thành tuỳ chọn** (tenant Dương Sinh không có rerank model). Mất bản vá này → **RAG cờ chết ngay** |
| `internal/application/service/agent_service.go` | Đăng ký 7 tool cờ, truyền `chessLibraryService`, getter depth/plies/limit, probe health engine |
| `internal/application/service/wiki_page.go` | `chessRefPrefixes` — thiếu một dòng thì `[[position/x]]` thành link wiki thường, **lỗi câm** |
| `internal/router/router.go`, `internal/container/container.go` | Các nhóm route `/chess/*` và wiring DI |
| `frontend/src/i18n/locales/zh-CN.ts` | Fork **cố ý xoá** file này. Merge sẽ báo modify/delete → chọn giữ xoá |
| `frontend/src/views/chat/components/ToolResultRenderer.vue`, `types/tool-results.ts` | Nhánh render `chess_board` + cờ `fen_invalid` |
| `go.sum`, `frontend/package-lock.json` | Đừng resolve tay — lấy bản upstream rồi `go mod tidy` / `npm install` |
| `internal/router/routes_chess.go` | 4 nhóm route cờ nằm ở đây (tách ra từ 0.7.2). `router.go` chỉ còn 4 field `Chess*Handler` + 4 dòng gọi — nếu upstream đổi `RouterParams` thì nối lại 2 chỗ đó |
| `internal/modelcontext/tool_policy_chess.go` | Policy cho 7 tool cờ, bơm bằng `init()`. 0 dòng sửa `tool_policy.go` của upstream |
| `mcp-server/**` | **Cố ý lấy nguyên bản upstream** — đây là CLI phụ trợ, giữ bản Việt hoá chỉ tạo nợ merge lặp lại |

## 4. Runbook DB production — chạy cho MỖI chặng deploy

1. **Backup**: `bash scripts/deploy/backup.sh` — không bỏ qua.
2. Ghi mốc hiện tại:
   ```sql
   SELECT version, dirty FROM schema_migrations;
   ```
3. Đặt lại version về mốc **ngay trước** nhóm migration upstream của chặng:

   | Chặng | Đặt version về | Runner sẽ chạy |
   |---|---|---|
   | 0.6.3 | `61` | 62, 63 → rồi 900–912 (no-op) |
   | 0.7.0 | `63` | 64–70 → rồi 900–912 |
   | 0.7.1 | `70` | 71–74 → rồi 900–912 |
   | 0.7.2 | `74` | 75–79 → rồi 900–912 |

   ```sql
   UPDATE schema_migrations SET version = 61, dirty = false;
   ```
   (hoặc `migrate … force 61` — CLI có sẵn trong image app.)

4. Deploy: `bash scripts/deploy/pull-deploy.sh`. App tự chạy migration khi khởi động.
5. **Nghiệm thu**:
   ```sql
   SELECT version, dirty FROM schema_migrations;          -- kỳ vọng 912, false
   SELECT count(*) FROM chess_books;                       -- dữ liệu cờ còn nguyên
   SELECT count(*) FROM chess_articles;
   SELECT to_regclass('storage_backends');                 -- bảng upstream mới có mặt
   ```

Migration lỗi giữa chừng: app **vẫn khởi động được** (thiết kế của WeKnora), trang
System Info hiện lỗi — xem [migration-troubleshooting.md](../migration-troubleshooting.md).
Rollback = restore backup bước 1 + `git revert` chặng đó.

## 4a. ⚠️ Kiểm `.env` production TRƯỚC khi khởi động bản mới

Upstream 0.7.x đổi nhiều biến trong `docker-compose.yml` từ **giá trị cứng** sang
**đọc `.env`**:

```diff
-      - DB_HOST=postgres
+      - DB_HOST=${DB_HOST:-postgres}
```

`.env` production có `DB_HOST=localhost` (giá trị dành cho chạy ngoài Docker; trước đây
vô hại vì compose ghi đè cứng thành `postgres`). Sau khi nâng cấp, app trong container
đọc đúng `localhost` và **không nối được Postgres** → panic lúc dựng DI:

```
DB Config: user=postgres host=localhost port=5432 dbname=WeKnora
dial tcp 127.0.0.1:5432: connect: connection refused
panic: could not build arguments for function ...registerModelConcurrencyLimiter
```

Đây là lỗi cấu hình, **không phải lỗi merge** — nhưng nó làm app không lên nổi.
Trước khi khởi động bản mới, đối chiếu mọi biến vừa chuyển sang `${VAR:-default}`:

```bash
git diff <commit-cũ> <commit-mới> -- docker-compose.yml | grep -E '^\+.*\$\{'
# rồi kiểm giá trị tương ứng trong .env production
grep -nE '^(DB_HOST|DB_PORT|REDIS_HOST|MINIO_ENDPOINT|DOCREADER_ADDR)=' .env
```

Trên VPS phải là tên service trong compose network: `DB_HOST=postgres`, không phải
`localhost`.

## 4b. Gộp nhiều chặng vào MỘT lần deploy

Nếu merge liền 4 chặng trong git rồi mới deploy (đúng cách đã làm ở đợt 0.6.2→0.7.2),
**chỉ cần đặt lại version MỘT lần** về mốc trước nhóm migration upstream sớm nhất:

```sql
UPDATE schema_migrations SET version = 61, dirty = false;
```

Runner sẽ chạy tuần tự 62 → 79 (upstream, lần đầu) rồi 900 → 912 (cờ, no-op vì idempotent).
Nghiệm thu vẫn là `912 / false`.

## 5. Sau mỗi chặng

- Chạy bộ kiểm: `gofmt -l .` · `go build ./...` · `go vet ./...` ·
  `go test ./internal/chess/ ./internal/agent/tools/` · `cd frontend && npm run build && npm test && npx vue-tsc --noEmit`
- Ghi vào `.claude/memory/04-nhat-ky-tuy-bien.md`: điểm rẽ nhánh mới, file dùng chung nào
  upstream viết lại, quyết định resolve nào đáng nhớ.
- Dựng lại inventory: `git diff --name-status -M upstream/main...HEAD`

> `go test` gói `internal/application/service` **crash SIGSEGV** trên Windows tại
> `internal/types.init()` → `gojieba.NewJieba`. Đây là lỗi môi trường, không phải lỗi code.
> Chạy tầng service trên WSL/Linux/CI mới coi là đã kiểm runtime.
