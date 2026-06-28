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
