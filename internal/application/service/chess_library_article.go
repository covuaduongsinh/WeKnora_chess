package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_library_article.go bổ sung nghiệp vụ cho "Ngân hàng bài viết"
// (chess_articles) vào CÙNG struct chessLibraryService (khai báo ở
// chess_library.go) — article là "ngăn" thứ năm cạnh games/puzzles/positions/
// thư viện sách, không phải service riêng, để container.go/router.go không
// phải mở rộng.
//
// Ranh giới: bài viết là trang tri thức ĐỘC LẬP (khái niệm/thuật ngữ/kinh
// nghiệm) — KHÔNG thuộc sách/khóa học nào, là đích [[article/<slug>]] cho mọi
// thực thể khác trỏ tới, đồng thời tự chứa wikilink ra ván/thế cờ/bài tập/
// sách. Chỉ status="published" được index vào KB tri thức cờ (cùng quy tắc
// sách) — xem reindexArticle.
//
// PHASE 1: CRUD + wikilink 2 chiều + RAG. CHƯA có chuyên mục (chess_article_
// topics), bí danh ghi vào chess_slug_aliases (cột Aliases hiện chỉ lưu văn
// bản hiển thị), ảnh chèn bài, lịch sử phiên bản — xem backlog trong
// .claude/memory/04-nhat-ky-tuy-bien.md.

func (s *chessLibraryService) ListArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]*types.ChessArticle, error) {
	return s.repo.ListArticles(ctx, tenantID, f)
}

func (s *chessLibraryService) SearchArticles(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessArticle, error) {
	return s.repo.SearchArticles(ctx, tenantID, keyword, limit)
}

func (s *chessLibraryService) GetArticle(ctx context.Context, tenantID uint64, id string) (*types.ChessArticle, error) {
	return s.repo.GetArticle(ctx, tenantID, id)
}

// GetArticleBySlug giải mã wikilink [[article/<slug>]]: khớp chính xác trước,
// không có thì thử alias (đổi tên HOẶC bí danh — cả hai cùng nằm trong
// chess_slug_aliases) rồi fuzzy.
func (s *chessLibraryService) GetArticleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticle, error) {
	a, err := s.repo.GetArticleBySlug(ctx, tenantID, slug)
	if err == nil {
		return a, nil
	}
	if resolved, ok := s.resolveAliasOrFuzzy(ctx, tenantID, types.ChessRefTypeArticle, slug, s.repo.ArticleSlugs); ok {
		return s.repo.GetArticleBySlug(ctx, tenantID, resolved)
	}
	return nil, err
}

func (s *chessLibraryService) GetArticleBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error) {
	if s.chessRefRepo == nil {
		return nil, nil
	}
	return s.chessRefRepo.ListBacklinks(ctx, tenantID, types.ChessRefTypeArticle, slug)
}

func (s *chessLibraryService) CreateArticle(ctx context.Context, article *types.ChessArticle) (*types.ChessArticle, error) {
	if strings.TrimSpace(article.Title) == "" {
		return nil, fmt.Errorf("tiêu đề bài viết không được để trống")
	}
	if article.Status == "" {
		article.Status = types.ChessArticleStatusDraft
	}
	article.ID = uuid.New().String()
	slug, err := ensureUniqueChessSlug(ctx, article.TenantID, articleSlugBase(article), article.ID, s.repo.ArticleSlugExists)
	if err != nil {
		return nil, err
	}
	article.Slug = slug
	if err := s.repo.CreateArticle(ctx, article); err != nil {
		return nil, err
	}
	article.Tags = s.applyChessTags(ctx, article.TenantID, types.ChessRefTypeArticle, article.ID, article.Tags)
	s.syncArticleChessRefs(ctx, article)
	s.syncArticleAliases(ctx, article)
	s.reindexArticle(ctx, article)
	return article, nil
}

// UpdateArticle cập nhật bài viết; nếu Title/Content thực sự đổi thì lưu bản
// TRƯỚC khi ghi đè vào lịch sử phiên bản (revisionNote là ghi chú thay đổi,
// tùy chọn — cùng khuôn UpdateChapter, tránh rác lịch sử khi chỉ sửa các
// trường phân loại như category/level/status/tags).
func (s *chessLibraryService) UpdateArticle(ctx context.Context, article *types.ChessArticle, revisionNote string) (*types.ChessArticle, error) {
	old, err := s.repo.GetArticle(ctx, article.TenantID, article.ID)
	if err != nil {
		return nil, err
	}
	if article.Status == "" {
		article.Status = types.ChessArticleStatusDraft
	}
	if old.Title != article.Title || old.Content != article.Content {
		next, _ := s.repo.CountArticleRevisions(ctx, old.TenantID, old.ID)
		_ = s.repo.CreateArticleRevision(ctx, &types.ChessArticleRevision{
			ID: uuid.New().String(), TenantID: old.TenantID, ArticleID: old.ID,
			RevisionNumber: int(next) + 1, Title: old.Title, Content: old.Content,
			Summary: revisionNote, CreatedBy: userIDOrEmpty(ctx),
		})
	}
	if err := s.repo.UpdateArticle(ctx, article); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetArticle(ctx, article.TenantID, article.ID)
	if err != nil || updated == nil {
		return updated, err
	}
	updated.Tags = s.applyChessTags(ctx, updated.TenantID, types.ChessRefTypeArticle, updated.ID, updated.Tags)
	s.syncArticleChessRefs(ctx, updated)
	s.syncArticleAliases(ctx, updated)
	s.reindexArticle(ctx, updated)
	return updated, nil
}

// reindexArticle đồng bộ lại KB tri thức cờ sau khi Status của bài có thể đã
// đổi: published → index; draft (kể cả vừa chuyển từ published) → GỠ khỏi KB.
// Đối xứng bắt buộc với reindexBookAndChapters — để IndexArticle tự trả nil
// khi draft là SAI, vì bài đã publish rồi hạ xuống draft sẽ để lại tri thức
// cũ trong KB (agent vẫn trích dẫn được nội dung chưa duyệt).
func (s *chessLibraryService) reindexArticle(ctx context.Context, a *types.ChessArticle) {
	if s.indexer == nil || a == nil {
		return
	}
	if a.Status == types.ChessArticleStatusPublished {
		_ = s.indexer.IndexArticle(ctx, a)
		return
	}
	s.indexer.Remove(ctx, a.TenantID, types.ChessRefTypeArticle, a.Slug)
}

// RenameArticleSlug đổi slug bài viết sang newSlug + ghi alias slug-cũ→mới
// (kind="rename", giữ vĩnh viễn) để wikilink [[article/<cũ>]] vẫn giải mã đúng.
func (s *chessLibraryService) RenameArticleSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessArticle, error) {
	a, err := s.repo.GetArticle(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	base := slugifyChess(newSlug)
	if base == "" {
		return nil, fmt.Errorf("slug mới không hợp lệ (cần chứa chữ/số)")
	}
	if base == a.Slug {
		return a, nil
	}
	unique, err := ensureUniqueChessSlug(ctx, tenantID, base, a.ID, s.repo.ArticleSlugExists)
	if err != nil {
		return nil, err
	}
	oldSlug := a.Slug
	if err := s.repo.UpdateArticleSlug(ctx, tenantID, id, unique); err != nil {
		return nil, err
	}
	if s.aliasRepo != nil && oldSlug != "" {
		_ = s.aliasRepo.AddAlias(ctx, tenantID, types.ChessRefTypeArticle, oldSlug, unique)
	}
	s.removeStaleIndex(ctx, tenantID, types.ChessRefTypeArticle, oldSlug)
	updated, err := s.repo.GetArticle(ctx, tenantID, id)
	if err != nil || updated == nil {
		return updated, err
	}
	// Chú giải/nội dung có thể tự tham chiếu slug cũ trong một số ngữ cảnh
	// hiển thị — đồng bộ lại nguồn ref dưới slug MỚI để backlink nhất quán.
	s.syncArticleChessRefs(ctx, updated)
	// Bí danh KHÔNG đổi khi đổi slug — nhưng target (new_slug trong bảng
	// alias) phải trỏ theo slug MỚI. ReplaceSynonyms dùng ON CONFLICT theo
	// old_slug (chuỗi bí danh không đổi) nên gọi lại đây TỰ SỬA new_slug của
	// các dòng hiện có từ slug cũ sang slug mới, không để mồ côi.
	s.syncArticleAliases(ctx, updated)
	s.reindexArticle(ctx, updated)
	return updated, nil
}

// DeleteArticle xóa bài viết VÀ cascade: ref TRỎ TỚI nó (backlink từ nơi
// khác) + ref TỪ nội dung của chính nó (nó từng trỏ đi đâu) + mọi alias
// (rename lẫn synonym) trỏ tới nó + pivot chuyên mục + ảnh chèn bài (cả bản
// ghi lẫn file vật lý) + gỡ khỏi KB tri thức cờ. (Lịch sử phiên bản thuộc
// Phase 3 — chưa có dữ liệu để dọn ở Phase 2.)
func (s *chessLibraryService) DeleteArticle(ctx context.Context, tenantID uint64, id string) error {
	a, _ := s.repo.GetArticle(ctx, tenantID, id)
	images, _ := s.repo.ListArticleImagesByArticle(ctx, tenantID, id)
	if err := s.repo.DeleteArticle(ctx, tenantID, id); err != nil {
		return err
	}
	_ = s.repo.RemoveArticleFromAllTopics(ctx, tenantID, id)
	_ = s.repo.DeleteArticleRevisionsByArticle(ctx, tenantID, id)
	s.removeChessTags(ctx, tenantID, types.ChessRefTypeArticle, id)
	if a != nil {
		s.pruneChessRefs(ctx, tenantID, types.ChessRefTypeArticle, a.Slug)
		if s.chessRefRepo != nil && a.Slug != "" {
			_ = s.chessRefRepo.DeleteForArticle(ctx, tenantID, a.Slug)
		}
		if s.aliasRepo != nil && a.Slug != "" {
			_ = s.aliasRepo.DeleteAliasesFor(ctx, tenantID, types.ChessRefTypeArticle, a.Slug)
		}
		if s.indexer != nil {
			s.indexer.Remove(ctx, tenantID, types.ChessRefTypeArticle, a.Slug)
		}
	}
	for _, img := range images {
		if s.fileService != nil && img.Path != "" {
			_ = s.fileService.DeleteFile(ctx, img.Path)
		}
		_ = s.repo.DeleteArticleImage(ctx, tenantID, img.ID)
	}
	return nil
}

// syncArticleChessRefs đồng bộ wiki_chess_refs theo nội dung của một bài viết
// (nguồn article) — TÁI DÙNG parseChessRefSlugs (chess_course.go, cùng
// package) như bài giảng/chương sách, để nội dung trở thành nguồn phát wikilink.
func (s *chessLibraryService) syncArticleChessRefs(ctx context.Context, a *types.ChessArticle) {
	if s.chessRefRepo == nil || a == nil || a.Slug == "" {
		return
	}
	refs := parseChessRefSlugs(a.Content)
	for i := range refs {
		refs[i].ID = uuid.New().String()
		refs[i].TenantID = a.TenantID
		refs[i].SourceType = types.ChessRefSourceArticle
		refs[i].PageSlug = a.Slug
	}
	_ = s.chessRefRepo.ReplaceForArticle(ctx, a.TenantID, a.Slug, refs)
}

// syncArticleAliases đồng bộ bí danh/từ đồng nghĩa (cột Aliases, CSV hiển
// thị người dùng gõ tay) vào chess_slug_aliases (kind="synonym") để
// [[article/<bí danh>]] giải mã đúng bài + autocomplete/RAG bắt trúng khi hỏi
// bằng từ đồng nghĩa. Bỏ qua (KHÔNG lưu) bí danh trùng slug THẬT của MỘT BÀI
// KHÁC đang tồn tại — thứ tự resolve là exact→alias nên bí danh trùng slug
// thật sẽ che khuất bài kia VĨNH VIỄN nếu lưu vào.
func (s *chessLibraryService) syncArticleAliases(ctx context.Context, a *types.ChessArticle) {
	if s.aliasRepo == nil || a == nil || a.Slug == "" {
		return
	}
	seen := make(map[string]bool)
	candidates := make([]string, 0, 4)
	for _, raw := range strings.Split(a.Aliases, ",") {
		slug := slugifyChess(raw)
		if slug == "" || slug == a.Slug || seen[slug] {
			continue
		}
		seen[slug] = true
		if exists, _ := s.repo.ArticleSlugExists(ctx, a.TenantID, slug); exists {
			continue // trùng slug thật của bài khác — bỏ qua, không che khuất
		}
		candidates = append(candidates, slug)
	}
	_ = s.aliasRepo.ReplaceSynonyms(ctx, a.TenantID, types.ChessRefTypeArticle, a.Slug, candidates)
}

// ---- Lịch sử phiên bản bài viết ----

func (s *chessLibraryService) ListArticleRevisions(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleRevision, error) {
	return s.repo.ListArticleRevisions(ctx, tenantID, articleID)
}

func (s *chessLibraryService) GetArticleRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessArticleRevision, error) {
	return s.repo.GetArticleRevision(ctx, tenantID, revisionID)
}

// RestoreArticleRevision ghi nội dung một bản cũ trở lại làm bản hiện tại.
// Tái dùng UpdateArticle để đi đúng luồng: bản ĐANG CÓ (trước khi khôi phục)
// cũng được lưu thành một phiên bản mới, đồng bộ ref/bí danh/index như sửa thường.
func (s *chessLibraryService) RestoreArticleRevision(ctx context.Context, tenantID uint64, articleID, revisionID string) (*types.ChessArticle, error) {
	rev, err := s.repo.GetArticleRevision(ctx, tenantID, revisionID)
	if err != nil {
		return nil, err
	}
	if rev.ArticleID != articleID {
		return nil, fmt.Errorf("bản phiên bản không thuộc bài viết này")
	}
	a, err := s.repo.GetArticle(ctx, tenantID, articleID)
	if err != nil {
		return nil, err
	}
	a.Title = rev.Title
	a.Content = rev.Content
	return s.UpdateArticle(ctx, a, fmt.Sprintf("Khôi phục từ bản #%d", rev.RevisionNumber))
}

// ---- Ảnh chèn trong bài viết ----

// UploadArticleImage lưu file qua FileService rồi ghi bản ghi ChessArticleImage
// (sao y UploadBookImage, book_id → article_id).
func (s *chessLibraryService) UploadArticleImage(ctx context.Context, tenantID uint64, articleID, fileName, mime string, data []byte) (*types.ChessArticleImage, error) {
	if s.fileService == nil {
		return nil, fmt.Errorf("dịch vụ lưu file chưa sẵn sàng")
	}
	if _, err := s.repo.GetArticle(ctx, tenantID, articleID); err != nil {
		return nil, fmt.Errorf("bài viết không tồn tại")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("file ảnh rỗng")
	}
	if len(data) > maxBookImageSize {
		return nil, fmt.Errorf("ảnh vượt quá %dMB", maxBookImageSize>>20)
	}
	ext, ok := bookImageExtByMime[strings.ToLower(mime)]
	if !ok {
		return nil, fmt.Errorf("định dạng ảnh không hỗ trợ (chỉ png/jpeg/gif/webp)")
	}
	id := uuid.New().String()
	storedName := fmt.Sprintf("article-images/%s%s", id, ext)
	path, err := s.fileService.SaveBytes(ctx, data, tenantID, storedName, false)
	if err != nil {
		return nil, fmt.Errorf("lưu ảnh thất bại: %w", err)
	}
	img := &types.ChessArticleImage{
		ID: id, TenantID: tenantID, ArticleID: articleID, Path: path,
		FileName: fileName, Mime: mime, Size: int64(len(data)),
	}
	if err := s.repo.CreateArticleImage(ctx, img); err != nil {
		return nil, err
	}
	return img, nil
}

// GetArticleImage trả metadata + luồng đọc nội dung ảnh (caller PHẢI Close()).
func (s *chessLibraryService) GetArticleImage(ctx context.Context, tenantID uint64, imageID string) (*types.ChessArticleImage, io.ReadCloser, error) {
	img, err := s.repo.GetArticleImage(ctx, tenantID, imageID)
	if err != nil {
		return nil, nil, err
	}
	if s.fileService == nil {
		return nil, nil, fmt.Errorf("dịch vụ lưu file chưa sẵn sàng")
	}
	rc, err := s.fileService.GetFile(ctx, img.Path)
	if err != nil {
		return nil, nil, err
	}
	return img, rc, nil
}

// ---- Export / Import ----

// ExportArticles xuất các bài viết (theo filter) để sao lưu/chia sẻ.
func (s *chessLibraryService) ExportArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]types.ChessArticleBundle, error) {
	articles, err := s.repo.ListArticles(ctx, tenantID, f)
	if err != nil {
		return nil, err
	}
	out := make([]types.ChessArticleBundle, 0, len(articles))
	for _, a := range articles {
		out = append(out, types.ChessArticleBundle{
			Title: a.Title, Summary: a.Summary, Aliases: a.Aliases, Category: a.Category,
			Level: a.Level, Tags: a.Tags, Status: a.Status, CoverURL: a.CoverURL, Content: a.Content,
		})
	}
	return out, nil
}

// ImportArticles nhập danh sách bài viết, LUÔN tạo mới trong tenant hiện tại
// (tái dùng CreateArticle để sinh ID/slug/đồng bộ ref/index). Trả số bài đã
// thêm; bỏ qua bản lỗi để 1 bản xấu không chặn cả lô.
func (s *chessLibraryService) ImportArticles(ctx context.Context, tenantID uint64, items []types.ChessArticleBundle) (int, error) {
	created := 0
	for i := range items {
		it := items[i]
		if strings.TrimSpace(it.Title) == "" {
			continue
		}
		_, err := s.CreateArticle(ctx, &types.ChessArticle{
			TenantID: tenantID, Title: it.Title, Summary: it.Summary, Aliases: it.Aliases,
			Category: it.Category, Level: it.Level, Tags: it.Tags, Status: it.Status,
			CoverURL: it.CoverURL, Content: it.Content,
		})
		if err != nil {
			continue
		}
		created++
	}
	if created == 0 && len(items) > 0 {
		return 0, fmt.Errorf("không nhập được bài viết nào (kiểm tra định dạng JSON)")
	}
	return created, nil
}
