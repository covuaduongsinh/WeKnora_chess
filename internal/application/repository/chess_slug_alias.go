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
		Kind:      "rename",
	}
	// Idempotent: trùng (tenant, loại, old_slug) thì cập nhật new_slug+kind —
	// một lần đổi slug PHẢI thắng một bí danh cũ tình cờ trùng chữ (hiếm).
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "chess_type"}, {Name: "old_slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"new_slug", "kind"}),
	}).Create(&a).Error
}

// ReplaceSynonyms ghi đè toàn bộ bí danh/từ đồng nghĩa (kind="synonym") của
// MỘT đối tượng cờ trong một giao dịch — xóa các dòng synonym cũ TRỎ TỚI
// targetSlug rồi chèn lại danh sách mới. KHÔNG đụng tới alias kind="rename"
// (lịch sử đổi slug). aliasSlugs PHẢI đã được chuẩn hóa (slugifyChess) + lọc
// rỗng/trùng chính targetSlug/trùng slug thật của đối tượng KHÁC ở tầng
// service TRƯỚC khi gọi vào đây — repository này không biết gì về bảng
// game/puzzle/position/... nên không tự kiểm tra được va chạm đó.
//
// Chèn qua ON CONFLICT (tenant_id, chess_type, old_slug) DO UPDATE new_slug+
// kind: nếu cùng một chuỗi bí danh vừa được gọi lại với targetSlug MỚI (vd
// sau khi đổi slug đối tượng), dòng cũ (new_slug=slug CŨ) được TỰ SỬA thành
// trỏ sang slug MỚI thay vì để mồ côi — xem RenameArticleSlug gọi lại
// syncArticleAliases với slug mới nhưng cùng danh sách bí danh.
func (r *chessSlugAliasRepository) ReplaceSynonyms(
	ctx context.Context, tenantID uint64, chessType, targetSlug string, aliasSlugs []string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND chess_type = ? AND kind = ? AND new_slug = ?",
			tenantID, chessType, "synonym", targetSlug).
			Delete(&types.ChessSlugAlias{}).Error; err != nil {
			return err
		}
		rows := make([]types.ChessSlugAlias, 0, len(aliasSlugs))
		for _, s := range aliasSlugs {
			if s == "" || s == targetSlug {
				continue
			}
			rows = append(rows, types.ChessSlugAlias{
				ID: uuid.New().String(), TenantID: tenantID, ChessType: chessType,
				OldSlug: s, NewSlug: targetSlug, Kind: "synonym",
			})
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "chess_type"}, {Name: "old_slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"new_slug", "kind"}),
		}).Create(&rows).Error
	})
}

// DeleteAliasesFor xóa TOÀN BỘ alias (mọi kind — cả rename lẫn synonym) đang
// TRỎ TỚI một đối tượng cờ — dùng khi xóa hẳn đối tượng đó: không còn gì để
// redirect tới nữa, giữ lại chỉ là rác vĩnh viễn không gỡ được qua UI.
func (r *chessSlugAliasRepository) DeleteAliasesFor(ctx context.Context, tenantID uint64, chessType, targetSlug string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND chess_type = ? AND new_slug = ?", tenantID, chessType, targetSlug).
		Delete(&types.ChessSlugAlias{}).Error
}
