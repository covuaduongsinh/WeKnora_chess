package tools

import "testing"

// setBoardFEN là cổng cuối cùng trước khi FEN rời backend ra frontend. Thư viện
// bàn cờ (cm-chessboard) lệch NGUYÊN MỘT HÀNG và nuốt hàng quân cuối khi chuỗi
// dư/thiếu token, mà không throw — nên mọi sai lệch ở đây đều là lỗi CÂM.
func TestSetBoardFEN(t *testing.T) {
	const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	cases := []struct {
		name    string
		in      string
		wantFEN string
		invalid bool
	}{
		{
			name:    "FEN đầy đủ giữ nguyên",
			in:      startFEN,
			wantFEN: startFEN,
		},
		{
			name: "FEN cụt được bù đủ 6 trường",
			in:   "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR",
			// LƯU Ý: quyền nhập thành bù bằng "-", KHÔNG phải "KQkq" — NormalizeFEN
			// không thể suy ra quyền nhập thành từ mỗi bố cục quân. Vô hại cho việc
			// hiển thị (cm-chessboard chỉ đọc trường bố cục quân), nhưng đừng dùng
			// chuỗi đã chuẩn hóa này làm đầu vào cho engine nếu cần đúng luật.
			wantFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
		},
		{
			name: "thế cờ KHÔNG có Vua vẫn hợp lệ (ngân hàng thế cờ cố ý cho phép)",
			// Nếu case này hỏng, các thế cờ giản lược dạy trẻ sẽ hiện hộp báo lỗi.
			in:      "8/8/8/3ppp2/3PPP2/8/8/8 w - - 0 1",
			wantFEN: "8/8/8/3ppp2/3PPP2/8/8/8 w - - 0 1",
		},
		{
			name:    "thiếu một hàng bị đánh dấu hỏng thay vì lọt ra ngoài",
			in:      "rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantFEN: "rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			invalid: true,
		},
		{
			name:    "chuỗi rác bị đánh dấu hỏng",
			in:      "khong-phai-fen",
			wantFEN: "khong-phai-fen",
			invalid: true,
		},
		{
			name:    "chuỗi rỗng bị đánh dấu hỏng",
			in:      "",
			wantFEN: "",
			invalid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{"display_type": "chess_board"}
			setBoardFEN(data, tc.in)

			if got := data["fen"]; got != tc.wantFEN {
				t.Errorf("fen = %q, muốn %q", got, tc.wantFEN)
			}
			_, marked := data["fen_invalid"]
			if marked != tc.invalid {
				t.Errorf("fen_invalid có mặt = %v, muốn %v", marked, tc.invalid)
			}
			if tc.invalid && data["fen_invalid"] != true {
				t.Errorf("fen_invalid = %v, muốn true", data["fen_invalid"])
			}
		})
	}
}
