package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// ChessLibraryHandler xử lý API kho ván đấu & ngân hàng bài tập cờ vua.
type ChessLibraryHandler struct {
	service interfaces.ChessLibraryService
}

// NewChessLibraryHandler tạo handler kho ván & bài tập.
func NewChessLibraryHandler(service interfaces.ChessLibraryService) *ChessLibraryHandler {
	return &ChessLibraryHandler{service: service}
}

func chessOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func chessFail(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"success": false, "error": err.Error()})
}

// chessOKPage trả danh sách KÈM tổng số. `data` vẫn là MẢNG như trước nên mọi
// caller cũ không đổi một dòng nào; phần đếm nằm ở khóa `meta` bổ sung.
//
// Có `meta` rồi thì giao diện mới hiện được "đang xem 20/137" — thay cho trần
// cứng Limit(500) trước đây vốn cắt âm thầm, không dấu hiệu gì.
func chessOKPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	meta := gin.H{"total": total, "page": page, "page_size": pageSize}
	if pageSize > 0 {
		meta["has_more"] = int64(page)*int64(pageSize) < total
	} else {
		meta["has_more"] = false
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "meta": meta})
}

// parseChessPagination đọc ?page & ?page_size cho các danh sách cờ.
//
// KHÁC parseListPagination của nền: THIẾU tham số nghĩa là KHÔNG phân trang
// (trả toàn bộ), chứ không mặc định trang 1 cỡ 20. Lý do: các endpoint này
// đang được dùng bởi picker chèn wikilink, export và script — mặc định cắt 20
// sẽ làm chúng thiếu dữ liệu âm thầm, đúng lỗi vừa sửa xong.
func parseChessPagination(c *gin.Context) (page, pageSize int, ok bool) {
	if strings.TrimSpace(c.Query("page")) == "" && strings.TrimSpace(c.Query("page_size")) == "" {
		return 0, 0, true // không phân trang
	}
	page, pageSize = 1, chessDefaultPageSize
	if v := strings.TrimSpace(c.Query("page")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			chessFail(c, http.StatusBadRequest, fmt.Errorf("page phải là số nguyên dương"))
			return 0, 0, false
		}
		page = n
	}
	if v := strings.TrimSpace(c.Query("page_size")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > chessMaxPageSize {
			chessFail(c, http.StatusBadRequest, fmt.Errorf("page_size phải trong khoảng 1..%d", chessMaxPageSize))
			return 0, 0, false
		}
		pageSize = n
	}
	return page, pageSize, true
}

const (
	chessDefaultPageSize = 50
	chessMaxPageSize     = 200
)

// ---- Ván đấu ----

// ListGames GET /chess/games?white=&black=&eco=&result=&level=&q=&tag=&tag_mode=
func (h *ChessLibraryHandler) ListGames(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	page, pageSize, ok := parseChessPagination(c)
	if !ok {
		return
	}
	filter := types.ChessGameFilter{
		White: c.Query("white"), Black: c.Query("black"),
		ECO: c.Query("eco"), Result: c.Query("result"),
		Level: c.Query("level"),
		// `q` được repo hỗ trợ từ đầu (ChessGameFilter.Keyword) nhưng handler
		// chưa bao giờ đọc — nối lại ở đây để ô tìm của Kho ván hoạt động.
		Keyword: c.Query("q"),
		Tags:    parseChessTagSelector(c),
		Page:    page, PageSize: pageSize,
	}
	games, err := h.service.ListGames(ctx, tenantID, filter)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	total, _ := h.service.CountGames(ctx, tenantID, filter)
	chessOKPage(c, games, total, page, pageSize)
}

// GetGame GET /chess/games/:id
func (h *ChessLibraryHandler) GetGame(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	g, err := h.service.GetGame(ctx, tenantID, c.Param("id"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, g)
}

// GetGameBySlug GET /chess/games/by-slug/:slug — giải mã wikilink [[game/<slug>]].
func (h *ChessLibraryHandler) GetGameBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	g, err := h.service.GetGameBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, g)
}

// GetGameBacklinks GET /chess/games/by-slug/:slug/backlinks — trang wiki trỏ tới ván.
func (h *ChessLibraryHandler) GetGameBacklinks(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	links, err := h.service.GetGameBacklinks(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, links)
}

type gameBody struct {
	White  string `json:"white"`
	Black  string `json:"black"`
	Result string `json:"result"`
	ECO    string `json:"eco"`
	Event  string `json:"event"`
	Date   string `json:"date"`
	PGN    string `json:"pgn"`
}

// CreateGame POST /chess/games
func (h *ChessLibraryHandler) CreateGame(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b gameBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	g, err := h.service.CreateGame(ctx, &types.ChessGame{
		TenantID: tenantID, White: b.White, Black: b.Black, Result: b.Result,
		ECO: b.ECO, Event: b.Event, Date: b.Date, PGN: b.PGN,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, g)
}

// UpdateGame PUT /chess/games/:id
func (h *ChessLibraryHandler) UpdateGame(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b gameBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	g, err := h.service.UpdateGame(ctx, &types.ChessGame{
		ID: c.Param("id"), TenantID: tenantID, White: b.White, Black: b.Black,
		Result: b.Result, ECO: b.ECO, Event: b.Event, Date: b.Date, PGN: b.PGN,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, g)
}

type slugBody struct {
	Slug string `json:"slug"`
}

// RenameGameSlug PUT /chess/games/:id/slug {slug} — đổi slug ván, giữ link cũ qua alias.
func (h *ChessLibraryHandler) RenameGameSlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b slugBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	g, err := h.service.RenameGameSlug(ctx, tenantID, c.Param("id"), b.Slug)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, g)
}

// RenamePuzzleSlug PUT /chess/puzzles/:id/slug {slug} — đổi slug bài tập, giữ link cũ.
func (h *ChessLibraryHandler) RenamePuzzleSlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b slugBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	p, err := h.service.RenamePuzzleSlug(ctx, tenantID, c.Param("id"), b.Slug)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, p)
}

// DeleteGame DELETE /chess/games/:id
func (h *ChessLibraryHandler) DeleteGame(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteGame(ctx, tenantID, c.Param("id")); err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, gin.H{"deleted": true})
}

// ImportGames POST /chess/games/import {pgn}
func (h *ChessLibraryHandler) ImportGames(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b struct {
		PGN string `json:"pgn"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	count, err := h.service.ImportGames(ctx, tenantID, b.PGN)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"imported": count})
}

// ExportGames GET /chess/games/export?white=&black=&eco=&result= — trả PGN nhiều ván.
func (h *ChessLibraryHandler) ExportGames(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	pgn, err := h.service.ExportGamesPGN(ctx, tenantID, types.ChessGameFilter{
		White: c.Query("white"), Black: c.Query("black"),
		ECO: c.Query("eco"), Result: c.Query("result"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, gin.H{"pgn": pgn})
}

// ---- Bài tập ----

// ExportPuzzles GET /chess/puzzles/export?theme=&difficulty= — danh sách bài tập (JSON).
func (h *ChessLibraryHandler) ExportPuzzles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	items, err := h.service.ExportPuzzles(ctx, tenantID, types.ChessPuzzleFilter{
		Theme: c.Query("theme"), Difficulty: c.Query("difficulty"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, items)
}

// ImportPuzzles POST /chess/puzzles/import {puzzles:[...]} — tạo mới; trả số đã thêm.
func (h *ChessLibraryHandler) ImportPuzzles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b struct {
		Puzzles []types.ChessPuzzleBundle `json:"puzzles"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	count, err := h.service.ImportPuzzles(ctx, tenantID, b.Puzzles)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, gin.H{"imported": count})
}

// ListPuzzles GET /chess/puzzles?theme=&difficulty=
func (h *ChessLibraryHandler) ListPuzzles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	page, pageSize, ok := parseChessPagination(c)
	if !ok {
		return
	}
	filter := types.ChessPuzzleFilter{
		Theme: c.Query("theme"), Difficulty: c.Query("difficulty"),
		Level: c.Query("level"),
		// `q` cũng chưa từng được nối ở đây — xem ghi chú tại ListGames.
		Keyword: c.Query("q"),
		Tags:    parseChessTagSelector(c),
		Page:    page, PageSize: pageSize,
	}
	puzzles, err := h.service.ListPuzzles(ctx, tenantID, filter)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	total, _ := h.service.CountPuzzles(ctx, tenantID, filter)
	chessOKPage(c, puzzles, total, page, pageSize)
}

// RandomPuzzle GET /chess/puzzles/random?theme=&difficulty=
func (h *ChessLibraryHandler) RandomPuzzle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	p, err := h.service.RandomPuzzle(ctx, tenantID, types.ChessPuzzleFilter{
		Theme: c.Query("theme"), Difficulty: c.Query("difficulty"),
	})
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, p)
}

// GetPuzzle GET /chess/puzzles/:id
func (h *ChessLibraryHandler) GetPuzzle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	p, err := h.service.GetPuzzle(ctx, tenantID, c.Param("id"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, p)
}

// GetPuzzleBySlug GET /chess/puzzles/by-slug/:slug — giải mã wikilink [[puzzle/<slug>]].
func (h *ChessLibraryHandler) GetPuzzleBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	p, err := h.service.GetPuzzleBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusNotFound, err)
		return
	}
	chessOK(c, p)
}

// GetPuzzleBacklinks GET /chess/puzzles/by-slug/:slug/backlinks — trang wiki trỏ tới thế cờ.
func (h *ChessLibraryHandler) GetPuzzleBacklinks(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	links, err := h.service.GetPuzzleBacklinks(ctx, tenantID, c.Param("slug"))
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, links)
}

type puzzleBody struct {
	Title      string `json:"title"`
	FEN        string `json:"fen"`
	Solution   string `json:"solution"`
	Theme      string `json:"theme"`
	Difficulty string `json:"difficulty"`
	Source     string `json:"source"`
}

// CreatePuzzle POST /chess/puzzles
func (h *ChessLibraryHandler) CreatePuzzle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b puzzleBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	p, err := h.service.CreatePuzzle(ctx, &types.ChessPuzzle{
		TenantID: tenantID, Title: b.Title, FEN: b.FEN, Solution: b.Solution,
		Theme: b.Theme, Difficulty: b.Difficulty, Source: b.Source,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, p)
}

// UpdatePuzzle PUT /chess/puzzles/:id
func (h *ChessLibraryHandler) UpdatePuzzle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b puzzleBody
	if err := c.ShouldBindJSON(&b); err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	p, err := h.service.UpdatePuzzle(ctx, &types.ChessPuzzle{
		ID: c.Param("id"), TenantID: tenantID, Title: b.Title, FEN: b.FEN,
		Solution: b.Solution, Theme: b.Theme, Difficulty: b.Difficulty, Source: b.Source,
	})
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	chessOK(c, p)
}

// DeletePuzzle DELETE /chess/puzzles/:id
func (h *ChessLibraryHandler) DeletePuzzle(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeletePuzzle(ctx, tenantID, c.Param("id")); err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, gin.H{"deleted": true})
}

// ---- Bảo trì KB ----

// ReindexKB POST /chess/library/reindex — đẩy lại toàn bộ ván+bài tập vào KB tri
// thức cờ. Dùng MỘT LẦN sau khi bật CHESS_KB_INDEX để index dữ liệu cũ (import
// hàng loạt không tự index). Trả lỗi rõ ràng nếu RAG cờ chưa bật hoặc KB cờ chưa
// có embedding model (fail-loud). Báo cáo trung thực: tổng / đã enqueue / lỗi.
func (h *ChessLibraryHandler) ReindexKB(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	res, err := h.service.ReindexAll(ctx, tenantID)
	if err != nil {
		chessFail(c, http.StatusBadRequest, err)
		return
	}
	// ⚠️ gin.H dựng THỦ CÔNG (không serialize thẳng res) để giữ 2 khóa tương
	// thích ngược bên dưới — hệ quả: mỗi khi thêm field vào ChessReindexResult
	// PHẢI thêm một dòng ở đây, nếu không field đó im lặng biến mất khỏi API.
	// (Đã dính đúng lỗi này: articles_total tính đúng ở service nhưng rơi mất
	// suốt từ đợt Ngân hàng bài viết Phase 1 tới khi phát hiện.)
	chessOK(c, gin.H{
		"games_total":     res.GamesTotal,
		"puzzles_total":   res.PuzzlesTotal,
		"positions_total": res.PositionsTotal,
		"books_total":     res.BooksTotal,
		"chapters_total":  res.ChaptersTotal,
		"articles_total":  res.ArticlesTotal,
		"enqueued":        res.Enqueued,
		"failed":          res.Failed,
		"purged":          res.Purged,
		"errors":          res.Errors,
		// Tương thích ngược với client/runbook cũ:
		"games_indexed":   res.GamesTotal,
		"puzzles_indexed": res.PuzzlesTotal,
		"note":            "đã enqueue để index; embedding chạy nền — kiểm tra GET /chess/library/index-status sau ~1 phút để xác nhận 'completed'",
	})
}

// IndexStatus GET /chess/library/index-status — báo trạng thái KB tri thức cờ để
// CHẨN ĐOÁN khi RAG rỗng (KB tồn tại?, có embedding model?, completed/pending/failed
// + mẫu lỗi). Luôn 200 (báo cáo trạng thái, không phải thao tác có thể thất bại).
func (h *ChessLibraryHandler) IndexStatus(c *gin.Context) {
	ctx := c.Request.Context()
	st, err := h.service.IndexStatus(ctx)
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, st)
}
