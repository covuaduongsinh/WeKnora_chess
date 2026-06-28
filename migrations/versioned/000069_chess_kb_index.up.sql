-- Migration: 000069_chess_kb_index
-- Description: Ánh xạ đối tượng cờ (game/puzzle/lesson) → bản ghi Knowledge đã được
--              index vào "KB tri thức cờ vua", để trợ lý/HLV truy hồi nội dung ván
--              cờ qua RAG. Một (tenant, loại, slug) ↔ một knowledge_id.
--              (Tính năng index gate sau env CHESS_KB_INDEX; bảng vẫn an toàn nếu tắt.)

DO $$ BEGIN RAISE NOTICE '[Migration 000069] Creating chess_kb_index'; END $$;

CREATE TABLE IF NOT EXISTS chess_kb_index (
    id           VARCHAR(36) PRIMARY KEY,
    tenant_id    BIGINT NOT NULL,
    chess_type   VARCHAR(16) NOT NULL,   -- 'game' | 'puzzle' | 'lesson'
    chess_slug   VARCHAR(255) NOT NULL,  -- slug trần
    knowledge_id VARCHAR(36) NOT NULL,   -- bản ghi Knowledge tương ứng
    kb_id        VARCHAR(36) NOT NULL,   -- KB chứa (KB tri thức cờ vua)
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_kb_index_edge
    ON chess_kb_index (tenant_id, chess_type, chess_slug);
CREATE INDEX IF NOT EXISTS idx_chess_kb_index_knowledge
    ON chess_kb_index (knowledge_id);
