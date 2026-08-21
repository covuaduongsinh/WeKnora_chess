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

// gameQuery dựng query lọc chung cho List/Count/Export — tách ra để số đếm
// và danh sách KHÔNG BAO GIỜ lệch điều kiện.
func (r *chessLibraryRepository) gameQuery(ctx context.Context, tenantID uint64, f types.ChessGameFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&types.ChessGame{}).Where("tenant_id = ?", tenantID)
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
	if f.Level != "" {
		q = q.Where("level = ?", f.Level)
	}
	q = applyChessKeyword(q, "chess_games.search_text", f.Keyword)
	q = r.applyTagFilter(q, tenantID, types.ChessRefTypeGame, "chess_games.id", f.Tags)
	return q
}

func (r *chessLibraryRepository) ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error) {
	var games []*types.ChessGame
	q := applyChessPaging(r.gameQuery(ctx, tenantID, f).Order("created_at DESC"), f.Page, f.PageSize)
	err := q.Find(&games).Error
	return games, err
}

func (r *chessLibraryRepository) CountGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) (int64, error) {
	var n int64
	err := r.gameQuery(ctx, tenantID, f).Count(&n).Error
	return n, err
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
	game.SearchText = gameSearchText(game)
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *chessLibraryRepository) CreateGames(ctx context.Context, games []*types.ChessGame) error {
	if len(games) == 0 {
		return nil
	}
	for _, g := range games {
		g.SearchText = gameSearchText(g)
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
			"pgn": game.PGN, "ply_count": game.PlyCount, "level": game.Level,
			"search_text": gameSearchText(game),
		}).Error
}

func (r *chessLibraryRepository) UpdateGameSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	if err := r.db.WithContext(ctx).
		Model(&types.ChessGame{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error; err != nil {
		return err
	}
	// Slug nằm trong search_text nên phải tính lại — nếu không, tìm theo
	// slug mới sẽ trượt (lỗi câm: bản ghi đúng, chỉ là tìm không ra).
	r.refreshGameSearchText(ctx, tenantID, id)
	return nil
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
	if f.Level != "" {
		q = q.Where("level = ?", f.Level)
	}
	q = applyChessKeyword(q, "chess_puzzles.search_text", f.Keyword)
	q = r.applyTagFilter(q, tenantID, types.ChessRefTypePuzzle, "chess_puzzles.id", f.Tags)
	return q
}

func (r *chessLibraryRepository) ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error) {
	var puzzles []*types.ChessPuzzle
	q := applyChessPaging(r.puzzleQuery(ctx, tenantID, f).Order("created_at DESC"), f.Page, f.PageSize)
	err := q.Find(&puzzles).Error
	return puzzles, err
}

func (r *chessLibraryRepository) CountPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (int64, error) {
	var n int64
	err := r.puzzleQuery(ctx, tenantID, f).Model(&types.ChessPuzzle{}).Count(&n).Error
	return n, err
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
	puzzle.SearchText = puzzleSearchText(puzzle)
	return r.db.WithContext(ctx).Create(puzzle).Error
}

func (r *chessLibraryRepository) UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessPuzzle{}).
		Where("tenant_id = ? AND id = ?", puzzle.TenantID, puzzle.ID).
		Updates(map[string]interface{}{
			"title": puzzle.Title, "fen": puzzle.FEN, "solution": puzzle.Solution,
			"theme": puzzle.Theme, "difficulty": puzzle.Difficulty, "source": puzzle.Source,
			"level": puzzle.Level, "search_text": puzzleSearchText(puzzle),
		}).Error
}

func (r *chessLibraryRepository) UpdatePuzzleSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	if err := r.db.WithContext(ctx).
		Model(&types.ChessPuzzle{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error; err != nil {
		return err
	}
	// Slug nằm trong search_text nên phải tính lại — nếu không, tìm theo
	// slug mới sẽ trượt (lỗi câm: bản ghi đúng, chỉ là tìm không ra).
	r.refreshPuzzleSearchText(ctx, tenantID, id)
	return nil
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
