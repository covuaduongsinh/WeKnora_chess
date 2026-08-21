package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// --- Fake gọn: embed interface để chỉ override method kệ/sách/chương/phiên
// bản cần dùng — cùng kỹ thuật fakePositionRepo (chess_library_position_test.go).
// Alias repo TÁI DÙNG fakePositionAliasRepo (cùng package, không cần định nghĩa lại).

type fakeBookRepo struct {
	// fakeTagBase (chess_library_tag_test.go) mang phần lưu trữ HỆ THẺ và tự
	// embed interface repo. Phải embed nó CHỨ KHÔNG embed interface trực tiếp:
	// mọi đường Create/Update/Delete nay đều gọi repo thẻ, còn embed cả hai ở
	// cùng độ sâu sẽ làm method thẻ nhập nhằng.
	fakeTagBase
	shelves    map[string]*types.ChessShelf
	shelfBooks map[string]map[string]int // shelfID -> bookID -> sortOrder
	books      map[string]*types.ChessBook
	chapters   map[string]*types.ChessBookChapter
	revisions  map[string]*types.ChessChapterRevision
	revOrder   []string // id các revision theo thứ tự tạo (để List trả mới nhất trước)
	revCounter map[string]int
}

func newFakeBookRepo() *fakeBookRepo {
	return &fakeBookRepo{
		fakeTagBase: newFakeTagBase(),
		shelves:     map[string]*types.ChessShelf{}, shelfBooks: map[string]map[string]int{},
		books: map[string]*types.ChessBook{}, chapters: map[string]*types.ChessBookChapter{},
		revisions: map[string]*types.ChessChapterRevision{}, revCounter: map[string]int{},
	}
}

// ---- Kệ ----

func (r *fakeBookRepo) ListShelves(ctx context.Context, tenantID uint64, f types.ChessShelfFilter) ([]*types.ChessShelf, error) {
	out := make([]*types.ChessShelf, 0, len(r.shelves))
	for _, s := range r.shelves {
		out = append(out, s)
	}
	return out, nil
}
func (r *fakeBookRepo) GetShelf(ctx context.Context, tenantID uint64, id string) (*types.ChessShelf, error) {
	s, ok := r.shelves[id]
	if !ok {
		return nil, fmt.Errorf("shelf not found: %s", id)
	}
	return s, nil
}
func (r *fakeBookRepo) GetShelfBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessShelf, error) {
	for _, s := range r.shelves {
		if s.Slug == slug {
			return s, nil
		}
	}
	return nil, fmt.Errorf("shelf not found for slug: %s", slug)
}
func (r *fakeBookRepo) ShelfSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	out := make([]string, 0, len(r.shelves))
	for _, s := range r.shelves {
		if s.Slug != "" {
			out = append(out, s.Slug)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) ShelfSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, s := range r.shelves {
		if s.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeBookRepo) CreateShelf(ctx context.Context, s *types.ChessShelf) error {
	cp := *s
	r.shelves[s.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateShelf(ctx context.Context, s *types.ChessShelf) error {
	existing, ok := r.shelves[s.ID]
	if !ok {
		return fmt.Errorf("shelf not found: %s", s.ID)
	}
	cp := *s
	cp.Slug = existing.Slug
	r.shelves[s.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateShelfSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	s, ok := r.shelves[id]
	if !ok {
		return fmt.Errorf("shelf not found: %s", id)
	}
	s.Slug = slug
	return nil
}
func (r *fakeBookRepo) DeleteShelf(ctx context.Context, tenantID uint64, id string) error {
	delete(r.shelves, id)
	return nil
}
func (r *fakeBookRepo) CountBooksOnShelf(ctx context.Context, tenantID uint64, shelfID string) (int64, error) {
	return int64(len(r.shelfBooks[shelfID])), nil
}

// ---- Nối kệ↔sách ----

func (r *fakeBookRepo) SetShelfBooks(ctx context.Context, tenantID uint64, shelfID string, bookIDs []string) error {
	m := map[string]int{}
	for i, bid := range bookIDs {
		if bid == "" {
			continue
		}
		m[bid] = i
	}
	r.shelfBooks[shelfID] = m
	return nil
}
func (r *fakeBookRepo) ListShelvesOfBook(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessShelf, error) {
	var out []*types.ChessShelf
	for shelfID, books := range r.shelfBooks {
		if _, ok := books[bookID]; ok {
			if s, ok := r.shelves[shelfID]; ok {
				out = append(out, s)
			}
		}
	}
	return out, nil
}
func (r *fakeBookRepo) RemoveBookFromAllShelves(ctx context.Context, tenantID uint64, bookID string) error {
	for _, books := range r.shelfBooks {
		delete(books, bookID)
	}
	return nil
}
func (r *fakeBookRepo) RemoveShelfBooks(ctx context.Context, tenantID uint64, shelfID string) error {
	delete(r.shelfBooks, shelfID)
	return nil
}

// ---- Sách ----

func (r *fakeBookRepo) ListBooks(ctx context.Context, tenantID uint64, f types.ChessBookFilter) ([]*types.ChessBook, error) {
	out := make([]*types.ChessBook, 0, len(r.books))
	for _, b := range r.books {
		if f.ShelfID != "" {
			if _, ok := r.shelfBooks[f.ShelfID][b.ID]; !ok {
				continue
			}
		}
		if f.Status != "" && b.Status != f.Status {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}
func (r *fakeBookRepo) GetBook(ctx context.Context, tenantID uint64, id string) (*types.ChessBook, error) {
	b, ok := r.books[id]
	if !ok {
		return nil, fmt.Errorf("book not found: %s", id)
	}
	return b, nil
}
func (r *fakeBookRepo) GetBookBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBook, error) {
	for _, b := range r.books {
		if b.Slug == slug {
			return b, nil
		}
	}
	return nil, fmt.Errorf("book not found for slug: %s", slug)
}
func (r *fakeBookRepo) BookSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	out := make([]string, 0, len(r.books))
	for _, b := range r.books {
		if b.Slug != "" {
			out = append(out, b.Slug)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) BookSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, b := range r.books {
		if b.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeBookRepo) CreateBook(ctx context.Context, b *types.ChessBook) error {
	cp := *b
	r.books[b.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateBook(ctx context.Context, b *types.ChessBook) error {
	existing, ok := r.books[b.ID]
	if !ok {
		return fmt.Errorf("book not found: %s", b.ID)
	}
	cp := *b
	cp.Slug = existing.Slug // UpdateBook cố ý không đụng slug
	r.books[b.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateBookSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	b, ok := r.books[id]
	if !ok {
		return fmt.Errorf("book not found: %s", id)
	}
	b.Slug = slug
	return nil
}
func (r *fakeBookRepo) DeleteBook(ctx context.Context, tenantID uint64, id string) error {
	delete(r.books, id)
	return nil
}
func (r *fakeBookRepo) CountChapters(ctx context.Context, tenantID uint64, bookID string) (int64, error) {
	var n int64
	for _, ch := range r.chapters {
		if ch.BookID == bookID {
			n++
		}
	}
	return n, nil
}

// ---- Chương ----

func (r *fakeBookRepo) ListChapters(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessBookChapter, error) {
	var out []*types.ChessBookChapter
	for _, ch := range r.chapters {
		if ch.BookID == bookID {
			out = append(out, ch)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) SearchChapters(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessBookChapter, error) {
	var out []*types.ChessBookChapter
	for _, ch := range r.chapters {
		if keyword == "" || strings.Contains(strings.ToLower(ch.Title), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(ch.Content), strings.ToLower(keyword)) {
			out = append(out, ch)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) GetChapter(ctx context.Context, tenantID uint64, id string) (*types.ChessBookChapter, error) {
	ch, ok := r.chapters[id]
	if !ok {
		return nil, fmt.Errorf("chapter not found: %s", id)
	}
	return ch, nil
}
func (r *fakeBookRepo) GetChapterBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBookChapter, error) {
	for _, ch := range r.chapters {
		if ch.Slug == slug {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("chapter not found for slug: %s", slug)
}
func (r *fakeBookRepo) ChapterSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	out := make([]string, 0, len(r.chapters))
	for _, ch := range r.chapters {
		if ch.Slug != "" {
			out = append(out, ch.Slug)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) ChapterSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, ch := range r.chapters {
		if ch.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeBookRepo) CreateChapter(ctx context.Context, ch *types.ChessBookChapter) error {
	cp := *ch
	r.chapters[ch.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateChapter(ctx context.Context, ch *types.ChessBookChapter) error {
	existing, ok := r.chapters[ch.ID]
	if !ok {
		return fmt.Errorf("chapter not found: %s", ch.ID)
	}
	cp := *ch
	cp.Slug = existing.Slug     // UpdateChapter cố ý không đụng slug
	cp.BookID = existing.BookID // và không đụng book_id
	r.chapters[ch.ID] = &cp
	return nil
}
func (r *fakeBookRepo) UpdateChapterSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	ch, ok := r.chapters[id]
	if !ok {
		return fmt.Errorf("chapter not found: %s", id)
	}
	ch.Slug = slug
	return nil
}
func (r *fakeBookRepo) DeleteChapter(ctx context.Context, tenantID uint64, id string) error {
	delete(r.chapters, id)
	return nil
}
func (r *fakeBookRepo) DeleteChaptersByBook(ctx context.Context, tenantID uint64, bookID string) error {
	for id, ch := range r.chapters {
		if ch.BookID == bookID {
			delete(r.chapters, id)
		}
	}
	return nil
}
func (r *fakeBookRepo) ReorderChapters(ctx context.Context, tenantID uint64, bookID string, chapterIDs []string) error {
	for i, id := range chapterIDs {
		if ch, ok := r.chapters[id]; ok {
			ch.SortOrder = i
		}
	}
	return nil
}

// ---- Lịch sử phiên bản chương ----

func (r *fakeBookRepo) CreateChapterRevision(ctx context.Context, rev *types.ChessChapterRevision) error {
	cp := *rev
	r.revisions[rev.ID] = &cp
	r.revOrder = append(r.revOrder, rev.ID)
	return nil
}
func (r *fakeBookRepo) ListChapterRevisions(ctx context.Context, tenantID uint64, chapterID string) ([]*types.ChessChapterRevision, error) {
	var out []*types.ChessChapterRevision
	// Duyệt ngược revOrder → mới nhất trước, khớp ORDER BY revision_number DESC của repo thật.
	for i := len(r.revOrder) - 1; i >= 0; i-- {
		rev := r.revisions[r.revOrder[i]]
		if rev != nil && rev.ChapterID == chapterID {
			out = append(out, rev)
		}
	}
	return out, nil
}
func (r *fakeBookRepo) GetChapterRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessChapterRevision, error) {
	rev, ok := r.revisions[revisionID]
	if !ok {
		return nil, fmt.Errorf("revision not found: %s", revisionID)
	}
	return rev, nil
}
func (r *fakeBookRepo) CountChapterRevisions(ctx context.Context, tenantID uint64, chapterID string) (int64, error) {
	var n int64
	for _, rev := range r.revisions {
		if rev.ChapterID == chapterID {
			n++
		}
	}
	return n, nil
}
func (r *fakeBookRepo) DeleteChapterRevisionsByChapter(ctx context.Context, tenantID uint64, chapterID string) error {
	for id, rev := range r.revisions {
		if rev.ChapterID == chapterID {
			delete(r.revisions, id)
		}
	}
	return nil
}

// ---- Ảnh (không tested chi tiết — DeleteBook cascade vẫn gọi ListBookImagesByBook) ----

func (r *fakeBookRepo) ListBookImagesByBook(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessBookImage, error) {
	return nil, nil
}
func (r *fakeBookRepo) DeleteBookImage(ctx context.Context, tenantID uint64, id string) error {
	return nil
}

// ---- Ref repo giả (mẫu fakePositionChessRefRepo) ----

type fakeBookChessRefRepo struct {
	interfaces.WikiChessRefRepository
}

func (fakeBookChessRefRepo) ReplaceForChapter(ctx context.Context, tenantID uint64, chapterSlug string, refs []types.WikiChessRef) error {
	return nil
}
func (fakeBookChessRefRepo) DeleteForChapter(ctx context.Context, tenantID uint64, chapterSlug string) error {
	return nil
}
func (fakeBookChessRefRepo) DeleteForChess(ctx context.Context, tenantID uint64, chessType, chessSlug string) error {
	return nil
}
func (fakeBookChessRefRepo) ListBacklinks(ctx context.Context, tenantID uint64, chessType, chessSlug string) ([]types.ChessBacklink, error) {
	return nil, nil
}

func newTestBookService(repo *fakeBookRepo, alias *fakePositionAliasRepo) *chessLibraryService {
	return &chessLibraryService{
		repo:         repo,
		chessRefRepo: fakeBookChessRefRepo{},
		aliasRepo:    alias,
		indexer:      nil, // CHESS_KB_INDEX tắt trong test → indexer nil an toàn (mọi lời gọi indexer đều guard nil)
		fileService:  nil, // không test ảnh ở đây — mọi nhánh ảnh guard nil
	}
}

// ---- Sách: slug + status mặc định ----

func TestCreateBook_GeneratesSlugAndDefaultsDraftStatus(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	b, err := svc.CreateBook(context.Background(), &types.ChessBook{TenantID: 1, Title: "STEP Trainer Tập 1"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	if b.Slug == "" {
		t.Errorf("slug phải được sinh tự động")
	}
	if b.Status != types.ChessBookStatusDraft {
		t.Errorf("status mặc định = %q, muốn %q", b.Status, types.ChessBookStatusDraft)
	}
}

func TestCreateBook_SlugCollision_AppendsSuffix(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	ctx := context.Background()
	first, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách trùng tên"})
	if err != nil {
		t.Fatalf("tạo lần 1 lỗi: %v", err)
	}
	second, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách trùng tên"})
	if err != nil {
		t.Fatalf("tạo lần 2 lỗi: %v", err)
	}
	if first.Slug == second.Slug {
		t.Fatalf("2 sách trùng tiêu đề phải có slug KHÁC nhau, cả hai đều là %q", first.Slug)
	}
}

func TestGetBookBySlug_ResolvesViaAlias(t *testing.T) {
	repo := newFakeBookRepo()
	alias := &fakePositionAliasRepo{}
	svc := newTestBookService(repo, alias)
	ctx := context.Background()
	created, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách gốc"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	if err := alias.AddAlias(ctx, 1, types.ChessRefTypeBook, "slug-cu", created.Slug); err != nil {
		t.Fatalf("AddAlias lỗi: %v", err)
	}
	got, err := svc.GetBookBySlug(ctx, 1, "slug-cu")
	if err != nil {
		t.Fatalf("GetBookBySlug qua alias lỗi: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("alias trả nhầm bản ghi: got id=%s want id=%s", got.ID, created.ID)
	}
}

func TestRenameBookSlug_WritesAliasAndKeepsOldSlugResolvable(t *testing.T) {
	repo := newFakeBookRepo()
	alias := &fakePositionAliasRepo{}
	svc := newTestBookService(repo, alias)
	ctx := context.Background()
	created, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách đổi tên"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	oldSlug := created.Slug

	renamed, err := svc.RenameBookSlug(ctx, 1, created.ID, "sach-moi")
	if err != nil {
		t.Fatalf("RenameBookSlug lỗi: %v", err)
	}
	if renamed.Slug != "sach-moi" {
		t.Errorf("slug mới = %q, muốn \"sach-moi\"", renamed.Slug)
	}

	got, err := svc.GetBookBySlug(ctx, 1, oldSlug)
	if err != nil {
		t.Fatalf("GetBookBySlug với slug cũ (qua alias) lỗi: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("slug cũ trả nhầm bản ghi sau khi đổi tên")
	}
}

// ---- Chương: ràng buộc sách + slug + phiên bản ----

func TestCreateChapter_RequiresExistingBook(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	_, err := svc.CreateChapter(context.Background(), &types.ChessBookChapter{
		TenantID: 1, BookID: "khong-ton-tai", Title: "Chương 1",
	})
	if err == nil {
		t.Fatalf("muốn lỗi \"sách không tồn tại\" nhưng không có")
	}
}

func TestCreateChapter_SlugIncludesBookSlug(t *testing.T) {
	repo := newFakeBookRepo()
	svc := newTestBookService(repo, &fakePositionAliasRepo{})
	ctx := context.Background()
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Tàn cuộc cơ bản"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	ch, err := svc.CreateChapter(ctx, &types.ChessBookChapter{TenantID: 1, BookID: book.ID, Title: "Vua và Tốt"})
	if err != nil {
		t.Fatalf("CreateChapter lỗi: %v", err)
	}
	if ch.Slug == "" {
		t.Errorf("slug chương phải được sinh tự động")
	}
	// Slug chương duy nhất TOÀN TENANT (không scope theo sách) — mang tiền tố
	// slug sách để đọc được và ít đụng nhau (xem chapterSlugBase).
	if !strings.HasPrefix(ch.Slug, book.Slug) {
		t.Errorf("slug chương = %q, muốn có tiền tố slug sách %q", ch.Slug, book.Slug)
	}
}

func TestUpdateChapter_CreatesRevisionOnlyWhenContentOrTitleChanges(t *testing.T) {
	repo := newFakeBookRepo()
	svc := newTestBookService(repo, &fakePositionAliasRepo{})
	ctx := context.Background()
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách test"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	ch, err := svc.CreateChapter(ctx, &types.ChessBookChapter{
		TenantID: 1, BookID: book.ID, Title: "Chương 1", Content: "Nội dung A",
	})
	if err != nil {
		t.Fatalf("CreateChapter lỗi: %v", err)
	}

	// Chỉ đổi sort_order (không đổi Title/Content) → KHÔNG được tạo revision.
	ch.SortOrder = 5
	if _, err := svc.UpdateChapter(ctx, ch, ""); err != nil {
		t.Fatalf("UpdateChapter (sort_order) lỗi: %v", err)
	}
	if revs, _ := svc.ListChapterRevisions(ctx, 1, ch.ID); len(revs) != 0 {
		t.Fatalf("sửa sort_order không đổi content/title không được tạo revision, có %d", len(revs))
	}

	// Đổi Content → PHẢI tạo revision lưu bản TRƯỚC khi ghi đè (pre-image).
	ch.Content = "Nội dung B"
	if _, err := svc.UpdateChapter(ctx, ch, "sửa nội dung"); err != nil {
		t.Fatalf("UpdateChapter (content) lỗi: %v", err)
	}
	revs, _ := svc.ListChapterRevisions(ctx, 1, ch.ID)
	if len(revs) != 1 {
		t.Fatalf("muốn 1 revision sau khi đổi content, có %d", len(revs))
	}
	if revs[0].Content != "Nội dung A" {
		t.Errorf("revision phải lưu bản TRƯỚC khi ghi đè, got content=%q", revs[0].Content)
	}
	if revs[0].Summary != "sửa nội dung" {
		t.Errorf("summary = %q, muốn \"sửa nội dung\"", revs[0].Summary)
	}
}

func TestRestoreChapterRevision_RevertsContentAndCreatesNewRevision(t *testing.T) {
	repo := newFakeBookRepo()
	svc := newTestBookService(repo, &fakePositionAliasRepo{})
	ctx := context.Background()
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách test"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	ch, err := svc.CreateChapter(ctx, &types.ChessBookChapter{
		TenantID: 1, BookID: book.ID, Title: "Chương 1", Content: "Bản gốc",
	})
	if err != nil {
		t.Fatalf("CreateChapter lỗi: %v", err)
	}

	ch.Content = "Bản sửa"
	if _, err := svc.UpdateChapter(ctx, ch, ""); err != nil {
		t.Fatalf("UpdateChapter lỗi: %v", err)
	}

	revs, _ := svc.ListChapterRevisions(ctx, 1, ch.ID)
	if len(revs) != 1 {
		t.Fatalf("muốn 1 revision sau lần sửa đầu, có %d", len(revs))
	}
	revID := revs[0].ID

	restored, err := svc.RestoreChapterRevision(ctx, 1, ch.ID, revID)
	if err != nil {
		t.Fatalf("RestoreChapterRevision lỗi: %v", err)
	}
	if restored.Content != "Bản gốc" {
		t.Errorf("nội dung sau khôi phục = %q, muốn \"Bản gốc\"", restored.Content)
	}

	// Khôi phục CŨNG tạo thêm 1 revision (lưu bản "Bản sửa" trước khi khôi phục).
	revs2, _ := svc.ListChapterRevisions(ctx, 1, ch.ID)
	if len(revs2) != 2 {
		t.Fatalf("khôi phục phải tạo thêm 1 revision, có %d", len(revs2))
	}
}

// ---- Xóa sách: cascade ----

func TestDeleteBook_CascadesChaptersAndRevisions(t *testing.T) {
	repo := newFakeBookRepo()
	svc := newTestBookService(repo, &fakePositionAliasRepo{})
	ctx := context.Background()
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách xóa thử"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	ch, err := svc.CreateChapter(ctx, &types.ChessBookChapter{TenantID: 1, BookID: book.ID, Title: "Chương 1", Content: "A"})
	if err != nil {
		t.Fatalf("CreateChapter lỗi: %v", err)
	}
	ch.Content = "B"
	if _, err := svc.UpdateChapter(ctx, ch, ""); err != nil {
		t.Fatalf("UpdateChapter lỗi: %v", err)
	}

	if err := svc.DeleteBook(ctx, 1, book.ID); err != nil {
		t.Fatalf("DeleteBook lỗi: %v", err)
	}

	if _, err := svc.GetBook(ctx, 1, book.ID); err == nil {
		t.Errorf("sách vẫn còn sau khi xóa")
	}
	if _, err := svc.GetChapter(ctx, 1, ch.ID); err == nil {
		t.Errorf("chương vẫn còn sau khi xóa sách")
	}
	if revs, _ := repo.ListChapterRevisions(ctx, 1, ch.ID); len(revs) != 0 {
		t.Errorf("lịch sử phiên bản của chương phải bị xóa theo, còn %d", len(revs))
	}
}

// ---- Kệ: nhiều-nhiều ----

func TestSetShelfBooks_BookCanBeOnMultipleShelves(t *testing.T) {
	repo := newFakeBookRepo()
	svc := newTestBookService(repo, &fakePositionAliasRepo{})
	ctx := context.Background()
	shelfA, err := svc.CreateShelf(ctx, &types.ChessShelf{TenantID: 1, Title: "Kệ Tàn cuộc", Kind: "phase"})
	if err != nil {
		t.Fatalf("CreateShelf A lỗi: %v", err)
	}
	shelfB, err := svc.CreateShelf(ctx, &types.ChessShelf{TenantID: 1, Title: "Kệ Sicilian", Kind: "theme"})
	if err != nil {
		t.Fatalf("CreateShelf B lỗi: %v", err)
	}
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách chung 2 kệ"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}

	if err := svc.SetShelfBooks(ctx, 1, shelfA.ID, []string{book.ID}); err != nil {
		t.Fatalf("SetShelfBooks A lỗi: %v", err)
	}
	if err := svc.SetShelfBooks(ctx, 1, shelfB.ID, []string{book.ID}); err != nil {
		t.Fatalf("SetShelfBooks B lỗi: %v", err)
	}

	shelves, err := svc.ListShelvesOfBook(ctx, 1, book.ID)
	if err != nil {
		t.Fatalf("ListShelvesOfBook lỗi: %v", err)
	}
	if len(shelves) != 2 {
		t.Fatalf("sách phải nằm trên 2 kệ (nhiều-nhiều), có %d", len(shelves))
	}
}

// ---- Giữ trạng thái xuất bản khi PUT thiếu trường status ----

// Trước đây UpdateBook ép Status="" về draft. Vì repo ghi cột status TƯỜNG MINH
// (không phải partial update), mọi PUT thiếu trường này sẽ âm thầm hạ sách xuống
// bản thảo VÀ gỡ cả sách lẫn mọi chương khỏi KB tri thức. Giao diện có gửi status
// nên không lộ, nhưng script/curl thì dính.
func TestUpdateBook_EmptyStatusKeepsCurrentStatus(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	ctx := context.Background()
	created, err := svc.CreateBook(ctx, &types.ChessBook{
		TenantID: 1, Title: "Tàn cuộc cơ bản", Status: types.ChessBookStatusPublished,
	})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}

	// PUT không kèm status (client cũ / script quên trường này).
	updated, err := svc.UpdateBook(ctx, &types.ChessBook{
		TenantID: 1, ID: created.ID, Title: "Tàn cuộc cơ bản (bản sửa)",
	})
	if err != nil {
		t.Fatalf("UpdateBook lỗi: %v", err)
	}
	if updated.Status != types.ChessBookStatusPublished {
		t.Errorf("status = %q, muốn giữ %q — status rỗng nghĩa là 'không đổi', không phải 'hạ bản thảo'",
			updated.Status, types.ChessBookStatusPublished)
	}
	if updated.Title != "Tàn cuộc cơ bản (bản sửa)" {
		t.Errorf("tiêu đề chưa được cập nhật: %q", updated.Title)
	}
}

func TestUpdateBook_ExplicitStatusStillWins(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	ctx := context.Background()
	created, err := svc.CreateBook(ctx, &types.ChessBook{
		TenantID: 1, Title: "Sách hạ bản thảo", Status: types.ChessBookStatusPublished,
	})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	updated, err := svc.UpdateBook(ctx, &types.ChessBook{
		TenantID: 1, ID: created.ID, Title: created.Title, Status: types.ChessBookStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpdateBook lỗi: %v", err)
	}
	if updated.Status != types.ChessBookStatusDraft {
		t.Errorf("hạ bản thảo TƯỜNG MINH phải có tác dụng, status = %q", updated.Status)
	}
}

// ---- Sắp xếp lại chương ----

func TestReorderChapters_PersistsNewOrder(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	ctx := context.Background()
	book, err := svc.CreateBook(ctx, &types.ChessBook{TenantID: 1, Title: "Sách xếp lại"})
	if err != nil {
		t.Fatalf("CreateBook lỗi: %v", err)
	}
	var ids []string
	for _, title := range []string{"Chương A", "Chương B", "Chương C"} {
		ch, cerr := svc.CreateChapter(ctx, &types.ChessBookChapter{TenantID: 1, BookID: book.ID, Title: title})
		if cerr != nil {
			t.Fatalf("CreateChapter %q lỗi: %v", title, cerr)
		}
		ids = append(ids, ch.ID)
	}

	// Đảo ngược: C, B, A
	reversed := []string{ids[2], ids[1], ids[0]}
	if err := svc.ReorderChapters(ctx, 1, book.ID, reversed); err != nil {
		t.Fatalf("ReorderChapters lỗi: %v", err)
	}
	for want, id := range reversed {
		ch, gerr := svc.GetChapter(ctx, 1, id)
		if gerr != nil {
			t.Fatalf("GetChapter lỗi: %v", gerr)
		}
		if ch.SortOrder != want {
			t.Errorf("chương %q sort_order = %d, muốn %d", ch.Title, ch.SortOrder, want)
		}
	}
}

func TestReorderChapters_RejectsUnknownBook(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	if err := svc.ReorderChapters(context.Background(), 1, "khong-ton-tai", nil); err == nil {
		t.Fatal("phải báo lỗi khi sách không tồn tại")
	}
}

// ---- Nhập sách ----

// Chống hồi quy cho việc ImportBooks đổi sang gọi createChapter (bản lõi, không
// làm mới mục lục từng lần) thay vì CreateChapter.
func TestImportBooks_CreatesChaptersUnderNewBook(t *testing.T) {
	svc := newTestBookService(newFakeBookRepo(), &fakePositionAliasRepo{})
	ctx := context.Background()
	n, err := svc.ImportBooks(ctx, 1, []types.ChessBookBundle{{
		Title: "Sách nhập", Author: "Thầy Tường",
		Chapters: []types.ChessChapterBundle{
			{Title: "Mở đầu", Content: "nội dung 1"},
			{Title: "Kết", Content: "nội dung 2"},
			{Title: "   ", Content: "bỏ qua vì không có tiêu đề"},
		},
	}})
	if err != nil {
		t.Fatalf("ImportBooks lỗi: %v", err)
	}
	if n != 1 {
		t.Fatalf("số sách nhập = %d, muốn 1", n)
	}
	books, err := svc.ListBooks(ctx, 1, types.ChessBookFilter{})
	if err != nil || len(books) != 1 {
		t.Fatalf("phải có đúng 1 sách, có %d (err=%v)", len(books), err)
	}
	chapters, err := svc.ListChapters(ctx, 1, books[0].ID)
	if err != nil {
		t.Fatalf("ListChapters lỗi: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("phải có 2 chương (chương không tiêu đề bị bỏ), có %d", len(chapters))
	}
	for _, ch := range chapters {
		if ch.Slug == "" {
			t.Errorf("chương %q thiếu slug — createChapter phải sinh slug như CreateChapter", ch.Title)
		}
	}
}
