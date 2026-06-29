# 04 — Nhật ký & Inventory tùy biến (vs upstream)

> **Mục đích:** theo dõi MỌI khác biệt so với `Tencent/WeKnora` để quản lý `git merge upstream` và bàn giao.
> **Quy tắc cho AI agent:** khi sửa **file dùng chung của upstream** hoặc **thêm migration/schema**, PHẢI thêm dòng vào bảng phù hợp trong cùng PR. Thêm/sửa file `*chess*` thuần thì khuyến khích ghi.

## Thông tin fork
- **Upstream:** `https://github.com/Tencent/WeKnora` · **Nền:** v0.6.2 · **Nhánh:** `main` · **Production:** `weknora.covuaduongsinh.com`
- **Remote khuyến nghị:**
  ```bash
  git remote add upstream https://github.com/Tencent/WeKnora.git
  git fetch upstream
  # đọc CHANGELOG trước:  git merge upstream/main
  ```

---

## A. Inventory lớp cờ (code RIÊNG — rủi ro conflict THẤP, dễ giữ qua merge)
Đây là vùng tự thêm, hầu như không đụng upstream:

| Vùng | Đường dẫn |
|---|---|
| Engine | `internal/chess/` (board, engine, uci_engine, http_engine, *_test) |
| Agent tools | `internal/agent/tools/chess_*.go` (6 tool + common + demo_test) |
| Repository | `internal/application/repository/chess_*`, `wiki_chess_ref.go` |
| Service | `internal/application/service/chess_*` (course, knowledge_indexer, knowledge_text, library, resolve, slug) |
| Handler | `internal/handler/chess_*` (course, library, ref) |
| Types | `internal/types/(interfaces/)chess_*`, `wiki_chess_ref.go` |
| Frontend | `frontend/src/views/chess/**`, `views/chat/components/tool-results/ChessBoardDisplay.vue`, `api/chess/`, `stores/chessWikiDraft.ts`, `utils/chessBlocks.ts`, `utils/chessRef.ts` |
| Docker engine | `docker-compose.chess.yml`, `docker/Dockerfile.chess-engine`, `docker/chess-engine/uci_http_bridge.py` |
| Deploy/Docs/Scripts | `scripts/deploy/weknora-chess.service`, `scripts/seed_chess_wikilink_demo.*`, `docs/chess-wikilink-demo.md` |

## B. Migrations cờ (đánh số NỐI TIẾP upstream — rủi ro: trùng số nếu upstream thêm migration mới)
`000062`–`000069`: courses · games_puzzles · slugs · wiki_chess_refs · course_slug · refs_source_type · slug_aliases · kb_index.
> Khi merge upstream: nếu upstream thêm migration cùng dải số → **đổi số migration cờ cho cao hơn**, kiểm tra thứ tự áp dụng.

## C. File DÙNG CHUNG đã/sẽ phải sửa (rủi ro conflict CAO — GHI Ở ĐÂY)
Những chỗ buộc phải đụng code upstream để gắn lớp cờ (đăng ký route, đăng ký tool, router frontend, i18n agent…). **Cập nhật khi phát hiện/khi sửa:**

| Ngày | File dùng chung | Thay đổi | Lý do | Ghi chú merge |
|---|---|---|---|---|
| _(đã có)_ | `config/builtin_agents.yaml` | Thêm agent `builtin-chess-coach` (block cuối file) | Agent "HLV Cờ vua" + 6 tool cờ | Upstream có thể đổi schema agent → kiểm tra field khi merge |
| _(rà soát)_ | nơi đăng ký agent tools (Go) | Đăng ký 6 `chess_*` tool vào registry | Để agent gọi được tool | Tìm chỗ register tool; ghi file chính xác khi xác nhận |
| _(rà soát)_ | router API backend | Mount route `/api/v1/chess/...` | Expose handler cờ | Ghi file router chính xác khi xác nhận |
| _(rà soát)_ | router + menu frontend | Thêm route/menu khu "Quản lý cờ vua" | Điều hướng UI | Ghi file chính xác khi xác nhận |
| _(rà soát)_ | `.env.example` | Thêm `WEKNORA_CHESS_*`, `CHESS_KB_INDEX` | Tài liệu cấu hình | Dễ merge (chỉ thêm dòng) |
| | | | | |

> **Việc nên làm:** một AI agent rảnh có thể rà `git diff upstream/main --stat` để hoàn thiện danh sách C (các file không-`chess` bị đổi).

---

## D. Quyết định kiến trúc đã chốt (ADR rút gọn)
- **Engine:** Arasan (MIT) chạy **sidecar HTTP** (UCI→HTTP bridge), gọi qua `WEKNORA_CHESS_*`. Không nhúng engine vào process app.
- **Agent HLV:** `kb_selection_mode: none` — mặc định KHÔNG dùng RAG, chỉ dùng 6 tool cờ + engine. *(Nâng cấp tùy chọn: bật `CHESS_KB_INDEX` + thêm `knowledge_search` vào allowed_tools để trích dẫn lý thuyết/sách.)*
- **RAG cờ:** gate sau `CHESS_KB_INDEX` (mặc định TẮT) vì cần embedding + vector + worker. Import PGN hàng loạt KHÔNG trigger index (tránh "bão" embedding).
- **Wikilink:** slug bất biến; resolve `exact → alias → fuzzy` (bigram-Jaccard ≥ 0.8); bảng `chess_slug_aliases` cho đổi tên/redirect.

## E. Backlog tùy biến
- [ ] Hoàn thiện mục **C** (rà `git diff upstream/main` để liệt kê đủ file dùng chung đã sửa).
- [ ] Áp nhận diện thương hiệu Dương Sinh (`#2B3990` + gold, logo) vào `frontend/`.
- [ ] (Tùy chọn) Bật `CHESS_KB_INDEX` trên full stack + nối KB "Tri thức cờ vua" vào agent HLV.
- [ ] Nút "đổi tên slug" để tận dụng `chess_slug_aliases` (hiện bảng còn trống).
- [ ] Bộ test recall cho KB cờ sau khi bật RAG.
