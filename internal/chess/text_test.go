package chess

import "testing"

// Bộ test này chạy được THẬT trên máy Windows (gói internal/chess không kéo
// internal/types nên không dính crash gojieba) — nên đây là nơi khóa hành vi
// khử dấu, thứ mà cả slug lẫn cột search_text đều phụ thuộc.

func TestFoldVN_StripsToneMarksAndDStroke(t *testing.T) {
	cases := map[string]string{
		"Khai cuộc":     "Khai cuoc",
		"Tàn cuộc":      "Tan cuoc",
		"Đòn mở":        "Don mo",
		"đóng đinh":     "dong dinh",
		"Chiếu hết":     "Chieu het",
		"Tốt thông":     "Tot thong",
		"Nhập thành":    "Nhap thanh",
		"Cờ vua Dương":  "Co vua Duong",
		"already ascii": "already ascii",
	}
	for in, want := range cases {
		if got := FoldVN(in); got != want {
			t.Errorf("FoldVN(%q) = %q, muốn %q", in, got, want)
		}
	}
}

// Đây là bất biến quan trọng nhất: gõ có dấu hay không dấu phải cho CÙNG chuỗi
// tìm kiếm, nếu không cột search_text và từ khóa sẽ không bao giờ gặp nhau.
func TestSearchNeedle_AccentedAndPlainConverge(t *testing.T) {
	pairs := [][2]string{
		{"Khai cuộc", "khai cuoc"},
		{"KHAI CUỘC", "Khai Cuoc"},
		{"Tàn cuộc Vua–Xe", "tan cuoc Vua–Xe"},
		{"  Ghim   ", "ghim"},
	}
	for _, p := range pairs {
		a, b := SearchNeedle(p[0]), SearchNeedle(p[1])
		if a != b {
			t.Errorf("SearchNeedle(%q)=%q khác SearchNeedle(%q)=%q", p[0], a, p[1], b)
		}
	}
}

func TestSearchText_JoinsFoldsAndCollapsesWhitespace(t *testing.T) {
	got := SearchText("tan-cuoc-vua-xe", "Tàn cuộc Vua–Xe", "", "  Cấp   độ  Xe ")
	want := "tan-cuoc-vua-xe tan cuoc vua–xe cap do xe"
	if got != want {
		t.Errorf("SearchText = %q, muốn %q", got, want)
	}
}

func TestSearchText_EmptyInputs(t *testing.T) {
	if got := SearchText(); got != "" {
		t.Errorf("không có mảnh nào phải trả rỗng, nhận %q", got)
	}
	if got := SearchText("", "  ", ""); got != "" {
		t.Errorf("toàn khoảng trắng phải trả rỗng, nhận %q", got)
	}
	if got := SearchNeedle("   "); got != "" {
		t.Errorf("từ khóa toàn khoảng trắng phải trả rỗng, nhận %q", got)
	}
}

// Cắt ở 8000 ký tự KHÔNG được tạo byte UTF-8 dở dang — chuỗi hỏng sẽ làm
// Postgres từ chối cả câu INSERT.
func TestSearchText_TruncatesOnRuneBoundary(t *testing.T) {
	long := ""
	for len(long) < 12000 {
		long += "cờ vua Dương Sinh "
	}
	got := SearchText(long)
	if len(got) > 8000 {
		t.Fatalf("độ dài = %d, phải <= 8000", len(got))
	}
	for i, r := range got {
		if r == '\uFFFD' {
			t.Fatalf("có rune hỏng ở vị trí %d", i)
		}
	}
}
