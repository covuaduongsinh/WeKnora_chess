package tools

import (
	"bufio"
	_ "embed"
	"strings"
)

// chess_openings_data.go nạp cơ sở dữ liệu khai cuộc ECO (nhúng) một lần lúc khởi
// động, dựng index tra cứu theo chuỗi nước SAN. Dataset nguồn: lichess-org/
// chess-openings (CC0) — xem data/NOTICE-eco.md.

//go:embed data/eco.tsv
var ecoTSV string

// openingIndex là index tra khai cuộc: key = chuỗi nước SAN chuẩn hoá nối bằng dấu
// cách (vd "e4 c5 Nf3"), value = (mã ECO, tên). Gộp dataset nhúng (tên tiếng Anh)
// với overlay `ecoOpenings` Việt hoá — overlay đè để khai cuộc phổ biến có tên
// tiếng Việt. Tra theo tiền tố dài nhất khớp (xem chess_lookup_opening.go).
var openingIndex = buildOpeningIndex()

// buildOpeningIndex parse TSV nhúng rồi overlay bảng Việt hoá lên trên.
func buildOpeningIndex() map[string]ecoEntry {
	idx := make(map[string]ecoEntry, 4096)

	sc := bufio.NewScanner(strings.NewReader(ecoTSV))
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	first := true
	for sc.Scan() {
		line := sc.Text()
		if first { // bỏ dòng header "eco\tname\tpgn"
			first = false
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 3 {
			continue
		}
		key := sanKeyFromPGN(cols[2])
		if key == "" {
			continue
		}
		idx[key] = ecoEntry{eco: strings.TrimSpace(cols[0]), name: strings.TrimSpace(cols[1])}
	}

	// Overlay bảng Việt hoá (ưu tiên tên tiếng Việt cho khai cuộc phổ biến).
	for k, v := range ecoOpenings {
		idx[k] = v
	}
	return idx
}

// sanKeyFromPGN biến cột pgn ("1. e4 c5 2. Nf3") thành key SAN chuẩn hoá ("e4 c5
// Nf3") — bỏ token số nước ("1." "2." …) và hậu tố !?+# trên từng nước.
func sanKeyFromPGN(pgn string) string {
	fields := strings.Fields(pgn)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f == "" || strings.HasSuffix(f, ".") { // "1." "2." "10." …
			continue
		}
		if n := normalizeSAN(f); n != "" {
			out = append(out, n)
		}
	}
	return strings.Join(out, " ")
}
