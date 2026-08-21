package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Tencent/WeKnora/internal/types"
)

// chess_library_tag.go bổ sung HỆ THẺ THỐNG NHẤT vào CÙNG struct
// chessLibraryService — thẻ là trục phân loại NGANG duy nhất phủ cả 8 loại
// nội dung cờ, không phải service riêng (container.go 0 dòng thay đổi).
//
// NGUỒN SỰ THẬT là bảng nối chess_tag_items. Ba cột CSV `tags` cũ
// (chess_positions/chess_books/chess_articles) trở thành BẢN HIỂN THỊ: sau
// mỗi lần gắn thẻ, service ghi lại CSV đã CHUẨN HÓA (tên hiển thị của thẻ đã
// resolve) vào cột đó. Nhờ vậy gõ "khai cuoc" sẽ tự hiện lại thành
// "Khai cuộc" — cùng khuôn chess_articles.aliases ↔ chess_slug_aliases.
//
// Chuẩn hóa tên thẻ dùng lại slugifyChess (chess_slug.go) vốn đã khử dấu
// tiếng Việt qua foldVN, nên "Khai cuộc" / "khai-cuoc" / "KHAI CUOC" quy về
// CÙNG một thẻ. Đây là lý do backfill nằm ở Go chứ không ở SQL: viết lại phép
// khử dấu bằng SQL sẽ nhân đôi logic và chắc chắn trôi lệch.

// chessTagName là một tên thẻ đã chuẩn hóa: Slug để so khớp, Name để hiển thị.
type chessTagName struct {
	Slug string
	Name string
}

// splitChessTagInput tách đầu vào thẻ của người dùng. Nhận cả CSV
// ("Ghim, Khai cuộc") lẫn danh sách phần tử rời, vì cột cũ lưu CSV còn giao
// diện mới gửi mảng.
func splitChessTagInput(raw []string) []chessTagName {
	out := make([]chessTagName, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, chunk := range raw {
		for _, piece := range strings.Split(chunk, ",") {
			name := strings.TrimSpace(piece)
			slug := slugifyChess(name)
			if slug == "" || seen[slug] {
				continue
			}
			seen[slug] = true
			out = append(out, chessTagName{Slug: slug, Name: name})
		}
	}
	return out
}

// chessTagsCSV dựng lại chuỗi CSV hiển thị từ danh sách thẻ đã resolve.
func chessTagsCSV(tags []*types.ChessTag) string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		if t != nil && t.Name != "" {
			names = append(names, t.Name)
		}
	}
	return strings.Join(names, ", ")
}

// ---- Từ điển thẻ ----

// EnsureChessTagGroups tạo 8 thẻ NHÓM NỘI DUNG dựng sẵn nếu tenant chưa có.
// Idempotent — gọi lại chỉ tạo phần còn thiếu. Trả số thẻ vừa tạo.
func (s *chessLibraryService) EnsureChessTagGroups(ctx context.Context, tenantID uint64) (int, error) {
	created := 0
	for _, seed := range types.ChessTagGroupSeeds {
		exists, err := s.repo.TagSlugExists(ctx, tenantID, seed.Slug)
		if err != nil {
			return created, err
		}
		if exists {
			continue
		}
		tag := &types.ChessTag{
			ID: uuid.New().String(), TenantID: tenantID,
			Slug: seed.Slug, Name: seed.Name, Kind: types.ChessTagKindGroup,
			Description: seed.Description, SortOrder: seed.SortOrder,
		}
		if err := s.repo.CreateTag(ctx, tag); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

// ListChessTags liệt kê thẻ, tự seed nhóm nội dung ở lần gọi đầu tiên để
// tenant mới không thấy bảng thẻ rỗng trơn (seed lỗi thì bỏ qua, vẫn trả
// danh sách — không chặn đường đọc vì một việc phụ trợ).
func (s *chessLibraryService) ListChessTags(ctx context.Context, tenantID uint64, f types.ChessTagFilter) ([]*types.ChessTag, error) {
	_, _ = s.EnsureChessTagGroups(ctx, tenantID)
	return s.repo.ListTags(ctx, tenantID, f)
}

func (s *chessLibraryService) GetChessTag(ctx context.Context, tenantID uint64, id string) (*types.ChessTag, error) {
	return s.repo.GetTag(ctx, tenantID, id)
}

func (s *chessLibraryService) GetChessTagBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessTag, error) {
	return s.repo.GetTagBySlug(ctx, tenantID, slugifyChess(slug))
}

// CreateChessTag tạo một thẻ mới. Trùng slug thì TRẢ VỀ thẻ sẵn có thay vì
// báo lỗi — người dùng gõ lại một thẻ đã tồn tại là chuyện thường, và mục
// đích cuối cùng (có thẻ đó để gắn) đã đạt.
func (s *chessLibraryService) CreateChessTag(ctx context.Context, tag *types.ChessTag) (*types.ChessTag, error) {
	name := strings.TrimSpace(tag.Name)
	slug := slugifyChess(name)
	if slug == "" {
		return nil, fmt.Errorf("tên thẻ không hợp lệ (cần chứa chữ hoặc số)")
	}
	if existing, err := s.repo.GetTagBySlug(ctx, tag.TenantID, slug); err == nil && existing != nil {
		return existing, nil
	}
	if tag.Kind != types.ChessTagKindGroup {
		tag.Kind = types.ChessTagKindFree
	}
	tag.ID = uuid.New().String()
	tag.Slug = slug
	tag.Name = name
	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// UpdateChessTag đổi tên/mô tả/màu của thẻ. Đổi tên có thể kéo theo đổi slug;
// nếu slug mới ĐÃ thuộc một thẻ khác thì GỘP vào thẻ đó thay vì báo lỗi trùng
// khóa — đây chính là thao tác người dùng mong đợi khi sửa "khai cuoc" thành
// "Khai cuộc" trong khi thẻ "Khai cuộc" đã tồn tại.
func (s *chessLibraryService) UpdateChessTag(ctx context.Context, tag *types.ChessTag) (*types.ChessTag, error) {
	current, err := s.repo.GetTag(ctx, tag.TenantID, tag.ID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(tag.Name)
	if name == "" {
		return nil, fmt.Errorf("tên thẻ không được để trống")
	}
	newSlug := slugifyChess(name)
	if newSlug == "" {
		return nil, fmt.Errorf("tên thẻ không hợp lệ (cần chứa chữ hoặc số)")
	}
	if newSlug != current.Slug {
		if other, err := s.repo.GetTagBySlug(ctx, tag.TenantID, newSlug); err == nil && other != nil && other.ID != current.ID {
			return s.MergeChessTags(ctx, tag.TenantID, current.ID, other.ID)
		}
		// Thẻ NHÓM giữ slug cố định: slug nhóm là từ vựng đóng, được backfill
		// và frontend tham chiếu theo hằng — đổi đi là gãy cả hai.
		if current.Kind != types.ChessTagKindGroup {
			if err := s.repo.UpdateTagSlug(ctx, tag.TenantID, current.ID, newSlug); err != nil {
				return nil, err
			}
		}
	}
	tag.Name = name
	if err := s.repo.UpdateTag(ctx, tag); err != nil {
		return nil, err
	}
	return s.repo.GetTag(ctx, tag.TenantID, tag.ID)
}

// DeleteChessTag xóa một thẻ tự do và mọi liên kết của nó. CHẶN xóa thẻ NHÓM:
// lần seed sau sẽ tạo lại thẻ đó, nên "xóa" chỉ tạo cảm giác sai và làm loạn
// thống kê.
func (s *chessLibraryService) DeleteChessTag(ctx context.Context, tenantID uint64, id string) error {
	tag, err := s.repo.GetTag(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if tag.Kind == types.ChessTagKindGroup {
		return fmt.Errorf("không thể xóa thẻ nhóm nội dung %q — đây là từ vựng hệ thống", tag.Name)
	}
	if err := s.repo.RemoveTagItems(ctx, tenantID, id); err != nil {
		return err
	}
	return s.repo.DeleteTag(ctx, tenantID, id)
}

// MergeChessTags gộp thẻ nguồn vào thẻ đích: mọi liên kết chuyển sang đích,
// thẻ nguồn bị xóa. Dùng để dọn các thẻ trùng nghĩa gõ lệch nhau.
func (s *chessLibraryService) MergeChessTags(ctx context.Context, tenantID uint64, fromID, toID string) (*types.ChessTag, error) {
	if fromID == toID {
		return s.repo.GetTag(ctx, tenantID, toID)
	}
	from, err := s.repo.GetTag(ctx, tenantID, fromID)
	if err != nil {
		return nil, err
	}
	// Chỉ cần xác nhận thẻ ĐÍCH có thật (và thuộc đúng tenant) trước khi dời
	// liên kết sang — nội dung của nó không dùng tới ở đây.
	if _, err := s.repo.GetTag(ctx, tenantID, toID); err != nil {
		return nil, err
	}
	if from.Kind == types.ChessTagKindGroup {
		return nil, fmt.Errorf("không thể gộp thẻ nhóm nội dung %q đi nơi khác", from.Name)
	}
	if err := s.repo.MergeTagItems(ctx, tenantID, fromID, toID); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteTag(ctx, tenantID, fromID); err != nil {
		return nil, err
	}
	_ = s.repo.RecountTagUsage(ctx, tenantID, []string{toID})
	// Cột CSV hiển thị của các mục vừa đổi thẻ nay lệch tên — đồng bộ lại.
	s.refreshTagCSVForTag(ctx, tenantID, toID)
	return s.repo.GetTag(ctx, tenantID, toID)
}

// RecountChessTags đếm lại toàn bộ usage_count từ bảng nối. usage_count là
// CACHE nên có thể lệch nếu một đường ghi nào đó quên gọi; đây là nút chữa.
func (s *chessLibraryService) RecountChessTags(ctx context.Context, tenantID uint64) error {
	return s.repo.RecountTagUsage(ctx, tenantID, nil)
}

// ---- Gắn thẻ cho nội dung ----

// SetChessTags GHI ĐÈ danh sách thẻ của MỘT mục nội dung (bất kỳ loại nào
// trong 8 loại). Tên thẻ chưa có sẽ được tạo mới (kind="free").
//
// Sau khi ghi pivot, hàm này ĐỒNG BỘ LẠI cột CSV hiển thị cho 3 loại có cột
// đó — nên tên thẻ trong cột luôn là tên chuẩn của từ điển, không phải chuỗi
// người dùng vừa gõ.
func (s *chessLibraryService) SetChessTags(
	ctx context.Context, tenantID uint64, chessType, chessID string, names []string,
) ([]*types.ChessTag, error) {
	if chessType == "" || chessID == "" {
		return nil, fmt.Errorf("thiếu loại nội dung hoặc mã mục cần gắn thẻ")
	}
	wanted := splitChessTagInput(names)

	// Thẻ ĐANG gắn — cần biết trước để đếm lại usage_count cho cả thẻ bị gỡ,
	// nếu không thẻ vừa bị bỏ sẽ giữ số đếm cũ vĩnh viễn.
	affected := map[string]bool{}
	if before, err := s.repo.TagsForMany(ctx, tenantID, chessType, []string{chessID}); err == nil {
		for _, t := range before[chessID] {
			affected[t.ID] = true
		}
	}

	tagIDs := make([]string, 0, len(wanted))
	resolved := make([]*types.ChessTag, 0, len(wanted))
	if len(wanted) > 0 {
		slugs := make([]string, 0, len(wanted))
		for _, w := range wanted {
			slugs = append(slugs, w.Slug)
		}
		existing, err := s.repo.GetTagsBySlugs(ctx, tenantID, slugs)
		if err != nil {
			return nil, err
		}
		bySlug := make(map[string]*types.ChessTag, len(existing))
		for _, t := range existing {
			bySlug[t.Slug] = t
		}
		// Giữ ĐÚNG thứ tự người dùng nhập (sort_order của pivot bám theo).
		for _, w := range wanted {
			tag, ok := bySlug[w.Slug]
			if !ok {
				created, err := s.CreateChessTag(ctx, &types.ChessTag{
					TenantID: tenantID, Name: w.Name, Kind: types.ChessTagKindFree,
				})
				if err != nil {
					return nil, err
				}
				tag = created
				bySlug[w.Slug] = tag
			}
			tagIDs = append(tagIDs, tag.ID)
			resolved = append(resolved, tag)
			affected[tag.ID] = true
		}
	}

	if err := s.repo.SetTagsFor(ctx, tenantID, chessType, chessID, tagIDs); err != nil {
		return nil, err
	}
	if len(affected) > 0 {
		ids := make([]string, 0, len(affected))
		for id := range affected {
			ids = append(ids, id)
		}
		_ = s.repo.RecountTagUsage(ctx, tenantID, ids)
	}
	_ = s.repo.UpdateEntityTagsCSV(ctx, tenantID, chessType, chessID, chessTagsCSV(resolved))
	return resolved, nil
}

// applyChessTags là đường vào dành cho các thực thể CÓ SẴN cột CSV `tags`
// (thế cờ/sách/bài viết): chúng vẫn nhận thẻ qua thân request cũ, và hàm này
// quy chuỗi đó về hệ thẻ rồi trả lại CSV đã CHUẨN HÓA.
//
// Caller PHẢI gán giá trị trả về ngược vào đối tượng trước khi index/trả về
// client — nếu không, client thấy chuỗi thô vừa gửi lên còn DB đã lưu bản
// chuẩn hóa, và văn bản index RAG cũng mang tên thẻ sai chính tả.
//
// Best-effort: hệ thẻ trục trặc thì trả nguyên chuỗi cũ, KHÔNG chặn việc lưu.
func (s *chessLibraryService) applyChessTags(ctx context.Context, tenantID uint64, chessType, chessID, csv string) string {
	if chessID == "" {
		return csv
	}
	tags, err := s.SetChessTags(ctx, tenantID, chessType, chessID, []string{csv})
	if err != nil {
		return csv
	}
	return chessTagsCSV(tags)
}

// removeChessTags gỡ một mục khỏi mọi thẻ (khi xóa mục). BẮT BUỘC gọi trong
// mọi đường Delete* — quên là để lại liên kết mồ côi và usage_count sai.
func (s *chessLibraryService) removeChessTags(ctx context.Context, tenantID uint64, chessType, chessID string) {
	if chessID == "" {
		return
	}
	var ids []string
	if before, err := s.repo.TagsForMany(ctx, tenantID, chessType, []string{chessID}); err == nil {
		for _, t := range before[chessID] {
			ids = append(ids, t.ID)
		}
	}
	_ = s.repo.RemoveAllTagsFor(ctx, tenantID, chessType, chessID)
	if len(ids) > 0 {
		_ = s.repo.RecountTagUsage(ctx, tenantID, ids)
	}
}

// refreshTagCSVForTag ghi lại cột CSV hiển thị cho mọi mục đang mang một thẻ
// (dùng sau khi gộp/đổi tên thẻ). Chỉ chạm 3 loại có cột CSV.
func (s *chessLibraryService) refreshTagCSVForTag(ctx context.Context, tenantID uint64, tagID string) {
	for _, chessType := range []string{types.ChessRefTypePosition, types.ChessRefTypeBook, types.ChessRefTypeArticle} {
		items, err := s.repo.ListTagItems(ctx, tenantID, tagID, chessType, 0, 500)
		if err != nil {
			continue
		}
		ids := make([]string, 0, len(items))
		for _, it := range items {
			ids = append(ids, it.ChessID)
		}
		byOwner, err := s.repo.TagsForMany(ctx, tenantID, chessType, ids)
		if err != nil {
			continue
		}
		for _, id := range ids {
			_ = s.repo.UpdateEntityTagsCSV(ctx, tenantID, chessType, id, chessTagsCSV(byOwner[id]))
		}
	}
}

// ChessTagsFor trả thẻ của NHIỀU mục cùng loại trong một truy vấn — để trang
// danh sách đính chip thẻ vào từng hàng mà không N+1.
func (s *chessLibraryService) ChessTagsFor(
	ctx context.Context, tenantID uint64, chessType string, ids []string,
) (map[string][]*types.ChessTag, error) {
	return s.repo.TagsForMany(ctx, tenantID, chessType, ids)
}

// ---- Tra nội dung theo thẻ (mục lục ngang xuyên loại) ----

// ListChessTagItems trả một TRANG nội dung mang thẻ, gộp mọi loại (hoặc lọc
// một loại). Đây là thứ biến thẻ thành công cụ tra cứu thật: bấm "Ghim" ra cả
// bài viết, thế cờ, bài tập lẫn ván minh họa trong một danh sách.
//
// Khác mọi list cờ hiện tại: CÓ phân trang thật và CÓ tổng số, nên không cắt
// câm khi nội dung nhiều lên.
func (s *chessLibraryService) ListChessTagItems(
	ctx context.Context, tenantID uint64, tagSlug, chessType string, page, pageSize int,
) (*types.ChessTagItemPage, error) {
	tag, err := s.repo.GetTagBySlug(ctx, tenantID, slugifyChess(tagSlug))
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	total, err := s.repo.CountTagItems(ctx, tenantID, tag.ID, chessType)
	if err != nil {
		return nil, err
	}
	byType, _ := s.repo.CountTagItemsByType(ctx, tenantID, tag.ID)
	if byType == nil {
		byType = map[string]int64{}
	}
	items, err := s.repo.ListTagItems(ctx, tenantID, tag.ID, chessType, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	refs := make([]types.ChessTagItemRef, 0, len(items))
	for _, it := range items {
		if ref, ok := s.describeChessItem(ctx, tenantID, it.ChessType, it.ChessID); ok {
			refs = append(refs, ref)
		}
	}
	return &types.ChessTagItemPage{
		Items: refs, Total: total, Page: page, PageSize: pageSize, ByType: byType,
	}, nil
}

// describeChessItem đọc tiêu đề/slug/cấp độ của một mục bất kỳ để hiển thị
// trong danh sách theo thẻ. Trả false khi mục đã bị xóa mà liên kết còn sót
// (liên kết mồ côi bị BỎ QUA khi hiển thị chứ không làm hỏng cả trang).
func (s *chessLibraryService) describeChessItem(
	ctx context.Context, tenantID uint64, chessType, id string,
) (types.ChessTagItemRef, bool) {
	ref := types.ChessTagItemRef{ChessType: chessType, ChessID: id}
	switch chessType {
	case types.ChessRefTypeGame:
		g, err := s.repo.GetGame(ctx, tenantID, id)
		if err != nil || g == nil {
			return ref, false
		}
		ref.Slug, ref.Level, ref.UpdatedAt = g.Slug, g.Level, g.UpdatedAt
		ref.Title = strings.TrimSpace(g.White + " – " + g.Black)
		ref.Subtitle = strings.TrimSpace(strings.Join(nonEmpty(g.Event, g.ECO, g.Result), " · "))
	case types.ChessRefTypePuzzle:
		p, err := s.repo.GetPuzzle(ctx, tenantID, id)
		if err != nil || p == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.UpdatedAt = p.Slug, p.Title, p.Level, p.UpdatedAt
		ref.Subtitle = strings.Join(nonEmpty(p.Theme, p.Difficulty), " · ")
	case types.ChessRefTypePosition:
		p, err := s.repo.GetPosition(ctx, tenantID, id)
		if err != nil || p == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.UpdatedAt = p.Slug, p.Title, p.Level, p.UpdatedAt
		ref.Subtitle = strings.Join(nonEmpty(p.Category, p.ECO), " · ")
	case types.ChessRefTypeBook:
		b, err := s.repo.GetBook(ctx, tenantID, id)
		if err != nil || b == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.Status, ref.UpdatedAt = b.Slug, b.Title, b.Level, b.Status, b.UpdatedAt
		ref.Subtitle = strings.Join(nonEmpty(b.Author, b.Phase), " · ")
	case types.ChessRefTypeChapter:
		ch, err := s.repo.GetChapter(ctx, tenantID, id)
		if err != nil || ch == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.UpdatedAt = ch.Slug, ch.Title, ch.Level, ch.UpdatedAt
		ref.Subtitle = ch.Part
		if b, err := s.repo.GetBook(ctx, tenantID, ch.BookID); err == nil && b != nil {
			ref.Subtitle = strings.Join(nonEmpty(b.Title, ch.Part), " · ")
			ref.Status = b.Status // chương thừa hưởng trạng thái xuất bản của sách
		}
	case types.ChessRefTypeArticle:
		a, err := s.repo.GetArticle(ctx, tenantID, id)
		if err != nil || a == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.Status, ref.UpdatedAt = a.Slug, a.Title, a.Level, a.Status, a.UpdatedAt
		ref.Subtitle = strings.Join(nonEmpty(a.Category, a.Summary), " · ")
	case types.ChessRefTypeCourse:
		if s.courseRepo == nil {
			return ref, false
		}
		c, err := s.courseRepo.GetCourse(ctx, tenantID, id)
		if err != nil || c == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.UpdatedAt = c.Slug, c.Title, c.Level, c.UpdatedAt
		ref.Subtitle = c.Description
	case types.ChessRefTypeLesson:
		if s.courseRepo == nil {
			return ref, false
		}
		l, err := s.courseRepo.GetLesson(ctx, tenantID, id)
		if err != nil || l == nil {
			return ref, false
		}
		ref.Slug, ref.Title, ref.Level, ref.UpdatedAt = l.Slug, l.Title, l.Level, l.UpdatedAt
		if c, err := s.courseRepo.GetCourse(ctx, tenantID, l.CourseID); err == nil && c != nil {
			ref.Subtitle = c.Title
		}
	default:
		return ref, false
	}
	if ref.Title == "" {
		ref.Title = ref.Slug
	}
	return ref, true
}

// nonEmpty lọc bỏ chuỗi rỗng/toàn khoảng trắng — dùng dựng dòng phụ đề gọn.
func nonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// ---- Backfill dữ liệu cũ ----

// chessBackfillGroupOfPositionCategory / ...OfBookPhase ánh xạ các trường
// phân loại CŨ sang thẻ NHÓM NỘI DUNG. Ánh xạ CỐ Ý thận trọng: chỉ những
// trường có nghĩa trùng khớp rõ ràng mới được suy ra, phần còn lại để trống
// và Thầy tự gắn — đoán bừa còn tệ hơn để trống.
var chessBackfillGroupOfPositionCategory = map[string]string{
	"tabiya":    types.ChessTagGroupOpening,
	"endgame":   types.ChessTagGroupEndgame,
	"structure": types.ChessTagGroupMiddlegame,
	"motif":     types.ChessTagGroupTactics,
	"model":     types.ChessTagGroupMiddlegame,
	"basic":     types.ChessTagGroupCurriculum,
}

var chessBackfillGroupOfBookPhase = map[string]string{
	"opening":    types.ChessTagGroupOpening,
	"middlegame": types.ChessTagGroupMiddlegame,
	"endgame":    types.ChessTagGroupEndgame,
	"tactics":    types.ChessTagGroupTactics,
	"rules":      types.ChessTagGroupRules,
}

// BackfillChessTags nạp dữ liệu phân loại CŨ vào hệ thẻ: tách 3 cột CSV `tags`
// thành thẻ thật, và suy ra thẻ nhóm nội dung từ position.category /
// book.phase. Chủ đề bài tập (puzzle.theme) vào làm thẻ TỰ DO vì đó là từ
// vựng mở, không phải một trong 8 nhóm.
//
// IDEMPOTENT: chạy lại nhiều lần cho cùng kết quả (SetChessTags ghi đè theo
// mục, không cộng dồn). An toàn để chạy lại sau mỗi đợt nhập liệu.
//
// GIỚI HẠN ĐÃ BIẾT: các hàm List* của tầng repository đang cắt cứng ở 500 bản
// ghi (nợ kỹ thuật sẽ xử lý ở đợt phân trang). Nếu một loại trả về đúng 500
// mục, backfill KHÔNG phủ hết và kết quả sẽ kèm cảnh báo.
func (s *chessLibraryService) BackfillChessTags(ctx context.Context, tenantID uint64) (*types.ChessTagBackfillResult, error) {
	res := &types.ChessTagBackfillResult{ByType: map[string]int{}}
	seeded, err := s.EnsureChessTagGroups(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res.GroupsSeeded = seeded

	tagsBefore, _ := s.repo.ListTags(ctx, tenantID, types.ChessTagFilter{})
	linksBefore := 0
	for _, t := range tagsBefore {
		linksBefore += t.UsageCount
	}

	const repoListCap = 500
	warn := func(kind string, n int) {
		if n >= repoListCap {
			res.Warnings = append(res.Warnings, fmt.Sprintf(
				"%s: đọc được %d mục — chạm trần %d của tầng lưu trữ, có thể còn mục CHƯA được gắn thẻ; chạy lại sau khi có phân trang",
				kind, n, repoListCap))
		}
	}

	// Thế cờ: CSV tags + suy nhóm từ category.
	positions, err := s.repo.ListPositions(ctx, tenantID, types.ChessPositionFilter{})
	if err != nil {
		return nil, err
	}
	warn("thế cờ", len(positions))
	for _, p := range positions {
		names := []string{p.Tags}
		if g, ok := chessBackfillGroupOfPositionCategory[p.Category]; ok {
			names = append(names, chessTagNameOfGroup(g))
		}
		if _, err := s.SetChessTags(ctx, tenantID, types.ChessRefTypePosition, p.ID, names); err != nil {
			return nil, err
		}
		res.ByType[types.ChessRefTypePosition]++
	}

	// Sách: CSV tags + suy nhóm từ phase.
	books, err := s.repo.ListBooks(ctx, tenantID, types.ChessBookFilter{})
	if err != nil {
		return nil, err
	}
	warn("sách", len(books))
	for _, b := range books {
		names := []string{b.Tags}
		if g, ok := chessBackfillGroupOfBookPhase[b.Phase]; ok {
			names = append(names, chessTagNameOfGroup(g))
		}
		if _, err := s.SetChessTags(ctx, tenantID, types.ChessRefTypeBook, b.ID, names); err != nil {
			return nil, err
		}
		res.ByType[types.ChessRefTypeBook]++
	}

	// Bài viết: CSV tags (category của bài viết là "thể loại trình bày" —
	// khái niệm/thuật ngữ/kinh nghiệm — KHÔNG phải nhóm nội dung, nên không suy).
	articles, err := s.repo.ListArticles(ctx, tenantID, types.ChessArticleFilter{})
	if err != nil {
		return nil, err
	}
	warn("bài viết", len(articles))
	for _, a := range articles {
		if _, err := s.SetChessTags(ctx, tenantID, types.ChessRefTypeArticle, a.ID, []string{a.Tags}); err != nil {
			return nil, err
		}
		res.ByType[types.ChessRefTypeArticle]++
	}

	// Bài tập: theme thành thẻ tự do (từ vựng mở: fork/pin/skewer/mate...).
	puzzles, err := s.repo.ListPuzzles(ctx, tenantID, types.ChessPuzzleFilter{})
	if err != nil {
		return nil, err
	}
	warn("bài tập", len(puzzles))
	for _, p := range puzzles {
		if strings.TrimSpace(p.Theme) == "" {
			continue
		}
		if _, err := s.SetChessTags(ctx, tenantID, types.ChessRefTypePuzzle, p.ID, []string{p.Theme}); err != nil {
			return nil, err
		}
		res.ByType[types.ChessRefTypePuzzle]++
	}

	_ = s.repo.RecountTagUsage(ctx, tenantID, nil)
	tagsAfter, _ := s.repo.ListTags(ctx, tenantID, types.ChessTagFilter{})
	linksAfter := 0
	for _, t := range tagsAfter {
		linksAfter += t.UsageCount
	}
	res.TagsCreated = len(tagsAfter) - len(tagsBefore)
	res.LinksCreated = linksAfter - linksBefore
	return res, nil
}

// chessTagNameOfGroup trả tên hiển thị của một thẻ nhóm theo slug.
func chessTagNameOfGroup(slug string) string {
	for _, seed := range types.ChessTagGroupSeeds {
		if seed.Slug == slug {
			return seed.Name
		}
	}
	return slug
}
