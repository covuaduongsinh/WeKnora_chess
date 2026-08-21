package types

import "time"

// Trạng thái xuất bản NỘI BỘ của bài viết (cột chess_articles.status) — cùng
// khuôn ChessBookStatus*.
const (
	// ChessArticleStatusDraft là bản thảo — KHÔNG được index vào KB tri thức cờ.
	ChessArticleStatusDraft = "draft"
	// ChessArticleStatusPublished là đã duyệt — được index (nếu CHESS_KB_INDEX bật).
	ChessArticleStatusPublished = "published"
)

// chess_article.go định nghĩa thực thể "Ngân hàng bài viết" — thực thể cờ thứ
// 7 (bên cạnh game/puzzle/lesson/course/position/book+chapter). Khác mọi thực
// thể khác: bài viết là trang tri thức ĐỘC LẬP về khái niệm/thuật ngữ/kinh
// nghiệm — KHÔNG thuộc sách hay khóa học nào. Đây là ontology "Concept" trong
// 02-mien-co-vua.md §2.5: bài viết vừa là ĐÍCH cho [[article/<slug>]] từ mọi
// thực thể khác, vừa là NGUỒN (nội dung có thể chứa wikilink trỏ ra ván/thế
// cờ/bài tập/sách).
//
// Tổ chức = 2 trục vuông góc: ChessArticleTopic (cây chuyên mục tối đa 2 tầng,
// điều hướng DỌC — xem chess_library_article.go) cộng Category/Level/Tags/
// Status (lọc NGANG). Chỉ status="published" được index RAG (cùng quy tắc
// sách — xem ChessKnowledgeIndexer.IndexArticle).
type ChessArticle struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là định danh thân thiện (duy nhất theo tenant) làm đích wikilink
	// [[article/<slug>]]. Sinh ở tầng service khi tạo; ổn định sau đó.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(255)"`
	// Title là tiêu đề bài viết (vd "Ghim (Pin) là gì?").
	Title string `json:"title" gorm:"type:varchar(255)"`
	// Summary là tóm tắt ngắn — hiện ở danh sách/popup wikilink, và đưa vào
	// văn bản index RAG làm mồi tìm kiếm.
	Summary string `json:"summary" gorm:"type:varchar(1000)"`
	// Aliases là bí danh/từ đồng nghĩa DẠNG HIỂN THỊ (CSV, vd "Pin, Đóng
	// đinh") — nguồn THẬT cho resolve wikilink/fuzzy là bảng chess_slug_aliases
	// (kind="synonym", đồng bộ ở tầng service qua syncArticleAliases), cột
	// này chỉ để hiển thị + đưa vào văn bản index RAG.
	Aliases string `json:"aliases" gorm:"type:varchar(500)"`
	// Category là phân loại: "khai-niem"|"thuat-ngu"|"kinh-nghiem"|
	// "huong-dan"|"phuong-phap-day"|... (xem chessArticleOptions.ts).
	Category string `json:"category" gorm:"type:varchar(32);index"`
	// Level là cấp độ trong lộ trình 6 cấp Dương Sinh: "tot"|"ma"|"tuong"|"xe"|
	// "hau"|"vua" (rỗng = mọi cấp).
	Level string `json:"level" gorm:"type:varchar(16);index"`
	// Tags là CSV nhẹ (khớp mẫu Tags của ChessPosition/ChessBook).
	Tags string `json:"tags" gorm:"type:varchar(255)"`
	// SearchText là bản KHỬ DẤU + hạ chữ thường của các trường tìm kiếm được,
	// do repository dựng lại ở MỖI lần ghi (chess.SearchText). Nhờ nó ô tìm
	// hoạt động khi gõ không dấu. Không lộ ra API — đây là chi tiết lưu trữ.
	SearchText string `json:"-" gorm:"column:search_text;type:text"`
	// Status là trạng thái xuất bản NỘI BỘ: "draft" (bản thảo, KHÔNG index RAG)
	// | "published" (đã duyệt, index vào KB tri thức cờ nếu CHESS_KB_INDEX bật).
	Status string `json:"status" gorm:"type:varchar(16);default:draft;index"`
	// CoverURL là ảnh bìa bài viết (tùy chọn).
	CoverURL string `json:"cover_url" gorm:"column:cover_url;type:varchar(512)"`
	// Content là nội dung markdown (có thể chứa khối ```chess + wikilink) —
	// ĐƯỢC chứa wikilink [[...]]/![[...]] để bài viết trở thành nguồn phát
	// liên kết (xem syncArticleChessRefs).
	Content string `json:"content" gorm:"type:text"`
	// SortOrder là thứ tự hiển thị mặc định (trong danh sách không lọc theo chuyên mục).
	SortOrder int `json:"sort_order" gorm:"default:0"`
	// CreatedAt / UpdatedAt là thời gian tạo/cập nhật.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_articles.
func (ChessArticle) TableName() string { return "chess_articles" }

// ChessArticleTopic là một chuyên mục bài viết — cây TỐI ĐA 2 TẦNG qua
// ParentID (ép ở tầng service, không CHECK ở DB). KHÔNG phải đích wikilink
// (chỉ điều hướng UI, giống ChessShelf).
type ChessArticleTopic struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID    uint64    `json:"tenant_id" gorm:"index"`
	ParentID    string    `json:"parent_id" gorm:"column:parent_id;type:varchar(36)"` // "" = chuyên mục gốc
	Slug        string    `json:"slug" gorm:"column:slug;type:varchar(255)"`
	Title       string    `json:"title" gorm:"type:varchar(255)"`
	Description string    `json:"description" gorm:"type:varchar(1000)"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// ArticleCount là số bài viết trong chuyên mục (tính toán, không lưu).
	ArticleCount int64 `json:"article_count" gorm:"-"`
}

// TableName ánh xạ tới bảng chess_article_topics.
func (ChessArticleTopic) TableName() string { return "chess_article_topics" }

// ChessArticleTopicItem là bản ghi nối chuyên mục↔bài viết (nhiều-nhiều, mẫu
// ChessShelfBook).
type ChessArticleTopicItem struct {
	TenantID  uint64    `json:"tenant_id" gorm:"index"`
	TopicID   string    `json:"topic_id" gorm:"column:topic_id;type:varchar(36);primaryKey"`
	ArticleID string    `json:"article_id" gorm:"column:article_id;type:varchar(36);primaryKey"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_article_topic_items.
func (ChessArticleTopicItem) TableName() string { return "chess_article_topic_items" }

// ChessArticleImage là một ảnh đã upload để chèn vào nội dung bài viết (sao y
// ChessBookImage — markdown tham chiếu qua URL ổn định
// GET /chess/articles/images/:id, KHÔNG dùng presigned URL có hạn dùng).
type ChessArticleImage struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID  uint64    `json:"tenant_id" gorm:"index"`
	ArticleID string    `json:"article_id" gorm:"column:article_id;type:varchar(36);index"`
	Path      string    `json:"-" gorm:"type:varchar(512)"` // provider:// path (FileService) — không lộ ra API
	FileName  string    `json:"file_name" gorm:"column:file_name;type:varchar(255)"`
	Mime      string    `json:"mime" gorm:"type:varchar(64)"`
	Size      int64     `json:"size" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_article_images.
func (ChessArticleImage) TableName() string { return "chess_article_images" }

// ChessArticleRevision là một bản PRE-IMAGE của bài viết (lưu TRƯỚC khi ghi
// đè), chỉ tạo khi Title/Content thực sự đổi — sao y ChessChapterRevision.
type ChessArticleRevision struct {
	ID             string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID       uint64    `json:"tenant_id" gorm:"index"`
	ArticleID      string    `json:"article_id" gorm:"column:article_id;type:varchar(36);index"`
	RevisionNumber int       `json:"revision_number" gorm:"column:revision_number;default:0"`
	Title          string    `json:"title" gorm:"type:varchar(255)"`
	Content        string    `json:"content" gorm:"type:text"`
	Summary        string    `json:"summary" gorm:"type:varchar(255)"`
	CreatedBy      string    `json:"created_by" gorm:"column:created_by;type:varchar(64)"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_article_revisions.
func (ChessArticleRevision) TableName() string { return "chess_article_revisions" }

// ---- Filter ----

// ChessArticleFilter là bộ lọc khi liệt kê bài viết.
type ChessArticleFilter struct {
	// TopicID lọc bài thuộc MỘT chuyên mục cụ thể (qua bảng nối
	// chess_article_topic_items). Rỗng = không lọc theo chuyên mục.
	TopicID  string
	Category string
	Level    string
	Status   string
	// Keyword là tìm kiếm tự do (ILIKE trên slug/title/aliases/summary/tags) —
	// dùng cho autocomplete wikilink. Rỗng = không lọc.
	Keyword string
	// Tags lọc theo hệ thẻ thống nhất (chess_tag_items). Rỗng = không lọc.
	Tags ChessTagSelector
	// Page / PageSize: phân trang thật. PageSize <= 0 nghĩa là KHÔNG phân
	// trang (trả toàn bộ) — giữ nguyên hành vi cho các nơi cần đủ dữ liệu như
	// export, backfill và picker chèn wikilink.
	Page     int
	PageSize int
}

// ChessArticleTopicFilter là bộ lọc khi liệt kê chuyên mục.
type ChessArticleTopicFilter struct {
	// ParentID lọc theo chuyên mục cha; "" (mặc định zero-value) không lọc —
	// dùng ParentIDSet để phân biệt "lọc theo gốc" (ParentID="", ParentIDSet=true).
	ParentID    string
	ParentIDSet bool
	Keyword     string
}

// ---- Bundle export/import (không kèm ID/slug/tenant để import luôn tạo mới) ----

// ChessArticleBundle là gói export/import 1 bài viết.
type ChessArticleBundle struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Aliases  string `json:"aliases"`
	Category string `json:"category"`
	Level    string `json:"level"`
	Tags     string `json:"tags"`
	Status   string `json:"status"`
	CoverURL string `json:"cover_url"`
	Content  string `json:"content"`
}
