-- Migration: 000066_chess_course_slug
-- Description: Thêm cột slug (duy nhất theo tenant) cho chess_courses để khóa học
--              trở thành đích wikilink [[course/<slug>]].

DO $$ BEGIN RAISE NOTICE '[Migration 000066] Adding slug column to chess_courses'; END $$;

ALTER TABLE chess_courses ADD COLUMN IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT '';

-- Backfill an toàn cho hàng cũ: 'c-' + 8 hex đầu của UUID.
UPDATE chess_courses SET slug = 'c-' || left(replace(id, '-', ''), 8) WHERE slug = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_courses_tenant_slug
    ON chess_courses (tenant_id, slug) WHERE slug <> '';
