package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_article_topic.go bổ sung API chuyên mục bài viết vào CÙNG
// ChessLibraryHandler — router.go chỉ cần mở rộng RegisterChessLibraryRoutes
// đã có, không phải thêm hàm Register* mới.

// ListArticleTopics GET /chess/article-topics?parent_id=&q=
// parent_id="" (tham số CÓ MẶT nhưng rỗng) = lọc chuyên mục GỐC; tham số
// VẮNG MẶT = không lọc theo cha (trả cả gốc lẫn con).
func (h *ChessLibraryHandler) ListArticleTopics(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	f := types.ChessArticleTopicFilter{Keyword: c.Query("q")}
	if _, ok := c.GetQuery("parent_id"); ok {
		f.ParentIDSet = true
		f.ParentID = c.Query("parent_id")
	}
	topics, err := h.service.ListArticleTopics(ctx, tenantID, f)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, topics)
}

// GetArticleTopic GET /chess/article-topics/:id
func (h *ChessLibraryHandler) GetArticleTopic(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	t, err := h.service.GetArticleTopic(ctx, tenantID, c.Param("id"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, t)
}

// GetArticleTopicBySlug GET /chess/article-topics/by-slug/:slug
func (h *ChessLibraryHandler) GetArticleTopicBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	t, err := h.service.GetArticleTopicBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, t)
}

type articleTopicBody struct {
	Title       string `json:"title"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// CreateArticleTopic POST /chess/article-topics
func (h *ChessLibraryHandler) CreateArticleTopic(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b articleTopicBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	t, err := h.service.CreateArticleTopic(ctx, &types.ChessArticleTopic{
		TenantID: tenantID, Title: b.Title, ParentID: b.ParentID,
		Description: b.Description, SortOrder: b.SortOrder,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, t)
}

// UpdateArticleTopic PUT /chess/article-topics/:id
func (h *ChessLibraryHandler) UpdateArticleTopic(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b articleTopicBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	t, err := h.service.UpdateArticleTopic(ctx, &types.ChessArticleTopic{
		ID: c.Param("id"), TenantID: tenantID, Title: b.Title, ParentID: b.ParentID,
		Description: b.Description, SortOrder: b.SortOrder,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, t)
}

// RenameArticleTopicSlug PUT /chess/article-topics/:id/slug {slug}
func (h *ChessLibraryHandler) RenameArticleTopicSlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b slugBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	t, err := h.service.RenameArticleTopicSlug(ctx, tenantID, c.Param("id"), b.Slug)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, t)
}

// DeleteArticleTopic DELETE /chess/article-topics/:id — chặn nếu còn chuyên mục con.
func (h *ChessLibraryHandler) DeleteArticleTopic(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteArticleTopic(ctx, tenantID, c.Param("id")); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"deleted": true})
}

// SetTopicArticles PUT /chess/article-topics/:id/articles {article_ids:[...]}
// — ghi đè toàn bộ bài viết trong chuyên mục theo đúng thứ tự truyền vào.
func (h *ChessLibraryHandler) SetTopicArticles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b struct {
		ArticleIDs []string `json:"article_ids"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	if err := h.service.SetTopicArticles(ctx, tenantID, c.Param("id"), b.ArticleIDs); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"saved": true})
}
