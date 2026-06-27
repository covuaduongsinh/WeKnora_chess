package chess

import (
	"fmt"
	"strings"

	notnil "github.com/notnil/chess"
)

// board.go gói các thao tác bàn cờ dựa trên github.com/notnil/chess (MIT):
// kiểm tra FEN hợp lệ, chuyển đổi giữa ký hiệu UCI và SAN, và bóc tách các thế
// cờ từ một ván PGN. Không có engine GPL nào ở đây — chỉ logic luật cờ thuần.

// ValidateFEN trả về nil nếu chuỗi FEN hợp lệ, ngược lại trả về ErrInvalidFEN.
func ValidateFEN(fen string) error {
	if strings.TrimSpace(fen) == "" {
		return fmt.Errorf("%w: chuỗi rỗng", ErrInvalidFEN)
	}
	if _, err := notnil.FEN(fen); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFEN, err)
	}
	return nil
}

// gameFromFEN dựng một ván cờ notnil bắt đầu từ thế cờ FEN cho trước.
func gameFromFEN(fen string) (*notnil.Game, error) {
	fenOpt, err := notnil.FEN(fen)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFEN, err)
	}
	return notnil.NewGame(fenOpt), nil
}

// sideToMove trả về "w" hoặc "b" từ trường thứ hai của FEN.
// Dùng cách tách chuỗi nhẹ để không phải dựng cả ván cờ.
func sideToMove(fen string) string {
	fields := strings.Fields(fen)
	if len(fields) >= 2 && (fields[1] == "w" || fields[1] == "b") {
		return fields[1]
	}
	return "w"
}

// UCIToSAN chuyển một nước đi ký hiệu UCI ("e2e4", "e7e8q") sang SAN ("e4",
// "e8=Q") trong ngữ cảnh thế cờ FEN.
func UCIToSAN(fen, uciMove string) (string, error) {
	g, err := gameFromFEN(fen)
	if err != nil {
		return "", err
	}
	pos := g.Position()
	mv, err := notnil.UCINotation{}.Decode(pos, uciMove)
	if err != nil {
		return "", fmt.Errorf("chess: nước UCI không hợp lệ %q: %v", uciMove, err)
	}
	// UCINotation.Decode chỉ bóc tọa độ, KHÔNG kiểm tra luật → tự xác thực
	// nước đi nằm trong danh sách hợp lệ của thế cờ.
	legal := findLegalMove(pos, mv)
	if legal == nil {
		return "", fmt.Errorf("chess: nước %q không hợp lệ ở thế cờ này", uciMove)
	}
	return notnil.AlgebraicNotation{}.Encode(pos, legal), nil
}

// findLegalMove tìm nước hợp lệ khớp ô đi/ô đến/phong cấp với mv, hoặc nil.
func findLegalMove(pos *notnil.Position, mv *notnil.Move) *notnil.Move {
	for _, vm := range pos.ValidMoves() {
		if vm.S1() == mv.S1() && vm.S2() == mv.S2() && vm.Promo() == mv.Promo() {
			return vm
		}
	}
	return nil
}

// SANToUCI chuyển một nước đi SAN ("Nf3", "e4", "O-O") sang ký hiệu UCI trong
// ngữ cảnh thế cờ FEN.
func SANToUCI(fen, sanMove string) (string, error) {
	g, err := gameFromFEN(fen)
	if err != nil {
		return "", err
	}
	pos := g.Position()
	mv, err := notnil.AlgebraicNotation{}.Decode(pos, sanMove)
	if err != nil {
		return "", fmt.Errorf("chess: nước SAN không hợp lệ %q: %v", sanMove, err)
	}
	return notnil.UCINotation{}.Encode(pos, mv), nil
}

// FENAfterMove trả về FEN sau khi thực hiện một nước (UCI hoặc SAN) từ thế cờ
// FEN cho trước. Phát hiện định dạng tự động.
func FENAfterMove(fen, move string) (string, error) {
	g, err := gameFromFEN(fen)
	if err != nil {
		return "", err
	}
	pos := g.Position()
	mv, decErr := notnil.UCINotation{}.Decode(pos, move)
	if decErr != nil {
		mv, decErr = notnil.AlgebraicNotation{}.Decode(pos, move)
		if decErr != nil {
			return "", fmt.Errorf("chess: nước không hợp lệ %q", move)
		}
	}
	if err := g.Move(mv); err != nil {
		return "", fmt.Errorf("chess: không đi được nước %q: %v", move, err)
	}
	return g.Position().String(), nil
}

// GamePly là một nước trong ván cờ kèm thế cờ trước và sau nước đó.
type GamePly struct {
	// MoveNumber là số nước đầy đủ (1, 1, 2, 2, ...).
	MoveNumber int `json:"move_number"`
	// Side là bên đi nước này: "w" hoặc "b".
	Side string `json:"side"`
	// SAN là nước đi ở dạng SAN.
	SAN string `json:"san"`
	// UCI là nước đi ở dạng UCI.
	UCI string `json:"uci"`
	// FENBefore là thế cờ trước khi đi nước này.
	FENBefore string `json:"fen_before"`
	// FENAfter là thế cờ sau khi đi nước này.
	FENAfter string `json:"fen_after"`
}

// GameInfo là một ván cờ đã phân tách: metadata + danh sách nước đi.
type GameInfo struct {
	// Tags là các thẻ PGN (Event, White, Black, Result, ECO...).
	Tags map[string]string `json:"tags,omitempty"`
	// Plies là danh sách nước đi theo thứ tự.
	Plies []GamePly `json:"plies"`
	// Outcome là kết quả ván ("1-0", "0-1", "1/2-1/2", "*").
	Outcome string `json:"outcome"`
}

// ParsePGN phân tích một ván PGN duy nhất thành GameInfo, kèm FEN trước/sau cho
// mỗi nước — tiện cho việc chấm ván bằng engine và hiển thị bàn cờ.
func ParsePGN(pgn string) (*GameInfo, error) {
	pgnOpt, err := notnil.PGN(strings.NewReader(pgn))
	if err != nil {
		return nil, fmt.Errorf("chess: PGN không hợp lệ: %v", err)
	}
	g := notnil.NewGame(pgnOpt)

	positions := g.Positions() // gồm cả thế cờ ban đầu → len = số nước + 1
	moves := g.Moves()

	info := &GameInfo{
		Tags:    map[string]string{},
		Outcome: string(g.Outcome()),
		Plies:   make([]GamePly, 0, len(moves)),
	}
	for _, tp := range g.TagPairs() {
		info.Tags[tp.Key] = tp.Value
	}

	for i, mv := range moves {
		before := positions[i]
		after := positions[i+1]
		side := "w"
		if before.Turn() == notnil.Black {
			side = "b"
		}
		info.Plies = append(info.Plies, GamePly{
			MoveNumber: i/2 + 1,
			Side:       side,
			SAN:        notnil.AlgebraicNotation{}.Encode(before, mv),
			UCI:        notnil.UCINotation{}.Encode(before, mv),
			FENBefore:  before.String(),
			FENAfter:   after.String(),
		})
	}
	return info, nil
}

// ImportedGame là một ván cờ rút gọn dùng cho việc import kho ván:
// metadata (thẻ PGN), PGN chuẩn hóa, số nước và kết quả.
type ImportedGame struct {
	Tags     map[string]string `json:"tags"`
	PGN      string            `json:"pgn"`
	PlyCount int               `json:"ply_count"`
	Outcome  string            `json:"outcome"`
}

// ParseMultiPGN tách một chuỗi PGN chứa NHIỀU ván thành danh sách ImportedGame,
// dùng notnil GamesFromPGN (MIT). Tiện cho việc import kho ván đấu.
func ParseMultiPGN(pgn string) ([]ImportedGame, error) {
	games, err := notnil.GamesFromPGN(strings.NewReader(pgn))
	if err != nil {
		return nil, fmt.Errorf("chess: PGN không hợp lệ: %v", err)
	}
	out := make([]ImportedGame, 0, len(games))
	for _, g := range games {
		tags := map[string]string{}
		for _, tp := range g.TagPairs() {
			tags[tp.Key] = tp.Value
		}
		out = append(out, ImportedGame{
			Tags:     tags,
			PGN:      g.String(),
			PlyCount: len(g.Moves()),
			Outcome:  string(g.Outcome()),
		})
	}
	return out, nil
}

// TagOr trả về giá trị thẻ PGN theo key, hoặc fallback nếu thiếu.
func (ig ImportedGame) TagOr(key, fallback string) string {
	if v, ok := ig.Tags[key]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
