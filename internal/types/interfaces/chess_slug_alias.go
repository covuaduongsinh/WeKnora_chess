package interfaces

import "context"

// ChessSlugAliasRepository quản lý alias/redirect slug đối tượng cờ.
type ChessSlugAliasRepository interface {
	// ResolveAlias trả new_slug nếu old_slug có alias (theo tenant + loại).
	ResolveAlias(ctx context.Context, tenantID uint64, chessType, oldSlug string) (string, bool, error)
	// AddAlias ghi một alias (old_slug -> new_slug). Idempotent theo khóa duy nhất.
	AddAlias(ctx context.Context, tenantID uint64, chessType, oldSlug, newSlug string) error
}
