package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

// chess_article.go bổ sung thao tác lưu trữ cho "Ngân hàng bài viết"
// (chess_articles) vào CÙNG struct chessLibraryRepository (khai báo ở
// chess_library.go) — article là "ngăn" thứ năm cạnh games/puzzles/positions/
// thư viện sách, không phải repository riêng, để container.go không cần
// Provide mới.
//
// PHASE 1: chỉ CRUD bài viết (chưa có chuyên mục/ảnh/lịch sử phiên bản — xem
// backlog trong .claude/memory/04-nhat-ky-tuy-bien.md).

// articleQuery dựng query lọc chung cho List/Export/tìm trùng — cùng khuôn
// positionQuery/bookQuery.
func (r *chessLibraryRepository) articleQuery(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Level != "" {
		q = q.Where("level = ?", f.Level)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("slug ILIKE ? OR title ILIKE ? OR aliases ILIKE ? OR summary ILIKE ? OR tags ILIKE ?",
			kw, kw, kw, kw, kw)
	}
	if f.TopicID != "" {
		q = q.Joins("JOIN chess_article_topic_items ti ON ti.article_id = chess_articles.id "+
			"AND ti.tenant_id = ? AND ti.topic_id = ?", tenantID, f.TopicID)
	}
	return q
}

func (r *chessLibraryRepository) ListArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]*types.ChessArticle, error) {
	var articles []*types.ChessArticle
	q := r.articleQuery(ctx, tenantID, f).Model(&types.ChessArticle{})
	if f.TopicID != "" {
		q = q.Order("chess_articles.sort_order ASC, chess_articles.created_at DESC")
	} else {
		q = q.Order("sort_order ASC, created_at DESC")
	}
	err := q.Limit(500).Find(&articles).Error
	return articles, err
}

// SearchArticles tìm bài viết theo từ khóa (slug/title/aliases/summary, VÀ
// content trên Postgres qua full-text) — dùng cho autocomplete wikilink, cùng
// khuôn SearchChapters. Chỉ chọn cột nhẹ (không tải Content).
func (r *chessLibraryRepository) SearchArticles(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessArticle, error) {
	if limit <= 0 {
		limit = 10
	}
	q := r.db.WithContext(ctx).Model(&types.ChessArticle{}).
		Select("id", "tenant_id", "slug", "title", "summary", "aliases", "category", "level", "created_at").
		Where("tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		if r.db.Dialector != nil && r.db.Dialector.Name() == "postgres" {
			q = q.Where("slug ILIKE ? OR title ILIKE ? OR aliases ILIKE ? OR "+
				"to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(aliases,'') || ' ' || "+
				"coalesce(summary,'') || ' ' || coalesce(content,'')) @@ plainto_tsquery('simple', ?)",
				kw, kw, kw, keyword)
		} else {
			// SQLite ("lite"): chưa có tsvector/pg_trgm — fallback ILIKE thường trên content.
			q = q.Where("slug ILIKE ? OR title ILIKE ? OR aliases ILIKE ? OR content ILIKE ?", kw, kw, kw, kw)
		}
	}
	var articles []*types.ChessArticle
	err := q.Order("created_at DESC").Limit(limit).Find(&articles).Error
	return articles, err
}

func (r *chessLibraryRepository) GetArticle(ctx context.Context, tenantID uint64, id string) (*types.ChessArticle, error) {
	var a types.ChessArticle
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *chessLibraryRepository) GetArticleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticle, error) {
	var a types.ChessArticle
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// ArticleSlugs trả toàn bộ slug bài viết "sống" của tenant — pool ứng viên fuzzy-resolve.
func (r *chessLibraryRepository) ArticleSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Model(&types.ChessArticle{}).
		Where("tenant_id = ? AND slug <> ''", tenantID).
		Pluck("slug", &slugs).Error
	return slugs, err
}

func (r *chessLibraryRepository) ArticleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessArticle{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessLibraryRepository) CreateArticle(ctx context.Context, article *types.ChessArticle) error {
	return r.db.WithContext(ctx).Create(article).Error
}

// UpdateArticle cố ý KHÔNG đụng cột slug (như UpdatePosition/UpdateBook) — đổi
// slug đi qua UpdateArticleSlug để luôn kèm ghi alias ở tầng service.
func (r *chessLibraryRepository) UpdateArticle(ctx context.Context, article *types.ChessArticle) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessArticle{}).
		Where("tenant_id = ? AND id = ?", article.TenantID, article.ID).
		Updates(map[string]interface{}{
			"title": article.Title, "summary": article.Summary, "aliases": article.Aliases,
			"category": article.Category, "level": article.Level, "tags": article.Tags,
			"status": article.Status, "cover_url": article.CoverURL, "content": article.Content,
			"sort_order": article.SortOrder,
		}).Error
}

func (r *chessLibraryRepository) UpdateArticleSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessArticle{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error
}

func (r *chessLibraryRepository) DeleteArticle(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessArticle{}).Error
}

// ---- Ảnh chèn trong bài viết (sao y chess_book.go phần ảnh, book_id → article_id) ----

func (r *chessLibraryRepository) CreateArticleImage(ctx context.Context, img *types.ChessArticleImage) error {
	return r.db.WithContext(ctx).Create(img).Error
}

func (r *chessLibraryRepository) GetArticleImage(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleImage, error) {
	var img types.ChessArticleImage
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

// ListArticleImagesByArticle phục vụ xóa file vật lý khi cascade delete bài viết.
func (r *chessLibraryRepository) ListArticleImagesByArticle(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleImage, error) {
	var imgs []*types.ChessArticleImage
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND article_id = ?", tenantID, articleID).Find(&imgs).Error
	return imgs, err
}

func (r *chessLibraryRepository) DeleteArticleImage(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessArticleImage{}).Error
}

// ---- Lịch sử phiên bản bài viết (sao y chess_book.go phần lịch sử chương) ----

func (r *chessLibraryRepository) CreateArticleRevision(ctx context.Context, rev *types.ChessArticleRevision) error {
	return r.db.WithContext(ctx).Create(rev).Error
}

func (r *chessLibraryRepository) ListArticleRevisions(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleRevision, error) {
	var revs []*types.ChessArticleRevision
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND article_id = ?", tenantID, articleID).
		Order("revision_number DESC").
		Find(&revs).Error
	return revs, err
}

func (r *chessLibraryRepository) GetArticleRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessArticleRevision, error) {
	var rev types.ChessArticleRevision
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, revisionID).First(&rev).Error; err != nil {
		return nil, err
	}
	return &rev, nil
}

// CountArticleRevisions phục vụ tính revision_number tiếp theo (bắt đầu từ 1).
func (r *chessLibraryRepository) CountArticleRevisions(ctx context.Context, tenantID uint64, articleID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessArticleRevision{}).
		Where("tenant_id = ? AND article_id = ?", tenantID, articleID).Count(&count).Error
	return count, err
}

// DeleteArticleRevisionsByArticle xóa toàn bộ lịch sử của một bài viết (khi xóa bài viết).
func (r *chessLibraryRepository) DeleteArticleRevisionsByArticle(ctx context.Context, tenantID uint64, articleID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND article_id = ?", tenantID, articleID).
		Delete(&types.ChessArticleRevision{}).Error
}
