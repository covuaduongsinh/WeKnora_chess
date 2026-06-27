-- Migration: 000062_chess_courses
-- Description: Bảng khóa học và bài học cờ vua (LMS) cho công ty Cờ vua Dương Sinh.
--              Khóa học (chess_courses) chứa nhiều bài học (chess_lessons).
--              Nội dung bài học là markdown, có thể nhúng khối ```chess (FEN/PGN).

DO $$ BEGIN RAISE NOTICE '[Migration 000062] Applying chess_courses + chess_lessons schema'; END $$;

CREATE TABLE IF NOT EXISTS chess_courses (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    level       VARCHAR(32) NOT NULL DEFAULT '',
    cover_url   VARCHAR(512) NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_courses_tenant
    ON chess_courses (tenant_id, sort_order);

CREATE TABLE IF NOT EXISTS chess_lessons (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    course_id   VARCHAR(36) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    fen         VARCHAR(128) NOT NULL DEFAULT '',
    pgn         TEXT NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_lessons_course
    ON chess_lessons (tenant_id, course_id, sort_order);
