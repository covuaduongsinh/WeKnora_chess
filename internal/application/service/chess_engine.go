package service

import (
	"context"
	"errors"
	"sync"

	"github.com/Tencent/WeKnora/internal/chess"
	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// ErrChessEngineDisabled báo engine cờ chưa được bật trong cấu hình.
var ErrChessEngineDisabled = errors.New("engine cờ chưa được bật")

// chessEngineService cung cấp trạng thái sức khỏe engine cờ cho tầng handler.
// Engine khởi tạo LƯỜI (chỉ khi cần) và tái dùng. Đây là service độc lập với
// agentService (vốn giữ engine riêng cho tool) — giữ thay đổi tối thiểu ở file
// dùng chung agent_service.go; engine http là wrapper rẻ nên không lo trùng.
type chessEngineService struct {
	cfg    *config.ChessConfig
	once   sync.Once
	client chess.EngineClient
}

// NewChessEngineService tạo service trạng thái engine cờ.
func NewChessEngineService(cfg *config.Config) interfaces.ChessEngineService {
	var cc *config.ChessConfig
	if cfg != nil {
		cc = cfg.Chess
	}
	return &chessEngineService{cfg: cc}
}

// Enabled cho biết engine cờ có được bật trong cấu hình không.
func (s *chessEngineService) Enabled() bool {
	return s.cfg != nil && s.cfg.Enabled
}

// engine khởi tạo (một lần) và trả về engine client; nil nếu chưa bật/cấu hình lỗi.
func (s *chessEngineService) engine() chess.EngineClient {
	s.once.Do(func() {
		if s.cfg == nil || !s.cfg.Enabled {
			return
		}
		client, err := chess.NewEngineClient(chess.Config{
			Mode:           s.cfg.Mode,
			EnginePath:     s.cfg.EnginePath,
			EngineEndpoint: s.cfg.EngineEndpoint,
			DefaultDepth:   s.cfg.DefaultDepth,
			MaxConcurrent:  s.cfg.MaxConcurrent,
			TimeoutSec:     s.cfg.TimeoutSec,
		})
		if err != nil {
			return
		}
		s.client = client
	})
	return s.client
}

// Health kiểm tra engine có sẵn sàng phản hồi không.
func (s *chessEngineService) Health(ctx context.Context) error {
	if !s.Enabled() {
		return ErrChessEngineDisabled
	}
	eng := s.engine()
	if eng == nil {
		return chess.ErrEngineUnavailable
	}
	// Chế độ http có probe /health; chế độ uci tạo được client coi như sẵn sàng.
	if hc, ok := eng.(interface{ Health(context.Context) error }); ok {
		return hc.Health(ctx)
	}
	return nil
}
