package types

import "time"

// ChessSlugAlias ánh xạ một slug cũ của đối tượng cờ về slug hiện hành, để wikilink
// [[<type>/<old_slug>]] vẫn giải mã đúng sau khi slug đổi (đổi tên/re-import/backfill).
type ChessSlugAlias struct {
	ID        string `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID  uint64 `json:"tenant_id" gorm:"index"`
	ChessType string `json:"chess_type" gorm:"type:varchar(16)"`
	OldSlug   string `json:"old_slug" gorm:"type:varchar(255)"`
	NewSlug   string `json:"new_slug" gorm:"type:varchar(255)"`
	// Kind phân biệt alias sinh do ĐỔI SLUG ("rename", giữ vĩnh viễn — xóa sẽ
	// làm gãy link cũ trong sách/bài giảng) với alias NGƯỜI DÙNG GÕ TAY làm
	// bí danh/từ đồng nghĩa ("synonym", sửa/xóa tự do qua ReplaceSynonyms).
	// Cột có DEFAULT 'rename' ở DB (migration 000910) — alias cũ trước khi có
	// cột này đều đúng là 'rename' (mọi alias trước đó đều sinh từ đổi slug).
	Kind      string    `json:"kind" gorm:"type:varchar(16);default:rename"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_slug_aliases.
func (ChessSlugAlias) TableName() string { return "chess_slug_aliases" }
