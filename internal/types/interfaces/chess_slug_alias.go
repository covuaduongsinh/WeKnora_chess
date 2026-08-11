package interfaces

import "context"

// ChessSlugAliasRepository quản lý alias/redirect slug đối tượng cờ.
type ChessSlugAliasRepository interface {
	// ResolveAlias trả new_slug nếu old_slug có alias (theo tenant + loại).
	ResolveAlias(ctx context.Context, tenantID uint64, chessType, oldSlug string) (string, bool, error)
	// AddAlias ghi một alias (old_slug -> new_slug), kind="rename". Idempotent
	// theo khóa duy nhất (tenant, loại, old_slug) — thắng bí danh cũ nếu trùng chữ.
	AddAlias(ctx context.Context, tenantID uint64, chessType, oldSlug, newSlug string) error
	// ReplaceSynonyms ghi đè toàn bộ bí danh (kind="synonym") của MỘT đối
	// tượng cờ — xóa-rồi-chèn-lại trong 1 giao dịch, KHÔNG đụng alias
	// kind="rename" (lịch sử đổi slug giữ vĩnh viễn).
	ReplaceSynonyms(ctx context.Context, tenantID uint64, chessType, targetSlug string, aliasSlugs []string) error
	// DeleteAliasesFor xóa toàn bộ alias (mọi kind) trỏ TỚI một đối tượng cờ
	// — dùng khi xóa hẳn đối tượng đó.
	DeleteAliasesFor(ctx context.Context, tenantID uint64, chessType, targetSlug string) error
}
