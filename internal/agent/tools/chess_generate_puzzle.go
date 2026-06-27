package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/utils"
)

var chessGeneratePuzzleTool = BaseTool{
	name: ToolChessGeneratePuzzle,
	description: `Ra một bài tập cờ vua (thế cờ để học viên tìm nước tốt nhất).

## Khi nào dùng
- Người dùng muốn luyện tập, xin một bài tập/thế cờ để giải.
- Có thể lọc theo chủ đề (theme) và độ khó (difficulty).

## Đầu vào
- theme: chủ đề (tùy chọn), ví dụ "chiếu hết", "chiến thuật", "khai cuộc", "tàn cuộc".
- difficulty: độ khó (tùy chọn): "dễ" | "trung bình" | "khó".

Trả về một thế cờ (bàn cờ tương tác) kèm gợi ý. KHÔNG kèm lời giải — học viên tự
tìm; có thể dùng chess_analyze_position/chess_best_move để kiểm tra đáp án.`,
	schema: utils.GenerateSchema[ChessGeneratePuzzleInput](),
}

// ChessGeneratePuzzleInput là tham số cho tool chess_generate_puzzle.
type ChessGeneratePuzzleInput struct {
	Theme      string `json:"theme,omitempty" jsonschema:"Chủ đề bài tập (tùy chọn)"`
	Difficulty string `json:"difficulty,omitempty" jsonschema:"Độ khó: dễ | trung bình | khó (tùy chọn)"`
}

// ChessGeneratePuzzleTool ra bài tập từ bộ thế cờ nhúng sẵn.
type ChessGeneratePuzzleTool struct {
	BaseTool
}

// NewChessGeneratePuzzleTool tạo tool chess_generate_puzzle.
func NewChessGeneratePuzzleTool() *ChessGeneratePuzzleTool {
	return &ChessGeneratePuzzleTool{BaseTool: chessGeneratePuzzleTool}
}

// chessPuzzle là một bài tập nhúng sẵn (thế cờ + chủ đề + gợi ý).
type chessPuzzle struct {
	fen        string
	theme      string
	difficulty string
	hint       string
}

// embeddedPuzzles — bộ bài tập tối giản; FEN được kiểm tra hợp lệ bằng test.
// Không lưu lời giải để học viên tự tìm; engine có thể kiểm tra đáp án.
var embeddedPuzzles = []chessPuzzle{
	{
		fen:        "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
		theme:      "chiếu hết",
		difficulty: "dễ",
		hint:       "Trắng đi. Tốt f7 của Đen chỉ có Vua bảo vệ — tìm đòn chiếu hết ngay.",
	},
	{
		fen:        "rnbqkbnr/ppp2ppp/8/3pp3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 3",
		theme:      "khai cuộc",
		difficulty: "dễ",
		hint:       "Trắng đi. Đen vừa đẩy d5 mở trung tâm — tìm nước phát triển/ăn trung tâm tốt nhất.",
	},
	{
		fen:        "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 6 5",
		theme:      "chiến thuật",
		difficulty: "trung bình",
		hint:       "Trắng đi (Ván Ý). Tìm kế hoạch trung tâm/đòn đánh điểm yếu f7.",
	},
	{
		fen:        "8/8/8/4k3/8/4K3/4P3/8 w - - 0 1",
		theme:      "tàn cuộc",
		difficulty: "trung bình",
		hint:       "Trắng đi. Tàn cuộc Vua-Tốt: tìm cách giành đối diện (đối ngẫu) để phong cấp.",
	},
	{
		fen:        "6k1/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1",
		theme:      "chiến thuật",
		difficulty: "dễ",
		hint:       "Trắng đi. Xe trên cột mở — tìm nước chiếm hàng ngang/cột tốt nhất.",
	},
	{
		fen:        "r2qkbnr/ppp2ppp/2np4/4p3/2B1P1b1/3P1N2/PPP2PPP/RNBQK2R w KQkq - 2 5",
		theme:      "chiến thuật",
		difficulty: "khó",
		hint:       "Trắng đi. Tượng g4 ghim Mã f3 — tìm cách hóa giải hoặc khai thác.",
	},
}

// Execute ra một bài tập phù hợp bộ lọc.
func (t *ChessGeneratePuzzleTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessGeneratePuzzleInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}

	theme := strings.ToLower(strings.TrimSpace(input.Theme))
	diff := strings.ToLower(strings.TrimSpace(input.Difficulty))

	// Lọc theo theme/difficulty (khớp lỏng); rỗng = lấy tất cả.
	var candidates []chessPuzzle
	for _, p := range embeddedPuzzles {
		if theme != "" && !strings.Contains(p.theme, theme) {
			continue
		}
		if diff != "" && p.difficulty != diff {
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		candidates = embeddedPuzzles // không khớp → trả bất kỳ
	}

	p := candidates[rand.Intn(len(candidates))]

	side := "Trắng"
	if fenSide(p.fen) == "b" {
		side = "Đen"
	}

	output := fmt.Sprintf("Bài tập (%s · %s) — %s đi:\n%s\n\nHãy tìm nước tốt nhất. "+
		"Bạn có thể nhờ kiểm tra đáp án bằng cách yêu cầu phân tích thế cờ này.",
		p.theme, p.difficulty, side, p.hint)

	data := map[string]interface{}{
		"display_type": "chess_board",
		"fen":          p.fen,
		"side_to_move": fenSide(p.fen),
		"caption":      fmt.Sprintf("Bài tập: %s (%s)", p.theme, p.difficulty),
	}

	return &types.ToolResult{Success: true, Output: output, Data: data}, nil
}
