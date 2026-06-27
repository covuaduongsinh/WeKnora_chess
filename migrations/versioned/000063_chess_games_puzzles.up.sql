-- Migration: 000063_chess_games_puzzles
-- Description: Kho ván đấu (chess_games) và ngân hàng bài tập (chess_puzzles) cờ vua.

DO $$ BEGIN RAISE NOTICE '[Migration 000063] Applying chess_games + chess_puzzles schema'; END $$;

CREATE TABLE IF NOT EXISTS chess_games (
    id         VARCHAR(36) PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    white      VARCHAR(128) NOT NULL DEFAULT '',
    black      VARCHAR(128) NOT NULL DEFAULT '',
    result     VARCHAR(16) NOT NULL DEFAULT '',
    eco        VARCHAR(8) NOT NULL DEFAULT '',
    event      VARCHAR(255) NOT NULL DEFAULT '',
    date       VARCHAR(32) NOT NULL DEFAULT '',
    pgn        TEXT NOT NULL DEFAULT '',
    ply_count  INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_games_tenant ON chess_games (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chess_games_eco ON chess_games (tenant_id, eco);

CREATE TABLE IF NOT EXISTS chess_puzzles (
    id         VARCHAR(36) PRIMARY KEY,
    tenant_id  BIGINT NOT NULL,
    title      VARCHAR(255) NOT NULL DEFAULT '',
    fen        VARCHAR(128) NOT NULL,
    solution   VARCHAR(255) NOT NULL DEFAULT '',
    theme      VARCHAR(64) NOT NULL DEFAULT '',
    difficulty VARCHAR(32) NOT NULL DEFAULT '',
    source     VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_puzzles_tenant ON chess_puzzles (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chess_puzzles_filter ON chess_puzzles (tenant_id, theme, difficulty);
