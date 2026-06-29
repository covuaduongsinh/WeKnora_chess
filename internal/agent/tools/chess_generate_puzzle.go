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

Ưu tiên lấy từ ngân hàng bài tập của tenant (DB); nếu kho trống/không khớp thì
dùng bộ bài tập nhúng sẵn. Trả về một thế cờ (bàn cờ tương tác) kèm gợi ý. KHÔNG
kèm lời giải — học viên tự tìm; có thể dùng chess_analyze_position/chess_best_move
để kiểm tra đáp án.`,
	schema: utils.GenerateSchema[ChessGeneratePuzzleInput](),
}

// ChessGeneratePuzzleInput là tham số cho tool chess_generate_puzzle.
type ChessGeneratePuzzleInput struct {
	Theme      string `json:"theme,omitempty" jsonschema:"Chủ đề bài tập (tùy chọn)"`
	Difficulty string `json:"difficulty,omitempty" jsonschema:"Độ khó: dễ | trung bình | khó (tùy chọn)"`
}

// PuzzleSource là nguồn bài tập từ ngân hàng bài tập (DB). Tách thành interface tối
// giản để tool không phụ thuộc trực tiếp tầng service và dễ test bằng fake.
type PuzzleSource interface {
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)
}

// ChessGeneratePuzzleTool ra bài tập từ ngân hàng bài tập (DB), fallback bộ nhúng.
type ChessGeneratePuzzleTool struct {
	BaseTool
	source PuzzleSource
}

// NewChessGeneratePuzzleTool tạo tool chess_generate_puzzle. source có thể nil
// (vd khi engine/library chưa cấu hình) → tool dùng bộ bài tập nhúng sẵn.
func NewChessGeneratePuzzleTool(source PuzzleSource) *ChessGeneratePuzzleTool {
	return &ChessGeneratePuzzleTool{BaseTool: chessGeneratePuzzleTool, source: source}
}

// chessPuzzle là một bài tập nhúng sẵn (thế cờ + chủ đề + gợi ý).
type chessPuzzle struct {
	fen        string
	theme      string
	difficulty string
	hint       string
}

// embeddedPuzzles — bộ bài tập tối giản dùng làm FALLBACK khi ngân hàng bài tập
// trống/không khớp. FEN được kiểm tra hợp lệ bằng test. Không lưu lời giải để học
// viên tự tìm; engine có thể kiểm tra đáp án.
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

// difficultyToDB ánh xạ độ khó người dùng nhập ("dễ"/"trung bình"/"khó") về dạng
// slug lưu trong DB ("de"/"trung-binh"/"kho"). Trả "" nếu không xác định → không
// lọc theo độ khó.
func difficultyToDB(diff string) string {
	switch strings.ToLower(strings.TrimSpace(diff)) {
	case "dễ", "de", "easy":
		return "de"
	case "trung bình", "trung-binh", "trungbinh", "medium", "tb":
		return "trung-binh"
	case "khó", "kho", "hard":
		return "kho"
	default:
		return ""
	}
}

// prettyDifficulty đổi slug độ khó về nhãn hiển thị tiếng Việt.
func prettyDifficulty(diff string) string {
	switch strings.ToLower(strings.TrimSpace(diff)) {
	case "de":
		return "dễ"
	case "trung-binh":
		return "trung bình"
	case "kho":
		return "khó"
	default:
		return diff
	}
}

// Execute ra một bài tập: ưu tiên ngân hàng bài tập thật, fallback bộ nhúng.
func (t *ChessGeneratePuzzleTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessGeneratePuzzleInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}

	// 1) Ưu tiên kho thật (DB) — lọc nới dần để tối đa cơ hội trúng.
	if p := t.pickFromBank(ctx, strings.TrimSpace(input.Theme), difficultyToDB(input.Difficulty)); p != nil {
		theme := p.Theme
		if theme == "" {
			theme = "tổng hợp"
		}
		return puzzleResult(p.FEN, theme, prettyDifficulty(p.Difficulty), bankHint(p)), nil
	}

	// 2) Fallback: bộ bài tập nhúng sẵn (giữ tool không bao giờ "câm").
	return t.pickEmbedded(input.Theme, input.Difficulty), nil
}

// pickFromBank thử lấy ngẫu nhiên một bài tập từ kho thật theo bộ lọc nới dần:
// (theme + độ khó) → (độ khó) → (bất kỳ). Trả nil nếu không có nguồn/không có
// tenant/kho trống.
func (t *ChessGeneratePuzzleTool) pickFromBank(ctx context.Context, theme, diffDB string) *types.ChessPuzzle {
	if t.source == nil {
		return nil
	}
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok || tenantID == 0 {
		return nil
	}
	filters := []types.ChessPuzzleFilter{
		{Theme: theme, Difficulty: diffDB},
		{Difficulty: diffDB},
		{},
	}
	tried := map[string]bool{}
	for _, f := range filters {
		key := f.Theme + "\x00" + f.Difficulty
		if tried[key] {
			continue
		}
		tried[key] = true
		if p, err := t.source.RandomPuzzle(ctx, tenantID, f); err == nil && p != nil && strings.TrimSpace(p.FEN) != "" {
			return p
		}
	}
	return nil
}

// bankHint dựng gợi ý ngắn cho bài tập lấy từ kho (KHÔNG lộ lời giải).
func bankHint(p *types.ChessPuzzle) string {
	side := "Trắng"
	if fenSide(p.FEN) == "b" {
		side = "Đen"
	}
	if strings.TrimSpace(p.Title) != "" {
		return fmt.Sprintf("%s đi. %s", side, strings.TrimSpace(p.Title))
	}
	return fmt.Sprintf("%s đi. Tìm nước tốt nhất.", side)
}

// pickEmbedded chọn một bài tập từ bộ nhúng sẵn (khớp lỏng theme/difficulty).
func (t *ChessGeneratePuzzleTool) pickEmbedded(theme, difficulty string) *types.ToolResult {
	th := strings.ToLower(strings.TrimSpace(theme))
	diff := strings.ToLower(strings.TrimSpace(difficulty))

	var candidates []chessPuzzle
	for _, p := range embeddedPuzzles {
		if th != "" && !strings.Contains(p.theme, th) {
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
	return puzzleResult(p.fen, p.theme, p.difficulty, p.hint)
}

// puzzleResult dựng ToolResult chuẩn (bàn cờ tương tác + lời nhắc) cho một bài tập.
func puzzleResult(fen, theme, difficulty, hint string) *types.ToolResult {
	side := "Trắng"
	if fenSide(fen) == "b" {
		side = "Đen"
	}
	output := fmt.Sprintf("Bài tập (%s · %s) — %s đi:\n%s\n\nHãy tìm nước tốt nhất. "+
		"Bạn có thể nhờ kiểm tra đáp án bằng cách yêu cầu phân tích thế cờ này.",
		theme, difficulty, side, hint)

	data := map[string]interface{}{
		"display_type": "chess_board",
		"fen":          fen,
		"side_to_move": fenSide(fen),
		"caption":      fmt.Sprintf("Bài tập: %s (%s)", theme, difficulty),
	}

	return &types.ToolResult{Success: true, Output: output, Data: data}
}
