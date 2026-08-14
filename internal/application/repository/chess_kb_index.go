package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// chessKBIndexRepository lưu ánh xạ đối tượng cờ ↔ Knowledge trên GORM.
type chessKBIndexRepository struct {
	db *gorm.DB
}

// NewChessKBIndexRepository tạo repository ánh xạ index cờ.
func NewChessKBIndexRepository(db *gorm.DB) interfaces.ChessKBIndexRepository {
	return &chessKBIndexRepository{db: db}
}

func (r *chessKBIndexRepository) Get(ctx context.Context, tenantID uint64, chessType, chessSlug string) (*types.ChessKBIndex, error) {
	var m types.ChessKBIndex
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND chess_type = ? AND chess_slug = ?", tenantID, chessType, chessSlug).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *chessKBIndexRepository) Upsert(ctx context.Context, m *types.ChessKBIndex) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "chess_type"}, {Name: "chess_slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"knowledge_id", "kb_id", "updated_at"}),
	}).Create(m).Error
}

func (r *chessKBIndexRepository) Delete(ctx context.Context, tenantID uint64, chessType, chessSlug string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND chess_type = ? AND chess_slug = ?", tenantID, chessType, chessSlug).
		Delete(&types.ChessKBIndex{}).Error
}

// CountByType đếm mapping đã index theo từng loại thực thể cờ. Trả map rỗng
// (không nil) khi tenant chưa index gì — caller lặp thẳng không cần kiểm nil.
func (r *chessKBIndexRepository) CountByType(ctx context.Context, tenantID uint64) (map[string]int64, error) {
	var rows []struct {
		ChessType string
		N         int64
	}
	err := r.db.WithContext(ctx).
		Model(&types.ChessKBIndex{}).
		Select("chess_type, count(*) AS n").
		Where("tenant_id = ?", tenantID).
		Group("chess_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.ChessType] = row.N
	}
	return out, nil
}
