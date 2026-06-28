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
	repo         interfaces.ChessLibraryRepository
	chessRefRepo interfaces.WikiChessRefRepository
	aliasRepo    interfaces.ChessSlugAliasRepository
	indexer      *ChessKnowledgeIndexer
}

// NewChessLibraryService tạo service kho ván & bài tập cờ vua.
func NewChessLibraryService(
	repo interfaces.ChessLibraryRepository,
	chessRefRepo interfaces.WikiChessRefRepository,
	aliasRepo interfaces.ChessSlugAliasRepository,
	indexer *ChessKnowledgeIndexer,
) interfaces.ChessLibraryService {
	return &chessLibraryService{repo: repo, chessRefRepo: chessRefRepo, aliasRepo: aliasRepo, indexer: indexer}
}

// pruneChessRefs xóa các backlink wiki trỏ tới đối tượng cờ vừa bị xóa (best-effort).
func (s *chessLibraryService) pruneChessRefs(ctx context.Context, tenantID uint64, chessType, slug string) {
	if s.chessRefRepo == nil || slug == "" {
		return
	}
	_ = s.chessRefRepo.DeleteForChess(ctx, tenantID, chessType, slug)
}

// ---- Ván đấu ----

func (s *chessLibraryService) ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error) {
	return s.repo.ListGames(ctx, tenantID, f)
}

func (s *chessLibraryService) GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error) {
	return s.repo.GetGame(ctx, tenantID, id)
}

func (s *chessLibraryService) GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error) {
	g, err := s.repo.GetGameBySlug(ctx, tenantID, slug)
	if err == nil {
		return g, nil
	}
	// Không khớp chính xác → thử alias rồi fuzzy (giữ wikilink cũ/sai nhẹ sống).
	if resolved, ok := s.resolveAliasOrFuzzy(ctx, tenantID, types.ChessRefTypeGame, slug, s.repo.GameSlugs); ok {
		return s.repo.GetGameBySlug(ctx, tenantID, resolved)
	}
	return nil, err
}

// resolveAliasOrFuzzy nắn một slug "chết" về slug sống: ưu tiên alias (đổi tên/
// re-import), sau đó fuzzy trên pool slug do listSlugs cung cấp.
func (s *chessLibraryService) resolveAliasOrFuzzy(
	ctx context.Context, tenantID uint64, chessType, slug string,
	listSlugs func(context.Context, uint64) ([]string, error),
) (string, bool) {
	if s.aliasRepo != nil {
		if ns, ok, _ := s.aliasRepo.ResolveAlias(ctx, tenantID, chessType, slug); ok {
			return ns, true
		}
	}
	slugs, err := listSlugs(ctx, tenantID)
	if err != nil {
		return "", false
	}
	return resolveChessSlugFuzzy(slug, slugs)
}

func (s *chessLibraryService) GetGameBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error) {
	if s.chessRefRepo == nil {
		return nil, nil
	}
	return s.chessRefRepo.ListBacklinks(ctx, tenantID, types.ChessRefTypeGame, slug)
}

func (s *chessLibraryService) CreateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error) {
	game.ID = uuid.New().String()
	slug, err := ensureUniqueChessSlug(ctx, game.TenantID, gameSlugBase(game), game.ID, s.repo.GameSlugExists)
	if err != nil {
		return nil, err
	}
	game.Slug = slug
	if err := s.repo.CreateGame(ctx, game); err != nil {
		return nil, err
	}
	if s.indexer != nil {
		s.indexer.IndexGame(ctx, game)
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
	updated, err := s.repo.GetGame(ctx, game.TenantID, game.ID)
	if err == nil && updated != nil && s.indexer != nil {
		s.indexer.IndexGame(ctx, updated)
	}
	return updated, err
}

func (s *chessLibraryService) DeleteGame(ctx context.Context, tenantID uint64, id string) error {
	g, _ := s.repo.GetGame(ctx, tenantID, id)
	if err := s.repo.DeleteGame(ctx, tenantID, id); err != nil {
		return err
	}
	if g != nil {
		s.pruneChessRefs(ctx, tenantID, types.ChessRefTypeGame, g.Slug)
		if s.indexer != nil {
			s.indexer.Remove(ctx, tenantID, types.ChessRefTypeGame, g.Slug)
		}
	}
	return nil
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
		id := uuid.New().String()
		games = append(games, &types.ChessGame{
			ID:       id,
			TenantID: tenantID,
			// Import hàng loạt: gán slug xác định "g-<id8>" để tránh N lần dò
			// trùng. Có thể humanize sau bằng backfill; link vẫn hoạt động.
			Slug:     "g-" + id8(id),
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

func (s *chessLibraryService) GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error) {
	p, err := s.repo.GetPuzzleBySlug(ctx, tenantID, slug)
	if err == nil {
		return p, nil
	}
	if resolved, ok := s.resolveAliasOrFuzzy(ctx, tenantID, types.ChessRefTypePuzzle, slug, s.repo.PuzzleSlugs); ok {
		return s.repo.GetPuzzleBySlug(ctx, tenantID, resolved)
	}
	return nil, err
}

func (s *chessLibraryService) GetPuzzleBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error) {
	if s.chessRefRepo == nil {
		return nil, nil
	}
	return s.chessRefRepo.ListBacklinks(ctx, tenantID, types.ChessRefTypePuzzle, slug)
}

func (s *chessLibraryService) CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error) {
	if strings.TrimSpace(puzzle.FEN) == "" {
		return nil, fmt.Errorf("thiếu thế cờ FEN")
	}
	if err := chess.ValidateFEN(puzzle.FEN); err != nil {
		return nil, fmt.Errorf("FEN không hợp lệ: %v", err)
	}
	puzzle.ID = uuid.New().String()
	slug, err := ensureUniqueChessSlug(ctx, puzzle.TenantID, puzzleSlugBase(puzzle), puzzle.ID, s.repo.PuzzleSlugExists)
	if err != nil {
		return nil, err
	}
	puzzle.Slug = slug
	if err := s.repo.CreatePuzzle(ctx, puzzle); err != nil {
		return nil, err
	}
	if s.indexer != nil {
		s.indexer.IndexPuzzle(ctx, puzzle)
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
	updated, err := s.repo.GetPuzzle(ctx, puzzle.TenantID, puzzle.ID)
	if err == nil && updated != nil && s.indexer != nil {
		s.indexer.IndexPuzzle(ctx, updated)
	}
	return updated, err
}

func (s *chessLibraryService) DeletePuzzle(ctx context.Context, tenantID uint64, id string) error {
	p, _ := s.repo.GetPuzzle(ctx, tenantID, id)
	if err := s.repo.DeletePuzzle(ctx, tenantID, id); err != nil {
		return err
	}
	if p != nil {
		s.pruneChessRefs(ctx, tenantID, types.ChessRefTypePuzzle, p.Slug)
		if s.indexer != nil {
			s.indexer.Remove(ctx, tenantID, types.ChessRefTypePuzzle, p.Slug)
		}
	}
	return nil
}

func (s *chessLibraryService) RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error) {
	return s.repo.RandomPuzzle(ctx, tenantID, f)
}

// ---- Export / Import (sao lưu & chia sẻ) ----

// ExportGamesPGN xuất các ván (theo filter) thành một chuỗi PGN nhiều ván.
func (s *chessLibraryService) ExportGamesPGN(ctx context.Context, tenantID uint64, f types.ChessGameFilter) (string, error) {
	games, err := s.repo.ListGames(ctx, tenantID, f)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(games))
	for _, g := range games {
		if strings.TrimSpace(g.PGN) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(g.PGN))
	}
	return strings.Join(parts, "\n\n"), nil
}

// ExportPuzzles xuất các bài tập (theo filter) để sao lưu/chia sẻ.
func (s *chessLibraryService) ExportPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]types.ChessPuzzleBundle, error) {
	puzzles, err := s.repo.ListPuzzles(ctx, tenantID, f)
	if err != nil {
		return nil, err
	}
	out := make([]types.ChessPuzzleBundle, 0, len(puzzles))
	for _, p := range puzzles {
		out = append(out, types.ChessPuzzleBundle{
			Title: p.Title, FEN: p.FEN, Solution: p.Solution,
			Theme: p.Theme, Difficulty: p.Difficulty, Source: p.Source,
		})
	}
	return out, nil
}

// ImportPuzzles nhập danh sách bài tập, LUÔN tạo mới trong tenant hiện tại (tái dùng
// CreatePuzzle để validate FEN + sinh ID/slug). Trả số bài đã thêm; bỏ qua bài FEN
// lỗi để không chặn cả lô.
func (s *chessLibraryService) ImportPuzzles(ctx context.Context, tenantID uint64, items []types.ChessPuzzleBundle) (int, error) {
	created := 0
	for i := range items {
		it := items[i]
		if strings.TrimSpace(it.FEN) == "" {
			continue
		}
		_, err := s.CreatePuzzle(ctx, &types.ChessPuzzle{
			TenantID: tenantID, Title: it.Title, FEN: it.FEN, Solution: it.Solution,
			Theme: it.Theme, Difficulty: it.Difficulty, Source: it.Source,
		})
		if err != nil {
			continue
		}
		created++
	}
	if created == 0 && len(items) > 0 {
		return 0, fmt.Errorf("không nhập được bài tập nào (kiểm tra FEN/định dạng JSON)")
	}
	return created, nil
}
