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

const (
	// Giới hạn mặc định để bó chi phí gọi engine trong vòng lặp agent.
	defaultEvaluateMaxPlies = 60
	defaultEvaluateDepth    = 12
	// Quy đổi điểm chiếu hết thành centipawn lớn để so sánh độ chênh.
	mateScoreCp = 100000
)

var chessEvaluateGameTool = BaseTool{
	name: ToolChessEvaluateGame,
	description: `Chấm điểm cả một ván cờ (PGN): đánh giá từng nước và chỉ ra sai lầm.

## Khi nào dùng
- Người dùng dán một ván cờ (PGN) và muốn biết đã đi sai ở đâu, nước nào là sai lầm nghiêm trọng.
- Phân tích ván đấu của học viên để dạy.

## Đầu vào
- pgn: nội dung ván cờ dạng PGN (bắt buộc).
- depth: độ sâu engine cho mỗi nước (tùy chọn, mặc định nhỏ để nhanh).
- max_plies: số nửa-nước tối đa phân tích (tùy chọn).

Trả về danh sách sai lầm (sai lầm nghiêm trọng/sai lầm/thiếu chính xác) kèm bàn cờ
tương tác để xem lại toàn bộ ván.`,
	schema: utils.GenerateSchema[ChessEvaluateGameInput](),
}

// ChessEvaluateGameInput là tham số cho tool chess_evaluate_game.
type ChessEvaluateGameInput struct {
	PGN      string `json:"pgn" jsonschema:"Nội dung ván cờ dạng PGN"`
	Depth    int    `json:"depth,omitempty" jsonschema:"Độ sâu engine cho mỗi nước (tùy chọn)"`
	MaxPlies int    `json:"max_plies,omitempty" jsonschema:"Số nửa-nước tối đa phân tích (tùy chọn)"`
}

// ChessEvaluateGameTool chấm điểm cả ván.
type ChessEvaluateGameTool struct {
	BaseTool
	engine chess.EngineClient
}

// NewChessEvaluateGameTool tạo tool chess_evaluate_game.
func NewChessEvaluateGameTool(engine chess.EngineClient) *ChessEvaluateGameTool {
	return &ChessEvaluateGameTool{BaseTool: chessEvaluateGameTool, engine: engine}
}

// whiteScore quy một Analysis về điểm centipawn theo góc nhìn Trắng để so sánh.
func whiteScore(a *chess.Analysis) int {
	if a.IsMate {
		whiteMating := (a.MateIn >= 0) == (a.SideToMove != "b")
		if whiteMating {
			return mateScoreCp
		}
		return -mateScoreCp
	}
	return a.WhiteCentipawns()
}

// classifyLoss phân loại mức độ sai lầm theo centipawn mất đi (góc nhìn người đi).
func classifyLoss(lossCp int) string {
	switch {
	case lossCp >= 300:
		return "?? sai lầm nghiêm trọng"
	case lossCp >= 150:
		return "? sai lầm"
	case lossCp >= 80:
		return "?! thiếu chính xác"
	default:
		return ""
	}
}

// Execute chấm điểm cả ván.
func (t *ChessEvaluateGameTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	var input ChessEvaluateGameInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("Không đọc được tham số: %v", err)}, err
	}
	if strings.TrimSpace(input.PGN) == "" {
		return &types.ToolResult{Success: false, Error: "Thiếu tham số pgn"}, nil
	}

	game, err := chess.ParsePGN(input.PGN)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("PGN không hợp lệ: %v", err)}, nil
	}
	if len(game.Plies) == 0 {
		return &types.ToolResult{Success: false, Error: "Ván cờ không có nước đi nào"}, nil
	}

	depth := input.Depth
	if depth <= 0 {
		depth = defaultEvaluateDepth
	}
	maxPlies := input.MaxPlies
	if maxPlies <= 0 || maxPlies > defaultEvaluateMaxPlies {
		maxPlies = defaultEvaluateMaxPlies
	}

	// Điểm Trắng trước nước đầu tiên (thế cờ ban đầu của ván).
	startFEN := game.Plies[0].FENBefore
	prevWhite := 0
	if a0, err := t.engine.Analyze(ctx, startFEN, depth); err == nil {
		prevWhite = whiteScore(a0)
	}

	type blunder struct {
		moveLabel string
		tag       string
		lossCp    int
		fen       string
	}
	var blunders []blunder
	analyzed := 0

	for i, ply := range game.Plies {
		if i >= maxPlies {
			break
		}
		if ctx.Err() != nil {
			break // hết thời gian/bị hủy → dừng, dùng kết quả đã có
		}

		a, err := t.engine.Analyze(ctx, ply.FENAfter, depth)
		if err != nil {
			break // engine lỗi/timeout → dừng, dùng kết quả đã có
		}
		analyzed++
		curWhite := whiteScore(a)

		// Mất điểm của bên vừa đi (góc nhìn của họ).
		var loss int
		if ply.Side == "w" {
			loss = prevWhite - curWhite
		} else {
			loss = curWhite - prevWhite
		}
		if tag := classifyLoss(loss); tag != "" {
			label := fmt.Sprintf("%d.%s %s", ply.MoveNumber, sidePrefix(ply.Side), ply.SAN)
			blunders = append(blunders, blunder{moveLabel: label, tag: tag, lossCp: loss, fen: ply.FENAfter})
		}
		prevWhite = curWhite
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Đã phân tích %d/%d nước (độ sâu %d).\n", analyzed, len(game.Plies), depth)
	if white, ok := game.Tags["White"]; ok {
		fmt.Fprintf(&b, "Trắng: %s — Đen: %s — Kết quả: %s\n", white, game.Tags["Black"], game.Outcome)
	}
	if len(blunders) == 0 {
		b.WriteString("Không phát hiện sai lầm đáng kể trong phạm vi đã phân tích.\n")
	} else {
		fmt.Fprintf(&b, "\nCác nước có vấn đề (%d):\n", len(blunders))
		for _, bl := range blunders {
			fmt.Fprintf(&b, "- %s %s (mất ~%.2f tốt)\n", bl.moveLabel, bl.tag, float64(bl.lossCp)/100)
		}
	}

	// Dữ liệu bàn cờ: gửi toàn bộ plies để người dùng xem lại ván.
	plies := make([]map[string]interface{}, 0, len(game.Plies))
	for _, p := range game.Plies {
		plies = append(plies, map[string]interface{}{
			"move_number": p.MoveNumber,
			"side":        p.Side,
			"san":         p.SAN,
			"uci":         p.UCI,
			"fen_before":  p.FENBefore,
			"fen_after":   p.FENAfter,
		})
	}
	data := map[string]interface{}{
		"display_type": "chess_board",
		"fen":          startFEN,
		"plies":        plies,
		"caption":      "Xem lại ván cờ",
	}

	return &types.ToolResult{Success: true, Output: b.String(), Data: data}, nil
}

// sidePrefix trả về phần hậu tố số nước cho ký hiệu (Trắng: "", Đen: "..").
func sidePrefix(side string) string {
	if side == "b" {
		return ".."
	}
	return ""
}
