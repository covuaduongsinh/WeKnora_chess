-- Migration: 000064_chess_slugs
-- Description: Thêm cột slug (định danh thân thiện, duy nhất theo tenant) cho
--              chess_games, chess_puzzles, chess_lessons để hỗ trợ wikilink
--              [[game/<slug>]], [[puzzle/<slug>]], [[lesson/<slug>]].

DO $$ BEGIN RAISE NOTICE '[Migration 000064] Adding slug columns to chess objects'; END $$;

ALTER TABLE chess_games   ADD COLUMN IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE chess_puzzles ADD COLUMN IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE chess_lessons ADD COLUMN IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT '';

-- Backfill an toàn cho hàng cũ: tiền tố loại + 8 hex đầu của UUID. Chỉ nhằm
-- bảo đảm slug KHÔNG rỗng và duy nhất để dựng được unique index; slug "đẹp"
-- (humanize) áp dụng cho hàng MỚI ở tầng service.
UPDATE chess_games   SET slug = 'g-' || left(replace(id, '-', ''), 8) WHERE slug = '';
UPDATE chess_puzzles SET slug = 'p-' || left(replace(id, '-', ''), 8) WHERE slug = '';
UPDATE chess_lessons SET slug = 'l-' || left(replace(id, '-', ''), 8) WHERE slug = '';

-- Duy nhất theo (tenant_id, slug) cho mỗi loại. Partial index loại trừ chuỗi
-- rỗng để hàng vừa thêm cột (chưa gán slug) không vi phạm ràng buộc.
CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_games_tenant_slug
    ON chess_games (tenant_id, slug) WHERE slug <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_puzzles_tenant_slug
    ON chess_puzzles (tenant_id, slug) WHERE slug <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_lessons_tenant_slug
    ON chess_lessons (tenant_id, slug) WHERE slug <> '';
