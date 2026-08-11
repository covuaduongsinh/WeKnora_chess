package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_article.go bổ sung API "Ngân hàng bài viết" vào CÙNG ChessLibraryHandler
// (khai báo ở chess_library.go) — article là "ngăn" thứ năm cạnh games/puzzles/
// positions/thư viện sách, không phải handler riêng, để router.go chỉ cần mở
// rộng RegisterChessLibraryRoutes đã có thay vì thêm hàm Register* mới.

// ListArticles GET /chess/articles?category=&level=&status=&q=&topic_id=
func (h *ChessLibraryHandler) ListArticles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	articles, err := h.service.ListArticles(ctx, tenantID, types.ChessArticleFilter{
		TopicID: c.Query("topic_id"), Category: c.Query("category"), Level: c.Query("level"),
		Status: c.Query("status"), Keyword: c.Query("q"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, articles)
}

// GetArticle GET /chess/articles/:id
func (h *ChessLibraryHandler) GetArticle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	a, err := h.service.GetArticle(ctx, tenantID, c.Param("id"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, a)
}

// GetArticleBySlug GET /chess/articles/by-slug/:slug — giải mã wikilink [[article/<slug>]].
func (h *ChessLibraryHandler) GetArticleBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	a, err := h.service.GetArticleBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, a)
}

// GetArticleBacklinks GET /chess/articles/by-slug/:slug/backlinks — trang wiki/
// bài giảng/thế cờ/chương/bài viết khác trỏ tới bài viết này.
func (h *ChessLibraryHandler) GetArticleBacklinks(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	links, err := h.service.GetArticleBacklinks(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, links)
}

type articleBody struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Aliases  string `json:"aliases"`
	Category string `json:"category"`
	Level    string `json:"level"`
	Tags     string `json:"tags"`
	Status   string `json:"status"`
	CoverURL string `json:"cover_url"`
	Content  string `json:"content"`
}

func articleFromBody(id string, tenantID uint64, b articleBody) *types.ChessArticle {
	return &types.ChessArticle{
		ID: id, TenantID: tenantID, Title: b.Title, Summary: b.Summary, Aliases: b.Aliases,
		Category: b.Category, Level: b.Level, Tags: b.Tags, Status: b.Status,
		CoverURL: b.CoverURL, Content: b.Content,
	}
}

// CreateArticle POST /chess/articles
func (h *ChessLibraryHandler) CreateArticle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b articleBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	a, err := h.service.CreateArticle(ctx, articleFromBody("", tenantID, b))
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, a)
}

// UpdateArticle PUT /chess/articles/:id
func (h *ChessLibraryHandler) UpdateArticle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b articleBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	a, err := h.service.UpdateArticle(ctx, articleFromBody(c.Param("id"), tenantID, b))
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, a)
}

// RenameArticleSlug PUT /chess/articles/:id/slug {slug} — đổi slug bài viết, giữ link cũ qua alias.
func (h *ChessLibraryHandler) RenameArticleSlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b slugBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	a, err := h.service.RenameArticleSlug(ctx, tenantID, c.Param("id"), b.Slug)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, a)
}

// DeleteArticle DELETE /chess/articles/:id
func (h *ChessLibraryHandler) DeleteArticle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteArticle(ctx, tenantID, c.Param("id")); err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, gin.H{"deleted": true})
}

// ExportArticles GET /chess/articles/export?category=&level=&status= — danh sách bài viết (JSON).
func (h *ChessLibraryHandler) ExportArticles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	items, err := h.service.ExportArticles(ctx, tenantID, types.ChessArticleFilter{
		Category: c.Query("category"), Level: c.Query("level"), Status: c.Query("status"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, items)
}

// ImportArticles POST /chess/articles/import {articles:[...]} — tạo mới; trả số đã thêm.
func (h *ChessLibraryHandler) ImportArticles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b struct {
		Articles []types.ChessArticleBundle `json:"articles"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	count, err := h.service.ImportArticles(ctx, tenantID, b.Articles)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"imported": count})
}
