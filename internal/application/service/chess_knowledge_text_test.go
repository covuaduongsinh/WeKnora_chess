package service

import (
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestExpandChessWikilinks(t *testing.T) {
	cases := map[string]string{
		"Xem [[game/morphy-opera|Morphy – Opera]] nhé": "Xem Morphy – Opera nhé",
		"Nhúng ![[puzzle/chieu-bi|Chiếu bí]] ở đây":    "Nhúng Chiếu bí ở đây",
		"Không nhãn [[game/paul-morphy-1858]]":         "Không nhãn paul morphy 1858",
		"Wiki thường [[entity/acme]]":                  "Wiki thường entity/acme", // không phải ref cờ → giữ tiền tố, đổi gạch nối
		"Không có link":                                "Không có link",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			if got := expandChessWikilinks(in); got != want {
				t.Errorf("expandChessWikilinks(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestBuildGameKnowledgeText(t *testing.T) {
	g := &types.ChessGame{
		White: "Paul Morphy", Black: "Duke Karl", Result: "1-0",
		ECO: "C41", Event: "Paris Opera", Date: "1858.01.01", PGN: "1. e4 e5 2. Nf3", PlyCount: 3,
	}
	title, content := buildGameKnowledgeText(g)
	if !strings.Contains(title, "Paul Morphy") || !strings.Contains(title, "Duke Karl") {
		t.Errorf("title thiếu tên đấu thủ: %q", title)
	}
	for _, want := range []string{"Paul Morphy", "Duke Karl", "1-0", "C41", "Paris Opera", "1. e4 e5"} {
		if !strings.Contains(content, want) {
			t.Errorf("content thiếu %q\n---\n%s", want, content)
		}
	}
}

func TestBuildPositionKnowledgeText(t *testing.T) {
	p := &types.ChessPosition{
		Title: "Vua+Xe đấu Vua", FEN: "8/8/8/8/8/2k5/8/R2K4 w - - 0 1",
		Category: "endgame", Level: "hau", ECO: "", Assessment: "Trắng thắng",
		Annotation: "Xem thêm [[game/morphy-opera|ván Morphy]] để so sánh kỹ thuật.",
		Tags:       "tan-cuoc, xe",
	}
	title, content := buildPositionKnowledgeText(p)
	if title != "Thế cờ: Vua+Xe đấu Vua" {
		t.Errorf("title = %q, muốn \"Thế cờ: Vua+Xe đấu Vua\"", title)
	}
	for _, want := range []string{"endgame", "Hậu", "Trắng thắng", "8/8/8/8/8/2k5/8/R2K4", "ván Morphy", "tan-cuoc, xe"} {
		if !strings.Contains(content, want) {
			t.Errorf("content thiếu %q\n---\n%s", want, content)
		}
	}
	if strings.Contains(content, "[[") {
		t.Errorf("chú giải embed vẫn còn cú pháp wikilink thô:\n%s", content)
	}
}

func TestBuildPositionKnowledgeText_NoKingStillProducesValidText(t *testing.T) {
	// Thế cờ giản lược KHÔNG có quân Vua (dạy trẻ mới học) vẫn phải sinh được
	// văn bản tri thức bình thường — builder KHÔNG được gọi bất kỳ đường xử lý
	// nào đòi hỏi thế cờ hợp lệ theo luật (notnil/engine).
	p := &types.ChessPosition{
		Title: "Cách Tốt ăn quân", FEN: "8/8/8/3p4/4P3/8/8/8 w - - 0 1", Category: "basic", Level: "tot",
	}
	title, content := buildPositionKnowledgeText(p)
	if title != "Thế cờ: Cách Tốt ăn quân" {
		t.Errorf("title = %q", title)
	}
	if !strings.Contains(content, "3p4/4P3") {
		t.Errorf("content thiếu FEN:\n%s", content)
	}
}

func TestBuildLessonKnowledgeText_ExpandsWikilinks(t *testing.T) {
	l := &types.ChessLesson{
		Title:   "Bài 1",
		Content: "Phân tích [[game/morphy-opera|Morphy – Opera]] để học chiến thuật.",
	}
	_, content := buildLessonKnowledgeText(l)
	if strings.Contains(content, "[[") {
		t.Errorf("nội dung embed vẫn còn cú pháp wikilink thô:\n%s", content)
	}
	if !strings.Contains(content, "Morphy – Opera") {
		t.Errorf("nội dung embed mất nhãn wikilink:\n%s", content)
	}
}

func TestBuildArticleKnowledgeText(t *testing.T) {
	a := &types.ChessArticle{
		Title:    "Ghim (Pin) là gì?",
		Slug:     "ghim-pin-la-gi",
		Aliases:  "Pin, Đóng đinh",
		Summary:  "Chiến thuật khiến một quân không thể di chuyển.",
		Category: "thuat-ngu",
		Level:    "ma",
		Tags:     "chien-thuat, co-ban",
		Content:  "Xem thêm [[position/ghim-tuyet-doi|thế cờ ghim tuyệt đối]].",
	}
	title, content := buildArticleKnowledgeText(a)
	if title != "Bài viết: Ghim (Pin) là gì?" {
		t.Errorf("title = %q", title)
	}
	// Bí danh PHẢI vào văn bản index — đây là thứ giúp RAG bắt trúng khi học
	// viên hỏi bằng từ tiếng Anh ("pin") thay vì tiêu đề tiếng Việt.
	if !strings.Contains(content, "Pin, Đóng đinh") {
		t.Errorf("content thiếu bí danh:\n%s", content)
	}
	if !strings.Contains(content, "Chiến thuật khiến một quân") {
		t.Errorf("content thiếu tóm tắt:\n%s", content)
	}
	// Wikilink phải được expand thành nhãn, không để lọt cú pháp thô vào embedding.
	if strings.Contains(content, "[[") {
		t.Errorf("nội dung vẫn còn cú pháp wikilink thô:\n%s", content)
	}
	if !strings.Contains(content, "thế cờ ghim tuyệt đối") {
		t.Errorf("nội dung mất nhãn wikilink:\n%s", content)
	}
}

func TestBuildArticleKnowledgeText_FallsBackToSlugWhenNoTitle(t *testing.T) {
	a := &types.ChessArticle{Slug: "khai-niem-x", Content: "nội dung"}
	title, _ := buildArticleKnowledgeText(a)
	if title != "Bài viết: khai-niem-x" {
		t.Errorf("title phải rơi về slug khi thiếu tiêu đề, nhận %q", title)
	}
}
