package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/WeKnora/internal/chess"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/utils"
)

var chessLookupOpeningTool = BaseTool{
	name: ToolChessLookupOpening,
	description: `Nhận diện tên khai cuộc và mã ECO từ một chuỗi nước đi mở đầu.

## Khi nào dùng
- Người dùng hỏi "đây là khai cuộc gì?" và đưa các nước đầu (PGN hoặc danh sách nước SAN).
- Cần gọi tên khai cuộc để giảng dạy.

## Đầu vào (cung cấp một trong hai)
- pgn: nội dung ván cờ dạng PGN.
- moves: danh sách nước đi dạng SAN, ví dụ ["e4","c5","Nf3"].

Không cần engine — dùng cơ sở dữ liệu khai cuộc nhúng sẵn. Trả về tên khai cuộc,
mã ECO và bàn cờ tương tác ở thế cờ sau các nước khai cuộc.`,
	schema: utils.GenerateSchema[ChessLookupOpeningInput](),
}

// ChessLookupOpeningInput là tham số cho tool chess_lookup_opening.
type ChessLookupOpeningInput struct {
	PGN   string   `json:"pgn,omitempty" jsonschema:"Ván cờ dạng PGN (cung cấp pgn hoặc moves)"`
	Moves []string `json:"moves,omitempty" jsonschema:"Danh sách nước đi SAN, ví dụ [\"e4\",\"c5\"]"`
}

// ChessLookupOpeningTool tra cứu khai cuộc theo bảng ECO nhúng.
type ChessLookupOpeningTool struct {
	BaseTool
}

// NewChessLookupOpeningTool tạo tool chess_lookup_opening.
func NewChessLookupOpeningTool() *ChessLookupOpeningTool {
	return &ChessLookupOpeningTool{BaseTool: chessLookupOpeningTool}
}

type ecoEntry struct {
	eco  string
	name string
}

// ecoOpenings là bảng khai cuộc Việt hoá (tên tiếng Việt cho các khai cuộc phổ
// biến), khóa là chuỗi nước SAN nối bằng dấu cách. Đây là OVERLAY: nó được đè lên
// dataset ECO nhúng đầy đủ (xem chess_openings_data.go) để tên tiếng Việt thắng ở
// các khai cuộc phổ biến; phần còn lại lấy tên tiếng Anh từ dataset. Việc tra cứu
// thực tế dùng openingIndex (đã gộp), KHÔNG dùng trực tiếp map này.
var ecoOpenings = map[string]ecoEntry{
	"e4":                                   {"B00", "Khai cuộc Vua tốt (1.e4)"},
	"e4 e5":                                {"C20", "Ván cờ mở (1.e4 e5)"},
	"e4 e5 Nf3":                            {"C40", "Khai cuộc Mã vua"},
	"e4 e5 Nf3 Nc6":                        {"C44", "Khai cuộc Mã vua kép"},
	"e4 e5 Nf3 Nc6 Bb5":                    {"C60", "Ván cờ Tây Ban Nha (Ruy Lopez)"},
	"e4 e5 Nf3 Nc6 Bc4":                    {"C50", "Ván cờ Ý (Italian Game)"},
	"e4 e5 Nf3 Nc6 Bc4 Bc5":                {"C50", "Ván cờ Ý — Giuoco Piano"},
	"e4 e5 Nf3 Nc6 d4":                     {"C44", "Ván cờ Scotch"},
	"e4 e5 Nf3 Nf6":                        {"C42", "Phòng thủ Petrov (Petroff)"},
	"e4 e5 f4":                             {"C30", "Gambit Vua (King's Gambit)"},
	"e4 e5 Nc3":                            {"C25", "Ván cờ Vienna"},
	"e4 c5":                                {"B20", "Phòng thủ Sicilian"},
	"e4 c5 Nf3":                            {"B27", "Phòng thủ Sicilian"},
	"e4 c5 Nf3 d6":                         {"B50", "Sicilian — biến d6"},
	"e4 c5 Nf3 Nc6":                        {"B30", "Sicilian — biến Nc6"},
	"e4 c5 Nf3 e6":                         {"B40", "Sicilian — biến e6"},
	"e4 c5 Nf3 d6 d4 cxd4 Nxd4 Nf6 Nc3":    {"B90", "Sicilian Najdorf (chuẩn bị)"},
	"e4 c5 Nf3 d6 d4 cxd4 Nxd4 Nf6 Nc3 a6": {"B90", "Sicilian Najdorf"},
	"e4 c6":                                {"B10", "Phòng thủ Caro-Kann"},
	"e4 c6 d4 d5":                          {"B12", "Caro-Kann — biến chính"},
	"e4 e6":                                {"C00", "Phòng thủ Pháp (French Defense)"},
	"e4 e6 d4 d5":                          {"C01", "Phòng thủ Pháp — biến chính"},
	"e4 d6":                                {"B07", "Phòng thủ Pirc"},
	"e4 d5":                                {"B01", "Phòng thủ Scandinavian"},
	"e4 g6":                                {"B06", "Phòng thủ Modern"},
	"e4 Nf6":                               {"B02", "Phòng thủ Alekhine"},
	"d4":                                   {"A40", "Khai cuộc Hậu tốt (1.d4)"},
	"d4 d5":                                {"D00", "Ván cờ Hậu tốt đóng"},
	"d4 d5 c4":                             {"D06", "Gambit Hậu (Queen's Gambit)"},
	"d4 d5 c4 e6":                          {"D30", "Gambit Hậu từ chối (QGD)"},
	"d4 d5 c4 dxc4":                        {"D20", "Gambit Hậu chấp nhận (QGA)"},
	"d4 d5 c4 c6":                          {"D10", "Phòng thủ Slav"},
	"d4 Nf6":                               {"A45", "Khai cuộc Indian"},
	"d4 Nf6 c4":                            {"A50", "Khai cuộc Indian — c4"},
	"d4 Nf6 c4 e6":                         {"E00", "Hệ Indian — e6"},
	"d4 Nf6 c4 e6 Nc3 Bb4":                 {"E20", "Phòng thủ Nimzo-Indian"},
	"d4 Nf6 c4 e6 Nf3 b6":                  {"E12", "Phòng thủ Queen's Indian"},
	"d4 Nf6 c4 g6":                         {"E60", "Phòng thủ King's Indian (chuẩn bị)"},
	"d4 Nf6 c4 g6 Nc3 Bg7":                 {"E70", "Phòng thủ King's Indian"},
	"d4 Nf6 c4 g6 Nc3 d5":                  {"D70", "Phòng thủ Grünfeld"},
	"d4 f5":                                {"A80", "Phòng thủ Hà Lan (Dutch)"},
	"c4":                                   {"A10", "Khai cuộc Anh (English)"},
	"Nf3":                                  {"A04", "Khai cuộc Réti"},
	"g3":                                   {"A00", "Khai cuộc Benko/Hungary"},
}

func normalizeSAN(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "+#!?")
	return s
}

// Execute tra cứu khai cuộc.
func (t *ChessLookupOpeningTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessLookupOpeningInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}

	// Lấy chuỗi nước SAN từ moves hoặc pgn.
	var sans []string
	if len(input.Moves) > 0 {
		for _, m := range input.Moves {
			if n := normalizeSAN(m); n != "" {
				sans = append(sans, n)
			}
		}
	} else if strings.TrimSpace(input.PGN) != "" {
		game, err := chess.ParsePGN(input.PGN)
		if err != nil {
			return &types.ToolResult{Success: false, Error: fmt.Sprintf("PGN không hợp lệ: %v", err)}, nil
		}
		for _, p := range game.Plies {
			sans = append(sans, normalizeSAN(p.SAN))
		}
	} else {
		return &types.ToolResult{Success: false, Error: "Cần cung cấp pgn hoặc moves"}, nil
	}

	if len(sans) == 0 {
		return &types.ToolResult{Success: false, Error: "Không có nước đi để tra cứu"}, nil
	}

	// Tìm tiền tố dài nhất khớp index khai cuộc (gộp dataset + overlay Việt hoá).
	// Giới hạn ~40 nước đầu để bắt được cả các biến lý thuyết sâu.
	limit := len(sans)
	if limit > 40 {
		limit = 40
	}
	var matched ecoEntry
	matchedLen := 0
	for i := 1; i <= limit; i++ {
		key := strings.Join(sans[:i], " ")
		if e, ok := openingIndex[key]; ok {
			matched = e
			matchedLen = i
		}
	}

	if matchedLen == 0 {
		return &types.ToolResult{
			Success: true,
			Output:  "Không nhận diện được khai cuộc trong cơ sở dữ liệu nhúng. Có thể là một biến ít phổ biến.",
		}, nil
	}

	// Tính FEN sau các nước khai cuộc đã khớp để hiển thị bàn cờ.
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	for i := 0; i < matchedLen; i++ {
		next, err := chess.FENAfterMove(fen, sans[i])
		if err != nil {
			break
		}
		fen = next
	}

	output := fmt.Sprintf("Khai cuộc: %s\nMã ECO: %s\nKhớp %d nước đầu: %s",
		matched.name, matched.eco, matchedLen, strings.Join(sans[:matchedLen], " "))

	stm := "w"
	if f := strings.Fields(fen); len(f) >= 2 {
		stm = f[1]
	}
	data := map[string]interface{}{
		"display_type": "chess_board",
		"fen":          fen,
		"side_to_move": stm,
		"caption":      fmt.Sprintf("%s (%s)", matched.name, matched.eco),
	}

	return &types.ToolResult{Success: true, Output: output, Data: data}, nil
}
