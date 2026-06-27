package chess

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// uciEngine triển khai EngineClient bằng cách spawn một hoặc nhiều tiến trình
// engine UCI cục bộ (mặc định: Arasan, MIT) và giao tiếp qua stdin/stdout.
//
// Bảo mật: EnginePath đến từ cấu hình triển khai, KHÔNG bao giờ từ input người
// dùng, nên không có nguy cơ chèn lệnh. Mỗi tiến trình chỉ phục vụ một lệnh
// Analyze tại một thời điểm; pool giới hạn tổng số tiến trình đồng thời.
type uciEngine struct {
	cfg  Config
	pool chan *uciProcess

	mu     sync.Mutex
	closed bool
}

// newUCIEngine dựng pool và khởi động trước một tiến trình để fail nhanh nếu
// binary engine không chạy được.
func newUCIEngine(cfg Config) (*uciEngine, error) {
	e := &uciEngine{
		cfg:  cfg,
		pool: make(chan *uciProcess, cfg.MaxConcurrent),
	}
	// Khởi động trước một tiến trình để xác thực binary; phần còn lại tạo lazy.
	proc, err := newUCIProcess(cfg.EnginePath)
	if err != nil {
		return nil, fmt.Errorf("%w: không khởi động được engine %q: %v",
			ErrEngineUnavailable, cfg.EnginePath, err)
	}
	e.pool <- proc
	for i := 1; i < cfg.MaxConcurrent; i++ {
		e.pool <- nil // ô trống, sẽ tạo tiến trình khi cần
	}
	return e, nil
}

// acquire lấy một tiến trình khỏe mạnh từ pool, tạo mới nếu ô đang trống/chết.
func (e *uciEngine) acquire(ctx context.Context) (*uciProcess, error) {
	select {
	case proc := <-e.pool:
		if proc != nil && proc.alive() {
			return proc, nil
		}
		// Ô trống hoặc tiến trình đã chết → tạo mới.
		newProc, err := newUCIProcess(e.cfg.EnginePath)
		if err != nil {
			// Trả lại ô trống để không làm hụt pool.
			e.pool <- nil
			return nil, fmt.Errorf("%w: %v", ErrEngineUnavailable, err)
		}
		return newProc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// release trả tiến trình về pool; tiến trình chết được thay bằng ô trống.
func (e *uciEngine) release(proc *uciProcess) {
	if proc != nil && !proc.alive() {
		proc.kill()
		proc = nil
	}
	e.pool <- proc
}

// Analyze thỏa EngineClient.
func (e *uciEngine) Analyze(ctx context.Context, fen string, depth int) (*Analysis, error) {
	if err := ValidateFEN(fen); err != nil {
		return nil, err
	}
	if depth <= 0 {
		depth = e.cfg.DefaultDepth
	}

	// Áp timeout cứng để không treo vòng lặp ReAct của agent.
	timeout := time.Duration(e.cfg.TimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	proc, err := e.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer e.release(proc)

	analysis, err := proc.analyze(ctx, fen, depth)
	if err != nil {
		return nil, err
	}
	// Bổ sung SAN cho nước tốt nhất (best effort, không chặn nếu lỗi).
	if analysis.BestMove != "" {
		if san, sErr := UCIToSAN(fen, analysis.BestMove); sErr == nil {
			analysis.BestMoveSAN = san
		}
	}
	analysis.FEN = fen
	analysis.SideToMove = sideToMove(fen)
	return analysis, nil
}

// Close đóng toàn bộ tiến trình trong pool.
func (e *uciEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	close(e.pool)
	for proc := range e.pool {
		if proc != nil {
			proc.kill()
		}
	}
	return nil
}

// uciProcess bọc một tiến trình engine UCI duy nhất.
type uciProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	mu     sync.Mutex
	dead   bool
}

// newUCIProcess spawn engine và hoàn tất bắt tay UCI (uci/uciok, isready/readyok).
func newUCIProcess(enginePath string) (*uciProcess, error) {
	cmd := exec.Command(enginePath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	p := &uciProcess{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
	}
	p.stdout.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Bắt tay UCI với thời hạn ngắn để fail nhanh nếu binary không phải engine UCI.
	if err := p.handshake(5 * time.Second); err != nil {
		p.kill()
		return nil, err
	}
	return p, nil
}

func (p *uciProcess) handshake(timeout time.Duration) error {
	if err := p.send("uci"); err != nil {
		return err
	}
	if err := p.waitFor("uciok", timeout); err != nil {
		return err
	}
	if err := p.send("isready"); err != nil {
		return err
	}
	return p.waitFor("readyok", timeout)
}

// send ghi một dòng lệnh UCI tới engine.
func (p *uciProcess) send(line string) error {
	_, err := io.WriteString(p.stdin, line+"\n")
	if err != nil {
		p.dead = true
	}
	return err
}

// waitFor đọc stdout cho tới khi gặp token (đầu dòng) hoặc hết hạn.
func (p *uciProcess) waitFor(token string, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		for p.stdout.Scan() {
			if strings.HasPrefix(strings.TrimSpace(p.stdout.Text()), token) {
				done <- nil
				return
			}
		}
		if err := p.stdout.Err(); err != nil {
			done <- err
			return
		}
		done <- io.EOF
	}()
	select {
	case err := <-done:
		if err != nil {
			p.dead = true
		}
		return err
	case <-time.After(timeout):
		p.dead = true
		return fmt.Errorf("chess: engine không phản hồi %q trong %s", token, timeout)
	}
}

// analyze chạy một lệnh tìm kiếm và phân tích kết quả.
func (p *uciProcess) analyze(ctx context.Context, fen string, depth int) (*Analysis, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.send("ucinewgame"); err != nil {
		return nil, err
	}
	if err := p.send("position fen " + fen); err != nil {
		return nil, err
	}
	if err := p.send(fmt.Sprintf("go depth %d", depth)); err != nil {
		return nil, err
	}

	type scanResult struct {
		analysis *Analysis
		err      error
	}
	resCh := make(chan scanResult, 1)
	go func() {
		a := &Analysis{Depth: depth}
		for p.stdout.Scan() {
			line := strings.TrimSpace(p.stdout.Text())
			if strings.HasPrefix(line, "info ") {
				parseInfoLine(line, a)
				continue
			}
			if strings.HasPrefix(line, "bestmove") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[1] != "(none)" {
					a.BestMove = fields[1]
				}
				resCh <- scanResult{analysis: a}
				return
			}
		}
		if err := p.stdout.Err(); err != nil {
			resCh <- scanResult{err: err}
			return
		}
		resCh <- scanResult{err: io.EOF}
	}()

	select {
	case res := <-resCh:
		if res.err != nil {
			p.dead = true
			return nil, res.err
		}
		return res.analysis, nil
	case <-ctx.Done():
		// Yêu cầu engine dừng ngay; nếu vẫn treo, đánh dấu chết để thay thế.
		_ = p.send("stop")
		select {
		case res := <-resCh:
			if res.err == nil && res.analysis.BestMove != "" {
				return res.analysis, nil
			}
		case <-time.After(500 * time.Millisecond):
		}
		p.dead = true
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrAnalysisTimeout
		}
		return nil, ctx.Err()
	}
}

// parseInfoLine bóc score và pv từ một dòng "info ...".
func parseInfoLine(line string, a *Analysis) {
	fields := strings.Fields(line)
	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "depth":
			if i+1 < len(fields) {
				if d, err := strconv.Atoi(fields[i+1]); err == nil {
					a.Depth = d
				}
			}
		case "score":
			if i+2 < len(fields) {
				kind, valStr := fields[i+1], fields[i+2]
				val, err := strconv.Atoi(valStr)
				if err == nil {
					switch kind {
					case "cp":
						a.IsMate = false
						a.EvalCP = val
					case "mate":
						a.IsMate = true
						a.MateIn = val
					}
				}
			}
		case "pv":
			// Phần còn lại của dòng là principal variation.
			if i+1 < len(fields) {
				a.PV = append([]string(nil), fields[i+1:]...)
			}
			return
		}
	}
}

func (p *uciProcess) alive() bool {
	return p != nil && !p.dead
}

func (p *uciProcess) kill() {
	p.dead = true
	_ = p.send("quit")
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}
}
