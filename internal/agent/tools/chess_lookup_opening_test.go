package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// Dataset ECO nhúng phải nạp được nhiều khai cuộc, và overlay Việt hoá phải thắng
// ở các khai cuộc phổ biến.
func TestOpeningIndexLoaded(t *testing.T) {
	if len(openingIndex) < 1000 {
		t.Fatalf("openingIndex quá nhỏ (%d) — dataset ECO chưa nạp?", len(openingIndex))
	}
	// "b4" KHÔNG có trong bảng Việt hoá gốc → chỉ nhận diện được nhờ dataset nhúng.
	if _, ok := openingIndex["b4"]; !ok {
		t.Errorf("kỳ vọng nhận diện 1.b4 từ dataset nhúng")
	}
	// Overlay Việt hoá: 1.e4 c5 phải giữ tên có chữ "Sicilian".
	if e := openingIndex["e4 c5"]; !strings.Contains(e.name, "Sicilian") {
		t.Errorf("kỳ vọng tên Việt hoá cho 1.e4 c5 chứa 'Sicilian', nhận %q", e.name)
	}
}

// sanKeyFromPGN phải bỏ số nước và hậu tố, chỉ giữ chuỗi SAN.
func TestSanKeyFromPGN(t *testing.T) {
	cases := map[string]string{
		"1. e4 c5 2. Nf3":            "e4 c5 Nf3",
		"1. d4 Nf6 2. c4 g6 3. Nc3":  "d4 Nf6 c4 g6 Nc3",
		"1. e4 e5 2. Nf3 Nc6 3. Bb5": "e4 e5 Nf3 Nc6 Bb5",
	}
	for in, want := range cases {
		if got := sanKeyFromPGN(in); got != want {
			t.Errorf("sanKeyFromPGN(%q) = %q, muốn %q", in, got, want)
		}
	}
}

// Tool phải nhận diện một khai cuộc NGOÀI bảng gốc (vd 1.f4 Bird) nhờ dataset.
func TestLookupOpening_FromDataset(t *testing.T) {
	tool := NewChessLookupOpeningTool()
	args, _ := json.Marshal(ChessLookupOpeningInput{Moves: []string{"f4"}})
	res, err := tool.Execute(context.Background(), args)
	if err != nil || res == nil || !res.Success {
		t.Fatalf("lookup lỗi: %v / %+v", err, res)
	}
	if !strings.Contains(res.Output, "Mã ECO") {
		t.Errorf("kỳ vọng nhận diện 1.f4 từ dataset, nhận output:\n%s", res.Output)
	}
}
