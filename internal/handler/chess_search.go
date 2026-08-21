package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_search.go bổ sung API TÌM KIẾM HỢP NHẤT vào CÙNG ChessLibraryHandler.
//
// Đây là ô tìm dành cho NGƯỜI DÙNG (tra cứu), khác /chess/refs/search vốn phục
// vụ autocomplete khi gõ "[[" trong ô soạn thảo.

// SearchChess GET /chess/search?q=&type=&level=&status=&tag=&tag_mode=&page=&page_size=
//
// `type` nhận nhiều giá trị (lặp tham số hoặc CSV) để lọc theo loại nội dung.
func (h *ChessLibraryHandler) SearchChess(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)

	page, pageSize, ok := parseChessPagination(c)
	if !ok {
		return
	}
	if pageSize <= 0 {
		// Tìm kiếm LUÔN phân trang: không có ngữ cảnh nào cần đổ toàn bộ kết
		// quả ra một lần, và mặc định "trả tất cả" ở đây dễ thành cú truy vấn
		// nặng vô ích.
		page, pageSize = 1, 20
	}

	res, err := h.service.SearchChessAll(ctx, tenantID, types.ChessSearchQuery{
		Keyword:  c.Query("q"),
		Types:    splitChessCSVParam(c.QueryArray("type")),
		Level:    c.Query("level"),
		Status:   c.Query("status"),
		Tags:     parseChessTagSelector(c),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, res)
}

// splitChessCSVParam gộp tham số lặp (?type=a&type=b) và dạng CSV (?type=a,b)
// về một danh sách phẳng — mỗi thư viện HTTP phía client sinh một kiểu.
func splitChessCSVParam(raw []string) []string {
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, chunk := range raw {
		for _, piece := range strings.Split(chunk, ",") {
			v := strings.TrimSpace(piece)
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
