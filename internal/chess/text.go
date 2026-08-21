package chess

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// text.go giữ phép KHỬ DẤU TIẾNG VIỆT dùng chung cho cả tầng service (sinh
// slug) lẫn tầng repository (dựng cột search_text).
//
// Trước đây logic này chỉ nằm trong service/chess_slug.go. Khi repository cũng
// cần nó để dựng cột tìm kiếm, chép sang là chắc chắn trôi lệch: slug và
// search_text phải khử dấu Y HỆT nhau, nếu không gõ "khai cuoc" sẽ khớp thẻ
// nhưng trượt tiêu đề (hoặc ngược lại). Một bản cài đặt, hai nơi dùng.
//
// internal/chess CỐ Ý không import internal/types — giữ gói này thuần để cả
// repository lẫn service đều nhập được mà không tạo vòng phụ thuộc.

// diacriticFold tách tổ hợp (NFD) → bỏ dấu thanh/dấu phụ (combining marks Mn)
// → NFC. Khử dấu ROBUST bất kể đầu vào ở dạng NFC hay NFD (không lệ thuộc bảng
// ký tự dựng sẵn vốn dễ trật khi normalize khác nhau).
var diacriticFold = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// FoldVN bỏ dấu tiếng Việt. Xử lý đ/Đ trước (đây là ký tự riêng, KHÔNG phải
// nguyên âm mang dấu tổ hợp, nên NFD không tách được) rồi mới tách dấu.
func FoldVN(s string) string {
	s = strings.NewReplacer("đ", "d", "Đ", "D", "ð", "d").Replace(s)
	out, _, err := transform.String(diacriticFold, s)
	if err != nil {
		return s
	}
	return out
}

// SearchText dựng chuỗi tìm kiếm đã CHUẨN HÓA từ nhiều mảnh nội dung: gộp lại,
// hạ chữ thường, khử dấu, rồi gom mọi khoảng trắng thừa về một dấu cách.
//
// Nhờ cột này, ô tìm hoạt động khi gõ KHÔNG DẤU ("tan cuoc vua xe" ra "Tàn
// cuộc Vua–Xe") — quan trọng nhất lúc tra nhanh trên điện thoại. Truy vấn phải
// dùng LIKE (không phải ILIKE) vì cả hai vế đều đã hạ chữ thường sẵn.
//
// Cắt ở 8000 ký tự: đủ để phủ tiêu đề + tóm tắt + phần đầu nội dung, mà không
// biến cột thành bản sao của cả chương sách. Tìm sâu trong nội dung dài là
// việc của index full-text (tsvector), không phải của cột này.
func SearchText(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p)
	}
	out := strings.ToLower(FoldVN(b.String()))
	out = strings.Join(strings.Fields(out), " ")
	if len(out) > 8000 {
		// Cắt theo ranh giới rune để không tạo byte UTF-8 dở dang.
		cut := 8000
		for cut > 0 && !utf8Start(out[cut]) {
			cut--
		}
		out = strings.TrimSpace(out[:cut])
	}
	return out
}

// utf8Start cho biết một byte có phải byte ĐẦU của một rune không.
func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

// SearchNeedle chuẩn hóa TỪ KHÓA người dùng gõ về cùng dạng với SearchText.
// Trả "" khi từ khóa rỗng sau khi chuẩn hóa — caller nên bỏ qua mệnh đề lọc.
func SearchNeedle(q string) string {
	return strings.Join(strings.Fields(strings.ToLower(FoldVN(q))), " ")
}
