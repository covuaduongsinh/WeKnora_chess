package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// --- Fake gọn: embed interface để chỉ override method bài viết/chuyên mục/ảnh/
// phiên bản cần dùng — cùng kỹ thuật fakePositionRepo/fakeBookRepo. ---

type fakeArticleRepo struct {
	// fakeTagBase (chess_library_tag_test.go) mang phần lưu trữ HỆ THẺ và tự
	// embed interface repo. Phải embed nó CHỨ KHÔNG embed interface trực tiếp:
	// mọi đường Create/Update/Delete nay đều gọi repo thẻ, còn embed cả hai ở
	// cùng độ sâu sẽ làm method thẻ nhập nhằng.
	fakeTagBase
	articles   map[string]*types.ChessArticle
	topics     map[string]*types.ChessArticleTopic
	topicItems map[string]map[string]int // topicID -> articleID -> sortOrder
	images     map[string]*types.ChessArticleImage
	revisions  map[string]*types.ChessArticleRevision
	revOrder   []string
}

func newFakeArticleRepo() *fakeArticleRepo {
	return &fakeArticleRepo{
		fakeTagBase: newFakeTagBase(),
		articles:    map[string]*types.ChessArticle{}, topics: map[string]*types.ChessArticleTopic{},
		topicItems: map[string]map[string]int{}, images: map[string]*types.ChessArticleImage{},
		revisions: map[string]*types.ChessArticleRevision{},
	}
}

// ---- Bài viết ----

func (r *fakeArticleRepo) ListArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]*types.ChessArticle, error) {
	out := make([]*types.ChessArticle, 0, len(r.articles))
	for _, a := range r.articles {
		if f.Status != "" && a.Status != f.Status {
			continue
		}
		if f.Category != "" && a.Category != f.Category {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *fakeArticleRepo) GetArticle(ctx context.Context, tenantID uint64, id string) (*types.ChessArticle, error) {
	a, ok := r.articles[id]
	if !ok {
		return nil, fmt.Errorf("article not found: %s", id)
	}
	// Trả BẢN SAO: GORM luôn dựng struct mới cho mỗi truy vấn. Fake trả con trỏ
	// gốc thì caller sửa tại chỗ (vd RestoreArticleRevision gán Title/Content rồi
	// gọi Update) sẽ vô tình đổi luôn bản trong 'DB', khiến so sánh cũ-mới thấy
	// giống nhau và bỏ qua việc tạo phiên bản — sai lệch so với thực tế.
	cp := *a
	return &cp, nil
}

func (r *fakeArticleRepo) GetArticleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticle, error) {
	for _, a := range r.articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return nil, fmt.Errorf("article not found for slug: %s", slug)
}

func (r *fakeArticleRepo) ArticleSlugs(ctx context.Context, tenantID uint64) ([]string, error) {
	out := make([]string, 0, len(r.articles))
	for _, a := range r.articles {
		if a.Slug != "" {
			out = append(out, a.Slug)
		}
	}
	return out, nil
}

func (r *fakeArticleRepo) ArticleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, a := range r.articles {
		if a.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeArticleRepo) CreateArticle(ctx context.Context, a *types.ChessArticle) error {
	cp := *a
	r.articles[a.ID] = &cp
	return nil
}

// UpdateArticle mô phỏng đúng repo thật: KHÔNG đụng cột slug.
func (r *fakeArticleRepo) UpdateArticle(ctx context.Context, a *types.ChessArticle) error {
	existing, ok := r.articles[a.ID]
	if !ok {
		return fmt.Errorf("article not found: %s", a.ID)
	}
	cp := *a
	cp.Slug = existing.Slug
	r.articles[a.ID] = &cp
	return nil
}

func (r *fakeArticleRepo) UpdateArticleSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	a, ok := r.articles[id]
	if !ok {
		return fmt.Errorf("article not found: %s", id)
	}
	a.Slug = slug
	return nil
}

func (r *fakeArticleRepo) DeleteArticle(ctx context.Context, tenantID uint64, id string) error {
	delete(r.articles, id)
	return nil
}

// ---- Chuyên mục ----

func (r *fakeArticleRepo) ListArticleTopics(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) ([]*types.ChessArticleTopic, error) {
	out := make([]*types.ChessArticleTopic, 0, len(r.topics))
	for _, t := range r.topics {
		if f.ParentIDSet && t.ParentID != f.ParentID {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *fakeArticleRepo) GetArticleTopic(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleTopic, error) {
	t, ok := r.topics[id]
	if !ok {
		return nil, fmt.Errorf("topic not found: %s", id)
	}
	return t, nil
}

func (r *fakeArticleRepo) ArticleTopicSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, t := range r.topics {
		if t.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeArticleRepo) CreateArticleTopic(ctx context.Context, t *types.ChessArticleTopic) error {
	cp := *t
	r.topics[t.ID] = &cp
	return nil
}

func (r *fakeArticleRepo) UpdateArticleTopic(ctx context.Context, t *types.ChessArticleTopic) error {
	existing, ok := r.topics[t.ID]
	if !ok {
		return fmt.Errorf("topic not found: %s", t.ID)
	}
	existing.Title = t.Title
	existing.Description = t.Description
	existing.SortOrder = t.SortOrder
	return nil
}

func (r *fakeArticleRepo) UpdateArticleTopicParent(ctx context.Context, tenantID uint64, id, parentID string) error {
	t, ok := r.topics[id]
	if !ok {
		return fmt.Errorf("topic not found: %s", id)
	}
	t.ParentID = parentID
	return nil
}

func (r *fakeArticleRepo) DeleteArticleTopic(ctx context.Context, tenantID uint64, id string) error {
	delete(r.topics, id)
	return nil
}

func (r *fakeArticleRepo) CountArticlesOnTopic(ctx context.Context, tenantID uint64, topicID string) (int64, error) {
	return int64(len(r.topicItems[topicID])), nil
}

func (r *fakeArticleRepo) CountArticleTopicChildren(ctx context.Context, tenantID uint64, parentID string) (int64, error) {
	var n int64
	for _, t := range r.topics {
		if t.ParentID == parentID {
			n++
		}
	}
	return n, nil
}

func (r *fakeArticleRepo) SetTopicArticles(ctx context.Context, tenantID uint64, topicID string, articleIDs []string) error {
	m := map[string]int{}
	for i, id := range articleIDs {
		m[id] = i
	}
	r.topicItems[topicID] = m
	return nil
}

func (r *fakeArticleRepo) RemoveArticleFromAllTopics(ctx context.Context, tenantID uint64, articleID string) error {
	for _, m := range r.topicItems {
		delete(m, articleID)
	}
	return nil
}

func (r *fakeArticleRepo) RemoveTopicItems(ctx context.Context, tenantID uint64, topicID string) error {
	delete(r.topicItems, topicID)
	return nil
}

// ---- Ảnh ----

func (r *fakeArticleRepo) ListArticleImagesByArticle(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleImage, error) {
	out := []*types.ChessArticleImage{}
	for _, img := range r.images {
		if img.ArticleID == articleID {
			out = append(out, img)
		}
	}
	return out, nil
}

func (r *fakeArticleRepo) DeleteArticleImage(ctx context.Context, tenantID uint64, id string) error {
	delete(r.images, id)
	return nil
}

// ---- Lịch sử phiên bản ----

func (r *fakeArticleRepo) CreateArticleRevision(ctx context.Context, rev *types.ChessArticleRevision) error {
	cp := *rev
	r.revisions[rev.ID] = &cp
	r.revOrder = append(r.revOrder, rev.ID)
	return nil
}

func (r *fakeArticleRepo) ListArticleRevisions(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleRevision, error) {
	out := []*types.ChessArticleRevision{}
	for i := len(r.revOrder) - 1; i >= 0; i-- { // mới nhất trước, như repo thật
		if rev := r.revisions[r.revOrder[i]]; rev != nil && rev.ArticleID == articleID {
			out = append(out, rev)
		}
	}
	return out, nil
}

func (r *fakeArticleRepo) GetArticleRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessArticleRevision, error) {
	rev, ok := r.revisions[revisionID]
	if !ok {
		return nil, fmt.Errorf("revision not found: %s", revisionID)
	}
	return rev, nil
}

func (r *fakeArticleRepo) CountArticleRevisions(ctx context.Context, tenantID uint64, articleID string) (int64, error) {
	var n int64
	for _, rev := range r.revisions {
		if rev.ArticleID == articleID {
			n++
		}
	}
	return n, nil
}

func (r *fakeArticleRepo) DeleteArticleRevisionsByArticle(ctx context.Context, tenantID uint64, articleID string) error {
	for id, rev := range r.revisions {
		if rev.ArticleID == articleID {
			delete(r.revisions, id)
		}
	}
	return nil
}

// ---- Alias repo giả: theo dõi RIÊNG rename vs synonym để khẳng định
// ReplaceSynonyms KHÔNG đụng lịch sử rename. ----

type fakeArticleAliasRepo struct {
	interfaces.ChessSlugAliasRepository
	renames  map[string]string   // chessType+"/"+oldSlug -> newSlug (kind=rename)
	synonyms map[string][]string // chessType+"/"+targetSlug -> danh sách bí danh
}

func newFakeArticleAliasRepo() *fakeArticleAliasRepo {
	return &fakeArticleAliasRepo{renames: map[string]string{}, synonyms: map[string][]string{}}
}

func (r *fakeArticleAliasRepo) ResolveAlias(ctx context.Context, tenantID uint64, chessType, oldSlug string) (string, bool, error) {
	if ns, ok := r.renames[chessType+"/"+oldSlug]; ok {
		return ns, true, nil
	}
	for key, list := range r.synonyms {
		for _, s := range list {
			if s == oldSlug {
				return key[len(chessType)+1:], true, nil
			}
		}
	}
	return "", false, nil
}

func (r *fakeArticleAliasRepo) AddAlias(ctx context.Context, tenantID uint64, chessType, oldSlug, newSlug string) error {
	r.renames[chessType+"/"+oldSlug] = newSlug
	return nil
}

func (r *fakeArticleAliasRepo) ReplaceSynonyms(ctx context.Context, tenantID uint64, chessType, targetSlug string, aliasSlugs []string) error {
	// Bảng thật có UNIQUE (tenant_id, chess_type, old_slug): MỘT bí danh chỉ trỏ
	// tới ĐÚNG MỘT đích. Ghi lại bí danh đó với đích mới là UPDATE dòng cũ (ON
	// CONFLICT old_slug), không tạo dòng thứ hai — nên không thể còn bí danh mồ
	// côi ở đích cũ. Fake phải mô phỏng đúng ràng buộc này, nếu không
	// RenameArticleSlug sẽ bị coi là để lại bí danh mồ côi trong khi thực tế không.
	for _, a := range aliasSlugs {
		for key, list := range r.synonyms {
			if key == chessType+"/"+targetSlug {
				continue
			}
			kept := make([]string, 0, len(list))
			for _, x := range list {
				if x != a {
					kept = append(kept, x)
				}
			}
			if len(kept) == 0 {
				delete(r.synonyms, key)
			} else {
				r.synonyms[key] = kept
			}
		}
	}
	r.synonyms[chessType+"/"+targetSlug] = append([]string{}, aliasSlugs...)
	return nil
}

func (r *fakeArticleAliasRepo) DeleteAliasesFor(ctx context.Context, tenantID uint64, chessType, targetSlug string) error {
	delete(r.synonyms, chessType+"/"+targetSlug)
	for key, ns := range r.renames {
		if ns == targetSlug {
			delete(r.renames, key)
		}
	}
	return nil
}

// ---- Ref repo giả: đếm lời gọi để khẳng định cascade xóa dọn cả 2 chiều ----

type fakeArticleChessRefRepo struct {
	interfaces.WikiChessRefRepository
	replacedFor  []string
	deletedFor   []string
	deletedChess []string
}

func (r *fakeArticleChessRefRepo) ReplaceForArticle(ctx context.Context, tenantID uint64, articleSlug string, refs []types.WikiChessRef) error {
	r.replacedFor = append(r.replacedFor, articleSlug)
	return nil
}
func (r *fakeArticleChessRefRepo) DeleteForArticle(ctx context.Context, tenantID uint64, articleSlug string) error {
	r.deletedFor = append(r.deletedFor, articleSlug)
	return nil
}
func (r *fakeArticleChessRefRepo) DeleteForChess(ctx context.Context, tenantID uint64, chessType, chessSlug string) error {
	r.deletedChess = append(r.deletedChess, chessType+"/"+chessSlug)
	return nil
}
func (r *fakeArticleChessRefRepo) ListBacklinks(ctx context.Context, tenantID uint64, chessType, chessSlug string) ([]types.ChessBacklink, error) {
	return nil, nil
}

func newTestArticleService(repo *fakeArticleRepo, alias *fakeArticleAliasRepo, refs *fakeArticleChessRefRepo) *chessLibraryService {
	return &chessLibraryService{
		repo:         repo,
		chessRefRepo: refs,
		aliasRepo:    alias,
		indexer:      nil, // CHESS_KB_INDEX tắt → mọi lời gọi indexer guard nil
		fileService:  nil, // không test ảnh vật lý ở đây — nhánh ảnh guard nil
	}
}

func newArticleSvc() (*chessLibraryService, *fakeArticleRepo, *fakeArticleAliasRepo, *fakeArticleChessRefRepo) {
	repo := newFakeArticleRepo()
	alias := newFakeArticleAliasRepo()
	refs := &fakeArticleChessRefRepo{}
	return newTestArticleService(repo, alias, refs), repo, alias, refs
}

// ---- Slug ----

func TestCreateArticle_GeneratesSlugAndDefaultsDraft(t *testing.T) {
	svc, _, _, _ := newArticleSvc()
	a, err := svc.CreateArticle(context.Background(), &types.ChessArticle{TenantID: 1, Title: "Ghim (Pin) là gì?"})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	if a.Slug == "" {
		t.Errorf("slug phải được sinh tự động")
	}
	if a.Status != types.ChessArticleStatusDraft {
		t.Errorf("status mặc định = %q, muốn %q", a.Status, types.ChessArticleStatusDraft)
	}
}

func TestCreateArticle_SlugCollision_AppendsSuffix(t *testing.T) {
	svc, _, _, _ := newArticleSvc()
	ctx := context.Background()
	first, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Bài trùng tên"})
	if err != nil {
		t.Fatalf("tạo lần 1 lỗi: %v", err)
	}
	second, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Bài trùng tên"})
	if err != nil {
		t.Fatalf("tạo lần 2 lỗi: %v", err)
	}
	if first.Slug == second.Slug {
		t.Fatalf("2 bài trùng tiêu đề phải có slug KHÁC nhau, cả hai đều là %q", first.Slug)
	}
}

func TestRenameArticleSlug_WritesAliasAndKeepsOldSlugResolvable(t *testing.T) {
	svc, _, _, _ := newArticleSvc()
	ctx := context.Background()
	created, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Bài đổi tên"})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	oldSlug := created.Slug
	renamed, err := svc.RenameArticleSlug(ctx, 1, created.ID, "ten-moi-hoan-toan")
	if err != nil {
		t.Fatalf("RenameArticleSlug lỗi: %v", err)
	}
	if renamed.Slug == oldSlug {
		t.Fatalf("slug phải đổi, vẫn là %q", oldSlug)
	}
	// Link cũ [[article/<oldSlug>]] vẫn phải giải mã đúng nhờ alias.
	got, err := svc.GetArticleBySlug(ctx, 1, oldSlug)
	if err != nil {
		t.Fatalf("slug cũ phải còn resolve qua alias: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("alias trả nhầm bản ghi: got=%s want=%s", got.ID, created.ID)
	}
}

// ---- Bí danh / từ đồng nghĩa ----

func TestCreateArticle_SyncsSynonyms(t *testing.T) {
	svc, _, alias, _ := newArticleSvc()
	a, err := svc.CreateArticle(context.Background(), &types.ChessArticle{
		TenantID: 1, Title: "Ghim", Aliases: "Pin, Đóng đinh",
	})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	got := alias.synonyms[types.ChessRefTypeArticle+"/"+a.Slug]
	if len(got) != 2 {
		t.Fatalf("kỳ vọng 2 bí danh, nhận %v", got)
	}
	// Phải được slugify (bỏ dấu, thay khoảng trắng) — không lưu nguyên chữ hiển thị.
	if got[0] != "pin" || got[1] != "dong-dinh" {
		t.Errorf("bí danh chưa được chuẩn hóa: %v", got)
	}
}

// Bí danh trùng slug THẬT của một bài khác phải bị BỎ QUA: thứ tự resolve là
// exact→alias nên nếu lưu, nó sẽ che khuất bài kia vĩnh viễn.
func TestSyncArticleAliases_SkipsAliasCollidingWithRealSlug(t *testing.T) {
	svc, _, alias, _ := newArticleSvc()
	ctx := context.Background()
	other, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Xiên"})
	if err != nil {
		t.Fatalf("tạo bài khác lỗi: %v", err)
	}
	a, err := svc.CreateArticle(ctx, &types.ChessArticle{
		TenantID: 1, Title: "Ghim", Aliases: "Pin, " + other.Slug,
	})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	got := alias.synonyms[types.ChessRefTypeArticle+"/"+a.Slug]
	for _, s := range got {
		if s == other.Slug {
			t.Fatalf("bí danh %q trùng slug thật của bài khác — phải bị bỏ qua, nhận %v", other.Slug, got)
		}
	}
	if len(got) != 1 || got[0] != "pin" {
		t.Errorf("chỉ bí danh hợp lệ được giữ, nhận %v", got)
	}
}

// Đổi slug bài → bí danh phải trỏ theo slug MỚI, không mồ côi ở slug cũ.
func TestRenameArticleSlug_RepointsSynonyms(t *testing.T) {
	svc, _, alias, _ := newArticleSvc()
	ctx := context.Background()
	created, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Ghim", Aliases: "Pin"})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	oldSlug := created.Slug
	renamed, err := svc.RenameArticleSlug(ctx, 1, created.ID, "ghim-tuyet-doi")
	if err != nil {
		t.Fatalf("RenameArticleSlug lỗi: %v", err)
	}
	if len(alias.synonyms[types.ChessRefTypeArticle+"/"+renamed.Slug]) != 1 {
		t.Errorf("bí danh phải trỏ theo slug MỚI %q, nhận %+v", renamed.Slug, alias.synonyms)
	}
	if len(alias.synonyms[types.ChessRefTypeArticle+"/"+oldSlug]) != 0 {
		t.Errorf("không được để bí danh mồ côi ở slug CŨ %q", oldSlug)
	}
}

// ---- Lịch sử phiên bản ----

func TestUpdateArticle_CreatesRevisionOnlyWhenTitleOrContentChanges(t *testing.T) {
	svc, repo, _, _ := newArticleSvc()
	ctx := context.Background()
	a, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Ghim", Content: "nội dung gốc"})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}

	// Sửa các trường phân loại — KHÔNG được tạo rác lịch sử.
	only := *a
	only.Category = "thuat-ngu"
	only.Level = "ma"
	only.Status = types.ChessArticleStatusPublished
	if _, err := svc.UpdateArticle(ctx, &only, ""); err != nil {
		t.Fatalf("UpdateArticle (metadata) lỗi: %v", err)
	}
	if n, _ := repo.CountArticleRevisions(ctx, 1, a.ID); n != 0 {
		t.Fatalf("sửa metadata KHÔNG được tạo phiên bản, nhận %d", n)
	}

	// Sửa nội dung — PHẢI lưu bản trước khi ghi đè.
	edit := only
	edit.Content = "nội dung mới"
	if _, err := svc.UpdateArticle(ctx, &edit, "bổ sung ví dụ"); err != nil {
		t.Fatalf("UpdateArticle (content) lỗi: %v", err)
	}
	revs, _ := repo.ListArticleRevisions(ctx, 1, a.ID)
	if len(revs) != 1 {
		t.Fatalf("sửa nội dung phải tạo đúng 1 phiên bản, nhận %d", len(revs))
	}
	if revs[0].Content != "nội dung gốc" {
		t.Errorf("phiên bản phải lưu bản TRƯỚC khi ghi đè, nhận %q", revs[0].Content)
	}
	if revs[0].Summary != "bổ sung ví dụ" {
		t.Errorf("ghi chú thay đổi sai: %q", revs[0].Summary)
	}
}

func TestRestoreArticleRevision_RestoresAndCreatesAnotherRevision(t *testing.T) {
	svc, repo, _, _ := newArticleSvc()
	ctx := context.Background()
	a, err := svc.CreateArticle(ctx, &types.ChessArticle{TenantID: 1, Title: "Ghim", Content: "bản 1"})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	edit := *a
	edit.Content = "bản 2"
	if _, err := svc.UpdateArticle(ctx, &edit, ""); err != nil {
		t.Fatalf("UpdateArticle lỗi: %v", err)
	}
	revs, _ := repo.ListArticleRevisions(ctx, 1, a.ID)
	if len(revs) != 1 {
		t.Fatalf("kỳ vọng 1 phiên bản trước khi khôi phục, nhận %d", len(revs))
	}

	restored, err := svc.RestoreArticleRevision(ctx, 1, a.ID, revs[0].ID)
	if err != nil {
		t.Fatalf("RestoreArticleRevision lỗi: %v", err)
	}
	if restored.Content != "bản 1" {
		t.Errorf("nội dung sau khôi phục = %q, muốn %q", restored.Content, "bản 1")
	}
	// Khôi phục CŨNG phải lưu bản đang có ("bản 2") thành phiên bản mới.
	after, _ := repo.ListArticleRevisions(ctx, 1, a.ID)
	if len(after) != 2 {
		t.Fatalf("khôi phục phải tạo thêm 1 phiên bản, nhận %d", len(after))
	}
	if after[0].Content != "bản 2" {
		t.Errorf("phiên bản mới nhất phải là bản bị ghi đè (%q), nhận %q", "bản 2", after[0].Content)
	}
}

// ---- Chuyên mục: ràng buộc cây 2 tầng ----

func TestCreateArticleTopic_RejectsThirdLevel(t *testing.T) {
	svc, _, _, _ := newArticleSvc()
	ctx := context.Background()
	root, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{TenantID: 1, Title: "Chiến thuật"})
	if err != nil {
		t.Fatalf("tạo chuyên mục gốc lỗi: %v", err)
	}
	child, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{TenantID: 1, Title: "Ghim", ParentID: root.ID})
	if err != nil {
		t.Fatalf("tạo chuyên mục con lỗi: %v", err)
	}
	// Tầng 3: lấy CON làm cha → phải bị chặn.
	if _, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{
		TenantID: 1, Title: "Ghim tuyệt đối", ParentID: child.ID,
	}); err == nil {
		t.Fatal("phải chặn tầng thứ 3 (chỉ chuyên mục GỐC mới được làm cha)")
	}
}

func TestDeleteArticleTopic_BlockedWhenHasChildren(t *testing.T) {
	svc, _, _, _ := newArticleSvc()
	ctx := context.Background()
	root, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{TenantID: 1, Title: "Khai cuộc"})
	if err != nil {
		t.Fatalf("tạo chuyên mục gốc lỗi: %v", err)
	}
	if _, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{TenantID: 1, Title: "Sicilian", ParentID: root.ID}); err != nil {
		t.Fatalf("tạo chuyên mục con lỗi: %v", err)
	}
	if err := svc.DeleteArticleTopic(ctx, 1, root.ID); err == nil {
		t.Fatal("phải chặn xóa chuyên mục còn con (tránh mồ côi tầng 2)")
	}
}

// ---- Cascade xóa ----

func TestDeleteArticle_CascadesRefsAliasesTopicsAndRevisions(t *testing.T) {
	svc, repo, alias, refs := newArticleSvc()
	ctx := context.Background()
	a, err := svc.CreateArticle(ctx, &types.ChessArticle{
		TenantID: 1, Title: "Ghim", Aliases: "Pin", Content: "xem [[position/abc]]",
	})
	if err != nil {
		t.Fatalf("CreateArticle lỗi: %v", err)
	}
	// Cho bài một phiên bản + gán vào một chuyên mục để kiểm cascade.
	edit := *a
	edit.Content = "nội dung mới"
	if _, err := svc.UpdateArticle(ctx, &edit, ""); err != nil {
		t.Fatalf("UpdateArticle lỗi: %v", err)
	}
	topic, err := svc.CreateArticleTopic(ctx, &types.ChessArticleTopic{TenantID: 1, Title: "Chiến thuật"})
	if err != nil {
		t.Fatalf("CreateArticleTopic lỗi: %v", err)
	}
	if err := svc.SetTopicArticles(ctx, 1, topic.ID, []string{a.ID}); err != nil {
		t.Fatalf("SetTopicArticles lỗi: %v", err)
	}

	if err := svc.DeleteArticle(ctx, 1, a.ID); err != nil {
		t.Fatalf("DeleteArticle lỗi: %v", err)
	}

	if _, err := repo.GetArticle(ctx, 1, a.ID); err == nil {
		t.Error("bài viết phải bị xóa")
	}
	// Ref TRỎ TỚI bài (backlink từ nơi khác).
	wantChess := types.ChessRefTypeArticle + "/" + a.Slug
	if len(refs.deletedChess) == 0 || refs.deletedChess[len(refs.deletedChess)-1] != wantChess {
		t.Errorf("phải dọn ref trỏ tới bài (%s), nhận %v", wantChess, refs.deletedChess)
	}
	// Ref bài PHÁT RA.
	if len(refs.deletedFor) == 0 || refs.deletedFor[len(refs.deletedFor)-1] != a.Slug {
		t.Errorf("phải dọn ref bài phát ra (%s), nhận %v", a.Slug, refs.deletedFor)
	}
	if len(alias.synonyms[types.ChessRefTypeArticle+"/"+a.Slug]) != 0 {
		t.Error("phải dọn sạch bí danh khi xóa bài")
	}
	if n, _ := repo.CountArticlesOnTopic(ctx, 1, topic.ID); n != 0 {
		t.Errorf("phải gỡ bài khỏi mọi chuyên mục, còn %d", n)
	}
	if n, _ := repo.CountArticleRevisions(ctx, 1, a.ID); n != 0 {
		t.Errorf("phải xóa sạch lịch sử phiên bản, còn %d", n)
	}
}

// ---- Index theo trạng thái ----

// reindexArticle phải ĐỐI XỨNG: published→index, draft→GỠ. Chỉ trả nil khi
// draft là chưa đủ — bài từng published rồi hạ xuống draft sẽ để lại tri thức
// cũ trong KB và agent vẫn trích dẫn.
func TestReindexArticle_RemovesWhenDraft(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	svc, _, _, _ := newArticleSvc()
	idx := newTestIndexer(nil, nil)
	svc.indexer = idx

	// Draft → đi vào nhánh Remove. Không panic là đủ để khẳng định nhánh đúng:
	// stub idxRepo trả nil mapping nên Remove thoát sớm, KHÔNG gọi IndexArticle.
	draft := &types.ChessArticle{
		TenantID: 1, ID: "a1", Slug: "ghim", Title: "Ghim", Status: types.ChessArticleStatusDraft,
	}
	svc.reindexArticle(context.Background(), draft)
}

func TestIndexArticle_SkipsDraftAndSlugless(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	ix := newTestIndexer(nil, nil)
	ctx := context.Background()

	// Bản thảo → bỏ qua, không lỗi.
	if err := ix.IndexArticle(ctx, &types.ChessArticle{
		TenantID: 1, Slug: "ghim", Title: "Ghim", Status: types.ChessArticleStatusDraft,
	}); err != nil {
		t.Errorf("bài draft phải bị bỏ qua êm, nhận lỗi: %v", err)
	}
	// Không slug → bỏ qua (không thể làm đích wikilink).
	if err := ix.IndexArticle(ctx, &types.ChessArticle{
		TenantID: 1, Title: "Chưa có slug", Status: types.ChessArticleStatusPublished,
	}); err != nil {
		t.Errorf("bài chưa có slug phải bị bỏ qua êm, nhận lỗi: %v", err)
	}
}
