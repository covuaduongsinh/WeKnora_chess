package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

// chess_position.go bổ sung thao tác lưu trữ cho "Ngân hàng thế cờ"
// (chess_positions) vào CÙNG struct chessLibraryRepository (khai báo ở
// chess_library.go) — position là "ngăn" thứ ba cạnh games/puzzles, không phải
// repository riêng, để container.go không cần Provide mới.

// positionQuery dựng query lọc chung cho List/Export/tìm trùng — cùng khuôn
// puzzleQuery ở chess_library.go.
func (r *chessLibraryRepository) positionQuery(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Level != "" {
		q = q.Where("level = ?", f.Level)
	}
	if f.ECO != "" {
		q = q.Where("eco ILIKE ?", f.ECO+"%")
	}
	if f.SourceGameID != "" {
		q = q.Where("source_game_id = ?", f.SourceGameID)
	}
	q = applyChessKeyword(q, "chess_positions.search_text", f.Keyword)
	q = r.applyTagFilter(q, tenantID, types.ChessRefTypePosition, "chess_positions.id", f.Tags)
	return q
}

func (r *chessLibraryRepository) ListPositions(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) ([]*types.ChessPosition, error) {
	var positions []*types.ChessPosition
	q := applyChessPaging(r.positionQuery(ctx, tenantID, f).Order("created_at DESC"), f.Page, f.PageSize)
	err := q.Find(&positions).Error
	return positions, err
}

func (r *chessLibraryRepository) CountPositions(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) (int64, error) {
	var n int64
	err := r.positionQuery(ctx, tenantID, f).Model(&types.ChessPosition{}).Count(&n).Error
	return n, err
}

func (r *chessLibraryRepository) GetPosition(ctx context.Context, tenantID uint64, id string) (*types.ChessPosition, error) {
	var p types.ChessPosition
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *chessLibraryRepository) GetPositionBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPosition, error) {
	var p types.ChessPosition
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// PositionSlugs trả toàn bộ slug thế cờ "sống" của tenant — pool ứng viên fuzzy-resolve.
func (r *chessLibraryRepository) PositionSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Model(&types.ChessPosition{}).
		Where("tenant_id = ? AND slug <> ''", tenantID).
		Pluck("slug", &slugs).Error
	return slugs, err
}

func (r *chessLibraryRepository) PositionSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessPosition{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessLibraryRepository) CreatePosition(ctx context.Context, position *types.ChessPosition) error {
	position.SearchText = positionSearchText(position)
	return r.db.WithContext(ctx).Create(position).Error
}

func (r *chessLibraryRepository) CreatePositions(ctx context.Context, positions []*types.ChessPosition) error {
	for _, p := range positions {
		p.SearchText = positionSearchText(p)
	}
	if len(positions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(positions, 100).Error
}

// UpdatePosition cố ý KHÔNG đụng cột slug (như UpdateGame/UpdatePuzzle) — đổi
// slug đi qua UpdatePositionSlug để luôn kèm ghi alias ở tầng service.
func (r *chessLibraryRepository) UpdatePosition(ctx context.Context, position *types.ChessPosition) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessPosition{}).
		Where("tenant_id = ? AND id = ?", position.TenantID, position.ID).
		Updates(map[string]interface{}{
			"title": position.Title, "fen": position.FEN, "fen_key": position.FENKey,
			"category": position.Category, "level": position.Level, "eco": position.ECO,
			"side_to_move": position.SideToMove, "orientation": position.Orientation,
			"assessment": position.Assessment, "annotation": position.Annotation,
			"tags": position.Tags, "source": position.Source,
			"source_game_id": position.SourceGameID, "source_ply": position.SourcePly,
			"search_text": positionSearchText(position),
		}).Error
}

func (r *chessLibraryRepository) UpdatePositionSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	if err := r.db.WithContext(ctx).
		Model(&types.ChessPosition{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error; err != nil {
		return err
	}
	// Slug nằm trong search_text nên phải tính lại — nếu không, tìm theo
	// slug mới sẽ trượt (lỗi câm: bản ghi đúng, chỉ là tìm không ra).
	r.refreshPositionSearchText(ctx, tenantID, id)
	return nil
}

func (r *chessLibraryRepository) DeletePosition(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessPosition{}).Error
}

// ListPositionsByGame liệt kê các thế cờ đã trích từ MỘT ván cụ thể (panel
// "Thế cờ đã trích" trong Kho ván) — mới nhất/theo nước trước.
func (r *chessLibraryRepository) ListPositionsByGame(ctx context.Context, tenantID uint64, gameID string) ([]*types.ChessPosition, error) {
	var positions []*types.ChessPosition
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND source_game_id = ?", tenantID, gameID).
		Order("source_ply ASC").Find(&positions).Error
	return positions, err
}

// FindByFENKey tìm các thế cờ có cùng fen_key — dùng để CẢNH BÁO (không chặn)
// trùng lặp khi tạo mới.
func (r *chessLibraryRepository) FindByFENKey(ctx context.Context, tenantID uint64, fenKey string) ([]*types.ChessPosition, error) {
	if fenKey == "" {
		return nil, nil
	}
	var positions []*types.ChessPosition
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND fen_key = ?", tenantID, fenKey).
		Limit(10).Find(&positions).Error
	return positions, err
}
