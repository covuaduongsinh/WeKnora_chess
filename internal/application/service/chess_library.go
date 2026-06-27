package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Tencent/WeKnora/internal/chess"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// chessLibraryService triển khai nghiệp vụ kho ván đấu & ngân hàng bài tập.
type chessLibraryService struct {
	repo interfaces.ChessLibraryRepository
}

// NewChessLibraryService tạo service kho ván & bài tập cờ vua.
func NewChessLibraryService(repo interfaces.ChessLibraryRepository) interfaces.ChessLibraryService {
	return &chessLibraryService{repo: repo}
}

// ---- Ván đấu ----

func (s *chessLibraryService) ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error) {
	return s.repo.ListGames(ctx, tenantID, f)
}

func (s *chessLibraryService) GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error) {
	return s.repo.GetGame(ctx, tenantID, id)
}

func (s *chessLibraryService) CreateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error) {
	game.ID = uuid.New().String()
	if err := s.repo.CreateGame(ctx, game); err != nil {
		return nil, err
	}
	return game, nil
}

func (s *chessLibraryService) UpdateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error) {
	if _, err := s.repo.GetGame(ctx, game.TenantID, game.ID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateGame(ctx, game); err != nil {
		return nil, err
	}
	return s.repo.GetGame(ctx, game.TenantID, game.ID)
}

func (s *chessLibraryService) DeleteGame(ctx context.Context, tenantID uint64, id string) error {
	return s.repo.DeleteGame(ctx, tenantID, id)
}

// ImportGames tách PGN nhiều ván → tạo nhiều ChessGame, trả số ván đã thêm.
func (s *chessLibraryService) ImportGames(ctx context.Context, tenantID uint64, pgn string) (int, error) {
	if strings.TrimSpace(pgn) == "" {
		return 0, fmt.Errorf("PGN rỗng")
	}
	imported, err := chess.ParseMultiPGN(pgn)
	if err != nil {
		return 0, err
	}
	if len(imported) == 0 {
		return 0, fmt.Errorf("không tìm thấy ván cờ nào trong PGN")
	}
	games := make([]*types.ChessGame, 0, len(imported))
	for _, ig := range imported {
		games = append(games, &types.ChessGame{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			White:    ig.TagOr("White", "?"),
			Black:    ig.TagOr("Black", "?"),
			Result:   ig.TagOr("Result", ig.Outcome),
			ECO:      ig.TagOr("ECO", ""),
			Event:    ig.TagOr("Event", ""),
			Date:     ig.TagOr("Date", ""),
			PGN:      ig.PGN,
			PlyCount: ig.PlyCount,
		})
	}
	if err := s.repo.CreateGames(ctx, games); err != nil {
		return 0, err
	}
	return len(games), nil
}

// ---- Bài tập ----

func (s *chessLibraryService) ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error) {
	return s.repo.ListPuzzles(ctx, tenantID, f)
}

func (s *chessLibraryService) GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error) {
	return s.repo.GetPuzzle(ctx, tenantID, id)
}

func (s *chessLibraryService) CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error) {
	if strings.TrimSpace(puzzle.FEN) == "" {
		return nil, fmt.Errorf("thiếu thế cờ FEN")
	}
	if err := chess.ValidateFEN(puzzle.FEN); err != nil {
		return nil, fmt.Errorf("FEN không hợp lệ: %v", err)
	}
	puzzle.ID = uuid.New().String()
	if err := s.repo.CreatePuzzle(ctx, puzzle); err != nil {
		return nil, err
	}
	return puzzle, nil
}

func (s *chessLibraryService) UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error) {
	if _, err := s.repo.GetPuzzle(ctx, puzzle.TenantID, puzzle.ID); err != nil {
		return nil, err
	}
	if puzzle.FEN != "" {
		if err := chess.ValidateFEN(puzzle.FEN); err != nil {
			return nil, fmt.Errorf("FEN không hợp lệ: %v", err)
		}
	}
	if err := s.repo.UpdatePuzzle(ctx, puzzle); err != nil {
		return nil, err
	}
	return s.repo.GetPuzzle(ctx, puzzle.TenantID, puzzle.ID)
}

func (s *chessLibraryService) DeletePuzzle(ctx context.Context, tenantID uint64, id string) error {
	return s.repo.DeletePuzzle(ctx, tenantID, id)
}

func (s *chessLibraryService) RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error) {
	return s.repo.RandomPuzzle(ctx, tenantID, f)
}
