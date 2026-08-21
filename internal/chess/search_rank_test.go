package chess

import "testing"

func TestScoreSearchHit_Ordering(t *testing.T) {
	n := SearchNeedle("tan cuoc")
	slugExact := ScoreSearchHit(n, "tan cuoc", "", "")
	titlePrefix := ScoreSearchHit(n, "khac", SearchNeedle("Tàn cuộc Vua–Xe"), "")
	titleWord := ScoreSearchHit(n, "khac", SearchNeedle("Bài về tan cuoc hay"), "")
	bodyOnly := ScoreSearchHit(n, "khac", "khong lien quan", SearchText("Chương này nói về tàn cuộc"))

	if !(slugExact > titlePrefix && titlePrefix > titleWord && titleWord > bodyOnly) {
		t.Errorf("thứ tự điểm sai: slug=%d prefix=%d word=%d body=%d",
			slugExact, titlePrefix, titleWord, bodyOnly)
	}
	if bodyOnly == 0 {
		t.Error("khớp trong nội dung vẫn phải có điểm")
	}
}

func TestScoreSearchHit_NoMatchIsZero(t *testing.T) {
	if got := ScoreSearchHit(SearchNeedle("ghim"), "khai-cuoc", "khai cuoc", "ly thuyet khai cuoc"); got != 0 {
		t.Errorf("không khớp phải trả 0, nhận %d", got)
	}
	if got := ScoreSearchHit("", "ghim", "ghim", "ghim"); got != 0 {
		t.Errorf("từ khóa rỗng phải trả 0, nhận %d", got)
	}
}

// Đây chính là cái bẫy substring của lọc thẻ kiểu `tags ILIKE '%kw%'`: thẻ
// "Mã" (slug "ma") khớp luôn mọi chuỗi có "ma" ở giữa từ — "mate", "tham",
// "nam"… Khớp TỪ TRỌN VẸN phải được điểm cao hơn hẳn khớp giữa từ.
func TestScoreSearchHit_WholeWordBeatsSubstring(t *testing.T) {
	n := SearchNeedle("Mã")
	whole := ScoreSearchHit(n, "", SearchNeedle("Bẫy khai cuộc cho Mã"), "")
	inside := ScoreSearchHit(n, "", SearchNeedle("Chiếu hết bằng mate"), "")
	if !(whole > inside) {
		t.Errorf("khớp trọn từ (%d) phải hơn khớp giữa từ (%d)", whole, inside)
	}
	if inside == 0 {
		t.Error("khớp giữa từ vẫn nên có điểm, chỉ là thấp hơn")
	}
}

func TestScoreSearchHit_AccentInsensitive(t *testing.T) {
	withMarks := ScoreSearchHit(SearchNeedle("Tàn cuộc"), "", SearchNeedle("Tàn cuộc Vua–Xe"), "")
	without := ScoreSearchHit(SearchNeedle("tan cuoc"), "", SearchNeedle("Tàn cuộc Vua–Xe"), "")
	if withMarks != without || withMarks == 0 {
		t.Errorf("gõ có dấu (%d) và không dấu (%d) phải cho cùng điểm khác 0", withMarks, without)
	}
}

func TestSnippet_KeepsAccentsAndMarksEllipsis(t *testing.T) {
	raw := "Phần mở đầu không liên quan. Ở đây ta bàn về đòn ghim tuyệt đối và cách khai thác nó trong tàn cuộc."
	got := Snippet(raw, SearchNeedle("ghim"), 60)
	if got == "" {
		t.Fatal("phải trả về đoạn trích")
	}
	if !containsAny(got, "ghim") {
		t.Errorf("đoạn trích phải chứa vùng khớp, nhận %q", got)
	}
	if []rune(got)[0] != '…' {
		t.Errorf("khớp ở giữa văn bản thì đoạn trích phải mở đầu bằng dấu lược, nhận %q", got)
	}
}

func TestSnippet_ShortTextReturnedWhole(t *testing.T) {
	raw := "Đòn ghim"
	if got := Snippet(raw, SearchNeedle("ghim"), 160); got != raw {
		t.Errorf("văn bản ngắn phải giữ nguyên, nhận %q", got)
	}
}

func containsAny(s, sub string) bool {
	return len(sub) == 0 || len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
