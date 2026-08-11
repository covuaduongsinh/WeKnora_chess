package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_library_article_topic.go bổ sung nghiệp vụ chuyên mục bài viết (cây
// TỐI ĐA 2 TẦNG) vào CÙNG struct chessLibraryService — không phải service
// riêng, để container.go/router.go chỉ mở rộng chỗ có sẵn.
//
// Ràng buộc 2 tầng ép Ở ĐÂY (không phải DB): một chuyên mục CHỈ được làm cha
// nếu bản thân nó KHÔNG có cha (tức là chuyên mục gốc) — nên không bao giờ có
// tầng thứ 3. Đây là quyết định thiết kế: với vài trăm bài viết, 2 tầng đã đủ
// phân loại; cây sâu hơn chỉ mời gọi ngồi sắp xếp lại thay vì viết.

func (s *chessLibraryService) ListArticleTopics(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) ([]*types.ChessArticleTopic, error) {
	topics, err := s.repo.ListArticleTopics(ctx, tenantID, f)
	if err != nil {
		return nil, err
	}
	for _, t := range topics {
		if n, err := s.repo.CountArticlesOnTopic(ctx, tenantID, t.ID); err == nil {
			t.ArticleCount = n
		}
	}
	return topics, nil
}

func (s *chessLibraryService) GetArticleTopic(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleTopic, error) {
	t, err := s.repo.GetArticleTopic(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if n, err := s.repo.CountArticlesOnTopic(ctx, tenantID, id); err == nil {
		t.ArticleCount = n
	}
	return t, nil
}

// GetArticleTopicBySlug giải mã slug chuyên mục. Chuyên mục KHÔNG phải đích
// wikilink (chỉ điều hướng UI nội bộ) nên KHÔNG cần alias/fuzzy resolve.
func (s *chessLibraryService) GetArticleTopicBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticleTopic, error) {
	return s.repo.GetArticleTopicBySlug(ctx, tenantID, slug)
}

// validateTopicParent kiểm ràng buộc 2 tầng: parentID rỗng luôn hợp lệ (tạo
// chuyên mục gốc); nếu có, chuyên mục cha PHẢI tồn tại VÀ bản thân nó phải là
// chuyên mục GỐC (parent_id rỗng) — chặn lồng tầng thứ 3.
func (s *chessLibraryService) validateTopicParent(ctx context.Context, tenantID uint64, parentID string) error {
	if parentID == "" {
		return nil
	}
	parent, err := s.repo.GetArticleTopic(ctx, tenantID, parentID)
	if err != nil {
		return fmt.Errorf("chuyên mục cha không tồn tại")
	}
	if parent.ParentID != "" {
		return fmt.Errorf("chỉ chuyên mục GỐC mới được làm cha (tối đa 2 tầng)")
	}
	return nil
}

func (s *chessLibraryService) CreateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) (*types.ChessArticleTopic, error) {
	if strings.TrimSpace(topic.Title) == "" {
		return nil, fmt.Errorf("tên chuyên mục không được để trống")
	}
	if err := s.validateTopicParent(ctx, topic.TenantID, topic.ParentID); err != nil {
		return nil, err
	}
	topic.ID = uuid.New().String()
	slug, err := ensureUniqueChessSlug(ctx, topic.TenantID, topicSlugBase(topic), topic.ID, s.repo.ArticleTopicSlugExists)
	if err != nil {
		return nil, err
	}
	topic.Slug = slug
	if err := s.repo.CreateArticleTopic(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

func (s *chessLibraryService) UpdateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) (*types.ChessArticleTopic, error) {
	existing, err := s.repo.GetArticleTopic(ctx, topic.TenantID, topic.ID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(topic.Title) == "" {
		return nil, fmt.Errorf("tên chuyên mục không được để trống")
	}
	if err := s.repo.UpdateArticleTopic(ctx, topic); err != nil {
		return nil, err
	}
	// Đổi cha (nếu có) — kiểm ràng buộc 2 tầng + chặn tự làm cha của chính mình.
	if topic.ParentID != existing.ParentID {
		if topic.ParentID == topic.ID {
			return nil, fmt.Errorf("chuyên mục không thể là cha của chính nó")
		}
		if err := s.validateTopicParent(ctx, topic.TenantID, topic.ParentID); err != nil {
			return nil, err
		}
		// Nếu bản thân topic đang LÀ cha của chuyên mục khác (có con), không
		// cho nó trở thành con (tránh tầng 3 gián tiếp).
		if n, _ := s.repo.CountArticleTopicChildren(ctx, topic.TenantID, topic.ID); n > 0 && topic.ParentID != "" {
			return nil, fmt.Errorf("chuyên mục đang có chuyên mục con — không thể trở thành chuyên mục con của mục khác")
		}
		if err := s.repo.UpdateArticleTopicParent(ctx, topic.TenantID, topic.ID, topic.ParentID); err != nil {
			return nil, err
		}
	}
	return s.GetArticleTopic(ctx, topic.TenantID, topic.ID)
}

// RenameArticleTopicSlug đổi slug chuyên mục. Không ghi alias (chuyên mục
// không phải đích wikilink).
func (s *chessLibraryService) RenameArticleTopicSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessArticleTopic, error) {
	t, err := s.repo.GetArticleTopic(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	base := slugifyChess(newSlug)
	if base == "" {
		return nil, fmt.Errorf("slug mới không hợp lệ (cần chứa chữ/số)")
	}
	if base == t.Slug {
		return t, nil
	}
	unique, err := ensureUniqueChessSlug(ctx, tenantID, base, t.ID, s.repo.ArticleTopicSlugExists)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateArticleTopicSlug(ctx, tenantID, id, unique); err != nil {
		return nil, err
	}
	return s.GetArticleTopic(ctx, tenantID, id)
}

// DeleteArticleTopic xóa chuyên mục — CHẶN nếu còn chuyên mục CON (tránh mồ
// côi tầng 2 mà không có gì làm cha); gỡ pivot bài viết trước khi xóa (bài
// viết KHÔNG bị xóa, chỉ gỡ khỏi chuyên mục này).
func (s *chessLibraryService) DeleteArticleTopic(ctx context.Context, tenantID uint64, id string) error {
	if n, err := s.repo.CountArticleTopicChildren(ctx, tenantID, id); err == nil && n > 0 {
		return fmt.Errorf("chuyên mục còn %d chuyên mục con — xóa chuyên mục con trước", n)
	}
	if err := s.repo.RemoveTopicItems(ctx, tenantID, id); err != nil {
		return err
	}
	return s.repo.DeleteArticleTopic(ctx, tenantID, id)
}

// SetTopicArticles GHI ĐÈ toàn bộ danh sách bài viết trong một chuyên mục
// theo đúng thứ tự articleIDs truyền vào.
func (s *chessLibraryService) SetTopicArticles(ctx context.Context, tenantID uint64, topicID string, articleIDs []string) error {
	if _, err := s.repo.GetArticleTopic(ctx, tenantID, topicID); err != nil {
		return fmt.Errorf("chuyên mục không tồn tại")
	}
	return s.repo.SetTopicArticles(ctx, tenantID, topicID, articleIDs)
}
