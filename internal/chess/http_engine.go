package chess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// httpEngine triển khai EngineClient bằng cách gọi một sidecar HTTP bọc engine
// UCI. Hữu ích khi muốn tách engine ra dịch vụ riêng để scale ngang hoặc chạy
// trên máy khác. Sidecar dự kiến nhận POST {fen, depth} và trả JSON Analysis.
type httpEngine struct {
	endpoint string
	depth    int
	client   *http.Client
}

func newHTTPEngine(cfg Config) *httpEngine {
	return &httpEngine{
		endpoint: strings.TrimRight(cfg.EngineEndpoint, "/"),
		depth:    cfg.DefaultDepth,
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
		},
	}
}

type httpAnalyzeRequest struct {
	FEN   string `json:"fen"`
	Depth int    `json:"depth"`
}

// Analyze thỏa EngineClient.
func (e *httpEngine) Analyze(ctx context.Context, fen string, depth int) (*Analysis, error) {
	if err := ValidateFEN(fen); err != nil {
		return nil, err
	}
	if depth <= 0 {
		depth = e.depth
	}

	body, err := json.Marshal(httpAnalyzeRequest{FEN: fen, Depth: depth})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.endpoint+"/analyze", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrAnalysisTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrEngineUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("chess: sidecar trả mã %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var analysis Analysis
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		return nil, fmt.Errorf("chess: không giải mã được phản hồi sidecar: %w", err)
	}

	// Đảm bảo các trường suy ra luôn nhất quán dù sidecar có điền hay không.
	analysis.FEN = fen
	if analysis.SideToMove == "" {
		analysis.SideToMove = sideToMove(fen)
	}
	if analysis.BestMoveSAN == "" && analysis.BestMove != "" {
		if san, sErr := UCIToSAN(fen, analysis.BestMove); sErr == nil {
			analysis.BestMoveSAN = san
		}
	}
	return &analysis, nil
}

// Close không cần làm gì với client HTTP.
func (e *httpEngine) Close() error { return nil }
