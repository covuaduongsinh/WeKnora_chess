package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// ChessEngineHandler expose trạng thái sức khỏe engine cờ cho vận hành/monitor.
type ChessEngineHandler struct {
	service interfaces.ChessEngineService
}

// NewChessEngineHandler tạo handler trạng thái engine cờ.
func NewChessEngineHandler(service interfaces.ChessEngineService) *ChessEngineHandler {
	return &ChessEngineHandler{service: service}
}

// Health GET /chess/engine/health — báo engine cờ có bật & có phản hồi không.
// Luôn trả 200 (đây là báo cáo trạng thái, không phải thao tác có thể "thất bại").
func (h *ChessEngineHandler) Health(c *gin.Context) {
	enabled := h.service.Enabled()
	healthy := false
	detail := ""
	if enabled {
		if err := h.service.Health(c.Request.Context()); err != nil {
			detail = err.Error()
		} else {
			healthy = true
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": enabled,
			"healthy": healthy,
			"detail":  detail,
		},
	})
}
