-- Rollback: 000068_chess_slug_aliases
DROP INDEX IF EXISTS idx_chess_slug_aliases_edge;
DROP TABLE IF EXISTS chess_slug_aliases;
