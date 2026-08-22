# 03 — Kiến trúc: WeKnora nền + lớp cờ vua

Tham chiếu để agent định vị code nhanh. (Nền: WeKnora v0.7.2 — đồng bộ 0.6.2 → 0.7.2 ngày 22/8/2026.)

## 3.1. Bản đồ thư mục nền (upstream)
| Thư mục/File | Vai trò |
|---|---|
| `cmd/` | Entry point binary Go |
| `internal/` | **Lõi backend Go** — retrieval, agent, wiki, RBAC, handler, service, repository, types |
| `config/` | Cấu hình app — gồm `builtin_agents.yaml` (định nghĩa agent dựng sẵn, i18n) |
| `frontend/` (và `web/`) | **Web UI** Vue + TS |
| `docreader/` | Parse tài liệu (Python, gRPC) — PDF/Word/Excel/ảnh… |
| `cli/` | CLI `weknora` (có `cli/AGENTS.md` riêng — không ghi đè) |
| `mcp-server/` | MCP server (`mcp-server/MCP_CONFIG.md`) |
| `migrations/versioned/` | Migration schema PostgreSQL (đánh số tăng dần) |
| `deploy/`, `docker/`, `helm/` | Hạ tầng Docker / K8s |
| `skills/`, `examples/skills/` | Agent Skills (sandboxed) |
| `docs/` | Tài liệu (gồm `docs/api/`, `ROADMAP.md`, `QA.md`, RBAC) |
| `Makefile` | `make dev-*`, build |

## 3.2. Luồng nền (pipeline)
```
Tài liệu → docreader (parse) → chunking (3-tier, parent-child)
        → embedding → vector store (pgvector) → retrieval (BM25/dense/GraphRAG, rerank)
        → LLM → RAG Q&A | ReAct Agent | Wiki Mode (sinh trang + graph)
```
Ba chế độ: **RAG Q&A** (nhanh), **ReAct Agent** (nhiều bước, tool calling), **Wiki Mode** (chưng cất tài liệu → Wiki + knowledge graph). RBAC 4 vai trò: Owner/Admin/Contributor/Viewer; sở hữu theo KB; audit log.

## 3.3. LỚP CỜ VUA (tùy biến của repo này)
Khái niệm: **Game / Position / Lesson / Course / Puzzle / Book+Chapter / Article / Tag / Chess Ref (wikilink)** — 8 loại nội dung.

**Backend Go:**
```
internal/chess/                       # engine: board, engine, uci_engine, http_engine (bọc Arasan)
                                       # + fen.go, text.go (khử dấu), search_rank.go (chấm điểm tìm kiếm)
internal/agent/tools/chess_*.go       # 7 tool agent + chess_common + chess_openings_data
internal/application/repository/chess_*  + wiki_chess_ref.go
internal/application/service/chess_*      # course, knowledge_indexer, knowledge_text, library(+book/article/
                                          # position/tag), resolve, slug, search
internal/handler/chess_*               # API: course, library, position, book, article, tag, search, ref, engine
internal/types/(interfaces/)chess_*    + wiki_chess_ref.go
internal/router/routes_chess.go       # ⚠️ 0.7.2 module hoá router: 4 nhóm route cờ TÁCH khỏi router.go.
                                       # router.go chỉ còn 4 field Chess*Handler + 4 dòng gọi.
internal/modelcontext/tool_policy_chess.go  # ⚠️ BẮT BUỘC từ 0.7.2: mọi tool lộ ra UI phải khai
                                       # model-handle policy, im lặng = go test ./internal/agent/tools/ ĐỎ.
                                       # Bơm bằng init() → 0 dòng sửa tool_policy.go của upstream.
                                       # Policy RỖNG là đúng cho cả 7 tool cờ (đầu vào FEN/PGN/slug,
                                       # đầu ra là bàn cờ do frontend render, không phải đoạn trích).
                                       # Thêm tool cờ mới → PHẢI thêm tên vào đây.
```

**7 tool cờ (đăng ký cho agent):** `chess_analyze_position`, `chess_best_move`, `chess_evaluate_game`, `chess_explain_move`, `chess_lookup_opening`, `chess_generate_puzzle`, `chess_lookup_position`.

**Migrations cờ:** `000900` courses · `000901` games_puzzles · `000902` slugs · `000903` wiki_chess_refs · `000904` course_slug · `000905` refs_source_type · `000906` slug_aliases · `000907` kb_index · `000908` chess_positions (Ngân hàng thế cờ) · `000909` chess_books (Thư viện sách: kệ/sách/chương/ảnh/phiên bản) · `000910` chess_articles (Ngân hàng bài viết: bài/chuyên mục/ảnh/phiên bản + cột `kind` cho slug alias) · `000911` chess_tags (hệ thẻ THỐNG NHẤT: từ điển thẻ + pivot ĐA HÌNH phủ cả 8 loại nội dung) · `000912` chess_search_text (cột `search_text` khử dấu + index GIN trigram cho cả 8 loại).
> Dải cờ là **`000900+`** từ 22/8/2026 (trước đó `000062`–`000074`, đụng khít với upstream). Migration cờ mới đánh tiếp từ **`000913`**; dải `000062`–`000899` thuộc upstream. Xem `docs/deploy/upstream-sync.md`.

**Frontend:**
```
frontend/src/views/chess/             # ChessCourses, ChessManage, GameLibrary, PuzzleBank,
                                       # PositionBank, BookLibrary, BookPrint
  components/                          # Backlinks, RefDialog, RefEmbed, RefMissing, WikiLinkSuggest,
                                       # ChessPositionEditor, ChessShelfManager, ChessChapterHistory
frontend/src/views/chat/components/tool-results/ChessBoardDisplay.vue   # bàn cờ tương tác
frontend/src/api/chess/ · stores/chessWikiDraft.ts · utils/chessBlocks.ts · utils/chessRef.ts
  · utils/chessPositionOptions.ts · utils/chessBookOptions.ts
```

**Engine (Arasan) — sidecar HTTP:**
```
docker-compose.chess.yml              # overlay: service chess-engine + biến WEKNORA_CHESS_*
docker/Dockerfile.chess-engine        # build Arasan (ARASAN_VERSION, ARASAN_BUILD=modern/avx2)
docker/chess-engine/uci_http_bridge.py  # cầu UCI → HTTP
```

**Agent HLV:** `config/builtin_agents.yaml` → `builtin-chess-coach` (avatar ♟️, system prompt tiếng Việt, `kb_selection_mode: none`, allowed_tools = 6 chess tools + `thinking`).

**Deploy/Docs:** `scripts/deploy/weknora-chess.service` (systemd), `scripts/seed_chess_wikilink_demo.*`, `docs/chess-wikilink-demo.md`.

## 3.4. Biến môi trường lớp cờ (`.env`)
- Engine: `WEKNORA_CHESS_ENABLED`, `WEKNORA_CHESS_MODE=http`, `WEKNORA_CHESS_ENGINE_ENDPOINT`, `WEKNORA_CHESS_DEFAULT_DEPTH`, `WEKNORA_CHESS_TIMEOUT_SEC`.
- RAG cờ: `CHESS_KB_INDEX` (**mặc định TẮT**) — bật để index ván/thế/bài giảng vào KB "Tri thức cờ vua".

## 3.5. API cờ (ví dụ)
- `GET /api/v1/chess/refs/search?q=` — tra cứu thực thể cờ (dùng cho autocomplete `[[`).
- Các handler course/library/ref khác trong `internal/handler/chess_*`.

## 3.6. Khi upstream cập nhật (CHIẾN LƯỢC SYNC)
Lớp cờ đụng sâu `internal/` + `migrations/` → merge upstream sẽ conflict. Trước khi `git merge upstream/main`:
1. `git fetch upstream` rồi đọc `CHANGELOG.md` (chú ý breaking ở schema/migrations, agent, router, RBAC).
2. Đối chiếu `04-nhat-ky-tuy-bien.md` — các file dùng chung đã sửa → điểm conflict dự kiến.
3. Backup DB trước khi chạy migration mới (production).
4. Giữ code cờ trong file `*chess*` riêng để merge dễ; resolve conflict ở file dùng chung theo nhật ký.
