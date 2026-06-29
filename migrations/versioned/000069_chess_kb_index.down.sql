-- Rollback: 000069_chess_kb_index
DROP INDEX IF EXISTS idx_chess_kb_index_knowledge;
DROP INDEX IF EXISTS idx_chess_kb_index_edge;
DROP TABLE IF EXISTS chess_kb_index;
