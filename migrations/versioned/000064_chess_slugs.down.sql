-- Rollback: 000064_chess_slugs
DROP INDEX IF EXISTS idx_chess_lessons_tenant_slug;
DROP INDEX IF EXISTS idx_chess_puzzles_tenant_slug;
DROP INDEX IF EXISTS idx_chess_games_tenant_slug;
ALTER TABLE chess_lessons DROP COLUMN IF EXISTS slug;
ALTER TABLE chess_puzzles DROP COLUMN IF EXISTS slug;
ALTER TABLE chess_games   DROP COLUMN IF EXISTS slug;
