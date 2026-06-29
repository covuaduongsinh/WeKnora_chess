package interfaces

import "context"

// ChessEngineService cung cấp trạng thái sức khỏe của engine cờ vua (Arasan) cho
// tầng handler — phục vụ endpoint vận hành GET /chess/engine/health (monitor/debug).
type ChessEngineService interface {
	// Enabled cho biết engine cờ có được bật trong cấu hình không.
	Enabled() bool
	// Health trả nil nếu engine bật và phản hồi; lỗi nếu tắt hoặc không phản hồi.
	Health(ctx context.Context) error
}
