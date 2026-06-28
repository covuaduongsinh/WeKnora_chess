package types

import "time"

// ChessGame là một ván cờ trong kho ván đấu (Kho ván đấu).
// Lưu PGN đầy đủ + metadata để tìm/lọc.
type ChessGame struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là định danh thân thiện (duy nhất theo tenant) làm đích wikilink
	// [[game/<slug>]]. Sinh ở tầng service khi tạo; ổn định sau đó.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(255)"`
	// White / Black là tên hai đấu thủ.
	White string `json:"white" gorm:"type:varchar(128)"`
	Black string `json:"black" gorm:"type:varchar(128)"`
	// Result là kết quả: "1-0" | "0-1" | "1/2-1/2" | "*".
	Result string `json:"result" gorm:"type:varchar(16)"`
	// ECO là mã khai cuộc.
	ECO string `json:"eco" gorm:"type:varchar(8)"`
	// Event là tên giải/sự kiện.
	Event string `json:"event" gorm:"type:varchar(255)"`
	// Date là ngày đấu (chuỗi PGN, vd "2026.06.27").
	Date string `json:"date" gorm:"type:varchar(32)"`
	// PGN là nội dung ván cờ đầy đủ.
	PGN string `json:"pgn" gorm:"type:text"`
	// PlyCount là số nửa-nước.
	PlyCount int `json:"ply_count" gorm:"default:0"`
	// CreatedAt / UpdatedAt là thời gian tạo/cập nhật.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_games.
func (ChessGame) TableName() string { return "chess_games" }

// ChessPuzzle là một bài tập cờ trong ngân hàng bài tập.
type ChessPuzzle struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là định danh thân thiện (duy nhất theo tenant) làm đích wikilink
	// [[puzzle/<slug>]]. Sinh ở tầng service khi tạo; ổn định sau đó.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(255)"`
	// Title là tiêu đề bài tập.
	Title string `json:"title" gorm:"type:varchar(255)"`
	// FEN là thế cờ của bài tập.
	FEN string `json:"fen" gorm:"type:varchar(128);not null"`
	// Solution là lời giải (SAN/UCI, tùy chọn).
	Solution string `json:"solution" gorm:"type:varchar(255)"`
	// Theme là chủ đề (vd "chiếu hết", "chiến thuật", "tàn cuộc").
	Theme string `json:"theme" gorm:"type:varchar(64);index"`
	// Difficulty là độ khó: "de" | "trung-binh" | "kho".
	Difficulty string `json:"difficulty" gorm:"type:varchar(32);index"`
	// Source là nguồn (tùy chọn).
	Source string `json:"source" gorm:"type:varchar(255)"`
	// CreatedAt / UpdatedAt là thời gian tạo/cập nhật.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_puzzles.
func (ChessPuzzle) TableName() string { return "chess_puzzles" }

// ChessGameFilter là bộ lọc khi liệt kê ván đấu.
type ChessGameFilter struct {
	White  string
	Black  string
	ECO    string
	Result string
	// Keyword là tìm kiếm tự do (ILIKE trên slug/white/black/event) — dùng cho
	// autocomplete wikilink. Rỗng = không lọc.
	Keyword string
}

// ChessPuzzleFilter là bộ lọc khi liệt kê bài tập.
type ChessPuzzleFilter struct {
	Theme      string
	Difficulty string
	// Keyword là tìm kiếm tự do (ILIKE trên slug/title/theme) — dùng cho
	// autocomplete wikilink. Rỗng = không lọc.
	Keyword string
}

// ChessRefSearchItem là một mục gợi ý khi tìm tham chiếu cờ cho autocomplete
// wikilink (gõ "[["). Gộp chung mọi loại đối tượng cờ.
type ChessRefSearchItem struct {
	// Type là loại đối tượng: "game" | "puzzle" | "lesson" | "course".
	Type string `json:"type"`
	// Slug là slug trần (không tiền tố).
	Slug string `json:"slug"`
	// Ref là chuỗi tham chiếu đầy đủ "<type>/<slug>" để chèn vào [[...]].
	Ref string `json:"ref"`
	// Title là nhãn hiển thị thân thiện.
	Title string `json:"title"`
	// Subtitle là thông tin phụ (ECO/sự kiện, theme/độ khó, trình độ...).
	Subtitle string `json:"subtitle"`
}

// ChessPuzzleBundle là gói export/import 1 bài tập (không kèm ID/slug/tenant để khi
// import luôn tạo mới trong tenant đích).
type ChessPuzzleBundle struct {
	Title      string `json:"title"`
	FEN        string `json:"fen"`
	Solution   string `json:"solution"`
	Theme      string `json:"theme"`
	Difficulty string `json:"difficulty"`
	Source     string `json:"source"`
}
