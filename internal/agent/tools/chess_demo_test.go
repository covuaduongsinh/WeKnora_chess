package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Tencent/WeKnora/internal/chess"
)

// stubEngine là EngineClient giả cho test tầng tool: trả đánh giá cố định nhưng
// vẫn đi qua đúng code path của tool (parse tham số → gọi engine → dựng
// ToolResult với display_type "chess_board").
type stubEngine struct{}

func (stubEngine) Analyze(_ context.Context, fen string, depth int) (*chess.Analysis, error) {
	// e2e4 hợp lệ ở thế cờ đầu → cho phép đổi SAN; nước nào không hợp lệ thì bỏ.
	best := "e2e4"
	san, _ := chess.UCIToSAN(fen, best)
	if san == "" {
		best = ""
	}
	return &chess.Analysis{
		FEN: fen, BestMove: best, BestMoveSAN: san,
		EvalCP: 35, Depth: depth, SideToMove: chessSideToMove(fen),
		PV: []string{"e2e4", "e7e5"},
	}, nil
}
func (stubEngine) Close() error { return nil }

// chessSideToMove đọc bên đi từ FEN (trường thứ 2).
func chessSideToMove(fen string) string {
	for i := 0; i < len(fen); i++ {
		if fen[i] == ' ' && i+1 < len(fen) {
			if fen[i+1] == 'b' {
				return "b"
			}
			return "w"
		}
	}
	return "w"
}

const demoStartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func TestDemoChessTools(t *testing.T) {
	ctx := context.Background()

	// 1) chess_analyze_position (qua stub engine) → kiểm tra display_type bàn cờ.
	t.Log("───── chess_analyze_position ─────")
	atool := NewChessAnalyzePositionTool(stubEngine{}, 14)
	args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: demoStartFEN})
	res, err := atool.Execute(ctx, args)
	if err != nil || !res.Success {
		t.Fatalf("analyze lỗi: %v / %s", err, res.Error)
	}
	if res.Data["display_type"] != "chess_board" {
		t.Fatalf("muốn display_type chess_board, nhận %v", res.Data["display_type"])
	}
	t.Logf("Output:\n%s", res.Output)
	t.Logf("Data (gửi frontend): display_type=%v best_move_san=%v eval_cp=%v",
		res.Data["display_type"], res.Data["best_move_san"], res.Data["eval_cp"])

	// 2) chess_lookup_opening — CHẠY THẬT, không cần engine.
	t.Log("───── chess_lookup_opening (Sicilian Najdorf) ─────")
	ltool := NewChessLookupOpeningTool()
	largs, _ := json.Marshal(ChessLookupOpeningInput{
		Moves: []string{"e4", "c5", "Nf3", "d6", "d4", "cxd4", "Nxd4", "Nf6", "Nc3", "a6"},
	})
	lres, err := ltool.Execute(ctx, largs)
	if err != nil || !lres.Success {
		t.Fatalf("lookup lỗi: %v / %s", err, lres.Error)
	}
	t.Logf("Output:\n%s", lres.Output)
	t.Logf("Data: caption=%v fen=%v", lres.Data["caption"], lres.Data["fen"])

	// 3) chess_evaluate_game (qua stub engine) — kiểm tra parse PGN + dựng plies.
	t.Log("───── chess_evaluate_game ─────")
	etool := NewChessEvaluateGameTool(stubEngine{})
	eargs, _ := json.Marshal(ChessEvaluateGameInput{
		PGN: "1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 1-0", Depth: 8,
	})
	eres, err := etool.Execute(ctx, eargs)
	if err != nil || !eres.Success {
		t.Fatalf("evaluate lỗi: %v / %s", err, eres.Error)
	}
	plies, _ := eres.Data["plies"].([]map[string]interface{})
	t.Logf("Output:\n%s", eres.Output)
	t.Logf("Data: display_type=%v số_nước=%d", eres.Data["display_type"], len(plies))
}
