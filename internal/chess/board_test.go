package chess

import "testing"

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func TestValidateFEN(t *testing.T) {
	if err := ValidateFEN(startFEN); err != nil {
		t.Fatalf("FEN khởi đầu phải hợp lệ, lỗi: %v", err)
	}
	if err := ValidateFEN(""); err == nil {
		t.Fatal("FEN rỗng phải bị từ chối")
	}
	if err := ValidateFEN("không phải fen"); err == nil {
		t.Fatal("FEN sai phải bị từ chối")
	}
}

func TestSideToMove(t *testing.T) {
	if got := sideToMove(startFEN); got != "w" {
		t.Fatalf("muốn w, nhận %q", got)
	}
	black := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	if got := sideToMove(black); got != "b" {
		t.Fatalf("muốn b, nhận %q", got)
	}
}

func TestUCIToSAN(t *testing.T) {
	san, err := UCIToSAN(startFEN, "e2e4")
	if err != nil {
		t.Fatalf("lỗi không mong đợi: %v", err)
	}
	if san != "e4" {
		t.Fatalf("muốn e4, nhận %q", san)
	}
	if _, err := UCIToSAN(startFEN, "e2e5"); err == nil {
		t.Fatal("nước không hợp lệ phải báo lỗi")
	}
}

func TestSANToUCI(t *testing.T) {
	uci, err := SANToUCI(startFEN, "Nf3")
	if err != nil {
		t.Fatalf("lỗi không mong đợi: %v", err)
	}
	if uci != "g1f3" {
		t.Fatalf("muốn g1f3, nhận %q", uci)
	}
}

func TestFENAfterMove(t *testing.T) {
	fen, err := FENAfterMove(startFEN, "e2e4")
	if err != nil {
		t.Fatalf("lỗi không mong đợi: %v", err)
	}
	if sideToMove(fen) != "b" {
		t.Fatalf("sau 1.e4 phải tới lượt Đen, FEN: %s", fen)
	}
}

func TestParsePGN(t *testing.T) {
	pgn := `[Event "Test"]
[White "A"]
[Black "B"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 1-0`
	info, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("lỗi không mong đợi: %v", err)
	}
	if len(info.Plies) != 4 {
		t.Fatalf("muốn 4 nước, nhận %d", len(info.Plies))
	}
	if info.Plies[0].SAN != "e4" || info.Plies[0].Side != "w" {
		t.Fatalf("nước đầu sai: %+v", info.Plies[0])
	}
	if info.Tags["White"] != "A" {
		t.Fatalf("thẻ White sai: %q", info.Tags["White"])
	}
	if info.Plies[2].MoveNumber != 2 {
		t.Fatalf("nước thứ 3 phải là nước đầy đủ số 2, nhận %d", info.Plies[2].MoveNumber)
	}
}

func TestWhiteCentipawns(t *testing.T) {
	a := &Analysis{EvalCP: 50, SideToMove: "b"}
	if got := a.WhiteCentipawns(); got != -50 {
		t.Fatalf("muốn -50, nhận %d", got)
	}
	a.SideToMove = "w"
	if got := a.WhiteCentipawns(); got != 50 {
		t.Fatalf("muốn 50, nhận %d", got)
	}
}
