package types

import "time"

// chess_search.go định nghĩa TÌM KIẾM HỢP NHẤT cho lớp cờ — một ô tìm duy nhất
// trả về kết quả của cả 8 loại nội dung, đã XẾP HẠNG chung.
//
// Khác /chess/refs/search (autocomplete khi gõ "[["): endpoint đó nối 8 khối
// theo thứ tự CỨNG trong code, cắt mỗi loại vài mục, và KHÔNG chấm điểm — nên
// một mục khớp đúng tiêu đề vẫn có thể nằm dưới một mục khớp mờ chỉ vì loại
// của nó được liệt kê trước. Nó phục vụ chèn wikilink, không phải tra cứu.

// ChessSearchQuery là tham số của một lần tìm hợp nhất.
type ChessSearchQuery struct {
	// Keyword là từ khóa thô người dùng gõ (có dấu hay không đều được).
	Keyword string
	// Types giới hạn theo loại nội dung; rỗng = mọi loại.
	Types []string
	// Level lọc theo cấp độ 6 bậc; rỗng = mọi cấp.
	Level string
	// Tags lọc theo hệ thẻ thống nhất.
	Tags ChessTagSelector
	// Status lọc theo trạng thái xuất bản (chỉ áp cho sách/bài viết).
	Status   string
	Page     int
	PageSize int
}

// ChessSearchHit là một kết quả đã chấm điểm.
type ChessSearchHit struct {
	ChessType string `json:"chess_type"`
	ChessID   string `json:"chess_id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	// Snippet là đoạn trích quanh vùng khớp, GIỮ NGUYÊN dấu tiếng Việt.
	Snippet string `json:"snippet"`
	Level   string `json:"level"`
	Status  string `json:"status"`
	// Score là điểm khớp (thang chung cho mọi loại — xem chess.ScoreSearchHit).
	Score     int         `json:"score"`
	Tags      []*ChessTag `json:"tags,omitempty"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ChessSearchPage là một trang kết quả tìm hợp nhất.
type ChessSearchPage struct {
	Items []ChessSearchHit `json:"items"`
	// Total là tổng số kết quả ĐÃ GOM được (sau khi chấm điểm và loại bỏ mục
	// không khớp), không phải ước lượng từ COUNT của từng bảng.
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	ByType   map[string]int   `json:"by_type"`
	Took     map[string]int64 `json:"-"`
	// Truncated báo có loại nào chạm trần quét mỗi-loại hay không — khi đó kết
	// quả vẫn đúng nhưng CÓ THỂ chưa đủ, và người dùng nên thu hẹp từ khóa.
	// Thà nói thật còn hơn im lặng như trần Limit(500) cũ.
	Truncated bool `json:"truncated"`
}
