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
	CreateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	UpdateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	DeleteGame(ctx context.Context, tenantID uint64, id string) error
	// ImportGames tách một PGN nhiều ván và tạo nhiều ChessGame; trả số ván đã thêm.
	ImportGames(ctx context.Context, tenantID uint64, pgn string) (int, error)

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)
}

// ChessLibraryRepository định nghĩa thao tác lưu trữ kho ván & bài tập.
type ChessLibraryRepository interface {
	// ---- Ván đấu ----
	ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error)
	GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error)
	CreateGame(ctx context.Context, game *types.ChessGame) error
	CreateGames(ctx context.Context, games []*types.ChessGame) error
	UpdateGame(ctx context.Context, game *types.ChessGame) error
	DeleteGame(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)
}
