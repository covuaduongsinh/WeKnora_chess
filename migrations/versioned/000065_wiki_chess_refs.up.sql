-- Migration: 000065_wiki_chess_refs
-- Description: Bảng liên kết wiki -> đối tượng cờ (game/puzzle/lesson) để hỗ trợ
--              backlink và đồ thị cho wikilink [[game/<slug>]].

DO $$ BEGIN RAISE NOTICE '[Migration 000065] Creating wiki_chess_refs'; END $$;

CREATE TABLE IF NOT EXISTS wiki_chess_refs (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    kb_id       VARCHAR(36) NOT NULL,
    page_slug   VARCHAR(255) NOT NULL,
    chess_type  VARCHAR(16) NOT NULL,   -- 'game' | 'puzzle' | 'lesson'
    chess_slug  VARCHAR(255) NOT NULL,  -- slug trần (không tiền tố)
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Một (trang, đích cờ) chỉ một dòng.
CREATE UNIQUE INDEX IF NOT EXISTS idx_wiki_chess_refs_edge
    ON wiki_chess_refs (kb_id, page_slug, chess_type, chess_slug);
-- Tra backlink theo đích cờ (trong phạm vi tenant).
CREATE INDEX IF NOT EXISTS idx_wiki_chess_refs_target
    ON wiki_chess_refs (tenant_id, chess_type, chess_slug);
-- Lấy toàn bộ tham chiếu cờ của một KB (dựng đồ thị).
CREATE INDEX IF NOT EXISTS idx_wiki_chess_refs_kb
    ON wiki_chess_refs (kb_id);
