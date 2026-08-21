-- Rollback: 000912_chess_search_text
-- Gỡ cột search_text + index kèm theo. KHÔI PHỤC lại hai index trigram trên
-- lower(title) đúng như trước (dù chúng chưa từng được planner chọn — rollback
-- phải trả schema về đúng trạng thái cũ, không phải trạng thái "tốt hơn").

DO $$ BEGIN RAISE NOTICE '[Migration 000912] Reverting chess search_text columns'; END $$;

DROP INDEX IF EXISTS idx_chess_positions_fulltext;
DROP INDEX IF EXISTS idx_chess_lessons_fulltext;

CREATE INDEX IF NOT EXISTS idx_chess_chapters_title_trgm
    ON chess_book_chapters USING GIN (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_chess_articles_title_trgm
    ON chess_articles USING GIN (lower(title) gin_trgm_ops);

DROP INDEX IF EXISTS idx_chess_lessons_search;
DROP INDEX IF EXISTS idx_chess_courses_search;
DROP INDEX IF EXISTS idx_chess_articles_search;
DROP INDEX IF EXISTS idx_chess_chapters_search;
DROP INDEX IF EXISTS idx_chess_books_search;
DROP INDEX IF EXISTS idx_chess_positions_search;
DROP INDEX IF EXISTS idx_chess_puzzles_search;
DROP INDEX IF EXISTS idx_chess_games_search;

ALTER TABLE chess_lessons       DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_courses       DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_articles      DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_book_chapters DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_books         DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_positions     DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_puzzles       DROP COLUMN IF EXISTS search_text;
ALTER TABLE chess_games         DROP COLUMN IF EXISTS search_text;
