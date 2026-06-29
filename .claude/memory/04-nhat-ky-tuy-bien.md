# 04 — Nhật ký & Inventory tùy biến (vs upstream)

> **Mục đích:** theo dõi MỌI khác biệt so với `Tencent/WeKnora` để quản lý `git merge upstream` và bàn giao.
> **Quy tắc cho AI agent:** khi sửa **file dùng chung của upstream** hoặc **thêm migration/schema**, PHẢI thêm dòng vào bảng phù hợp trong cùng PR.

## Thông tin fork
- **Upstream:** `https://github.com/Tencent/WeKnora` · **Origin:** `github.com/covuaduongsinh/WeKnora_chess`
- **Nền:** v0.6.2 · **Nhánh:** `main` · **Production:** `weknora.covuaduongsinh.com`
- **Điểm rẽ nhánh (merge-base) lúc lập inventory:** `5d0d317a` (2026-06-22). Upstream tip khi đó: `7d8a80ae` (2026-06-28).
- **Quy mô tùy biến (diff `upstream/main...HEAD`):** ~75 file lớp cờ (mới) · **76 file dùng chung bị SỬA** · 1 file bị XÓA · 39 file mới non-chess (deploy/lite/i18n).
- **Cập nhật lại inventory:**
  ```bash
  git fetch upstream main --no-tags
  git diff --name-status -M upstream/main...HEAD     # phần fork đã đổi so với điểm chung
  ```

---

## A. Lớp cờ — code RIÊNG (rủi ro conflict THẤP, ~75 file)
| Vùng | Đường dẫn |
|---|---|
| Engine | `internal/chess/` (board, engine, uci_engine, http_engine, *_test) |
| Agent tools | `internal/agent/tools/chess_*.go` (6 tool + common + demo_test) |
| Repository | `internal/application/repository/chess_*`, `wiki_chess_ref.go` |
| Service | `internal/application/service/chess_*` |
| Handler | `internal/handler/chess_*` |
| Types | `internal/types/(interfaces/)chess_*`, `wiki_chess_ref.go` |
| Frontend | `frontend/src/views/chess/**`, `views/chat/components/tool-results/ChessBoardDisplay.vue`, `api/chess/`, `stores/chessWikiDraft.ts`, `utils/chessBlocks.ts`, `utils/chessRef.ts` |
| Docker engine | `docker-compose.chess.yml`, `docker/Dockerfile.chess-engine`, `docker/chess-engine/uci_http_bridge.py` |
| Deploy/Docs | `scripts/deploy/weknora-chess.service`, `scripts/seed_chess_wikilink_demo.*`, `docs/chess-wikilink-demo.md` |

## B. Migrations cờ (NỐI TIẾP upstream) — `000062`–`000069`
courses · games_puzzles · slugs · wiki_chess_refs · course_slug · refs_source_type · slug_aliases · kb_index.
> Khi merge: nếu upstream thêm migration trùng dải số → đổi số migration cờ cho cao hơn.

---

## C. FILE DÙNG CHUNG ĐÃ SỬA (rủi ro conflict CAO) — 76 file
*(số "N×" = số lần xuất hiện chữ "chess" trong file, đo lúc lập inventory — càng cao càng là điểm móc lõi.)*

### C1 — Móc nối backend (BẮT BUỘC để ý khi merge)
| File | Vai trò móc nối cờ | Mức |
|---|---|---|
| `internal/application/service/agent_service.go` | Gắn agent/tool cờ vào luồng agent. **WS1 (nối Puzzle Bank):** +field `chessLibraryService` + param `NewAgentService` + truyền vào `NewChessGeneratePuzzleTool(s.chessLibraryService)` để tool ra bài tập từ DB. | 32× |
| `internal/config/config.go` | Đọc `WEKNORA_CHESS_*`, `CHESS_KB_INDEX` | 23× |
| `internal/agent/tools/definitions.go` | Đăng ký 6 tool `chess_*` vào registry | 21× |
| `internal/router/router.go` | Mount route `/api/v1/chess/...`. **WS2:** +nhóm `/chess/library` với `POST /reindex` (backfill KB cờ). | 17× |
| `internal/container/container.go` | Wiring DI service/repo cờ | 11× |
| `internal/types/custom_agent.go` | Field phục vụ agent cờ | 3× |
| `config/builtin_agents.yaml` | Agent `builtin-chess-coach` | (block cuối) |
| `internal/agent/act.go` | **0× chess** — sửa vì lý do khác (rà lại khi merge) | 0× |

### C2 — Tích hợp Wiki ↔ wikilink cờ
| File | Vai trò | Mức |
|---|---|---|
| `internal/application/service/wiki_page.go` | Trang wiki nhận diện/giải link cờ | 39× |
| `internal/types/wiki_page.go` | Kiểu dữ liệu wiki + ref cờ | 15× |
| `internal/application/service/wiki_ingest.go` | Expand wikilink khi embed | 1× |
| `internal/application/service/wiki_lint.go` | Lint wiki có ref cờ | 1× |
| `frontend/src/views/knowledge/wiki/WikiBrowser.vue` | Hiển thị node cờ trong graph | — |

### C3 — Móc nối frontend
| File | Vai trò | Mức |
|---|---|---|
| `frontend/src/types/tool-results.ts` | Kiểu kết quả tool cờ | 8× |
| `frontend/src/views/chat/components/ToolResultRenderer.vue` | Render bàn cờ từ kết quả tool | 4× |
| `frontend/src/router/index.ts` | Route khu "Quản lý cờ vua" | 3× |
| `frontend/src/components/menu.vue` | Mục menu cờ | 2× |
| `frontend/src/stores/menu.ts` | State menu | 1× |
| `frontend/src/utils/agent-tool-icons.ts` | Icon cho tool cờ | 1× |

### C4 — i18n & prompt templates (Việt hóa + bỏ tiếng Trung)
- **Bỏ** `frontend/src/i18n/locales/zh-CN.ts` (D) · **Thêm** `locales/vi-VN.ts` (A).
- Sửa: `i18n/index.ts`, `i18n/embed.ts`, `locales/en-US.ts`, `locales/ko-KR.ts`, `locales/ru-RU.ts`.
- Prompt: `config/prompt_templates/` → `agent_system_prompt.yaml`, `context_template.yaml`, `fallback.yaml`, `intent_prompts.yaml`, `rewrite.yaml`, `system_prompt.yaml`.
> Đây là vùng dễ conflict text khi upstream đổi prompt/i18n — thường tự giải quyết được, nhưng kiểm tra kỹ vi-VN không bị mất.

### C5 — Hạ tầng / build / deploy (khác biệt lớn ngoài lớp cờ)
- **Build/dep:** `Makefile`, `docker/Dockerfile.app`, `go.mod`, `go.sum`, `frontend/package.json`, `frontend/package-lock.json`, `frontend/index.html`, `frontend/env.d.ts`, `.gitattributes`, `.env.example`.
- **CI/CD & deploy riêng (mới, A):** `.github/workflows/cicd-deploy.yml`, `docker-compose.caddy.yml`, `docker-compose.ghcr.yml`, `docker-compose.override.yml`, `docker/caddy/Caddyfile`, `scripts/deploy/{backup,restore,dc,pull-deploy,redeploy,server-bootstrap}.sh`, `scripts/deploy/hetzner-cloud-init.yaml`, `scripts/dev/{local-deploy,local-status}.sh`, `scripts/dev/githooks/pre-push`, `docs/deploy/{backup-restore,cicd,dev-workflow,hetzner,https}.md`. Sửa: `.github/workflows/docker-image.yml`.
- **mcp-server:** sửa `main.py`, `run_server.py`, `run.py`, `setup.py`, `test_imports.py`, `test_module.py` + thêm `uv.lock`, `*.egg-info/*`.

### C6 — Bản chạy SQLite ("lite") — mới
- Sửa `migrations/sqlite/000000_init.up.sql`; **thêm** `migrations/sqlite/000001..000003*` (wiki_and_indexing, lite_missing_tables, messages_attachments), `build_sqliteshim/*.h`, `frontend/src/utils/fileTransfer.ts`.
- Liên quan: `.env.lite`, `weknora-lite.log` ở root.
> Fork có nhánh chạy **SQLite** thay PostgreSQL (bản nhẹ) — lưu ý khi merge migration upstream (chỉ có bản postgres).

### C7 — File dùng chung khác bị sửa (không chứa "chess")
`internal/im/{cmd_clear,cmd_help,cmd_info,cmd_search,cmd_stop,service}.go`, `internal/im/feishu/adapter.go`, `internal/errors/errors.go`, `internal/middleware/language.go`, `internal/types/{context_helpers,context_helpers_test,placeholder}.go`, `frontend/src/App.vue`, `frontend/src/views/auth/Login.vue`, `frontend/src/api/chat/streame.ts`, `frontend/src/api/embed/index.ts`, `frontend/src/components/{AgentEmbedChannelPanel,manual-knowledge-editor,MyInvitationsDialog}.vue`, `frontend/src/views/chat/components/{AgentStreamDisplay,botmsg}.vue`, `frontend/src/views/chat/components/tool-results/WebSearchResults.vue`, `frontend/src/views/settings/{GeneralSettings,TenantInfo,TenantMembers,UserProfile}.vue`, `frontend/src/views/system/SystemSettings.vue`, `frontend/src/utils/request.ts`, `data/weknora.db.bak-131441`.
> Phần lớn là Việt hóa/tinh chỉnh UI/IM. Rà từng cái khi conflict; ưu tiên giữ thay đổi tối thiểu.

---

## D. Quyết định kiến trúc (ADR rút gọn)
- **Engine:** Arasan (MIT) sidecar **HTTP** (UCI→HTTP bridge), gọi qua `WEKNORA_CHESS_*`.
- **Agent HLV:** `kb_selection_mode: none` — mặc định KHÔNG RAG, chỉ 6 tool cờ + engine. *(Bật `CHESS_KB_INDEX` + thêm `knowledge_search` để trích dẫn lý thuyết/sách.)*
- **RAG cờ:** gate `CHESS_KB_INDEX` (mặc định TẮT); import PGN hàng loạt KHÔNG trigger index.
- **Wikilink:** slug bất biến; resolve `exact → alias → fuzzy` (bigram-Jaccard ≥ 0.8); bảng `chess_slug_aliases`.
- **i18n:** bỏ `zh-CN`, chuẩn hóa **vi-VN** làm ngôn ngữ chính.
- **Triển khai:** thêm bản **SQLite "lite"** + bộ **deploy Caddy/Hetzner/CI-CD** riêng (ngoài upstream).

## E. Backlog tùy biến

### Tiến độ thực thi plan "nối dây + đổ xăng" (2026-06-29)
- [x] **WS1** — Nối Puzzle Bank → `chess_generate_puzzle` (tool đọc DB qua `PuzzleSource`/`RandomPuzzle`, fallback embedded). Sửa `agent_service.go` (C1). Test: `chess_generate_puzzle_test.go`.
- [x] **WS3** — Mở rộng khai cuộc: nhúng `data/eco.tsv` (3733 khai cuộc, lichess CC0) qua `chess_openings_data.go`; `openingIndex` gộp dataset + overlay Việt hoá; `chess_lookup_opening.go` dùng index. Test: `chess_lookup_opening_test.go`.
- [x] **WS2 (chuẩn bị, CHƯA bật)** — `ReindexAll` (service+interface chess) + route `POST /chess/library/reindex` (router.go C1) + runbook `docs/chess-rag-enable.md`. Bật runtime do Thầy theo runbook.
- [x] **WS4b** — thông báo lỗi engine thân thiện (`friendlyEngineError` trong `chess_common.go`, áp cho analyze/best_move/explain_move); `httpEngine.Health()` probe + cảnh báo sớm lúc init trong `getChessEngine` (agent_service.go C1). Test: `chess_engine_error_test.go`. (Resolve fuzzy đã có test sẵn `chess_resolve_test.go`.) *Còn nợ:* endpoint `/chess/engine/health` (cần refactor engine thành service DI) — để sau.
- [ ] **WS4a** — áp thương hiệu (chờ file logo).

### Backlog cũ
- [ ] Áp nhận diện thương hiệu Dương Sinh (`#2B3990` + gold, logo) vào `frontend/`.
- [ ] (Tùy chọn) Bật `CHESS_KB_INDEX` full stack + nối KB "Tri thức cờ vua" vào agent HLV — **runbook đã có:** `docs/chess-rag-enable.md`.
- [ ] Nút "đổi tên slug" để dùng `chess_slug_aliases` (bảng đang trống).
- [ ] Khi merge upstream lần tới: ưu tiên rà C1 (móc lõi) + C4 (i18n/prompt) + C6 (migration sqlite).
