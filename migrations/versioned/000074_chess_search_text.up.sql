-- Migration: 000074_chess_search_text
-- Description: Cột `search_text` KHỬ DẤU + index trigram cho mọi loại nội dung
--              cờ, và SỬA hai index trigram cũ vốn không bao giờ được dùng.
--
--              VẤN ĐỀ 1 — tìm kiếm không chạy được khi gõ không dấu.
--              Mọi ô tìm của lớp cờ đang là `col ILIKE '%kw%'` trên chuỗi gốc
--              có dấu, nên "tan cuoc" KHÔNG khớp "Tàn cuộc". Đây là cách gõ
--              phổ biến nhất khi tra nhanh trên điện thoại. PostgreSQL có
--              extension `unaccent` nhưng (a) chưa được cài trong repo này,
--              (b) `unaccent()` không IMMUTABLE nên không index thẳng được,
--              (c) bản chạy SQLite ("lite") không có nó. Giải pháp: cột
--              `search_text` do tầng Go dựng sẵn bằng chess.SearchText()
--              (hạ chữ thường + khử dấu, CÙNG một phép khử dấu với slug), rồi
--              truy vấn bằng LIKE — chạy giống hệt trên cả Postgres lẫn SQLite.
--
--              VẤN ĐỀ 2 — hai index trigram đang là gánh nặng chết.
--              000071:111 và 000072:66 tạo index trên BIỂU THỨC `lower(title)`,
--              trong khi truy vấn viết là `title ILIKE ?` (chess_book.go:283,
--              chess_article.go:69). Postgres chỉ dùng index biểu thức khi câu
--              truy vấn chứa ĐÚNG biểu thức đó, nên hai index này chưa từng
--              được chọn — vẫn seq scan, mà vẫn tốn ghi và tốn đĩa. Bỏ hẳn:
--              index trên `search_text` bên dưới thay thế đúng vai trò của
--              chúng và còn khử dấu.
--
--              Cột để rỗng sau khi migrate. Chạy POST /chess/tags/backfill
--              (nút "Nạp thẻ từ dữ liệu cũ") một lần để điền cho dữ liệu cũ;
--              bản ghi tạo/sửa từ đây trở đi được điền tự động ở repository.

DO $$ BEGIN RAISE NOTICE '[Migration 000074] Applying chess search_text columns'; END $$;

ALTER TABLE chess_games         ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_puzzles       ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_positions     ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_books         ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_book_chapters ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_articles      ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_courses       ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';
ALTER TABLE chess_lessons       ADD COLUMN IF NOT EXISTS search_text TEXT NOT NULL DEFAULT '';

-- Trigram trên CỘT TRẦN (không phải biểu thức): pg_trgm tự hạ chữ thường khi
-- tách trigram nên index này phục vụ được cả LIKE lẫn ILIKE, và quan trọng hơn
-- là planner khớp được vì truy vấn dùng đúng tên cột.
-- pg_trgm đã được cài vô điều kiện ở migration 000041.
CREATE INDEX IF NOT EXISTS idx_chess_games_search      ON chess_games         USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_puzzles_search    ON chess_puzzles       USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_positions_search  ON chess_positions     USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_books_search      ON chess_books         USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_chapters_search   ON chess_book_chapters USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_articles_search   ON chess_articles      USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_courses_search    ON chess_courses       USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_lessons_search    ON chess_lessons       USING GIN (search_text gin_trgm_ops);

-- Bỏ hai index trigram chưa từng được planner chọn (xem VẤN ĐỀ 2 ở trên).
DROP INDEX IF EXISTS idx_chess_chapters_title_trgm;
DROP INDEX IF EXISTS idx_chess_articles_title_trgm;

-- Full-text cho hai loại còn thiếu, để tìm SÂU trong nội dung dài (cột
-- search_text bị cắt ở 8000 ký tự nên không thay được vai trò này).
-- Cùng khuôn idx_chess_chapters_fulltext (000071) và idx_chess_articles_fulltext (000072).
CREATE INDEX IF NOT EXISTS idx_chess_lessons_fulltext
    ON chess_lessons USING GIN (to_tsvector('simple',
        coalesce(title, '') || ' ' || coalesce(content, '')));
CREATE INDEX IF NOT EXISTS idx_chess_positions_fulltext
    ON chess_positions USING GIN (to_tsvector('simple',
        coalesce(title, '') || ' ' || coalesce(annotation, '')));
