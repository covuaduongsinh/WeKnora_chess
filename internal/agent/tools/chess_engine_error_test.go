package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/chess"
)

// errEngine là EngineClient luôn trả một lỗi cho trước — để kiểm tra tool gọi engine
// dịch lỗi sang thông báo thân thiện.
type errEngine struct{ err error }

func (e errEngine) Analyze(_ context.Context, _ string, _ int) (*chess.Analysis, error) {
	return nil, e.err
}
func (errEngine) Close() error { return nil }

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// Khi engine "câm" (ErrEngineUnavailable) → tool trả thông báo thân thiện, KHÔNG lộ
// lỗi kỹ thuật thô.
func TestEngineToolFriendlyError(t *testing.T) {
	tool := NewChessAnalyzePositionTool(errEngine{err: chess.ErrEngineUnavailable}, 12)
	args, _ := json.Marshal(ChessAnalyzePositionInput{FEN: startFEN})
	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute trả err: %v", err)
	}
	if res.Success {
		t.Fatalf("kỳ vọng Success=false khi engine lỗi")
	}
	if strings.Contains(res.Error, "Engine lỗi") || !strings.Contains(res.Error, "không phản hồi") {
		t.Errorf("kỳ vọng thông báo thân thiện, nhận: %q", res.Error)
	}
}

// friendlyEngineError map đúng các lỗi engine đã biết.
func TestFriendlyEngineErrorMapping(t *testing.T) {
	if got := friendlyEngineError(chess.ErrAnalysisTimeout); !strings.Contains(got, "hết thời gian") {
		t.Errorf("timeout map sai: %q", got)
	}
	if got := friendlyEngineError(chess.ErrEngineUnavailable); !strings.Contains(got, "không phản hồi") {
		t.Errorf("unavailable map sai: %q", got)
	}
	if got := friendlyEngineError(chess.ErrInvalidFEN); !strings.Contains(got, "FEN") {
		t.Errorf("invalid-fen map sai: %q", got)
	}
}
