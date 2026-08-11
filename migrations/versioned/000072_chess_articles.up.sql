-- Migration: 000072_chess_articles
-- Description: Ngân hàng bài viết (chess_articles) — thực thể cờ thứ 7 (bên
--              cạnh game/puzzle/lesson/course/position/book+chapter). Mỗi bài
--              là một trang tri thức ĐỘC LẬP về khái niệm/thuật ngữ/kinh
--              nghiệm — KHÔNG thuộc sách hay khóa học nào — vừa là ĐÍCH
--              wikilink [[article/<slug>]] vừa là NGUỒN (nội dung có thể chứa
--              wikilink trỏ sang ván/thế cờ/bài tập/sách khác).
--
--              Tổ chức = 2 trục vuông góc: chess_article_topics (cây chuyên
--              mục tối đa 2 tầng qua parent_id, KHÔNG phải đích wikilink —
--              giống chess_shelves) cho điều hướng DỌC, cộng category/level/
--              tags/status cho lọc NGANG. chess_article_topic_items là bảng
--              nối nhiều-nhiều (mẫu chess_shelf_books).
--
--              chess_article_images sao y chess_book_images (ảnh chèn trong
--              bài, URL ổn định). chess_article_revisions sao y
--              chess_chapter_revisions (pre-image, chỉ tạo khi Title/Content
--              đổi).
--
--              Chỉ bài status='published' được index vào KB tri thức cờ (bản
--              thảo không rò vào câu trả lời agent) — cùng quy tắc sách.
--
--              Bí danh/từ đồng nghĩa: thêm cột `kind` vào chess_slug_aliases
--              (bảng dùng chung, migration 000068) để phân biệt alias sinh do
--              ĐỔI SLUG ('rename', không bao giờ xóa — link cũ trong sách sẽ
--              gãy) với alias người dùng GÕ TAY ('synonym', sửa/xóa tự do qua
--              UI). Alias cũ (trước migration này) mặc định 'rename' — đúng
--              vì mọi alias hiện có đều sinh từ đổi slug.

DO $$ BEGIN RAISE NOTICE '[Migration 000072] Applying chess_articles schema'; END $$;

CREATE TABLE IF NOT EXISTS chess_articles (
    id           VARCHAR(36) PRIMARY KEY,
    tenant_id    BIGINT NOT NULL,
    slug         VARCHAR(255)  NOT NULL DEFAULT '',
    title        VARCHAR(255)  NOT NULL DEFAULT '',
    summary      VARCHAR(1000) NOT NULL DEFAULT '', -- tóm tắt ngắn (danh sách/popup wikilink/mồi RAG)
    aliases      VARCHAR(500)  NOT NULL DEFAULT '', -- "Pin, Đóng đinh" — DẠNG HIỂN THỊ, CSV; nguồn thật cho
                                                      -- resolve wikilink là chess_slug_aliases (kind='synonym')
    category     VARCHAR(32)   NOT NULL DEFAULT '', -- khai-niem|thuat-ngu|kinh-nghiem|huong-dan|phuong-phap-day|...
    level        VARCHAR(16)   NOT NULL DEFAULT '', -- tot|ma|tuong|xe|hau|vua (6 cấp Dương Sinh)
    tags         VARCHAR(255)  NOT NULL DEFAULT '', -- CSV nhẹ, khớp mẫu Tags của chess_positions
    status       VARCHAR(16)   NOT NULL DEFAULT 'draft', -- draft|published — chỉ published được index RAG
    cover_url    VARCHAR(512)  NOT NULL DEFAULT '',
    content      TEXT          NOT NULL DEFAULT '', -- markdown, CÓ THỂ chứa wikilink + khối ```chess
    sort_order   INTEGER       NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_articles_tenant_slug
    ON chess_articles (tenant_id, slug) WHERE slug <> '';
CREATE INDEX IF NOT EXISTS idx_chess_articles_tenant
    ON chess_articles (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chess_articles_filter
    ON chess_articles (tenant_id, category, level);
CREATE INDEX IF NOT EXISTS idx_chess_articles_status
    ON chess_articles (tenant_id, status);
-- Full-text + trigram cho tìm kiếm nội dung bài (mẫu chess_book_chapters,
-- migration 000071). pg_trgm đã cài ở migration 000002.
CREATE INDEX IF NOT EXISTS idx_chess_articles_fulltext
    ON chess_articles USING GIN (to_tsvector('simple',
        coalesce(title, '') || ' ' || coalesce(aliases, '') || ' ' ||
        coalesce(summary, '') || ' ' || coalesce(content, '')));
CREATE INDEX IF NOT EXISTS idx_chess_articles_title_trgm
    ON chess_articles USING GIN (lower(title) gin_trgm_ops);

-- Chuyên mục bài viết — cây TỐI ĐA 2 TẦNG qua parent_id (ép ở tầng service,
-- không CHECK ở DB). KHÔNG phải đích wikilink (chỉ điều hướng UI, giống
-- chess_shelves) — xem 04-nhat-ky-tuy-bien.md.
CREATE TABLE IF NOT EXISTS chess_article_topics (
    id           VARCHAR(36) PRIMARY KEY,
    tenant_id    BIGINT NOT NULL,
    parent_id    VARCHAR(36)  NOT NULL DEFAULT '', -- '' = chuyên mục gốc (tầng 1)
    slug         VARCHAR(255) NOT NULL DEFAULT '',
    title        VARCHAR(255) NOT NULL DEFAULT '',
    description  VARCHAR(1000) NOT NULL DEFAULT '',
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chess_article_topics_tenant_slug
    ON chess_article_topics (tenant_id, slug) WHERE slug <> '';
CREATE INDEX IF NOT EXISTS idx_chess_article_topics_parent
    ON chess_article_topics (tenant_id, parent_id, sort_order);

-- Bảng nối NHIỀU-NHIỀU chuyên mục↔bài viết (mẫu chess_shelf_books).
CREATE TABLE IF NOT EXISTS chess_article_topic_items (
    tenant_id    BIGINT NOT NULL,
    topic_id     VARCHAR(36) NOT NULL,
    article_id   VARCHAR(36) NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (topic_id, article_id)
);

CREATE INDEX IF NOT EXISTS idx_chess_article_topic_items_article
    ON chess_article_topic_items (tenant_id, article_id);
CREATE INDEX IF NOT EXISTS idx_chess_article_topic_items_topic
    ON chess_article_topic_items (tenant_id, topic_id, sort_order);

-- Ảnh chèn trong nội dung bài (upload qua POST /chess/articles/:id/images) —
-- sao y chess_book_images. `path` là provider:// path trả về từ FileService.
CREATE TABLE IF NOT EXISTS chess_article_images (
    id           VARCHAR(36) PRIMARY KEY,
    tenant_id    BIGINT NOT NULL,
    article_id   VARCHAR(36)  NOT NULL DEFAULT '',
    path         VARCHAR(512) NOT NULL DEFAULT '',
    file_name    VARCHAR(255) NOT NULL DEFAULT '',
    mime         VARCHAR(64)  NOT NULL DEFAULT '',
    size         BIGINT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_article_images_article
    ON chess_article_images (tenant_id, article_id);

-- Lịch sử phiên bản bài viết — sao y chess_chapter_revisions: lưu bản TRƯỚC
-- khi ghi đè (pre-image), chỉ khi Title/Content thực sự đổi.
CREATE TABLE IF NOT EXISTS chess_article_revisions (
    id               VARCHAR(36) PRIMARY KEY,
    tenant_id        BIGINT NOT NULL,
    article_id       VARCHAR(36)  NOT NULL DEFAULT '',
    revision_number  INTEGER      NOT NULL DEFAULT 0,
    title            VARCHAR(255) NOT NULL DEFAULT '',
    content          TEXT         NOT NULL DEFAULT '',
    summary          VARCHAR(255) NOT NULL DEFAULT '',
    created_by       VARCHAR(64)  NOT NULL DEFAULT '',
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chess_article_revisions_article
    ON chess_article_revisions (tenant_id, article_id, revision_number DESC);

-- ---- Bí danh/từ đồng nghĩa (chess_slug_aliases, bảng dùng chung 000068) ----
-- 'rename' = sinh tự động khi đổi slug (giữ VĨNH VIỄN, không xóa qua UI).
-- 'synonym' = người dùng gõ tay (sửa/xóa tự do qua ReplaceSynonyms).
ALTER TABLE chess_slug_aliases ADD COLUMN IF NOT EXISTS kind VARCHAR(16) NOT NULL DEFAULT 'rename';
CREATE INDEX IF NOT EXISTS idx_chess_slug_aliases_kind
    ON chess_slug_aliases (tenant_id, chess_type, kind, new_slug);
