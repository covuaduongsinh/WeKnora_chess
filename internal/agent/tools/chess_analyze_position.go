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

var chessAnalyzePositionTool = BaseTool{
	name: ToolChessAnalyzePosition,
	description: `Phân tích một thế cờ vua bằng engine.

## Khi nào dùng
- Người dùng đưa một thế cờ (chuỗi FEN) và muốn biết: ai đang ưu thế, nước đi tốt nhất, hay đánh giá thế cờ.
- Cần con số đánh giá khách quan (centipawn) thay vì cảm tính.

## Đầu vào
- fen: chuỗi FEN của thế cờ cần phân tích (bắt buộc).
- depth: độ sâu tìm kiếm (tùy chọn; để trống dùng mặc định).

Kết quả trả về kèm bàn cờ tương tác (hiển thị thế cờ và nước tốt nhất).`,
	schema: utils.GenerateSchema[ChessAnalyzePositionInput](),
}

// ChessAnalyzePositionInput là tham số cho tool chess_analyze_position.
type ChessAnalyzePositionInput struct {
	FEN   string `json:"fen" jsonschema:"Chuỗi FEN của thế cờ cần phân tích"`
	Depth int    `json:"depth,omitempty" jsonschema:"Độ sâu tìm kiếm của engine (tùy chọn)"`
}

// ChessAnalyzePositionTool phân tích thế cờ bằng engine.
type ChessAnalyzePositionTool struct {
	BaseTool
	engine       chess.EngineClient
	defaultDepth int
}

// NewChessAnalyzePositionTool tạo tool chess_analyze_position.
func NewChessAnalyzePositionTool(engine chess.EngineClient, defaultDepth int) *ChessAnalyzePositionTool {
	return &ChessAnalyzePositionTool{
		BaseTool:     chessAnalyzePositionTool,
		engine:       engine,
		defaultDepth: defaultDepth,
	}
}

// Execute thực thi phân tích thế cờ.
func (t *ChessAnalyzePositionTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessAnalyzePositionInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}
	fen, bad := validateFENArg(input.FEN)
	if bad != nil {
		return bad, nil
	}
	input.FEN = fen

	depth := input.Depth
	if depth <= 0 {
		depth = t.defaultDepth
	}

	analysis, err := t.engine.Analyze(ctx, input.FEN, depth)
	if err != nil {
		return &types.ToolResult{Success: false, Error: friendlyEngineError(err)}, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Đánh giá: %s\n", describeEval(analysis))
	if analysis.BestMoveSAN != "" || analysis.BestMove != "" {
		mv := analysis.BestMoveSAN
		if mv == "" {
			mv = analysis.BestMove
		}
		fmt.Fprintf(&b, "Nước tốt nhất: %s\n", mv)
	}
	fmt.Fprintf(&b, "Độ sâu: %d\n", analysis.Depth)
	if len(analysis.PV) > 0 {
		fmt.Fprintf(&b, "Biến chính (PV): %s\n", strings.Join(analysis.PV, " "))
	}

	return &types.ToolResult{
		Success: true,
		Output:  b.String(),
		Data:    chessBoardData(analysis, "Phân tích thế cờ"),
	}, nil
}
