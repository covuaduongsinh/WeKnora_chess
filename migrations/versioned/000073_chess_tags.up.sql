-- Migration: 000073_chess_tags
-- Description: Hệ THẺ + phân loại THỐNG NHẤT cho toàn bộ lớp cờ vua.
--
--              VẤN ĐỀ ĐANG SỬA: trước migration này, `tags` là chuỗi CSV
--              VARCHAR(255) và CHỈ tồn tại ở 3/9 loại (chess_positions:26,
--              chess_books:57, chess_articles:42). Ván cờ, bài tập, bài giảng,
--              khóa học, chương, kệ KHÔNG có thẻ nào. Lọc theo thẻ phải dùng
--              `tags ILIKE '%kw%'` — không index, và khớp SUBSTRING nên thẻ
--              "Mã" (slug "ma") kéo về cả "mate", "tham", "nam"… Cộng thêm:
--              "Khai cuộc"/"khai-cuoc"/"khai cuoc" thành 3 thẻ khác nhau,
--              không gộp lại được.
--
--              GIẢI PHÁP: bảng thẻ thật `chess_tags` + bảng nối ĐA HÌNH
--              `chess_tag_items` phủ MỌI loại nội dung cờ. Bảng nối là NGUỒN
--              SỰ THẬT; 3 cột CSV `tags` cũ giữ lại làm bản HIỂN THỊ/tương
--              thích (service ghi lại từ pivot sau mỗi lần lưu, KHÔNG bao giờ
--              đọc ngược) — đúng tiền lệ chess_articles.aliases ↔
--              chess_slug_aliases (migration 000072).
--
--              BA TRỤC PHÂN LOẠI sau migration này:
--                1. Cấp độ  — cột `level`, từ vựng ĐÓNG 6 bậc Dương Sinh
--                             (tot|ma|tuong|xe|hau|vua). Đã có ở courses/
--                             positions/books/chapters/articles; migration này
--                             thêm cho games/puzzles/lessons.
--                2. Nhóm nội dung — THẺ HỆ THỐNG (chess_tags.kind='group'),
--                             từ vựng ĐÓNG 8 nhóm theo .claude/memory/
--                             02-mien-co-vua.md §2.1. CỐ Ý không làm thành cột
--                             riêng trên 9 bảng: đi qua chính hệ thẻ thì chỉ
--                             có MỘT cơ chế lưu trữ, MỘT bảng nối, MỘT đường
--                             lọc — và thêm loại nội dung thứ 10 sau này = 0
--                             thay đổi schema.
--                3. Thẻ tự do — chess_tags.kind='free'.
--
--              Trường phân loại cũ (category/theme/phase/part) GIỮ NGUYÊN,
--              không xóa, không đổi.
--
--              KHÔNG có backfill trong file SQL này — CỐ Ý. Tách CSV cũ thành
--              thẻ đòi slug hóa có KHỬ DẤU tiếng Việt; viết lại logic đó bằng
--              SQL sẽ nhân đôi foldVN/slugifyChess (internal/application/
--              service/chess_slug.go) và chắc chắn trôi lệch. Backfill nằm ở
--              tầng service, chạy idempotent qua POST /chess/tags/backfill.

DO $$ BEGIN RAISE NOTICE '[Migration 000073] Applying chess_tags schema'; END $$;

-- Từ điển thẻ. slug là dạng đã slug hóa + KHỬ DẤU (slugifyChess) nên
-- "Khai cuộc", "khai-cuoc", "KHAI CUOC" cùng quy về một thẻ; name giữ nguyên
-- dấu để hiển thị.
CREATE TABLE IF NOT EXISTS chess_tags (
    id           VARCHAR(36) PRIMARY KEY,
    tenant_id    BIGINT NOT NULL,
    slug         VARCHAR(64)  NOT NULL DEFAULT '',
    name         VARCHAR(128) NOT NULL DEFAULT '', -- "Khai cuộc" (có dấu, để hiển thị)
    kind         VARCHAR(16)  NOT NULL DEFAULT 'free', -- 'group' (hệ thống, 8 nhóm, không xóa được) | 'free'
    description  VARCHAR(500) NOT NULL DEFAULT '',
    color        VARCHAR(16)  NOT NULL DEFAULT '',
    usage_count  INTEGER      NOT NULL DEFAULT 0, -- CACHE đếm số mục đang gắn; nguồn thật là chess_tag_items
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_tags_tenant_slug
    ON chess_tags (tenant_id, slug) WHERE slug <> '';
CREATE INDEX IF NOT EXISTS idx_chess_tags_kind
    ON chess_tags (tenant_id, kind, sort_order);

-- Bảng nối ĐA HÌNH thẻ↔nội dung (mẫu chess_article_topic_items / chess_shelf_books).
-- chess_type để VARCHAR(16) TỰ DO, KHÔNG CHECK/enum — đúng tiền lệ
-- wiki_chess_refs.chess_type (000065) và chess_kb_index (000069), nên thêm
-- loại nội dung mới KHÔNG cần migration.
CREATE TABLE IF NOT EXISTS chess_tag_items (
    tenant_id    BIGINT NOT NULL,
    tag_id       VARCHAR(36) NOT NULL,
    chess_type   VARCHAR(16) NOT NULL, -- game|puzzle|lesson|course|position|book|chapter|article
    chess_id     VARCHAR(36) NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tag_id, chess_type, chess_id)
);

-- "thẻ của mục này" (hiển thị chip trên từng hàng danh sách, xóa cascade)
CREATE INDEX IF NOT EXISTS idx_chess_tag_items_target
    ON chess_tag_items (tenant_id, chess_type, chess_id);
-- "mọi mục mang thẻ này" (lọc theo thẻ + trang mục lục ngang xuyên loại)
CREATE INDEX IF NOT EXISTS idx_chess_tag_items_tag
    ON chess_tag_items (tenant_id, tag_id, chess_type, sort_order);

-- Trục CẤP ĐỘ cho 3 bảng còn thiếu (courses/positions/books/chapters/articles
-- đã có sẵn). Từ vựng: tot|ma|tuong|xe|hau|vua — rỗng = mọi cấp.
ALTER TABLE chess_games   ADD COLUMN IF NOT EXISTS level VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE chess_puzzles ADD COLUMN IF NOT EXISTS level VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE chess_lessons ADD COLUMN IF NOT EXISTS level VARCHAR(16) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_chess_games_level   ON chess_games   (tenant_id, level);
CREATE INDEX IF NOT EXISTS idx_chess_puzzles_level ON chess_puzzles (tenant_id, level);
CREATE INDEX IF NOT EXISTS idx_chess_lessons_level ON chess_lessons (tenant_id, level);
