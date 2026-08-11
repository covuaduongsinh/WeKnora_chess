package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

// chess_article_topic.go bổ sung thao tác lưu trữ cho chuyên mục bài viết
// (chess_article_topics + pivot chess_article_topic_items) vào CÙNG struct
// chessLibraryRepository — không phải repository riêng, để container.go
// không cần Provide mới. Chuyên mục là cây TỐI ĐA 2 TẦNG qua ParentID (ép ở
// tầng service), KHÔNG phải đích wikilink (chỉ điều hướng UI, giống
// chess_shelves).

// ---- Chuyên mục ----

func (r *chessLibraryRepository) topicQuery(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.ParentIDSet {
		q = q.Where("parent_id = ?", f.ParentID)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("slug ILIKE ? OR title ILIKE ?", kw, kw)
	}
	return q
}

func (r *chessLibraryRepository) ListArticleTopics(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) ([]*types.ChessArticleTopic, error) {
	var topics []*types.ChessArticleTopic
	err := r.topicQuery(ctx, tenantID, f).Order("sort_order ASC, created_at DESC").Find(&topics).Error
	return topics, err
}

func (r *chessLibraryRepository) GetArticleTopic(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleTopic, error) {
	var t types.ChessArticleTopic
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *chessLibraryRepository) GetArticleTopicBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticleTopic, error) {
	var t types.ChessArticleTopic
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// ArticleTopicSlugs trả toàn bộ slug chuyên mục "sống" của tenant.
func (r *chessLibraryRepository) ArticleTopicSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND slug <> ''", tenantID).
		Pluck("slug", &slugs).Error
	return slugs, err
}

func (r *chessLibraryRepository) ArticleTopicSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessLibraryRepository) CreateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

// UpdateArticleTopic cố ý KHÔNG đụng slug/parent_id — đổi slug đi qua
// UpdateArticleTopicSlug; parent_id đổi qua UpdateArticleTopicParent (tách
// riêng vì service cần kiểm ràng buộc 2 tầng trước khi cho phép).
func (r *chessLibraryRepository) UpdateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND id = ?", topic.TenantID, topic.ID).
		Updates(map[string]interface{}{
			"title": topic.Title, "description": topic.Description, "sort_order": topic.SortOrder,
		}).Error
}

func (r *chessLibraryRepository) UpdateArticleTopicSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error
}

func (r *chessLibraryRepository) UpdateArticleTopicParent(ctx context.Context, tenantID uint64, id, parentID string) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("parent_id", parentID).Error
}

func (r *chessLibraryRepository) DeleteArticleTopic(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessArticleTopic{}).Error
}

func (r *chessLibraryRepository) CountArticlesOnTopic(ctx context.Context, tenantID uint64, topicID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessArticleTopicItem{}).
		Where("tenant_id = ? AND topic_id = ?", tenantID, topicID).Count(&count).Error
	return count, err
}

// CountArticleTopicChildren đếm số chuyên mục CON của một chuyên mục — dùng
// để chặn xóa chuyên mục còn con (giữ đúng ràng buộc cây 2 tầng, tránh mồ côi).
func (r *chessLibraryRepository) CountArticleTopicChildren(ctx context.Context, tenantID uint64, parentID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessArticleTopic{}).
		Where("tenant_id = ? AND parent_id = ?", tenantID, parentID).Count(&count).Error
	return count, err
}

// ---- Nối chuyên mục↔bài viết (nhiều-nhiều) ----

// SetTopicArticles xóa toàn bộ liên kết cũ của MỘT chuyên mục rồi chèn lại
// theo đúng thứ tự articleIDs (transaction) — cùng khuôn SetShelfBooks (gán
// TỪ PHÍA chuyên mục, đơn giản hơn gán từ phía bài viết vì không phải diff
// nhiều chuyên mục cùng lúc).
func (r *chessLibraryRepository) SetTopicArticles(ctx context.Context, tenantID uint64, topicID string, articleIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND topic_id = ?", tenantID, topicID).
			Delete(&types.ChessArticleTopicItem{}).Error; err != nil {
			return err
		}
		rows := make([]types.ChessArticleTopicItem, 0, len(articleIDs))
		for i, aid := range articleIDs {
			if aid == "" {
				continue
			}
			rows = append(rows, types.ChessArticleTopicItem{TenantID: tenantID, TopicID: topicID, ArticleID: aid, SortOrder: i})
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

// RemoveArticleFromAllTopics gỡ một bài viết khỏi mọi chuyên mục (khi xóa bài viết).
func (r *chessLibraryRepository) RemoveArticleFromAllTopics(ctx context.Context, tenantID uint64, articleID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND article_id = ?", tenantID, articleID).
		Delete(&types.ChessArticleTopicItem{}).Error
}

// RemoveTopicItems xóa toàn bộ liên kết của một chuyên mục (khi xóa chuyên
// mục) — KHÔNG đụng bài viết.
func (r *chessLibraryRepository) RemoveTopicItems(ctx context.Context, tenantID uint64, topicID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND topic_id = ?", tenantID, topicID).
		Delete(&types.ChessArticleTopicItem{}).Error
}
