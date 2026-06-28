package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// chessSlugAliasRepository lưu alias/redirect slug đối tượng cờ trên GORM.
type chessSlugAliasRepository struct {
	db *gorm.DB
}

// NewChessSlugAliasRepository tạo repository alias slug cờ.
func NewChessSlugAliasRepository(db *gorm.DB) interfaces.ChessSlugAliasRepository {
	return &chessSlugAliasRepository{db: db}
}

func (r *chessSlugAliasRepository) ResolveAlias(ctx context.Context, tenantID uint64, chessType, oldSlug string) (string, bool, error) {
	if oldSlug == "" {
		return "", false, nil
	}
	var a types.ChessSlugAlias
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND chess_type = ? AND old_slug = ?", tenantID, chessType, oldSlug).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return a.NewSlug, a.NewSlug != "", nil
}

func (r *chessSlugAliasRepository) AddAlias(ctx context.Context, tenantID uint64, chessType, oldSlug, newSlug string) error {
	if oldSlug == "" || newSlug == "" || oldSlug == newSlug {
		return nil
	}
	a := types.ChessSlugAlias{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		ChessType: chessType,
		OldSlug:   oldSlug,
		NewSlug:   newSlug,
	}
	// Idempotent: trùng (tenant, loại, old_slug) thì cập nhật new_slug.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "chess_type"}, {Name: "old_slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"new_slug"}),
	}).Create(&a).Error
}
