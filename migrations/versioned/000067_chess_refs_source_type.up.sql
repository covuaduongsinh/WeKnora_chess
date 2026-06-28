-- Migration: 000067_chess_refs_source_type
-- Description: Cho phép BÀI GIẢNG (chess_lessons) làm nguồn tham chiếu cờ, không
--              chỉ trang wiki. Thêm cột source_type ('wiki' | 'lesson') vào
--              wiki_chess_refs. Bài giảng lưu source_type='lesson', kb_id='',
--              page_slug=<lesson slug>.

DO $$ BEGIN RAISE NOTICE '[Migration 000067] Adding source_type to wiki_chess_refs'; END $$;

ALTER TABLE wiki_chess_refs ADD COLUMN IF NOT EXISTS source_type VARCHAR(16) NOT NULL DEFAULT 'wiki';

-- Unique mới gồm source_type để (nguồn, đích) là duy nhất kể cả khi kb_id rỗng (lesson).
DROP INDEX IF EXISTS idx_wiki_chess_refs_edge;
CREATE UNIQUE INDEX IF NOT EXISTS idx_wiki_chess_refs_edge
    ON wiki_chess_refs (source_type, kb_id, page_slug, chess_type, chess_slug);

-- Tra cứu ref nguồn theo bài giảng (xóa/đồng bộ).
CREATE INDEX IF NOT EXISTS idx_wiki_chess_refs_source
    ON wiki_chess_refs (source_type, page_slug);
