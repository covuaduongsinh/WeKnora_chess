package types

import "time"

// ChessCourse là một khóa học cờ vua (đơn vị đào tạo cấp cao nhất).
// Thuộc về một tenant; chứa nhiều ChessLesson.
type ChessCourse struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu khóa học.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là định danh thân thiện (duy nhất theo tenant) làm đích wikilink
	// [[course/<slug>]]. Sinh ở tầng service khi tạo; ổn định sau đó.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(255)"`
	// Title là tên khóa học.
	Title string `json:"title" gorm:"type:varchar(255);not null"`
	// Description là mô tả ngắn.
	Description string `json:"description" gorm:"type:text"`
	// Level là trình độ: "co-ban" | "trung-cap" | "nang-cao".
	Level string `json:"level" gorm:"type:varchar(32)"`
	// CoverURL là ảnh bìa (tùy chọn).
	CoverURL string `json:"cover_url" gorm:"type:varchar(512)"`
	// SortOrder là thứ tự hiển thị.
	SortOrder int `json:"sort_order" gorm:"default:0"`
	// CreatedAt / UpdatedAt là thời gian tạo/cập nhật.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// LessonCount là số bài học (tính toán, không lưu).
	LessonCount int64 `json:"lesson_count" gorm:"-"`
}

// TableName ánh xạ tới bảng chess_courses.
func (ChessCourse) TableName() string { return "chess_courses" }

// ChessLesson là một bài học trong khóa học.
// Nội dung là markdown — có thể nhúng khối ```chess (FEN/PGN) để hiển thị bàn cờ.
type ChessLesson struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là định danh thân thiện (duy nhất theo tenant) làm đích wikilink
	// [[lesson/<slug>]]. Sinh ở tầng service khi tạo; ổn định sau đó.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(255)"`
	// CourseID là khóa học chứa bài học này.
	CourseID string `json:"course_id" gorm:"type:varchar(36);index"`
	// Title là tên bài học.
	Title string `json:"title" gorm:"type:varchar(255);not null"`
	// Content là nội dung markdown (có thể chứa khối ```chess).
	Content string `json:"content" gorm:"type:text"`
	// FEN là thế cờ chính của bài (tùy chọn).
	FEN string `json:"fen" gorm:"type:varchar(128)"`
	// PGN là ván minh họa (tùy chọn).
	PGN string `json:"pgn" gorm:"type:text"`
	// SortOrder là thứ tự bài trong khóa.
	SortOrder int `json:"sort_order" gorm:"default:0"`
	// CreatedAt / UpdatedAt là thời gian tạo/cập nhật.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_lessons.
func (ChessLesson) TableName() string { return "chess_lessons" }

// ChessLessonBundle là gói export/import 1 bài học (chỉ nội dung, KHÔNG kèm ID/slug
// để khi import luôn tạo mới trong tenant đích).
type ChessLessonBundle struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	FEN       string `json:"fen"`
	PGN       string `json:"pgn"`
	SortOrder int    `json:"sort_order"`
}

// ChessCourseBundle là gói export/import 1 khóa học kèm bài học. Dùng cho cả xuất
// (sao lưu/chia sẻ) lẫn nhập (tạo mới). Không chứa ID/slug/tenant.
type ChessCourseBundle struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Level       string              `json:"level"`
	CoverURL    string              `json:"cover_url"`
	SortOrder   int                 `json:"sort_order"`
	Lessons     []ChessLessonBundle `json:"lessons"`
}
