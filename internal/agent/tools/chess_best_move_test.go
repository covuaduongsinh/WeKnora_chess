package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/chess"
)

func TestChessBestMove(t *testing.T) {
	ctx := context.Background()

	t.Run("FEN hợp lệ → trả nước tốt nhất + bàn cờ", func(t *testing.T) {
		tool := NewChessBestMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessBestMoveInput{FEN: demoStartFEN})
		res, err := tool.Execute(ctx, args)
		if err != nil || !res.Success {
			t.Fatalf("muốn thành công, nhận err=%v error=%s", err, res.Error)
		}
		if res.Data["display_type"] != "chess_board" {
			t.Fatalf("muốn display_type chess_board, nhận %v", res.Data["display_type"])
		}
		if res.Data["best_move_san"] == nil {
			t.Fatalf("muốn có best_move_san trong data")
		}
	})

	t.Run("thế cờ kết thúc (không có nước) → vẫn thành công, nhắn rõ", func(t *testing.T) {
		tool := NewChessBestMoveTool(emptyMoveEngine{}, 12)
		args, _ := json.Marshal(ChessBestMoveInput{FEN: demoStartFEN})
		res, _ := tool.Execute(ctx, args)
		if !res.Success {
			t.Fatalf("muốn thành công khi không có nước, nhận %+v", res)
		}
		if !strings.Contains(res.Output, "Không có nước") {
			t.Fatalf("muốn nhắn không có nước đi, nhận %q", res.Output)
		}
	})

	t.Run("FEN sai → thất bại", func(t *testing.T) {
		tool := NewChessBestMoveTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessBestMoveInput{FEN: "xxx"})
		res, _ := tool.Execute(ctx, args)
		if res.Success {
			t.Fatalf("muốn thất bại với FEN sai")
		}
	})

	t.Run("engine lỗi → thông báo thân thiện", func(t *testing.T) {
		tool := NewChessBestMoveTool(errEngine{chess.ErrAnalysisTimeout}, 12)
		args, _ := json.Marshal(ChessBestMoveInput{FEN: demoStartFEN})
		res, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("không nên trả error Go: %v", err)
		}
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại có thông báo khi engine lỗi, nhận %+v", res)
		}
	})
}
