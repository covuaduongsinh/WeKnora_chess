package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

// ChessLibraryService định nghĩa nghiệp vụ kho ván đấu & ngân hàng bài tập cờ vua.
type ChessLibraryService interface {
	// ---- Ván đấu ----
	ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error)
	GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error)
	// GetGameBySlug giải mã wikilink [[game/<slug>]] về ván cờ.
	GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error)
	// GetGameBacklinks liệt kê trang wiki/bài giảng trỏ tới ván cờ này.
	GetGameBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	UpdateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	DeleteGame(ctx context.Context, tenantID uint64, id string) error
	// ImportGames tách một PGN nhiều ván và tạo nhiều ChessGame; trả số ván đã thêm.
	ImportGames(ctx context.Context, tenantID uint64, pgn string) (int, error)

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	// GetPuzzleBySlug giải mã wikilink [[puzzle/<slug>]] về thế cờ/bài tập.
	GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error)
	// GetPuzzleBacklinks liệt kê trang wiki/bài giảng trỏ tới thế cờ này.
	GetPuzzleBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)

	// ---- Export / Import ----
	// ExportGamesPGN xuất các ván (theo filter) thành một PGN nhiều ván.
	ExportGamesPGN(ctx context.Context, tenantID uint64, f types.ChessGameFilter) (string, error)
	// ExportPuzzles xuất các bài tập (theo filter) để sao lưu/chia sẻ.
	ExportPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]types.ChessPuzzleBundle, error)
	// ImportPuzzles nhập danh sách bài tập (luôn tạo mới); trả số bài đã thêm.
	ImportPuzzles(ctx context.Context, tenantID uint64, items []types.ChessPuzzleBundle) (int, error)

	// ReindexAll đẩy lại toàn bộ ván + bài tập của tenant vào KB tri thức cờ (chỉ
	// tác dụng khi CHESS_KB_INDEX bật). Trả (số ván, số bài tập) đã index.
	ReindexAll(ctx context.Context, tenantID uint64) (int, int, error)
}

// ChessLibraryRepository định nghĩa thao tác lưu trữ kho ván & bài tập.
type ChessLibraryRepository interface {
	// ---- Ván đấu ----
	ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error)
	GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error)
	GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error)
	// GameSlugs trả mọi slug ván sống của tenant (pool fuzzy-resolve).
	GameSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	GameSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateGame(ctx context.Context, game *types.ChessGame) error
	CreateGames(ctx context.Context, games []*types.ChessGame) error
	UpdateGame(ctx context.Context, game *types.ChessGame) error
	DeleteGame(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error)
	// PuzzleSlugs trả mọi slug bài tập sống của tenant (pool fuzzy-resolve).
	PuzzleSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	PuzzleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)
}
