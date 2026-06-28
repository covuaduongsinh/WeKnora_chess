package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

// ChessKBIndexRepository quản lý ánh xạ đối tượng cờ ↔ bản ghi Knowledge (RAG).
type ChessKBIndexRepository interface {
	// Get trả mapping cho (tenant, loại, slug) nếu có.
	Get(ctx context.Context, tenantID uint64, chessType, chessSlug string) (*types.ChessKBIndex, error)
	// Upsert ghi/đè mapping theo khóa (tenant, loại, slug).
	Upsert(ctx context.Context, m *types.ChessKBIndex) error
	// Delete xóa mapping theo (tenant, loại, slug).
	Delete(ctx context.Context, tenantID uint64, chessType, chessSlug string) error
}
