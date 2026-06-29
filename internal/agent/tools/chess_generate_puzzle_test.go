package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

// fakePuzzleSource là PuzzleSource giả cho test: trả một bài cố định (hoặc lỗi) và
// ghi lại bộ lọc của LẦN GỌI ĐẦU để kiểm tra ánh xạ độ khó.
type fakePuzzleSource struct {
	puzzle      *types.ChessPuzzle
	err         error
	calls       int
	firstFilter types.ChessPuzzleFilter
}

func (f *fakePuzzleSource) RandomPuzzle(_ context.Context, _ uint64, flt types.ChessPuzzleFilter) (*types.ChessPuzzle, error) {
	f.calls++
	if f.calls == 1 {
		f.firstFilter = flt
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.puzzle, nil
}

func ctxWithTenant(id uint64) context.Context {
	return context.WithValue(context.Background(), types.TenantIDContextKey, id)
}

func runGeneratePuzzle(t *testing.T, tool *ChessGeneratePuzzleTool, ctx context.Context, in ChessGeneratePuzzleInput) map[string]interface{} {
	t.Helper()
	args, _ := json.Marshal(in)
	res, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Execute lỗi: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("kỳ vọng Success, nhận: %+v", res)
	}
	if res.Data["display_type"] != "chess_board" {
		t.Fatalf("kỳ vọng display_type chess_board, nhận %v", res.Data["display_type"])
	}
	return res.Data
}

func isEmbeddedFEN(fen string) bool {
	for _, p := range embeddedPuzzles {
		if p.fen == fen {
			return true
		}
	}
	return false
}

// Khi ngân hàng bài tập có dữ liệu → tool phải dùng bài từ DB (không phải bộ nhúng),
// và ánh xạ độ khó "khó" → "kho" khi lọc.
func TestGeneratePuzzle_UsesBankWhenAvailable(t *testing.T) {
	dbFEN := "8/8/8/3k4/8/3K4/3P4/8 w - - 0 1"
	src := &fakePuzzleSource{puzzle: &types.ChessPuzzle{
		FEN: dbFEN, Title: "Tốt thông", Theme: "tàn cuộc", Difficulty: "kho",
	}}
	tool := NewChessGeneratePuzzleTool(src)

	data := runGeneratePuzzle(t, tool, ctxWithTenant(1),
		ChessGeneratePuzzleInput{Theme: "tàn cuộc", Difficulty: "khó"})

	if data["fen"] != dbFEN {
		t.Errorf("kỳ vọng FEN từ kho %q, nhận %q", dbFEN, data["fen"])
	}
	if src.calls == 0 {
		t.Errorf("kỳ vọng có gọi RandomPuzzle")
	}
	if src.firstFilter.Difficulty != "kho" {
		t.Errorf("độ khó lọc DB kỳ vọng 'kho', nhận %q", src.firstFilter.Difficulty)
	}
	if src.firstFilter.Theme != "tàn cuộc" {
		t.Errorf("theme lọc DB kỳ vọng 'tàn cuộc', nhận %q", src.firstFilter.Theme)
	}
}

// Kho lỗi/trống → fallback bộ bài tập nhúng (tool không "câm").
func TestGeneratePuzzle_FallbackWhenBankEmpty(t *testing.T) {
	src := &fakePuzzleSource{err: errors.New("không có bản ghi")}
	tool := NewChessGeneratePuzzleTool(src)

	data := runGeneratePuzzle(t, tool, ctxWithTenant(1), ChessGeneratePuzzleInput{})
	fen, _ := data["fen"].(string)
	if !isEmbeddedFEN(fen) {
		t.Errorf("kỳ vọng FEN từ bộ nhúng, nhận %q", fen)
	}
}

// Không có nguồn (source nil) → dùng bộ nhúng.
func TestGeneratePuzzle_NilSourceFallback(t *testing.T) {
	tool := NewChessGeneratePuzzleTool(nil)
	data := runGeneratePuzzle(t, tool, context.Background(), ChessGeneratePuzzleInput{})
	fen, _ := data["fen"].(string)
	if !isEmbeddedFEN(fen) {
		t.Errorf("kỳ vọng FEN từ bộ nhúng, nhận %q", fen)
	}
}

// Không có tenant trong ctx → KHÔNG truy vấn DB, fallback bộ nhúng.
func TestGeneratePuzzle_NoTenantSkipsBank(t *testing.T) {
	src := &fakePuzzleSource{puzzle: &types.ChessPuzzle{FEN: "8/8/8/8/4k3/8/4K3/8 w - - 0 1"}}
	tool := NewChessGeneratePuzzleTool(src)

	data := runGeneratePuzzle(t, tool, context.Background(), ChessGeneratePuzzleInput{}) // không tenant
	if src.calls != 0 {
		t.Errorf("không có tenant thì không nên gọi DB; calls=%d", src.calls)
	}
	fen, _ := data["fen"].(string)
	if !isEmbeddedFEN(fen) {
		t.Errorf("kỳ vọng FEN từ bộ nhúng, nhận %q", fen)
	}
}
