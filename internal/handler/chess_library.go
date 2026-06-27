package handler

import (
	"net/http"

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

// ---- Ván đấu ----

// ListGames GET /chess/games?white=&black=&eco=&result=
func (h *ChessLibraryHandler) ListGames(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	games, err := h.service.ListGames(ctx, tenantID, types.ChessGameFilter{
		White: c.Query("white"), Black: c.Query("black"),
		ECO: c.Query("eco"), Result: c.Query("result"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, games)
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

// ---- Bài tập ----

// ListPuzzles GET /chess/puzzles?theme=&difficulty=
func (h *ChessLibraryHandler) ListPuzzles(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	puzzles, err := h.service.ListPuzzles(ctx, tenantID, types.ChessPuzzleFilter{
		Theme: c.Query("theme"), Difficulty: c.Query("difficulty"),
	})
	if err != nil {
		chessFail(c, http.StatusInternalServerError, err)
		return
	}
	chessOK(c, puzzles)
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
