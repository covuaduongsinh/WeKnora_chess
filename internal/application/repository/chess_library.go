package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
)

// chessLibraryRepository lưu trữ kho ván đấu & bài tập cờ vua trên GORM.
type chessLibraryRepository struct {
	db *gorm.DB
}

// NewChessLibraryRepository tạo repository kho ván & bài tập.
func NewChessLibraryRepository(db *gorm.DB) interfaces.ChessLibraryRepository {
	return &chessLibraryRepository{db: db}
}

// ---- Ván đấu ----

func (r *chessLibraryRepository) ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.White != "" {
		q = q.Where("white ILIKE ?", "%"+f.White+"%")
	}
	if f.Black != "" {
		q = q.Where("black ILIKE ?", "%"+f.Black+"%")
	}
	if f.ECO != "" {
		q = q.Where("eco ILIKE ?", f.ECO+"%")
	}
	if f.Result != "" {
		q = q.Where("result = ?", f.Result)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("slug ILIKE ? OR white ILIKE ? OR black ILIKE ? OR event ILIKE ?", kw, kw, kw, kw)
	}
	var games []*types.ChessGame
	err := q.Order("created_at DESC").Limit(500).Find(&games).Error
	return games, err
}

func (r *chessLibraryRepository) GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error) {
	var g types.ChessGame
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *chessLibraryRepository) GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error) {
	var g types.ChessGame
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// GameSlugs trả toàn bộ slug ván "sống" của tenant — pool ứng viên fuzzy-resolve.
func (r *chessLibraryRepository) GameSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Model(&types.ChessGame{}).
		Where("tenant_id = ? AND slug <> ''", tenantID).
		Pluck("slug", &slugs).Error
	return slugs, err
}

func (r *chessLibraryRepository) GameSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessGame{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessLibraryRepository) CreateGame(ctx context.Context, game *types.ChessGame) error {
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *chessLibraryRepository) CreateGames(ctx context.Context, games []*types.ChessGame) error {
	if len(games) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(games, 100).Error
}

func (r *chessLibraryRepository) UpdateGame(ctx context.Context, game *types.ChessGame) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessGame{}).
		Where("tenant_id = ? AND id = ?", game.TenantID, game.ID).
		Updates(map[string]interface{}{
			"white": game.White, "black": game.Black, "result": game.Result,
			"eco": game.ECO, "event": game.Event, "date": game.Date,
			"pgn": game.PGN, "ply_count": game.PlyCount,
		}).Error
}

func (r *chessLibraryRepository) DeleteGame(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessGame{}).Error
}

// ---- Bài tập ----

func (r *chessLibraryRepository) puzzleQuery(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.Theme != "" {
		q = q.Where("theme = ?", f.Theme)
	}
	if f.Difficulty != "" {
		q = q.Where("difficulty = ?", f.Difficulty)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("slug ILIKE ? OR title ILIKE ? OR theme ILIKE ?", kw, kw, kw)
	}
	return q
}

func (r *chessLibraryRepository) ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error) {
	var puzzles []*types.ChessPuzzle
	err := r.puzzleQuery(ctx, tenantID, f).Order("created_at DESC").Limit(500).Find(&puzzles).Error
	return puzzles, err
}

func (r *chessLibraryRepository) GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error) {
	var p types.ChessPuzzle
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *chessLibraryRepository) GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error) {
	var p types.ChessPuzzle
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// PuzzleSlugs trả toàn bộ slug bài tập "sống" của tenant — pool ứng viên fuzzy-resolve.
func (r *chessLibraryRepository) PuzzleSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Model(&types.ChessPuzzle{}).
		Where("tenant_id = ? AND slug <> ''", tenantID).
		Pluck("slug", &slugs).Error
	return slugs, err
}

func (r *chessLibraryRepository) PuzzleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessPuzzle{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessLibraryRepository) CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error {
	return r.db.WithContext(ctx).Create(puzzle).Error
}

func (r *chessLibraryRepository) UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessPuzzle{}).
		Where("tenant_id = ? AND id = ?", puzzle.TenantID, puzzle.ID).
		Updates(map[string]interface{}{
			"title": puzzle.Title, "fen": puzzle.FEN, "solution": puzzle.Solution,
			"theme": puzzle.Theme, "difficulty": puzzle.Difficulty, "source": puzzle.Source,
		}).Error
}

func (r *chessLibraryRepository) DeletePuzzle(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessPuzzle{}).Error
}

func (r *chessLibraryRepository) RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error) {
	var p types.ChessPuzzle
	// random() là cú pháp Postgres (stack hiện tại dùng pgvector/postgres).
	if err := r.puzzleQuery(ctx, tenantID, f).Order("random()").First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
