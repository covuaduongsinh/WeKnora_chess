package types

import "time"

// ChessKBIndex ánh xạ một đối tượng cờ (game/puzzle/lesson) tới bản ghi Knowledge
// đã được index vào "KB tri thức cờ vua", để trợ lý truy hồi nội dung qua RAG.
type ChessKBIndex struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID    uint64    `json:"tenant_id" gorm:"index"`
	ChessType   string    `json:"chess_type" gorm:"type:varchar(16)"`
	ChessSlug   string    `json:"chess_slug" gorm:"type:varchar(255)"`
	KnowledgeID string    `json:"knowledge_id" gorm:"type:varchar(36)"`
	KBID        string    `json:"kb_id" gorm:"column:kb_id;type:varchar(36)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_kb_index.
func (ChessKBIndex) TableName() string { return "chess_kb_index" }
