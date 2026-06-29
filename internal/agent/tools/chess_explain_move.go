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

var chessExplainMoveTool = BaseTool{
	name: ToolChessExplainMove,
	description: `Giải thích một nước đi cờ vua dựa trên đánh giá engine trước và sau nước đi.

## Khi nào dùng
- Người dùng hỏi "nước đi X có tốt không / vì sao?" ở một thế cờ (FEN).
- Cần so sánh nước đã đi với nước tốt nhất của engine để dạy học.

## Đầu vào
- fen: thế cờ trước khi đi (bắt buộc).
- move: nước cần giải thích, dạng SAN ("Nf3", "e4") hoặc UCI ("g1f3").
- depth: độ sâu engine (tùy chọn).

Trả về so sánh đánh giá, nước tốt nhất của engine, và mức độ tốt/xấu của nước đã đi,
kèm bàn cờ tương tác.`,
	schema: utils.GenerateSchema[ChessExplainMoveInput](),
}

// ChessExplainMoveInput là tham số cho tool chess_explain_move.
type ChessExplainMoveInput struct {
	FEN   string `json:"fen" jsonschema:"Thế cờ trước khi đi (FEN)"`
	Move  string `json:"move" jsonschema:"Nước cần giải thích, dạng SAN hoặc UCI"`
	Depth int    `json:"depth,omitempty" jsonschema:"Độ sâu engine (tùy chọn)"`
}

// ChessExplainMoveTool giải thích một nước đi.
type ChessExplainMoveTool struct {
	BaseTool
	engine       chess.EngineClient
	defaultDepth int
}

// NewChessExplainMoveTool tạo tool chess_explain_move.
func NewChessExplainMoveTool(engine chess.EngineClient, defaultDepth int) *ChessExplainMoveTool {
	return &ChessExplainMoveTool{BaseTool: chessExplainMoveTool, engine: engine, defaultDepth: defaultDepth}
}

// Execute giải thích nước đi.
func (t *ChessExplainMoveTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessExplainMoveInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}
	input.FEN = strings.TrimSpace(input.FEN)
	input.Move = strings.TrimSpace(input.Move)
	if input.FEN == "" || input.Move == "" {
		return &types.ToolResult{Success: false, Error: "Cần cả fen và move"}, nil
	}
	if err := chess.ValidateFEN(input.FEN); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("FEN không hợp lệ: %v", err)}, nil
	}

	// Chuẩn hóa nước đi: chấp nhận cả SAN lẫn UCI.
	var san, uci string
	if u, err := chess.SANToUCI(input.FEN, input.Move); err == nil {
		san, uci = input.Move, u
	} else if s, err := chess.UCIToSAN(input.FEN, input.Move); err == nil {
		san, uci = s, input.Move
	} else {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Nước đi không hợp lệ ở thế cờ này: %q", input.Move)}, nil
	}

	fenAfter, err := chess.FENAfterMove(input.FEN, uci)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đi được nước: %v", err)}, nil
	}

	depth := input.Depth
	if depth <= 0 {
		depth = t.defaultDepth
	}

	before, err := t.engine.Analyze(ctx, input.FEN, depth)
	if err != nil {
		return &types.ToolResult{Success: false, Error: friendlyEngineError(err)}, nil
	}
	after, err := t.engine.Analyze(ctx, fenAfter, depth)
	if err != nil {
		return &types.ToolResult{Success: false, Error: friendlyEngineError(err)}, nil
	}

	mover := before.SideToMove
	whiteBefore := whiteScore(before)
	whiteAfter := whiteScore(after)
	var loss int
	if mover == "w" {
		loss = whiteBefore - whiteAfter
	} else {
		loss = whiteAfter - whiteBefore
	}

	bestSAN := before.BestMoveSAN
	if bestSAN == "" {
		bestSAN = before.BestMove
	}
	isBest := uci == before.BestMove

	var b strings.Builder
	fmt.Fprintf(&b, "Nước đã đi: %s\n", san)
	fmt.Fprintf(&b, "Đánh giá trước (nước tốt nhất %s): %s\n", bestSAN, describeEval(before))
	// Đánh giá sau nước đã đi, quy về góc nhìn người đi để dễ hiểu.
	fmt.Fprintf(&b, "Đánh giá sau nước đã đi: %s\n", formatEvalWhite(after))
	if isBest {
		b.WriteString("=> Đây chính là nước tốt nhất theo engine.\n")
	} else if tag := classifyLoss(loss); tag != "" {
		fmt.Fprintf(&b, "=> Nước này %s (mất ~%.2f tốt so với nước tốt nhất %s).\n",
			tag, float64(loss)/100, bestSAN)
	} else {
		fmt.Fprintf(&b, "=> Nước hợp lý (chênh không đáng kể so với %s).\n", bestSAN)
	}

	// Hiển thị thế cờ sau nước đã đi; gợi ý nước tốt nhất từ thế cờ ban đầu.
	data := chessBoardData(after, fmt.Sprintf("Sau nước %s", san))
	data["best_move"] = before.BestMove
	data["best_move_san"] = bestSAN

	return &types.ToolResult{Success: true, Output: b.String(), Data: data}, nil
}
