package types

import "time"

// ChessSlugAlias ánh xạ một slug cũ của đối tượng cờ về slug hiện hành, để wikilink
// [[<type>/<old_slug>]] vẫn giải mã đúng sau khi slug đổi (đổi tên/re-import/backfill).
type ChessSlugAlias struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID  uint64    `json:"tenant_id" gorm:"index"`
	ChessType string    `json:"chess_type" gorm:"type:varchar(16)"`
	OldSlug   string    `json:"old_slug" gorm:"type:varchar(255)"`
	NewSlug   string    `json:"new_slug" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_slug_aliases.
func (ChessSlugAlias) TableName() string { return "chess_slug_aliases" }
