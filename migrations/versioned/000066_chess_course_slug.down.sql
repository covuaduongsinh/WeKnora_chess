-- Rollback: 000066_chess_course_slug
DROP INDEX IF EXISTS idx_chess_courses_tenant_slug;
ALTER TABLE chess_courses DROP COLUMN IF EXISTS slug;
