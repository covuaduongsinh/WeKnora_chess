package handler

import (
	"io"
	"mime"
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
		Tags:    parseChessTagSelector(c),
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
	// RevisionNote là ghi chú thay đổi (tùy chọn) — chỉ dùng khi PUT, lưu kèm
	// bản phiên bản mới nếu title/content đổi. KHÔNG phải Summary (tóm tắt
	// ngắn của bài, một trường persisted khác hẳn).
	RevisionNote string `json:"revision_note"`
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
	a, err := h.service.UpdateArticle(ctx, articleFromBody(c.Param("id"), tenantID, b), b.RevisionNote)
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

// ---- Ảnh chèn trong bài viết ----

// UploadArticleImage POST /chess/articles/:id/images (multipart form field
// "file") → {id, url}. url là đường dẫn ổn định GET /chess/articles/images/:id
// (KHÔNG phải presigned URL có hạn dùng — nội dung bài viết lưu URL này lâu dài).
func (h *ChessLibraryHandler) UploadArticleImage(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	fh, err := c.FormFile("file")
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	f, err := fh.Open()
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	mimeType := fh.Header.Get("Content-Type")
	img, err := h.service.UploadArticleImage(ctx, tenantID, c.Param("id"), fh.Filename, mimeType, data)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"id": img.ID, "url": "/api/v1/chess/articles/images/" + img.ID})
}

// GetArticleImage GET /chess/articles/images/:imageId — stream ảnh (inline).
func (h *ChessLibraryHandler) GetArticleImage(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	img, rc, err := h.service.GetArticleImage(ctx, tenantID, c.Param("imageId"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	defer rc.Close()
	contentType := img.Mime
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": img.FileName}))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Stream(func(w io.Writer) bool {
		_, _ = io.Copy(w, rc)
		return false
	})
}

// ---- Lịch sử phiên bản bài viết ----

// ListArticleRevisions GET /chess/articles/:id/revisions
func (h *ChessLibraryHandler) ListArticleRevisions(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	revs, err := h.service.ListArticleRevisions(ctx, tenantID, c.Param("id"))
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, revs)
}

// GetArticleRevision GET /chess/articles/:id/revisions/:rev_id
func (h *ChessLibraryHandler) GetArticleRevision(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	rev, err := h.service.GetArticleRevision(ctx, tenantID, c.Param("rev_id"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, rev)
}

// RestoreArticleRevision POST /chess/articles/:id/revisions/:rev_id/restore
func (h *ChessLibraryHandler) RestoreArticleRevision(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	a, err := h.service.RestoreArticleRevision(ctx, tenantID, c.Param("id"), c.Param("rev_id"))
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, a)
}
