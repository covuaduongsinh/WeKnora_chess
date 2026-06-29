package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Tencent/WeKnora/internal/chess"
)

// emptyMoveEngine trả về phân tích hợp lệ nhưng KHÔNG có nước đi (thế cờ kết thúc).
// (errEngine dùng chung được khai báo ở chess_engine_error_test.go.)
type emptyMoveEngine struct{}

func (emptyMoveEngine) Analyze(_ context.Context, fen string, depth int) (*chess.Analysis, error) {
	return &chess.Analysis{FEN: fen, Depth: depth, SideToMove: fenSide(fen)}, nil
}
func (emptyMoveEngine) Close() error { return nil }

func TestChessAnalyzePosition(t *testing.T) {
	ctx := context.Background()

	t.Run("FEN hợp lệ → bàn cờ + thành công", func(t *testing.T) {
		tool := NewChessAnalyzePositionTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: demoStartFEN})
		res, err := tool.Execute(ctx, args)
		if err != nil || !res.Success {
			t.Fatalf("muốn thành công, nhận err=%v error=%s", err, res.Error)
		}
		if res.Data["display_type"] != "chess_board" {
			t.Fatalf("muốn display_type chess_board, nhận %v", res.Data["display_type"])
		}
	})

	t.Run("FEN rỗng → báo thiếu tham số", func(t *testing.T) {
		tool := NewChessAnalyzePositionTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: "   "})
		res, _ := tool.Execute(ctx, args)
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại khi thiếu fen, nhận %+v", res)
		}
	})

	t.Run("FEN sai → báo không hợp lệ", func(t *testing.T) {
		tool := NewChessAnalyzePositionTool(stubEngine{}, 12)
		args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: "khong-phai-fen"})
		res, _ := tool.Execute(ctx, args)
		if res.Success {
			t.Fatalf("muốn thất bại với FEN sai, nhận thành công")
		}
	})

	t.Run("engine lỗi → thông báo thân thiện, không panic", func(t *testing.T) {
		tool := NewChessAnalyzePositionTool(errEngine{chess.ErrEngineUnavailable}, 12)
		args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: demoStartFEN})
		res, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("không nên trả error Go, engine lỗi phải gói vào ToolResult: %v", err)
		}
		if res.Success || res.Error == "" {
			t.Fatalf("muốn thất bại có thông báo khi engine lỗi, nhận %+v", res)
		}
	})
}
