package chess

import "strings"

// search_rank.go chấm điểm mức khớp của một kết quả tìm kiếm.
//
// VÌ SAO CHẤM Ở GO chứ không dùng ts_rank của Postgres: tìm kiếm hợp nhất phải
// trộn kết quả từ 8 bảng khác nhau, mỗi bảng có cấu trúc và cột khác nhau. Một
// câu UNION có ts_rank sẽ vừa khó đọc vừa không chạy được trên bản SQLite
// ("lite"). Chấm bằng Go cho ra thang điểm CHUNG cho mọi loại, giải thích được
// bằng lời, và test được mà không cần cơ sở dữ liệu.
//
// Gói này cố ý không biết gì về kiểu dữ liệu cờ — chỉ nhận chuỗi.

// Thang điểm. Khoảng cách giữa các bậc đủ rộng để một kết quả khớp tiêu đề
// không bao giờ bị một kết quả chỉ khớp nội dung vượt qua.
const (
	// ScoreSlugExact: gõ đúng slug — gần như chắc chắn là thứ người dùng muốn.
	ScoreSlugExact = 100
	// ScoreTitleExact: tiêu đề trùng khít từ khóa.
	ScoreTitleExact = 90
	// ScoreTitlePrefix: tiêu đề BẮT ĐẦU bằng từ khóa ("tàn cuộc" → "Tàn cuộc Vua–Xe").
	ScoreTitlePrefix = 70
	// ScoreTitleWord: từ khóa là một TỪ trọn vẹn trong tiêu đề.
	ScoreTitleWord = 55
	// ScoreTitleContains: tiêu đề chứa từ khóa ở giữa một từ.
	ScoreTitleContains = 40
	// ScoreBodyWord: từ khóa là một từ trọn vẹn trong phần nội dung.
	ScoreBodyWord = 25
	// ScoreBodyContains: nội dung có chứa từ khóa.
	ScoreBodyContains = 10
)

// ScoreSearchHit chấm một kết quả. Cả ba tham số nội dung PHẢI đã được chuẩn
// hóa bằng SearchNeedle/SearchText (hạ chữ thường + khử dấu), needle cũng vậy.
// Trả 0 khi không khớp gì — caller nên loại bỏ.
//
// body thường là cột search_text; truyền "" nếu loại nội dung đó không có.
func ScoreSearchHit(needle, slug, title, body string) int {
	if needle == "" {
		return 0
	}
	score := 0
	if slug != "" {
		switch {
		case slug == needle:
			score = ScoreSlugExact
		case strings.HasPrefix(slug, needle):
			// Slug khớp tiền tố mạnh gần bằng tiêu đề khớp tiền tố, vì slug
			// vốn sinh RA TỪ tiêu đề.
			score = ScoreTitlePrefix
		}
	}
	if title != "" {
		switch {
		case title == needle:
			score = max2(score, ScoreTitleExact)
		case strings.HasPrefix(title, needle):
			score = max2(score, ScoreTitlePrefix)
		case containsWord(title, needle):
			score = max2(score, ScoreTitleWord)
		case strings.Contains(title, needle):
			score = max2(score, ScoreTitleContains)
		}
	}
	if body != "" {
		switch {
		case containsWord(body, needle):
			score = max2(score, ScoreBodyWord)
		case strings.Contains(body, needle):
			score = max2(score, ScoreBodyContains)
		}
	}
	return score
}

// containsWord cho biết needle xuất hiện trong hay như một TỪ TRỌN VẸN — tức
// hai đầu là ranh giới chuỗi hoặc khoảng trắng.
//
// Phân biệt được "ghim" trong "đòn ghim" với "ghim" trong "ghimbap" là điểm
// mấu chốt: đúng cái bẫy substring khiến bộ lọc thẻ cũ cho "pin" khớp cả
// "opening".
func containsWord(haystack, needle string) bool {
	if needle == "" {
		return false
	}
	from := 0
	for {
		i := strings.Index(haystack[from:], needle)
		if i < 0 {
			return false
		}
		i += from
		startOK := i == 0 || haystack[i-1] == ' '
		end := i + len(needle)
		endOK := end == len(haystack) || haystack[end] == ' '
		if startOK && endOK {
			return true
		}
		from = i + 1
		if from >= len(haystack) {
			return false
		}
	}
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Snippet trích một đoạn ngắn quanh vị trí khớp để hiển thị dưới tiêu đề.
//
// Nhận văn bản GỐC (còn dấu) nhưng dò vị trí trên bản đã khử dấu, nên đoạn
// trích giữ nguyên dấu tiếng Việt trong khi vẫn tìm đúng chỗ khi người dùng gõ
// không dấu. Hai chuỗi luôn cùng độ dài byte? KHÔNG — nên chỉ dùng chỉ số của
// bản đã fold để ước lượng, rồi cắt theo ranh giới rune trên chuỗi gốc.
func Snippet(raw, needle string, width int) string {
	if raw == "" || needle == "" {
		return ""
	}
	if width <= 0 {
		width = 160
	}
	folded := strings.ToLower(FoldVN(raw))
	idx := strings.Index(folded, needle)
	if idx < 0 {
		return truncateRunes(strings.TrimSpace(raw), width)
	}
	// Ước lượng vị trí tương ứng trên chuỗi gốc theo TỶ LỆ số rune, vì khử dấu
	// có thể đổi độ dài byte (đ→d) nhưng KHÔNG đổi số rune.
	runeIdx := len([]rune(folded[:idx]))
	rs := []rune(raw)
	start := runeIdx - width/3
	if start < 0 {
		start = 0
	}
	end := start + width
	if end > len(rs) {
		end = len(rs)
	}
	out := strings.TrimSpace(string(rs[start:end]))
	if start > 0 {
		out = "…" + out
	}
	if end < len(rs) {
		out += "…"
	}
	return out
}

func truncateRunes(s string, width int) string {
	rs := []rune(s)
	if len(rs) <= width {
		return s
	}
	return strings.TrimSpace(string(rs[:width])) + "…"
}
