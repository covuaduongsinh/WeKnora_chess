-- Rollback: 000067_chess_refs_source_type
DROP INDEX IF EXISTS idx_wiki_chess_refs_source;
DROP INDEX IF EXISTS idx_wiki_chess_refs_edge;
CREATE UNIQUE INDEX IF NOT EXISTS idx_wiki_chess_refs_edge
    ON wiki_chess_refs (kb_id, page_slug, chess_type, chess_slug);
ALTER TABLE wiki_chess_refs DROP COLUMN IF EXISTS source_type;
