package service

import "testing"

func TestResolveChessSlugFuzzy(t *testing.T) {
	live := []string{
		"paul-morphy-duke-karl-paris-opera-1858",
		"chieu-bi-hang-ngang",
		"bai-1-chip-vs-nhung",
	}
	cases := []struct {
		name     string
		dead     string
		wantSlug string
		wantOK   bool
	}{
		{"khớp chính xác", "chieu-bi-hang-ngang", "chieu-bi-hang-ngang", true},
		{"khác gạch nối", "chieubi-hangngang", "chieu-bi-hang-ngang", true},
		{"thiếu một gạch nối", "paul-morphy-duke-karl-paris-opera1858", "paul-morphy-duke-karl-paris-opera-1858", true},
		{"không liên quan", "khai-cuoc-sicilian", "", false},
		{"rỗng", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := resolveChessSlugFuzzy(c.dead, live)
			if ok != c.wantOK || got != c.wantSlug {
				t.Errorf("resolveChessSlugFuzzy(%q) = (%q, %v), want (%q, %v)",
					c.dead, got, ok, c.wantSlug, c.wantOK)
			}
		})
	}
}

func TestResolveChessSlugFuzzy_EmptyPool(t *testing.T) {
	if got, ok := resolveChessSlugFuzzy("anything", nil); ok || got != "" {
		t.Errorf("pool rỗng phải trả (\"\", false), nhận (%q, %v)", got, ok)
	}
}
