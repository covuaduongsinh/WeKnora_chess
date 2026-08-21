package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/chess"
	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

// chess_search_text.go dựng cột `search_text` — bản KHỬ DẤU + hạ chữ thường
// của các trường tìm kiếm được của mỗi loại nội dung cờ (migration 000912).
//
// VÌ SAO Ở TẦNG REPOSITORY chứ không phải service: đây là nơi DUY NHẤT mọi
// đường ghi đều đi qua. Đặt ở service thì mỗi lần thêm một lối tạo/sửa mới
// (import hàng loạt, khôi phục phiên bản, seed script...) lại phải nhớ gọi
// thêm — và quên là lỗi CÂM: bản ghi vẫn lưu, chỉ là tìm không ra.
//
// Toàn bộ đi qua chess.SearchText, CÙNG phép khử dấu với slugifyChess. Lệch
// nhau là gõ "khai cuoc" khớp thẻ nhưng trượt tiêu đề.

func gameSearchText(g *types.ChessGame) string {
	return chess.SearchText(g.Slug, g.White, g.Black, g.Event, g.ECO, g.Result, g.Date)
}

func puzzleSearchText(p *types.ChessPuzzle) string {
	return chess.SearchText(p.Slug, p.Title, p.Theme, p.Difficulty, p.Solution, p.Source, p.Level)
}

func positionSearchText(p *types.ChessPosition) string {
	return chess.SearchText(p.Slug, p.Title, p.Category, p.ECO, p.Tags, p.Assessment, p.Source, p.Annotation)
}

func bookSearchText(b *types.ChessBook) string {
	return chess.SearchText(b.Slug, b.Title, b.Subtitle, b.Author, b.Translator,
		b.Publisher, b.Year, b.ISBN, b.Tags, b.ECO, b.Description)
}

func chapterSearchText(ch *types.ChessBookChapter) string {
	return chess.SearchText(ch.Slug, ch.Title, ch.Part, ch.Content)
}

func articleSearchText(a *types.ChessArticle) string {
	return chess.SearchText(a.Slug, a.Title, a.Aliases, a.Summary, a.Tags, a.Category, a.Content)
}

func courseSearchText(c *types.ChessCourse) string {
	return chess.SearchText(c.Slug, c.Title, c.Description)
}

func lessonSearchText(l *types.ChessLesson) string {
	return chess.SearchText(l.Slug, l.Title, l.Content)
}

// applyChessKeyword thêm mệnh đề tìm theo từ khóa trên cột search_text.
//
// Dùng LIKE chứ KHÔNG phải ILIKE: cả hai vế đã được hạ chữ thường và khử dấu
// sẵn, nên ILIKE chỉ tốn thêm công mà không thêm gì. Quan trọng hơn, LIKE trên
// cột trần là dạng mà index GIN trigram (idx_chess_*_search) khớp được —
// đúng cái bẫy đã làm hai index trigram cũ thành vô dụng.
//
// column phải ĐỦ TÊN BẢNG ở những truy vấn có JOIN sẵn.
func applyChessKeyword(q *gorm.DB, column, keyword string) *gorm.DB {
	needle := chess.SearchNeedle(keyword)
	if needle == "" {
		return q
	}
	return q.Where(column+" LIKE ?", "%"+needle+"%")
}

// ---- Làm mới search_text sau khi ĐỔI SLUG ----
//
// Slug là một phần của search_text, nên đổi slug mà không tính lại sẽ để cột
// mang slug CŨ — tìm theo slug mới sẽ trượt. Lỗi câm: bản ghi vẫn đúng, chỉ là
// tìm không ra. Các hàm dưới đọc lại bản ghi (slug đã mới) rồi ghi đè cột.
//
// Best-effort: lỗi ở đây KHÔNG được làm hỏng thao tác đổi slug vốn đã thành
// công — cùng lắm là search_text lệch tới lần sửa/backfill kế tiếp.

func (r *chessLibraryRepository) refreshSearchText(ctx context.Context, model interface{}, tenantID uint64, id, text string) {
	_ = r.db.WithContext(ctx).Model(model).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("search_text", text).Error
}

func (r *chessLibraryRepository) refreshGameSearchText(ctx context.Context, tenantID uint64, id string) {
	if g, err := r.GetGame(ctx, tenantID, id); err == nil && g != nil {
		r.refreshSearchText(ctx, &types.ChessGame{}, tenantID, id, gameSearchText(g))
	}
}

func (r *chessLibraryRepository) refreshPuzzleSearchText(ctx context.Context, tenantID uint64, id string) {
	if p, err := r.GetPuzzle(ctx, tenantID, id); err == nil && p != nil {
		r.refreshSearchText(ctx, &types.ChessPuzzle{}, tenantID, id, puzzleSearchText(p))
	}
}

func (r *chessLibraryRepository) refreshPositionSearchText(ctx context.Context, tenantID uint64, id string) {
	if p, err := r.GetPosition(ctx, tenantID, id); err == nil && p != nil {
		r.refreshSearchText(ctx, &types.ChessPosition{}, tenantID, id, positionSearchText(p))
	}
}

func (r *chessLibraryRepository) refreshBookSearchText(ctx context.Context, tenantID uint64, id string) {
	if b, err := r.GetBook(ctx, tenantID, id); err == nil && b != nil {
		r.refreshSearchText(ctx, &types.ChessBook{}, tenantID, id, bookSearchText(b))
	}
}

func (r *chessLibraryRepository) refreshChapterSearchText(ctx context.Context, tenantID uint64, id string) {
	if ch, err := r.GetChapter(ctx, tenantID, id); err == nil && ch != nil {
		r.refreshSearchText(ctx, &types.ChessBookChapter{}, tenantID, id, chapterSearchText(ch))
	}
}

func (r *chessLibraryRepository) refreshArticleSearchText(ctx context.Context, tenantID uint64, id string) {
	if a, err := r.GetArticle(ctx, tenantID, id); err == nil && a != nil {
		r.refreshSearchText(ctx, &types.ChessArticle{}, tenantID, id, articleSearchText(a))
	}
}

// BackfillSearchText tính lại cột search_text cho TOÀN BỘ nội dung cờ của một
// tenant. Dùng sau migration 000912 (cột mới, các bản ghi cũ đang rỗng nên tìm
// không ra) và bất cứ khi nào nghi cột bị lệch.
//
// Trả số bản ghi đã ghi theo từng loại. Idempotent — chạy lại cho cùng kết quả.
//
// GHI CHÚ RANH GIỚI: hàm này chạm cả chess_courses/chess_lessons vốn thuộc
// chessCourseRepository. Cố ý: đây là đường BẢO TRÌ chạy một lần, và thêm một
// API bulk vào repository thứ hai chỉ để phục vụ nó sẽ nhân đôi bề mặt cho
// đúng một lời gọi. Mọi đường ghi THÔNG THƯỜNG vẫn tôn trọng ranh giới.
func (r *chessLibraryRepository) BackfillSearchText(ctx context.Context, tenantID uint64) (map[string]int, error) {
	out := map[string]int{}

	var games []*types.ChessGame
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&games).Error; err != nil {
		return out, err
	}
	for _, g := range games {
		r.refreshSearchText(ctx, &types.ChessGame{}, tenantID, g.ID, gameSearchText(g))
		out[types.ChessRefTypeGame]++
	}

	var puzzles []*types.ChessPuzzle
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&puzzles).Error; err != nil {
		return out, err
	}
	for _, p := range puzzles {
		r.refreshSearchText(ctx, &types.ChessPuzzle{}, tenantID, p.ID, puzzleSearchText(p))
		out[types.ChessRefTypePuzzle]++
	}

	var positions []*types.ChessPosition
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&positions).Error; err != nil {
		return out, err
	}
	for _, p := range positions {
		r.refreshSearchText(ctx, &types.ChessPosition{}, tenantID, p.ID, positionSearchText(p))
		out[types.ChessRefTypePosition]++
	}

	var books []*types.ChessBook
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&books).Error; err != nil {
		return out, err
	}
	for _, b := range books {
		r.refreshSearchText(ctx, &types.ChessBook{}, tenantID, b.ID, bookSearchText(b))
		out[types.ChessRefTypeBook]++
	}

	var chapters []*types.ChessBookChapter
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&chapters).Error; err != nil {
		return out, err
	}
	for _, ch := range chapters {
		r.refreshSearchText(ctx, &types.ChessBookChapter{}, tenantID, ch.ID, chapterSearchText(ch))
		out[types.ChessRefTypeChapter]++
	}

	var articles []*types.ChessArticle
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&articles).Error; err != nil {
		return out, err
	}
	for _, a := range articles {
		r.refreshSearchText(ctx, &types.ChessArticle{}, tenantID, a.ID, articleSearchText(a))
		out[types.ChessRefTypeArticle]++
	}

	var courses []*types.ChessCourse
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&courses).Error; err != nil {
		return out, err
	}
	for _, c := range courses {
		r.refreshSearchText(ctx, &types.ChessCourse{}, tenantID, c.ID, courseSearchText(c))
		out[types.ChessRefTypeCourse]++
	}

	var lessons []*types.ChessLesson
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&lessons).Error; err != nil {
		return out, err
	}
	for _, l := range lessons {
		r.refreshSearchText(ctx, &types.ChessLesson{}, tenantID, l.ID, lessonSearchText(l))
		out[types.ChessRefTypeLesson]++
	}

	return out, nil
}
