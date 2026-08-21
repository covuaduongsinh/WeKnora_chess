-- Rollback: 000073_chess_tags
-- Gỡ hệ thẻ thống nhất. Cột CSV `tags` của chess_positions/chess_books/
-- chess_articles KHÔNG bị đụng ở migration này nên rollback không mất dữ liệu
-- thẻ cũ — chỉ mất các thẻ gắn cho 6 loại chưa từng có cột CSV.

DO $$ BEGIN RAISE NOTICE '[Migration 000073] Reverting chess_tags schema'; END $$;

DROP INDEX IF EXISTS idx_chess_lessons_level;
DROP INDEX IF EXISTS idx_chess_puzzles_level;
DROP INDEX IF EXISTS idx_chess_games_level;

ALTER TABLE chess_lessons DROP COLUMN IF EXISTS level;
ALTER TABLE chess_puzzles DROP COLUMN IF EXISTS level;
ALTER TABLE chess_games   DROP COLUMN IF EXISTS level;

DROP INDEX IF EXISTS idx_chess_tag_items_tag;
DROP INDEX IF EXISTS idx_chess_tag_items_target;
DROP TABLE IF EXISTS chess_tag_items;

DROP INDEX IF EXISTS idx_chess_tags_kind;
DROP INDEX IF EXISTS idx_chess_tags_tenant_slug;
DROP TABLE IF EXISTS chess_tags;
