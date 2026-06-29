package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestChessEvaluateGame(t *testing.T) {
	ctx := context.Background()

	t.Run("PGN hợp lệ → bàn cờ + danh sách nước", func(t *testing.T) {
		tool := NewChessEvaluateGameTool(stubEngine{}, 0, 0)
		args, _ := json.Marshal(ChessEvaluateGameInput{
			PGN: "1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 1-0", Depth: 8,
		})
		res, err := tool.Execute(ctx, args)
		if err != nil || !res.Success {
			t.Fatalf("muốn thành công, nhận err=%v error=%s", err, res.Error)
		}
		if res.Data["display_type"] != "chess_board" {
			t.Fatalf("muốn display_type chess_board, nhận %v", res.Data["display_type"])
		}
		plies, _ := res.Data["plies"].([]map[string]interface{})
		if len(plies) == 0 {
			t.Fatalf("muốn có danh sách plies")
		}
	})

	t.Run("PGN rỗng → báo thiếu tham số", func(t *testing.T) {
		tool := NewChessEvaluateGameTool(stubEngine{}, 0, 0)
		args, _ := json.Marshal(ChessEvaluateGameInput{PGN: "  "})
		res, _ := tool.Execute(ctx, args)
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại khi thiếu pgn, nhận %+v", res)
		}
	})

	t.Run("PGN sai cú pháp → báo không hợp lệ", func(t *testing.T) {
		tool := NewChessEvaluateGameTool(stubEngine{}, 0, 0)
		args, _ := json.Marshal(ChessEvaluateGameInput{PGN: "1. zz9 qq8 đây không phải PGN"})
		res, _ := tool.Execute(ctx, args)
		if res.Success {
			t.Fatalf("muốn thất bại với PGN sai")
		}
	})
}
