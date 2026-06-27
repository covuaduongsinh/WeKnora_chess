package chess

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Mã nguồn một "engine UCI giả" tối giản (chỉ stdlib) để kiểm thử đường ống
// uci_engine thật: bắt tay uci/uciok, isready/readyok, và trả bestmove + score.
const mockEngineSource = `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "uci":
			fmt.Println("id name MockEngine")
			fmt.Println("uciok")
		case line == "isready":
			fmt.Println("readyok")
		case strings.HasPrefix(line, "go"):
			fmt.Println("info depth 12 score cp 35 pv e2e4 e7e5 g1f3")
			fmt.Println("bestmove e2e4")
		case line == "quit":
			return
		}
	}
}
`

// buildMockEngine biên dịch engine giả ra một binary tạm và trả về đường dẫn.
func buildMockEngine(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(mockEngineSource), 0o644); err != nil {
		t.Fatalf("ghi nguồn engine giả lỗi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module mockengine\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("ghi go.mod lỗi: %v", err)
	}
	out := filepath.Join(dir, "mockengine")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build engine giả lỗi: %v\n%s", err, b)
	}
	return out
}

// TestUCIEngineEndToEnd kiểm thử toàn bộ đường ống engine UCI với một engine giả:
// spawn tiến trình → bắt tay → gửi position/go → parse bestmove + điểm số.
func TestUCIEngineEndToEnd(t *testing.T) {
	enginePath := buildMockEngine(t)

	client, err := NewEngineClient(Config{
		Mode:          "uci",
		EnginePath:    enginePath,
		DefaultDepth:  12,
		MaxConcurrent: 1,
		TimeoutSec:    10,
	})
	if err != nil {
		t.Fatalf("không tạo được engine client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	a, err := client.Analyze(ctx, startFEN, 12)
	if err != nil {
		t.Fatalf("Analyze lỗi: %v", err)
	}

	if a.BestMove != "e2e4" {
		t.Fatalf("muốn bestmove e2e4, nhận %q", a.BestMove)
	}
	if a.BestMoveSAN != "e4" {
		t.Fatalf("muốn SAN e4, nhận %q", a.BestMoveSAN)
	}
	if a.EvalCP != 35 {
		t.Fatalf("muốn eval 35cp, nhận %d", a.EvalCP)
	}
	if a.SideToMove != "w" {
		t.Fatalf("muốn bên đi w, nhận %q", a.SideToMove)
	}

	t.Logf("✅ Engine UCI (giả) chạy: bestmove=%s SAN=%s eval=%dcp depth=%d pv=%v",
		a.BestMove, a.BestMoveSAN, a.EvalCP, a.Depth, a.PV)
}
