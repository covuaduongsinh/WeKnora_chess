// Package chess cung cấp khả năng phân tích cờ vua cho WeKnora.
//
// Toàn bộ gói này sử dụng giấy phép permissive:
//   - Engine phân tích mặc định: Arasan (MIT) chạy như tiến trình UCI riêng.
//   - Logic bàn cờ (FEN/PGN/SAN/nước hợp lệ): github.com/notnil/chess (MIT).
//
// Engine GPL (Stockfish/Lc0) KHÔNG được import vào đây. Mọi engine đều giao
// tiếp qua giao thức UCI (stdin/stdout) hoặc HTTP, nên nằm sau ranh giới tiến
// trình — code WeKnora không bị ràng buộc giấy phép của engine. Tầng engine
// được trừu tượng hóa sau interface EngineClient để có thể thay engine bất kỳ
// lúc nào mà không sửa các tool/agent gọi nó.
package chess

import (
	"context"
	"errors"
	"fmt"
)

// Các lỗi chuẩn của gói.
var (
	// ErrEngineUnavailable nghĩa là không có engine nào được cấu hình/khởi động.
	ErrEngineUnavailable = errors.New("chess: engine không khả dụng")
	// ErrInvalidFEN nghĩa là chuỗi FEN không hợp lệ.
	ErrInvalidFEN = errors.New("chess: FEN không hợp lệ")
	// ErrAnalysisTimeout nghĩa là engine vượt quá thời gian cho phép.
	ErrAnalysisTimeout = errors.New("chess: phân tích vượt thời gian cho phép")
)

// Analysis là kết quả phân tích một thế cờ từ engine.
//
// Quy ước điểm số: EvalCP và MateIn được tính theo góc nhìn của BÊN ĐANG ĐI
// (đúng theo giao thức UCI). Giá trị dương = bên đang đi có lợi. Dùng
// WhiteCentipawns() để quy về góc nhìn cố định của Trắng khi cần hiển thị.
type Analysis struct {
	// FEN là thế cờ đã phân tích (giữ nguyên đầu vào).
	FEN string `json:"fen"`
	// BestMove là nước tốt nhất ở dạng UCI, ví dụ "e2e4", "e7e8q".
	BestMove string `json:"best_move"`
	// BestMoveSAN là nước tốt nhất ở dạng SAN, ví dụ "e4", "Nf3", "e8=Q".
	// Có thể rỗng nếu không chuyển đổi được.
	BestMoveSAN string `json:"best_move_san,omitempty"`
	// EvalCP là đánh giá theo centipawn (1 tốt = 100), góc nhìn bên đang đi.
	// Bỏ qua nếu IsMate = true.
	EvalCP int `json:"eval_cp"`
	// IsMate cho biết đây là thế chiếu hết bắt buộc.
	IsMate bool `json:"is_mate"`
	// MateIn là số nước tới chiếu hết (dương: bên đang đi chiếu hết;
	// âm: bên đang đi bị chiếu hết). Chỉ có nghĩa khi IsMate = true.
	MateIn int `json:"mate_in,omitempty"`
	// Depth là độ sâu engine đã tìm.
	Depth int `json:"depth"`
	// PV (principal variation) là chuỗi nước đi tốt nhất dạng UCI.
	PV []string `json:"pv,omitempty"`
	// SideToMove là bên đang đi: "w" hoặc "b".
	SideToMove string `json:"side_to_move"`
}

// WhiteCentipawns quy đổi EvalCP về góc nhìn cố định của quân Trắng
// (dương = Trắng có lợi), tiện cho hiển thị thanh đánh giá.
func (a *Analysis) WhiteCentipawns() int {
	if a.SideToMove == "b" {
		return -a.EvalCP
	}
	return a.EvalCP
}

// EngineClient là giao diện trừu tượng cho một engine cờ vua.
// Mọi triển khai (UCI tiến trình con, HTTP sidecar, hay engine Go nhúng) đều
// thỏa giao diện này, nên việc thay engine không ảnh hưởng tới tầng gọi.
type EngineClient interface {
	// Analyze phân tích thế cờ FEN tới độ sâu depth và trả về kết quả.
	// depth <= 0 nghĩa là dùng độ sâu mặc định của triển khai.
	// Tôn trọng việc hủy/timeout qua ctx.
	Analyze(ctx context.Context, fen string, depth int) (*Analysis, error)
	// Close giải phóng tài nguyên (tiến trình con, kết nối...).
	Close() error
}

// Config là cấu hình tối thiểu để dựng một EngineClient.
// Map 1-1 với config.ChessConfig của ứng dụng; tách riêng để gói chess không
// phụ thuộc ngược vào gói config.
type Config struct {
	// Mode: "uci" (mặc định, spawn binary cục bộ) hoặc "http" (gọi sidecar).
	Mode string
	// EnginePath là đường dẫn tới binary engine UCI (chế độ "uci").
	EnginePath string
	// EngineEndpoint là URL sidecar HTTP (chế độ "http").
	EngineEndpoint string
	// DefaultDepth là độ sâu tìm kiếm mặc định khi caller không chỉ định.
	DefaultDepth int
	// MaxConcurrent là số tiến trình engine tối đa trong pool (chế độ "uci").
	MaxConcurrent int
	// TimeoutSec là thời gian tối đa cho một lệnh Analyze (giây).
	TimeoutSec int
}

// Các giá trị mặc định an toàn cho cấu hình engine.
const (
	DefaultDepth         = 14
	DefaultMaxConcurrent = 2
	DefaultTimeoutSec    = 15
)

// withDefaults trả về bản sao Config đã điền giá trị mặc định cho ô còn trống.
func (c Config) withDefaults() Config {
	if c.Mode == "" {
		c.Mode = "uci"
	}
	if c.DefaultDepth <= 0 {
		c.DefaultDepth = DefaultDepth
	}
	if c.MaxConcurrent <= 0 {
		c.MaxConcurrent = DefaultMaxConcurrent
	}
	if c.TimeoutSec <= 0 {
		c.TimeoutSec = DefaultTimeoutSec
	}
	return c
}

// NewEngineClient dựng một EngineClient theo cấu hình.
// Trả về ErrEngineUnavailable nếu cấu hình không đủ để dựng engine (ví dụ
// chế độ uci nhưng thiếu EnginePath) — caller nên coi khả năng phân tích là
// tắt và xử lý nhẹ nhàng thay vì lỗi cứng.
func NewEngineClient(cfg Config) (EngineClient, error) {
	cfg = cfg.withDefaults()
	switch cfg.Mode {
	case "uci":
		if cfg.EnginePath == "" {
			return nil, fmt.Errorf("%w: thiếu engine_path cho chế độ uci", ErrEngineUnavailable)
		}
		return newUCIEngine(cfg)
	case "http":
		if cfg.EngineEndpoint == "" {
			return nil, fmt.Errorf("%w: thiếu engine_endpoint cho chế độ http", ErrEngineUnavailable)
		}
		return newHTTPEngine(cfg), nil
	default:
		return nil, fmt.Errorf("%w: chế độ engine không hỗ trợ %q", ErrEngineUnavailable, cfg.Mode)
	}
}
