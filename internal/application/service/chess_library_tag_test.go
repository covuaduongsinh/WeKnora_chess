package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// fakeTagBase là phần lưu trữ HỆ THẺ dùng chung cho mọi fake repo của gói này.
//
// Vì sao là một struct riêng để EMBED chứ không viết lại trong từng fake: kể từ
// khi hệ thẻ ra đời, mọi đường Create/Update/Delete của thế cờ/sách/bài viết
// đều đi qua repo thẻ, nên fakeArticleRepo/fakeBookRepo/fakePositionRepo đều
// cần các method này. Đồng thời fakeTagBase phải EMBED interface (chứ không để
// các fake kia embed song song) — hai chỗ embed cùng độ sâu sẽ làm method thẻ
// bị nhập nhằng và fake không còn thỏa interface.
type fakeTagBase struct {
	interfaces.ChessLibraryRepository
	tags  map[string]*types.ChessTag // id -> thẻ
	links map[string][]string        // "<chessType>:<chessID>" -> []tagID
	csv   map[string]string          // "<chessType>:<chessID>" -> CSV đã ghi lại
}

func newFakeTagBase() fakeTagBase {
	return fakeTagBase{
		tags:  map[string]*types.ChessTag{},
		links: map[string][]string{},
		csv:   map[string]string{},
	}
}

func tagKey(chessType, chessID string) string { return chessType + ":" + chessID }

func (r *fakeTagBase) ListTags(ctx context.Context, tenantID uint64, f types.ChessTagFilter) ([]*types.ChessTag, error) {
	out := make([]*types.ChessTag, 0, len(r.tags))
	for _, t := range r.tags {
		if f.Kind != "" && t.Kind != f.Kind {
			continue
		}
		if f.OnlyUsed && t.UsageCount == 0 {
			continue
		}
		if f.Keyword != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(f.Keyword)) {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out, nil
}

func (r *fakeTagBase) GetTag(ctx context.Context, tenantID uint64, id string) (*types.ChessTag, error) {
	t, ok := r.tags[id]
	if !ok {
		return nil, fmt.Errorf("tag not found: %s", id)
	}
	return t, nil
}

func (r *fakeTagBase) GetTagBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessTag, error) {
	for _, t := range r.tags {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tag not found: %s", slug)
}

func (r *fakeTagBase) GetTagsBySlugs(ctx context.Context, tenantID uint64, slugs []string) ([]*types.ChessTag, error) {
	want := map[string]bool{}
	for _, s := range slugs {
		want[s] = true
	}
	out := []*types.ChessTag{}
	for _, t := range r.tags {
		if want[t.Slug] {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *fakeTagBase) CreateTag(ctx context.Context, tag *types.ChessTag) error {
	r.tags[tag.ID] = tag
	return nil
}

func (r *fakeTagBase) UpdateTag(ctx context.Context, tag *types.ChessTag) error {
	cur, ok := r.tags[tag.ID]
	if !ok {
		return fmt.Errorf("tag not found: %s", tag.ID)
	}
	cur.Name, cur.Description, cur.Color, cur.SortOrder = tag.Name, tag.Description, tag.Color, tag.SortOrder
	return nil
}

func (r *fakeTagBase) UpdateTagSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	if t, ok := r.tags[id]; ok {
		t.Slug = slug
	}
	return nil
}

func (r *fakeTagBase) DeleteTag(ctx context.Context, tenantID uint64, id string) error {
	delete(r.tags, id)
	return nil
}

func (r *fakeTagBase) TagSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	for _, t := range r.tags {
		if t.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

// hasTagID thay cho một hàm contains cục bộ: gói service đã có sẵn contains()
// ở knowledge_span_tracker.go, trùng tên là lỗi biên dịch.
func hasTagID(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func (r *fakeTagBase) RecountTagUsage(ctx context.Context, tenantID uint64, tagIDs []string) error {
	counts := map[string]int{}
	for _, ids := range r.links {
		for _, id := range ids {
			counts[id]++
		}
	}
	for id, t := range r.tags {
		if len(tagIDs) > 0 && !hasTagID(tagIDs, id) {
			continue
		}
		t.UsageCount = counts[id]
	}
	return nil
}

func (r *fakeTagBase) SetTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string, tagIDs []string) error {
	k := tagKey(chessType, chessID)
	if len(tagIDs) == 0 {
		delete(r.links, k)
		return nil
	}
	r.links[k] = append([]string{}, tagIDs...)
	return nil
}

func (r *fakeTagBase) RemoveAllTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string) error {
	delete(r.links, tagKey(chessType, chessID))
	return nil
}

func (r *fakeTagBase) RemoveTagItems(ctx context.Context, tenantID uint64, tagID string) error {
	for k, ids := range r.links {
		kept := ids[:0]
		for _, id := range ids {
			if id != tagID {
				kept = append(kept, id)
			}
		}
		if len(kept) == 0 {
			delete(r.links, k)
		} else {
			r.links[k] = kept
		}
	}
	return nil
}

func (r *fakeTagBase) MergeTagItems(ctx context.Context, tenantID uint64, fromTagID, toTagID string) error {
	for k, ids := range r.links {
		out := []string{}
		for _, id := range ids {
			if id == fromTagID {
				id = toTagID
			}
			if !hasTagID(out, id) { // dedupe: mục có thể đã mang sẵn thẻ đích
				out = append(out, id)
			}
		}
		r.links[k] = out
	}
	return nil
}

func (r *fakeTagBase) TagsForMany(ctx context.Context, tenantID uint64, chessType string, chessIDs []string) (map[string][]*types.ChessTag, error) {
	out := map[string][]*types.ChessTag{}
	for _, id := range chessIDs {
		for _, tid := range r.links[tagKey(chessType, id)] {
			if t, ok := r.tags[tid]; ok {
				out[id] = append(out[id], t)
			}
		}
	}
	return out, nil
}

func (r *fakeTagBase) UpdateEntityTagsCSV(ctx context.Context, tenantID uint64, chessType, chessID, csv string) error {
	r.csv[tagKey(chessType, chessID)] = csv
	return nil
}

func (r *fakeTagBase) CountTagItems(ctx context.Context, tenantID uint64, tagID, chessType string) (int64, error) {
	var n int64
	for k, ids := range r.links {
		if chessType != "" && !strings.HasPrefix(k, chessType+":") {
			continue
		}
		if hasTagID(ids, tagID) {
			n++
		}
	}
	return n, nil
}

func (r *fakeTagBase) CountTagItemsByType(ctx context.Context, tenantID uint64, tagID string) (map[string]int64, error) {
	out := map[string]int64{}
	for k, ids := range r.links {
		if !hasTagID(ids, tagID) {
			continue
		}
		out[strings.SplitN(k, ":", 2)[0]]++
	}
	return out, nil
}

func (r *fakeTagBase) ListTagItems(ctx context.Context, tenantID uint64, tagID, chessType string, offset, limit int) ([]*types.ChessTagItem, error) {
	keys := make([]string, 0, len(r.links))
	for k := range r.links {
		keys = append(keys, k)
	}
	sort.Strings(keys) // thứ tự ổn định để test không phụ thuộc thứ tự map
	items := []*types.ChessTagItem{}
	for _, k := range keys {
		if !hasTagID(r.links[k], tagID) {
			continue
		}
		if chessType != "" && !strings.HasPrefix(k, chessType+":") {
			continue
		}
		parts := strings.SplitN(k, ":", 2)
		items = append(items, &types.ChessTagItem{TenantID: tenantID, TagID: tagID, ChessType: parts[0], ChessID: parts[1]})
	}
	if offset >= len(items) {
		return nil, nil
	}
	items = items[offset:]
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

// ---- Dàn dựng service chỉ có phần thẻ ----

type fakeTagRepo struct{ fakeTagBase }

func newTagSvc() (*chessLibraryService, *fakeTagRepo) {
	repo := &fakeTagRepo{fakeTagBase: newFakeTagBase()}
	return &chessLibraryService{repo: repo}, repo
}

func mustSetTags(t *testing.T, svc *chessLibraryService, chessType, id string, names ...string) []*types.ChessTag {
	t.Helper()
	tags, err := svc.SetChessTags(context.Background(), 1, chessType, id, names)
	if err != nil {
		t.Fatalf("SetChessTags lỗi: %v", err)
	}
	return tags
}

// ---- Chuẩn hóa tên thẻ ----

// Đây là lý do tồn tại của cả hệ thẻ: trước đó "Khai cuộc", "khai-cuoc" và
// "KHAI CUOC" nằm trong cột CSV thành ba chuỗi khác nhau, không gộp được.
func TestSetChessTags_FoldsVietnameseDiacriticsIntoOneTag(t *testing.T) {
	svc, repo := newTagSvc()
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Khai cuộc")
	mustSetTags(t, svc, types.ChessRefTypeBook, "b1", "khai-cuoc")
	mustSetTags(t, svc, types.ChessRefTypePosition, "p1", "KHAI CUOC")

	if len(repo.tags) != 1 {
		names := []string{}
		for _, tg := range repo.tags {
			names = append(names, tg.Name)
		}
		t.Fatalf("ba biến thể phải quy về MỘT thẻ, đang có %d: %v", len(repo.tags), names)
	}
	for _, tg := range repo.tags {
		if tg.Slug != "khai-cuoc" {
			t.Errorf("slug = %q, muốn %q", tg.Slug, "khai-cuoc")
		}
	}
}

func TestSetChessTags_SplitsCSVAndDedupes(t *testing.T) {
	svc, repo := newTagSvc()
	tags := mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim, Khai cuộc , ghim", "  ")
	if len(tags) != 2 {
		t.Fatalf("muốn 2 thẻ sau khi khử trùng, nhận %d", len(tags))
	}
	if len(repo.links[tagKey(types.ChessRefTypeArticle, "a1")]) != 2 {
		t.Errorf("bảng nối phải có đúng 2 liên kết")
	}
}

// Cột CSV là BẢN HIỂN THỊ được ghi lại từ hệ thẻ — gõ không dấu phải hiện lại
// thành tên chuẩn có dấu.
func TestSetChessTags_RewritesDisplayCSVToCanonicalNames(t *testing.T) {
	svc, repo := newTagSvc()
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Khai cuộc")
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a2", "khai cuoc")

	if got := repo.csv[tagKey(types.ChessRefTypeArticle, "a2")]; got != "Khai cuộc" {
		t.Errorf("CSV hiển thị = %q, muốn %q (tên chuẩn của thẻ đã có)", got, "Khai cuộc")
	}
}

// Ghi ĐÈ, không cộng dồn: bỏ một thẻ khỏi ô nhập là thẻ đó phải rời khỏi mục.
func TestSetChessTags_OverwritesAndRecountsRemovedTag(t *testing.T) {
	svc, repo := newTagSvc()
	first := mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim, Xiên")
	removed := first[1]

	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim")

	ids := repo.links[tagKey(types.ChessRefTypeArticle, "a1")]
	if len(ids) != 1 {
		t.Fatalf("muốn còn 1 liên kết, nhận %d", len(ids))
	}
	if repo.tags[removed.ID].UsageCount != 0 {
		t.Errorf("thẻ bị gỡ phải được đếm lại về 0, đang là %d", repo.tags[removed.ID].UsageCount)
	}
}

func TestSetChessTags_EmptyClearsAllTags(t *testing.T) {
	svc, repo := newTagSvc()
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim")
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "")
	if _, ok := repo.links[tagKey(types.ChessRefTypeArticle, "a1")]; ok {
		t.Errorf("gửi danh sách rỗng phải gỡ hết thẻ khỏi mục")
	}
}

// ---- Thẻ nhóm nội dung (từ vựng đóng) ----

func TestEnsureChessTagGroups_IsIdempotent(t *testing.T) {
	svc, repo := newTagSvc()
	n1, err := svc.EnsureChessTagGroups(context.Background(), 1)
	if err != nil {
		t.Fatalf("seed lỗi: %v", err)
	}
	if n1 != len(types.ChessTagGroupSeeds) {
		t.Fatalf("lần đầu phải tạo %d nhóm, nhận %d", len(types.ChessTagGroupSeeds), n1)
	}
	n2, _ := svc.EnsureChessTagGroups(context.Background(), 1)
	if n2 != 0 {
		t.Errorf("gọi lại không được tạo thêm, nhận %d", n2)
	}
	if len(repo.tags) != len(types.ChessTagGroupSeeds) {
		t.Errorf("tổng số thẻ = %d, muốn %d", len(repo.tags), len(types.ChessTagGroupSeeds))
	}
}

func TestDeleteChessTag_BlocksGroupTags(t *testing.T) {
	svc, repo := newTagSvc()
	if _, err := svc.EnsureChessTagGroups(context.Background(), 1); err != nil {
		t.Fatalf("seed lỗi: %v", err)
	}
	var groupID string
	for id, tg := range repo.tags {
		if tg.Slug == types.ChessTagGroupOpening {
			groupID = id
		}
	}
	if groupID == "" {
		t.Fatal("không tìm thấy thẻ nhóm khai cuộc sau khi seed")
	}
	if err := svc.DeleteChessTag(context.Background(), 1, groupID); err == nil {
		t.Error("phải CHẶN xóa thẻ nhóm — lần seed sau sẽ tạo lại và sinh trùng lặp")
	}
}

// Gõ tên trùng một nhóm sẵn có KHÔNG được đẻ ra thẻ tự do song song.
func TestSetChessTags_ReusesExistingGroupTagInsteadOfCreatingFreeTag(t *testing.T) {
	svc, repo := newTagSvc()
	if _, err := svc.EnsureChessTagGroups(context.Background(), 1); err != nil {
		t.Fatalf("seed lỗi: %v", err)
	}
	before := len(repo.tags)
	tags := mustSetTags(t, svc, types.ChessRefTypeBook, "b1", "khai cuoc")
	if len(repo.tags) != before {
		t.Errorf("không được tạo thẻ mới, số thẻ %d -> %d", before, len(repo.tags))
	}
	if len(tags) != 1 || tags[0].Kind != types.ChessTagKindGroup {
		t.Errorf("phải dùng lại chính thẻ NHÓM, nhận %+v", tags)
	}
}

// ---- Gộp & đổi tên ----

func TestMergeChessTags_MovesLinksAndDeletesSource(t *testing.T) {
	svc, repo := newTagSvc()
	a := mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim")[0]
	b := mustSetTags(t, svc, types.ChessRefTypePosition, "p1", "Đóng đinh")[0]

	to, err := svc.MergeChessTags(context.Background(), 1, b.ID, a.ID)
	if err != nil {
		t.Fatalf("MergeChessTags lỗi: %v", err)
	}
	if to.ID != a.ID {
		t.Errorf("phải trả về thẻ ĐÍCH")
	}
	if _, ok := repo.tags[b.ID]; ok {
		t.Error("thẻ nguồn phải bị xóa sau khi gộp")
	}
	if !hasTagID(repo.links[tagKey(types.ChessRefTypePosition, "p1")], a.ID) {
		t.Error("liên kết của thẻ nguồn phải chuyển sang thẻ đích")
	}
	if repo.tags[a.ID].UsageCount != 2 {
		t.Errorf("usage_count sau gộp = %d, muốn 2", repo.tags[a.ID].UsageCount)
	}
}

// Đổi tên thành tên của một thẻ đã tồn tại = GỘP, không phải lỗi trùng khóa.
// Đây là thao tác dọn thẻ thường gặp nhất khi từ điển đã lỡ có bản gõ sai.
func TestUpdateChessTag_RenameIntoExistingSlugMerges(t *testing.T) {
	svc, repo := newTagSvc()
	keep := mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Khai cuộc")[0]
	dup := mustSetTags(t, svc, types.ChessRefTypeArticle, "a2", "Khai cuoc bay")[0]

	got, err := svc.UpdateChessTag(context.Background(), &types.ChessTag{
		ID: dup.ID, TenantID: 1, Name: "Khai cuộc",
	})
	if err != nil {
		t.Fatalf("UpdateChessTag lỗi: %v", err)
	}
	if got.ID != keep.ID {
		t.Errorf("phải gộp vào thẻ sẵn có, nhận id %s", got.ID)
	}
	if _, ok := repo.tags[dup.ID]; ok {
		t.Error("thẻ trùng phải biến mất sau khi gộp")
	}
}

func TestUpdateChessTag_KeepsGroupSlugStable(t *testing.T) {
	svc, repo := newTagSvc()
	if _, err := svc.EnsureChessTagGroups(context.Background(), 1); err != nil {
		t.Fatalf("seed lỗi: %v", err)
	}
	var group *types.ChessTag
	for _, tg := range repo.tags {
		if tg.Slug == types.ChessTagGroupTactics {
			group = tg
		}
	}
	if _, err := svc.UpdateChessTag(context.Background(), &types.ChessTag{
		ID: group.ID, TenantID: 1, Name: "Đòn chiến thuật",
	}); err != nil {
		t.Fatalf("UpdateChessTag lỗi: %v", err)
	}
	if group.Slug != types.ChessTagGroupTactics {
		t.Errorf("slug thẻ nhóm phải giữ nguyên %q, nhận %q", types.ChessTagGroupTactics, group.Slug)
	}
	if group.Name != "Đòn chiến thuật" {
		t.Errorf("tên hiển thị vẫn phải đổi được, nhận %q", group.Name)
	}
}

// ---- Tra nội dung theo thẻ (mục lục ngang) ----

func TestListChessTagItems_CountsAcrossTypes(t *testing.T) {
	svc, _ := newTagSvc()
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Ghim")
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a2", "Ghim")
	mustSetTags(t, svc, types.ChessRefTypePosition, "p1", "Ghim")

	page, err := svc.ListChessTagItems(context.Background(), 1, "ghim", "", 1, 20)
	if err != nil {
		t.Fatalf("ListChessTagItems lỗi: %v", err)
	}
	if page.Total != 3 {
		t.Errorf("tổng = %d, muốn 3", page.Total)
	}
	if page.ByType[types.ChessRefTypeArticle] != 2 || page.ByType[types.ChessRefTypePosition] != 1 {
		t.Errorf("đếm theo loại sai: %+v", page.ByType)
	}
	// Các mục là ID giả không tra được thực thể → describeChessItem bỏ qua.
	// Điểm quan trọng: tổng số và thống kê theo loại vẫn ĐÚNG, trang không vỡ.
	if len(page.Items) != 0 {
		t.Errorf("mục mồ côi phải bị bỏ qua khi hiển thị, nhận %d", len(page.Items))
	}
}

// Tra bằng tên có dấu cũng phải ra đúng thẻ (handler chỉ hạ chữ thường; việc
// khử dấu do service làm).
func TestListChessTagItems_ResolvesSlugFromAccentedName(t *testing.T) {
	svc, _ := newTagSvc()
	mustSetTags(t, svc, types.ChessRefTypeArticle, "a1", "Khai cuộc")
	page, err := svc.ListChessTagItems(context.Background(), 1, "Khai cuộc", "", 1, 20)
	if err != nil {
		t.Fatalf("ListChessTagItems lỗi: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("tổng = %d, muốn 1", page.Total)
	}
}
