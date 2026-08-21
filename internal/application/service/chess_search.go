package service

import (
	"context"
	"sort"
	"strings"

	"github.com/Tencent/WeKnora/internal/chess"
	"github.com/Tencent/WeKnora/internal/types"
)

// chess_search.go triển khai TÌM KIẾM HỢP NHẤT: một từ khóa, kết quả của cả 8
// loại nội dung cờ, xếp hạng chung.
//
// CÁCH LÀM: quét từng bảng bằng chính các hàm List*/Search* đã có (nên không
// phát sinh SQL mới và mọi bộ lọc sẵn có vẫn đúng), rồi chấm điểm + trộn +
// phân trang trong Go.
//
// VÌ SAO KHÔNG UNION TRONG SQL: 8 bảng có cấu trúc khác nhau, một câu UNION có
// ts_rank sẽ khó đọc, khó sửa, và KHÔNG chạy trên bản SQLite ("lite"). Cách
// này đổi một chút hiệu năng lấy tính đúng đắn kiểm chứng được — với quy mô
// một tenant giáo dục thì quét trần mỗi loại vài trăm dòng là thoải mái.

// chessSearchPerType là trần quét MỖI LOẠI trước khi chấm điểm. Chạm trần thì
// kết quả được đánh dấu Truncated để giao diện nói thật với người dùng, thay
// vì im lặng như trần Limit(500) cũ.
const chessSearchPerType = 200

// SearchChessAll tìm trên mọi loại nội dung cờ và trả về một trang đã xếp hạng.
func (s *chessLibraryService) SearchChessAll(
	ctx context.Context, tenantID uint64, q types.ChessSearchQuery,
) (*types.ChessSearchPage, error) {
	needle := chess.SearchNeedle(q.Keyword)
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	res := &types.ChessSearchPage{Page: page, PageSize: pageSize, ByType: map[string]int{}}
	if needle == "" && !q.Tags.Active() && q.Level == "" {
		// Không có tiêu chí nào: trả rỗng thay vì đổ cả kho ra.
		res.Items = []types.ChessSearchHit{}
		return res, nil
	}

	want := map[string]bool{}
	for _, t := range q.Types {
		if t = strings.TrimSpace(t); t != "" {
			want[t] = true
		}
	}
	enabled := func(t string) bool { return len(want) == 0 || want[t] }

	var hits []types.ChessSearchHit
	add := func(h types.ChessSearchHit, score int) {
		// Lọc THUẦN theo thẻ/cấp độ (không có từ khóa): mọi mục đã qua được bộ
		// lọc ở tầng SQL đều là kết quả hợp lệ. Không có điểm nền thì
		// ScoreSearchHit trả 0 cho từ khóa rỗng và trang kết quả rỗng trơn —
		// đúng lúc người dùng bấm một thẻ để duyệt.
		if needle == "" {
			score = 1
		}
		if score <= 0 {
			return
		}
		h.Score = score
		hits = append(hits, h)
		res.ByType[h.ChessType]++
	}
	// mark ghi nhận một loại đã chạm trần quét.
	mark := func(n int) {
		if n >= chessSearchPerType {
			res.Truncated = true
		}
	}

	if enabled(types.ChessRefTypeGame) {
		items, _ := s.repo.ListGames(ctx, tenantID, types.ChessGameFilter{
			Keyword: q.Keyword, Level: q.Level, Tags: q.Tags, Page: 1, PageSize: chessSearchPerType,
		})
		mark(len(items))
		for _, g := range items {
			title := strings.TrimSpace(g.White + " – " + g.Black)
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeGame, ChessID: g.ID, Slug: g.Slug, Title: title,
				Subtitle: strings.Join(nonEmpty(g.Event, g.ECO, g.Result), " · "),
				Level:    g.Level, UpdatedAt: g.UpdatedAt,
			}, chess.ScoreSearchHit(needle, g.Slug, chess.SearchNeedle(title), g.SearchText))
		}
	}

	if enabled(types.ChessRefTypePuzzle) {
		items, _ := s.repo.ListPuzzles(ctx, tenantID, types.ChessPuzzleFilter{
			Keyword: q.Keyword, Level: q.Level, Tags: q.Tags, Page: 1, PageSize: chessSearchPerType,
		})
		mark(len(items))
		for _, p := range items {
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypePuzzle, ChessID: p.ID, Slug: p.Slug, Title: p.Title,
				Subtitle: strings.Join(nonEmpty(p.Theme, p.Difficulty), " · "),
				Level:    p.Level, UpdatedAt: p.UpdatedAt,
			}, chess.ScoreSearchHit(needle, p.Slug, chess.SearchNeedle(p.Title), p.SearchText))
		}
	}

	if enabled(types.ChessRefTypePosition) {
		items, _ := s.repo.ListPositions(ctx, tenantID, types.ChessPositionFilter{
			Keyword: q.Keyword, Level: q.Level, Tags: q.Tags, Page: 1, PageSize: chessSearchPerType,
		})
		mark(len(items))
		for _, p := range items {
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypePosition, ChessID: p.ID, Slug: p.Slug, Title: p.Title,
				Subtitle: strings.Join(nonEmpty(p.Category, p.ECO), " · "),
				Snippet:  chess.Snippet(p.Annotation, needle, 160),
				Level:    p.Level, UpdatedAt: p.UpdatedAt,
			}, chess.ScoreSearchHit(needle, p.Slug, chess.SearchNeedle(p.Title), p.SearchText))
		}
	}

	if enabled(types.ChessRefTypeBook) {
		items, _ := s.repo.ListBooks(ctx, tenantID, types.ChessBookFilter{
			Keyword: q.Keyword, Level: q.Level, Status: q.Status, Tags: q.Tags,
			Page: 1, PageSize: chessSearchPerType,
		})
		mark(len(items))
		for _, b := range items {
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeBook, ChessID: b.ID, Slug: b.Slug, Title: b.Title,
				Subtitle: strings.Join(nonEmpty(b.Author, b.Phase), " · "),
				Snippet:  chess.Snippet(b.Description, needle, 160),
				Level:    b.Level, Status: b.Status, UpdatedAt: b.UpdatedAt,
			}, chess.ScoreSearchHit(needle, b.Slug, chess.SearchNeedle(b.Title), b.SearchText))
		}
	}

	if enabled(types.ChessRefTypeChapter) {
		// Chương không có filter riêng — dùng SearchChapters (đã có nhánh
		// search_text + full-text). Không lọc được theo thẻ ở tầng SQL nên lọc
		// sau khi lấy về, chấp nhận vì tập kết quả đã bị trần chặn.
		items, _ := s.repo.SearchChapters(ctx, tenantID, q.Keyword, chessSearchPerType)
		mark(len(items))
		keep := s.filterByTags(ctx, tenantID, types.ChessRefTypeChapter, q.Tags, chapterIDs(items))
		for _, ch := range items {
			if keep != nil && !keep[ch.ID] {
				continue
			}
			if q.Level != "" && ch.Level != q.Level {
				continue
			}
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeChapter, ChessID: ch.ID, Slug: ch.Slug, Title: ch.Title,
				Subtitle: ch.Part, Snippet: chess.Snippet(ch.Content, needle, 160),
				Level: ch.Level, UpdatedAt: ch.UpdatedAt,
			}, chess.ScoreSearchHit(needle, ch.Slug, chess.SearchNeedle(ch.Title), ch.SearchText))
		}
	}

	if enabled(types.ChessRefTypeArticle) {
		items, _ := s.repo.ListArticles(ctx, tenantID, types.ChessArticleFilter{
			Keyword: q.Keyword, Level: q.Level, Status: q.Status, Tags: q.Tags,
			Page: 1, PageSize: chessSearchPerType,
		})
		mark(len(items))
		for _, a := range items {
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeArticle, ChessID: a.ID, Slug: a.Slug, Title: a.Title,
				Subtitle: strings.Join(nonEmpty(a.Category, a.Summary), " · "),
				Snippet:  chess.Snippet(a.Content, needle, 160),
				Level:    a.Level, Status: a.Status, UpdatedAt: a.UpdatedAt,
			}, chess.ScoreSearchHit(needle, a.Slug, chess.SearchNeedle(a.Title), a.SearchText))
		}
	}

	if s.courseRepo != nil && enabled(types.ChessRefTypeLesson) {
		items, _ := s.courseRepo.SearchLessons(ctx, tenantID, q.Keyword, chessSearchPerType)
		mark(len(items))
		keep := s.filterByTags(ctx, tenantID, types.ChessRefTypeLesson, q.Tags, lessonIDs(items))
		for _, l := range items {
			if keep != nil && !keep[l.ID] {
				continue
			}
			if q.Level != "" && l.Level != q.Level {
				continue
			}
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeLesson, ChessID: l.ID, Slug: l.Slug, Title: l.Title,
				Snippet: chess.Snippet(l.Content, needle, 160),
				Level:   l.Level, UpdatedAt: l.UpdatedAt,
			}, chess.ScoreSearchHit(needle, l.Slug, chess.SearchNeedle(l.Title), l.SearchText))
		}
	}

	if s.courseRepo != nil && enabled(types.ChessRefTypeCourse) {
		// ListCourses không nhận bộ lọc nào — lọc trong Go. Số khóa học của
		// một tenant vốn nhỏ (hàng chục), nên không cần đường SQL riêng.
		items, _ := s.courseRepo.ListCourses(ctx, tenantID)
		keep := s.filterByTags(ctx, tenantID, types.ChessRefTypeCourse, q.Tags, courseIDs(items))
		for _, c := range items {
			if keep != nil && !keep[c.ID] {
				continue
			}
			if q.Level != "" && c.Level != q.Level {
				continue
			}
			add(types.ChessSearchHit{
				ChessType: types.ChessRefTypeCourse, ChessID: c.ID, Slug: c.Slug, Title: c.Title,
				Subtitle: c.Description, Level: c.Level, UpdatedAt: c.UpdatedAt,
			}, chess.ScoreSearchHit(needle, c.Slug, chess.SearchNeedle(c.Title), c.SearchText))
		}
	}

	// Xếp theo điểm giảm dần; đồng điểm thì mục sửa gần đây lên trước — khi
	// đang soạn dở, thứ vừa động tới gần như luôn là thứ đang tìm.
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].UpdatedAt.After(hits[j].UpdatedAt)
	})

	res.Total = len(hits)
	start := (page - 1) * pageSize
	if start > len(hits) {
		start = len(hits)
	}
	end := start + pageSize
	if end > len(hits) {
		end = len(hits)
	}
	res.Items = hits[start:end]
	if res.Items == nil {
		res.Items = []types.ChessSearchHit{}
	}
	s.attachSearchTags(ctx, tenantID, res.Items)
	return res, nil
}

// filterByTags trả tập id ĐƯỢC GIỮ khi lọc theo thẻ, hoặc nil nếu không lọc.
// Dùng cho các loại không lọc thẻ được ở tầng SQL (chương/bài giảng/khóa học).
func (s *chessLibraryService) filterByTags(
	ctx context.Context, tenantID uint64, chessType string, sel types.ChessTagSelector, ids []string,
) map[string]bool {
	if !sel.Active() || len(ids) == 0 {
		if !sel.Active() {
			return nil
		}
		return map[string]bool{}
	}
	byOwner, err := s.repo.TagsForMany(ctx, tenantID, chessType, ids)
	if err != nil {
		return map[string]bool{}
	}
	keep := map[string]bool{}
	for id, tags := range byOwner {
		have := map[string]bool{}
		for _, t := range tags {
			have[t.Slug] = true
		}
		matched := 0
		for _, want := range sel.TagSlugs {
			if have[want] {
				matched++
			}
		}
		if (sel.MatchAll() && matched == len(sel.TagSlugs)) || (!sel.MatchAll() && matched > 0) {
			keep[id] = true
		}
	}
	return keep
}

// attachSearchTags đính chip thẻ cho ĐÚNG trang đang hiển thị (không phải toàn
// bộ tập kết quả) — mỗi loại một truy vấn gộp, không N+1.
func (s *chessLibraryService) attachSearchTags(ctx context.Context, tenantID uint64, items []types.ChessSearchHit) {
	byType := map[string][]string{}
	for _, it := range items {
		byType[it.ChessType] = append(byType[it.ChessType], it.ChessID)
	}
	lookup := map[string]map[string][]*types.ChessTag{}
	for t, ids := range byType {
		if m, err := s.repo.TagsForMany(ctx, tenantID, t, ids); err == nil {
			lookup[t] = m
		}
	}
	for i := range items {
		if m, ok := lookup[items[i].ChessType]; ok {
			items[i].Tags = m[items[i].ChessID]
		}
	}
}

func chapterIDs(items []*types.ChessBookChapter) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.ID)
	}
	return out
}

func lessonIDs(items []*types.ChessLesson) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.ID)
	}
	return out
}

func courseIDs(items []*types.ChessCourse) []string {
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, x.ID)
	}
	return out
}
