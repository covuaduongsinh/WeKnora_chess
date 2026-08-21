package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_tag.go bổ sung API HỆ THẺ THỐNG NHẤT vào CÙNG ChessLibraryHandler
// (khai báo ở chess_library.go) — thẻ là trục phân loại NGANG phủ cả 8 loại
// nội dung cờ, nên router.go chỉ cần thêm một nhóm route trong
// RegisterChessLibraryRoutes đã có.

// parseChessTagSelector đọc bộ lọc thẻ từ query của MỌI endpoint danh sách
// nội dung: ?tag=ghim,khai-cuoc&tag_mode=all
//
// Chấp nhận cả tham số lặp (?tag=a&tag=b) lẫn CSV (?tag=a,b) vì mỗi thư viện
// HTTP phía client sinh một kiểu. Giá trị được chuẩn hóa nhẹ ở đây; việc khử
// dấu thật do tầng service làm (slugifyChess) nên "Khai cuộc" cũng khớp.
func parseChessTagSelector(c *gin.Context) types.ChessTagSelector {
	raw := append([]string{}, c.QueryArray("tag")...)
	raw = append(raw, c.QueryArray("tags")...)
	slugs := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, chunk := range raw {
		for _, piece := range strings.Split(chunk, ",") {
			v := strings.ToLower(strings.TrimSpace(piece))
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			slugs = append(slugs, v)
		}
	}
	mode := types.ChessTagModeAny
	if strings.EqualFold(c.Query("tag_mode"), types.ChessTagModeAll) {
		mode = types.ChessTagModeAll
	}
	return types.ChessTagSelector{TagSlugs: slugs, Mode: mode}
}

// ---- Từ điển thẻ ----

// ListChessTags GET /chess/tags?kind=&q=&only_used=
func (h *ChessLibraryHandler) ListChessTags(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	tags, err := h.service.ListChessTags(ctx, tenantID, types.ChessTagFilter{
		Kind:     c.Query("kind"),
		Keyword:  c.Query("q"),
		OnlyUsed: c.Query("only_used") == "true",
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, tags)
}

// GetChessTagBySlug GET /chess/tags/by-slug/:slug
func (h *ChessLibraryHandler) GetChessTagBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	tag, err := h.service.GetChessTagBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, tag)
}

// ListChessTagItems GET /chess/tags/by-slug/:slug/items?type=&page=&page_size=
//
// "Mục lục ngang": bấm một thẻ, ra MỌI loại nội dung mang thẻ đó. Đây là
// endpoint cờ ĐẦU TIÊN có phân trang thật kèm tổng số — các list cũ cắt cứng
// ở 500 bản ghi mà không báo gì.
func (h *ChessLibraryHandler) ListChessTagItems(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	page, pageSize, ok := parseListPagination(c)
	if !ok {
		return
	}
	res, err := h.service.ListChessTagItems(ctx, tenantID, c.Param("slug"), c.Query("type"), page, pageSize)
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, res)
}

type chessTagBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

// CreateChessTag POST /chess/tags
func (h *ChessLibraryHandler) CreateChessTag(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b chessTagBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	tag, err := h.service.CreateChessTag(ctx, &types.ChessTag{
		TenantID: tenantID, Name: b.Name, Description: b.Description,
		Color: b.Color, SortOrder: b.SortOrder, Kind: types.ChessTagKindFree,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, tag)
}

// UpdateChessTag PUT /chess/tags/:id — đổi tên có thể kéo theo GỘP vào thẻ
// cùng slug đã tồn tại (service xử lý), nên phản hồi có thể là thẻ ĐÍCH.
func (h *ChessLibraryHandler) UpdateChessTag(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b chessTagBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	tag, err := h.service.UpdateChessTag(ctx, &types.ChessTag{
		ID: c.Param("id"), TenantID: tenantID, Name: b.Name,
		Description: b.Description, Color: b.Color, SortOrder: b.SortOrder,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, tag)
}

// DeleteChessTag DELETE /chess/tags/:id — chặn với thẻ nhóm nội dung.
func (h *ChessLibraryHandler) DeleteChessTag(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteChessTag(ctx, tenantID, c.Param("id")); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"deleted": true})
}

type chessTagMergeBody struct {
	// TargetID là thẻ ĐÍCH — mọi liên kết của thẻ trong URL chuyển sang đây,
	// rồi thẻ trong URL bị xóa.
	TargetID string `json:"target_id"`
}

// MergeChessTags PUT /chess/tags/:id/merge
func (h *ChessLibraryHandler) MergeChessTags(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b chessTagMergeBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(b.TargetID) == "" {
		chessFail(c, http.StatusBadRequest, fmt.Errorf("thiếu thẻ đích để gộp vào"))
		return
	}
	tag, err := h.service.MergeChessTags(ctx, tenantID, c.Param("id"), b.TargetID)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, tag)
}

// ---- Gắn thẻ cho nội dung ----

type chessTagAssignBody struct {
	// ChessType là loại nội dung: game|puzzle|lesson|course|position|book|
	// chapter|article.
	ChessType string `json:"chess_type"`
	ChessID   string `json:"chess_id"`
	// Tags là danh sách tên thẻ; chấp nhận cả phần tử chứa CSV. Mảng RỖNG
	// nghĩa là gỡ hết thẻ của mục (ghi đè, không phải cộng dồn).
	Tags []string `json:"tags"`
}

// AssignChessTags PUT /chess/tags/assign
//
// Một endpoint duy nhất gắn thẻ cho MỌI loại nội dung — nhờ vậy 5 loại chưa
// từng có cột `tags` (ván/bài tập/bài giảng/khóa học/chương) dùng được hệ thẻ
// mà không phải đổi chữ ký Create/Update của chúng.
func (h *ChessLibraryHandler) AssignChessTags(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b chessTagAssignBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	if !chessTagAssignableTypes[b.ChessType] {
		chessFail(c, http.StatusBadRequest, fmt.Errorf("loại nội dung không hợp lệ: %q", b.ChessType))
		return
	}
	tags, err := h.service.SetChessTags(ctx, tenantID, b.ChessType, b.ChessID, b.Tags)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, tags)
}

// chessTagAssignableTypes chốt danh sách loại nội dung gắn thẻ được. Kệ sách
// và chuyên mục bài viết CỐ Ý không có mặt: chúng là công cụ điều hướng, bản
// thân đã là cách gom nhóm rồi.
var chessTagAssignableTypes = map[string]bool{
	types.ChessRefTypeGame:     true,
	types.ChessRefTypePuzzle:   true,
	types.ChessRefTypeLesson:   true,
	types.ChessRefTypeCourse:   true,
	types.ChessRefTypePosition: true,
	types.ChessRefTypeBook:     true,
	types.ChessRefTypeChapter:  true,
	types.ChessRefTypeArticle:  true,
}

// GetChessTagsOf GET /chess/tags/of/:type/:id — thẻ đang gắn cho một mục.
func (h *ChessLibraryHandler) GetChessTagsOf(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	chessType, id := c.Param("type"), c.Param("id")
	if !chessTagAssignableTypes[chessType] {
		chessFail(c, http.StatusBadRequest, fmt.Errorf("loại nội dung không hợp lệ: %q", chessType))
		return
	}
	byOwner, err := h.service.ChessTagsFor(ctx, tenantID, chessType, []string{id})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	tags := byOwner[id]
	if tags == nil {
		tags = []*types.ChessTag{} // trả [] chứ không phải null cho client
	}
	chessOK(c, tags)
}

type chessTagsOfBody struct {
	ChessType string   `json:"chess_type"`
	IDs       []string `json:"ids"`
}

// ListChessTagsOfMany POST /chess/tags/of — thẻ của NHIỀU mục cùng loại trong
// MỘT lượt gọi. Cần cho việc hiện chip thẻ trên từng hàng danh sách: gọi lẻ
// từng mục sẽ thành N+1 request cho một trang.
//
// Dùng POST vì danh sách id có thể dài quá giới hạn URL — đây là truy vấn, không
// phải thao tác ghi.
func (h *ChessLibraryHandler) ListChessTagsOfMany(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b chessTagsOfBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	if !chessTagAssignableTypes[b.ChessType] {
		chessFail(c, http.StatusBadRequest, fmt.Errorf("loại nội dung không hợp lệ: %q", b.ChessType))
		return
	}
	byOwner, err := h.service.ChessTagsFor(ctx, tenantID, b.ChessType, b.IDs)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	// Trả {} thay vì null cho client khi không có mục nào mang thẻ.
	if byOwner == nil {
		byOwner = map[string][]*types.ChessTag{}
	}
	chessOK(c, byOwner)
}

// ---- Bảo trì ----

// BackfillChessTags POST /chess/tags/backfill — nạp dữ liệu phân loại CŨ (3
// cột CSV + category/phase/theme) vào hệ thẻ. Idempotent, chạy lại được.
func (h *ChessLibraryHandler) BackfillChessTags(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	res, err := h.service.BackfillChessTags(ctx, tenantID)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, res)
}

// RecountChessTags POST /chess/tags/recount — đếm lại cache usage_count từ
// bảng nối. Nút chữa khi số trên chip thẻ trông không đúng.
func (h *ChessLibraryHandler) RecountChessTags(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.RecountChessTags(ctx, tenantID); err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, gin.H{"recounted": true})
}
