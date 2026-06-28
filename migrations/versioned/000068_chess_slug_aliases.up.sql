-- Migration: 000068_chess_slug_aliases
-- Description: Bảng alias/redirect cho slug đối tượng cờ. Khi một slug đổi (đổi tên,
--              re-import dedup, humanize backfill), ghi (old_slug -> new_slug) để
--              wikilink [[game/<old>]] cũ vẫn giải mã đúng. Resolve theo thứ tự:
--              exact slug -> alias -> fuzzy.

DO $$ BEGIN RAISE NOTICE '[Migration 000068] Creating chess_slug_aliases'; END $$;

CREATE TABLE IF NOT EXISTS chess_slug_aliases (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    chess_type  VARCHAR(16) NOT NULL,   -- 'game' | 'puzzle' | 'lesson' | 'course'
    old_slug    VARCHAR(255) NOT NULL,  -- slug cũ (đích wikilink cũ cần giữ sống)
    new_slug    VARCHAR(255) NOT NULL,  -- slug hiện hành
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Một (tenant, loại, old_slug) chỉ trỏ tới một new_slug.
CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_slug_aliases_edge
    ON chess_slug_aliases (tenant_id, chess_type, old_slug);
