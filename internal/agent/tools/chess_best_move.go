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

var chessBestMoveTool = BaseTool{
	name: ToolChessBestMove,
	description: `Tìm nước đi tốt nhất cho một thế cờ vua.

## Khi nào dùng
- Người dùng hỏi "nên đi nước nào?" ở một thế cờ (FEN).
- Cần gợi ý nước đi cho học viên.

## Đầu vào
- fen: chuỗi FEN của thế cờ (bắt buộc).
- depth: độ sâu tìm kiếm (tùy chọn).

Trả về nước tốt nhất (ký hiệu SAN và UCI) kèm bàn cờ tương tác.`,
	schema: utils.GenerateSchema[ChessBestMoveInput](),
}

// ChessBestMoveInput là tham số cho tool chess_best_move.
type ChessBestMoveInput struct {
	FEN   string `json:"fen" jsonschema:"Chuỗi FEN của thế cờ"`
	Depth int    `json:"depth,omitempty" jsonschema:"Độ sâu tìm kiếm (tùy chọn)"`
}

// ChessBestMoveTool tìm nước đi tốt nhất.
type ChessBestMoveTool struct {
	BaseTool
	engine       chess.EngineClient
	defaultDepth int
}

// NewChessBestMoveTool tạo tool chess_best_move.
func NewChessBestMoveTool(engine chess.EngineClient, defaultDepth int) *ChessBestMoveTool {
	return &ChessBestMoveTool{
		BaseTool:     chessBestMoveTool,
		engine:       engine,
		defaultDepth: defaultDepth,
	}
}

// Execute tìm nước tốt nhất.
func (t *ChessBestMoveTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessBestMoveInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}
	input.FEN = strings.TrimSpace(input.FEN)
	if input.FEN == "" {
		return &types.ToolResult{Success: false, Error: "Thiếu tham số fen"}, nil
	}
	if err := chess.ValidateFEN(input.FEN); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("FEN không hợp lệ: %v", err)}, nil
	}

	depth := input.Depth
	if depth <= 0 {
		depth = t.defaultDepth
	}

	analysis, err := t.engine.Analyze(ctx, input.FEN, depth)
	if err != nil {
		return &types.ToolResult{Success: false, Error: friendlyEngineError(err)}, nil
	}
	if analysis.BestMove == "" {
		return &types.ToolResult{
			Success: true,
			Output:  "Không có nước đi (thế cờ đã kết thúc: chiếu hết hoặc hòa).",
			Data:    chessBoardData(analysis, "Thế cờ"),
		}, nil
	}

	mv := analysis.BestMoveSAN
	if mv == "" {
		mv = analysis.BestMove
	}
	output := fmt.Sprintf("Nước tốt nhất: %s (UCI: %s)\nĐánh giá sau nước này: %s\nĐộ sâu: %d",
		mv, analysis.BestMove, describeEval(analysis), analysis.Depth)

	return &types.ToolResult{
		Success: true,
		Output:  output,
		Data:    chessBoardData(analysis, fmt.Sprintf("Nước tốt nhất: %s", mv)),
	}, nil
}
