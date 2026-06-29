package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Tencent/WeKnora/internal/chess"
)

func TestChessExplainMove(t *testing.T) {
	ctx := context.Background()

	t.Run("nước hợp lệ (SAN) → bàn cờ + thành công", func(t *testing.T) {
		tool := NewChessExplainMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessExplainMoveInput{FEN: demoStartFEN, Move: "e4"})
		res, err := tool.Execute(ctx, args)
		if err != nil || !res.Success {
			t.Fatalf("muốn thành công, nhận err=%v error=%s", err, res.Error)
		}
		if res.Data["display_type"] != "chess_board" {
			t.Fatalf("muốn display_type chess_board, nhận %v", res.Data["display_type"])
		}
	})

	t.Run("thiếu move → báo cần cả fen và move", func(t *testing.T) {
		tool := NewChessExplainMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessExplainMoveInput{FEN: demoStartFEN, Move: ""})
		res, _ := tool.Execute(ctx, args)
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại khi thiếu move, nhận %+v", res)
		}
	})

	t.Run("nước không hợp lệ ở thế cờ → thất bại", func(t *testing.T) {
		tool := NewChessExplainMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessExplainMoveInput{FEN: demoStartFEN, Move: "Qh5xh7"})
		res, _ := tool.Execute(ctx, args)
		if res.Success {
			t.Fatalf("muốn thất bại với nước không hợp lệ")
		}
	})

	t.Run("FEN sai → thất bại", func(t *testing.T) {
		tool := NewChessExplainMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessExplainMoveInput{FEN: "khong-hop-le", Move: "e4"})
		res, _ := tool.Execute(ctx, args)
		if res.Success {
			t.Fatalf("muốn thất bại với FEN sai")
		}
	})

	t.Run("engine lỗi → thông báo thân thiện", func(t *testing.T) {
		tool := NewChessExplainMoveTool(errEngine{chess.ErrEngineUnavailable}, 12)
		args, _ := json.Marshal(ChessExplainMoveInput{FEN: demoStartFEN, Move: "e4"})
		res, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("không nên trả error Go: %v", err)
		}
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại có thông báo khi engine lỗi, nhận %+v", res)
		}
	})
}
