package types

import "time"

// chess_tag.go định nghĩa HỆ THẺ THỐNG NHẤT cho toàn bộ lớp cờ vua — trục
// phân loại NGANG duy nhất phủ cả 8 loại nội dung (game/puzzle/lesson/course/
// position/book/chapter/article), thay cho cột CSV `tags` vốn chỉ có ở 3 loại
// và chỉ lọc được bằng ILIKE substring.
//
// Ba trục phân loại của lớp cờ sau migration 000911:
//  1. Cấp độ        — cột `level` trên từng bảng, từ vựng ĐÓNG 6 bậc Dương Sinh.
//  2. Nhóm nội dung — thẻ hệ thống (Kind="group"), từ vựng ĐÓNG 8 nhóm.
//  3. Thẻ tự do     — thẻ Kind="free".
//
// Nhóm nội dung CỐ Ý đi qua chính hệ thẻ thay vì thành cột riêng trên 9 bảng:
// một cơ chế lưu trữ, một bảng nối, một đường lọc, một trang quản lý — và
// thêm loại nội dung thứ 10 sau này không cần đổi schema. Trên giao diện nó
// vẫn hiện thành ô chọn "Nhóm nội dung" riêng.

// Phân loại thẻ (cột chess_tags.kind).
const (
	// ChessTagKindGroup là thẻ HỆ THỐNG "nhóm nội dung" — từ vựng đóng, seed
	// sẵn cho mỗi tenant, KHÔNG cho xóa qua UI (xóa xong lần seed sau tạo lại
	// thì sinh trùng lặp và loạn thống kê).
	ChessTagKindGroup = "group"
	// ChessTagKindFree là thẻ tự do người dùng gõ.
	ChessTagKindFree = "free"
)

// Tám nhóm nội dung — bám đúng .claude/memory/02-mien-co-vua.md muc 2.1 để
// phân loại khớp với cách tổ chức knowledge base đã thống nhất từ trước.
const (
	ChessTagGroupRules      = "luat"        // Luật cờ (FIDE + luật VN, trọng tài)
	ChessTagGroupOpening    = "khai-cuoc"   // Khai cuộc (lý thuyết, bẫy, ECO)
	ChessTagGroupMiddlegame = "trung-cuoc"  // Trung cuộc (chiến lược, kế hoạch, cấu trúc tốt)
	ChessTagGroupEndgame    = "tan-cuoc"    // Tàn cuộc (cơ bản đến lý thuyết)
	ChessTagGroupTactics    = "chien-thuat" // Chiến thuật (bắt đôi/ghim/xiên/chiếu hết)
	ChessTagGroupCurriculum = "giao-trinh"  // Giáo trình & giáo án 6 cấp
	ChessTagGroupCulture    = "van-hoa"     // Văn hóa & lịch sử (kỳ thủ, ván kinh điển)
	ChessTagGroupOps        = "van-hanh"    // Vận hành Dương Sinh (CLB, giải đấu)
)

// ChessTagGroupSeed là một nhóm nội dung dựng sẵn. Đây là NGUỒN SỰ THẬT của
// từ vựng đóng ở phía Go; bản đối ứng frontend là utils/chessTaxonomy.ts —
// sửa một bên PHẢI sửa bên kia.
type ChessTagGroupSeed struct {
	Slug        string
	Name        string
	Description string
	SortOrder   int
}

// ChessTagGroupSeeds liệt kê 8 nhóm nội dung theo đúng thứ tự hiển thị.
var ChessTagGroupSeeds = []ChessTagGroupSeed{
	{ChessTagGroupRules, "Luật cờ", "Luật FIDE và luật Việt Nam, cờ nhanh/chớp, xử lý trọng tài", 1},
	{ChessTagGroupOpening, "Khai cuộc", "Lý thuyết khai cuộc, bẫy, hệ thống theo mã ECO", 2},
	{ChessTagGroupMiddlegame, "Trung cuộc", "Chiến lược, kế hoạch, cấu trúc tốt", 3},
	{ChessTagGroupEndgame, "Tàn cuộc", "Tàn cuộc cơ bản đến tàn cuộc lý thuyết", 4},
	{ChessTagGroupTactics, "Chiến thuật", "Bắt đôi, ghim, xiên, đòn mở, chiếu hết", 5},
	{ChessTagGroupCurriculum, "Giáo trình", "Bài giảng 6 cấp Tốt đến Vua, giáo án, worksheet", 6},
	{ChessTagGroupCulture, "Văn hóa & lịch sử", "Kỳ thủ, ván kinh điển, chuyện truyền cảm hứng", 7},
	{ChessTagGroupOps, "Vận hành", "Quy trình câu lạc bộ, hợp tác trường, giải đấu", 8},
}

// ChessTag là một thẻ trong từ điển thẻ của tenant.
type ChessTag struct {
	// ID là định danh duy nhất (UUID).
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// TenantID là tenant sở hữu.
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	// Slug là dạng đã slug hóa + KHỬ DẤU của Name (slugifyChess) — khóa duy
	// nhất theo tenant, nên "Khai cuộc"/"khai-cuoc"/"KHAI CUOC" quy về một thẻ.
	Slug string `json:"slug" gorm:"column:slug;type:varchar(64)"`
	// Name là tên hiển thị GIỮ NGUYÊN DẤU (vd "Khai cuộc").
	Name string `json:"name" gorm:"type:varchar(128)"`
	// Kind là "group" (nhóm nội dung hệ thống) hoặc "free" (thẻ tự do).
	Kind string `json:"kind" gorm:"type:varchar(16);default:free;index"`
	// Description là mô tả ngắn (chủ yếu cho thẻ nhóm).
	Description string `json:"description" gorm:"type:varchar(500)"`
	// Color là màu chip tùy chọn (hex, hoặc rỗng = màu mặc định theo Kind).
	Color string `json:"color" gorm:"type:varchar(16)"`
	// UsageCount là CACHE số mục đang gắn thẻ này. Nguồn thật là bảng
	// chess_tag_items — dùng RecountTagUsage khi nghi ngờ lệch.
	UsageCount int `json:"usage_count" gorm:"column:usage_count;default:0"`
	// SortOrder là thứ tự hiển thị (thẻ nhóm dùng thứ tự cố định 1..8).
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName ánh xạ tới bảng chess_tags.
func (ChessTag) TableName() string { return "chess_tags" }

// ChessTagItem là bản ghi nối thẻ với một mục nội dung bất kỳ (đa hình qua
// ChessType, mẫu ChessArticleTopicItem/ChessShelfBook). ĐÂY là nguồn sự thật
// của việc gắn thẻ; cột CSV `tags` trên chess_positions/chess_books/
// chess_articles chỉ là bản hiển thị được ghi lại TỪ bảng này.
type ChessTagItem struct {
	TenantID uint64 `json:"tenant_id" gorm:"index"`
	TagID    string `json:"tag_id" gorm:"column:tag_id;type:varchar(36);primaryKey"`
	// ChessType là loại nội dung — dùng hằng ChessRefType* (game|puzzle|
	// lesson|course|position|book|chapter|article). Gõ chuỗi tay là lỗi CÂM.
	ChessType string    `json:"chess_type" gorm:"column:chess_type;type:varchar(16);primaryKey"`
	ChessID   string    `json:"chess_id" gorm:"column:chess_id;type:varchar(36);primaryKey"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName ánh xạ tới bảng chess_tag_items.
func (ChessTagItem) TableName() string { return "chess_tag_items" }

// ---- Filter ----

// ChessTagFilter là bộ lọc khi liệt kê thẻ.
type ChessTagFilter struct {
	// Kind lọc theo loại thẻ ("group" hoặc "free"); rỗng = cả hai.
	Kind string
	// Keyword tìm theo slug/name.
	Keyword string
	// OnlyUsed chỉ trả thẻ đang có ít nhất một mục gắn (UsageCount > 0).
	OnlyUsed bool
}

// ---- Lọc theo thẻ trên danh sách nội dung ----

// Chế độ khớp nhiều thẻ (ChessTagSelector.Mode).
const (
	// ChessTagModeAny khớp mục mang ÍT NHẤT MỘT thẻ trong danh sách (mặc định).
	ChessTagModeAny = "any"
	// ChessTagModeAll khớp mục mang ĐỦ TẤT CẢ thẻ trong danh sách.
	ChessTagModeAll = "all"
)

// ChessTagSelector là mệnh đề lọc theo thẻ, nhúng vào mọi filter danh sách
// nội dung cờ (ChessGameFilter, ChessPuzzleFilter...). Rỗng = không lọc.
type ChessTagSelector struct {
	// TagSlugs là các slug thẻ cần khớp (đã slug hóa ở tầng handler).
	TagSlugs []string
	// Mode là "any" (mặc định) hoặc "all".
	Mode string
}

// Active cho biết selector có thực sự lọc gì không.
func (s ChessTagSelector) Active() bool { return len(s.TagSlugs) > 0 }

// MatchAll cho biết có yêu cầu khớp ĐỦ mọi thẻ không.
func (s ChessTagSelector) MatchAll() bool { return s.Mode == ChessTagModeAll }

// ---- Kết quả tra cứu xuyên loại ----

// ChessTagItemRef là một mục nội dung mang thẻ, đã kèm đủ thông tin để hiển
// thị và điều hướng mà KHÔNG cần gọi thêm API theo từng loại — phục vụ trang
// "mục lục ngang" (bấm một thẻ, ra mọi loại nội dung mang thẻ đó).
type ChessTagItemRef struct {
	ChessType string    `json:"chess_type"`
	ChessID   string    `json:"chess_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Level     string    `json:"level"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChessTagItemPage là một trang kết quả của việc tra nội dung theo thẻ.
type ChessTagItemPage struct {
	Items []ChessTagItemRef `json:"items"`
	// Total là tổng số mục mang thẻ (theo bộ lọc loại nếu có) — cho phép
	// frontend hiện "đang xem 20/137" thay vì cắt câm như các list cờ hiện tại.
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	ByType   map[string]int64 `json:"by_type"`
}

// ChessTagBackfillResult là báo cáo của một lần chạy backfill CSV sang thẻ.
type ChessTagBackfillResult struct {
	// GroupsSeeded là số thẻ nhóm vừa được tạo mới (0 nếu đã seed từ trước).
	GroupsSeeded int `json:"groups_seeded"`
	// TagsCreated là số thẻ tự do sinh ra từ CSV cũ.
	TagsCreated int `json:"tags_created"`
	// LinksCreated là số liên kết thẻ với nội dung vừa ghi.
	LinksCreated int `json:"links_created"`
	// ByType đếm số mục đã xử lý theo từng loại nội dung.
	ByType map[string]int `json:"by_type"`
	// SearchTextByType đếm số bản ghi đã được tính lại cột search_text (khử
	// dấu). Chạy cùng lượt với backfill thẻ vì cả hai đều là "nạp lại dữ liệu
	// cũ" — người vận hành chỉ phải bấm MỘT nút.
	SearchTextByType map[string]int `json:"search_text_by_type"`
	// Warnings cảnh báo vận hành — đáng chú ý nhất là trường hợp một loại nội
	// dung trả về đúng trần 500 bản ghi của tầng repository: backfill khi đó
	// KHÔNG phủ hết dữ liệu và phải chạy lại sau khi có phân trang thật.
	Warnings []string `json:"warnings"`
}
