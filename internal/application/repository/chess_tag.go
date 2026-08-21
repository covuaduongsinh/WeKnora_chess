package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

// chess_tag.go bổ sung thao tác lưu trữ cho HỆ THẺ THỐNG NHẤT (chess_tags +
// pivot đa hình chess_tag_items) vào CÙNG struct chessLibraryRepository —
// không phải repository riêng, để container.go không cần Provide mới (đúng
// khuôn chess_article_topic.go / chess_book.go).
//
// Bảng nối là NGUỒN SỰ THẬT của việc gắn thẻ. Ba cột CSV `tags` cũ
// (chess_positions/chess_books/chess_articles) chỉ là bản hiển thị, được tầng
// service ghi lại TỪ pivot — không bao giờ đọc ngược.

// ---- Từ điển thẻ ----

func (r *chessLibraryRepository) tagQuery(ctx context.Context, tenantID uint64, f types.ChessTagFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&types.ChessTag{}).Where("tenant_id = ?", tenantID)
	if f.Kind != "" {
		q = q.Where("kind = ?", f.Kind)
	}
	if f.OnlyUsed {
		q = q.Where("usage_count > 0")
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("slug ILIKE ? OR name ILIKE ?", kw, kw)
	}
	return q
}

// ListTags liệt kê thẻ. Thứ tự: thẻ NHÓM trước (kind='group' — từ vựng đóng,
// luôn muốn thấy đầu bảng) rồi tới thẻ tự do xếp theo mức dùng giảm dần, để
// thẻ hay dùng nổi lên trên trong đám mây thẻ và ô gợi ý.
func (r *chessLibraryRepository) ListTags(ctx context.Context, tenantID uint64, f types.ChessTagFilter) ([]*types.ChessTag, error) {
	var tags []*types.ChessTag
	err := r.tagQuery(ctx, tenantID, f).
		Order("CASE WHEN kind = 'group' THEN 0 ELSE 1 END ASC, sort_order ASC, usage_count DESC, name ASC").
		Find(&tags).Error
	return tags, err
}

func (r *chessLibraryRepository) GetTag(ctx context.Context, tenantID uint64, id string) (*types.ChessTag, error) {
	var t types.ChessTag
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *chessLibraryRepository) GetTagBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessTag, error) {
	var t types.ChessTag
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTagsBySlugs tra nhiều thẻ trong MỘT truy vấn — dùng khi lưu danh sách thẻ
// của một mục (tránh N truy vấn cho N thẻ).
func (r *chessLibraryRepository) GetTagsBySlugs(ctx context.Context, tenantID uint64, slugs []string) ([]*types.ChessTag, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	var tags []*types.ChessTag
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND slug IN ?", tenantID, slugs).Find(&tags).Error
	return tags, err
}

func (r *chessLibraryRepository) CreateTag(ctx context.Context, tag *types.ChessTag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// UpdateTag cố ý KHÔNG đụng slug (đổi slug đi qua UpdateTagSlug, vì đổi slug
// có thể phải GỘP với thẻ sẵn có — nghiệp vụ nằm ở tầng service) và KHÔNG đụng
// usage_count (cache, chỉ RecountTagUsage được ghi).
func (r *chessLibraryRepository) UpdateTag(ctx context.Context, tag *types.ChessTag) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessTag{}).
		Where("tenant_id = ? AND id = ?", tag.TenantID, tag.ID).
		Updates(map[string]interface{}{
			"name": tag.Name, "description": tag.Description,
			"color": tag.Color, "sort_order": tag.SortOrder,
		}).Error
}

func (r *chessLibraryRepository) UpdateTagSlug(ctx context.Context, tenantID uint64, id, slug string) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessTag{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("slug", slug).Error
}

func (r *chessLibraryRepository) DeleteTag(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.ChessTag{}).Error
}

func (r *chessLibraryRepository) TagSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessTag{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

// RecountTagUsage đồng bộ lại cache usage_count từ NGUỒN THẬT (chess_tag_items)
// cho các tagID truyền vào; tagIDs rỗng = đếm lại TOÀN BỘ thẻ của tenant.
// Một câu UPDATE có subquery tương quan, không lặp trong Go.
//
// Viết SQL thẳng thay vì dựng bằng GORM: Update() nhận subquery là vùng xám
// giữa các phiên bản GORM, còn câu này chạy nguyên vẹn trên cả Postgres lẫn
// SQLite ("lite").
func (r *chessLibraryRepository) RecountTagUsage(ctx context.Context, tenantID uint64, tagIDs []string) error {
	sql := `UPDATE chess_tags SET usage_count = (
			SELECT COUNT(*) FROM chess_tag_items ti
			WHERE ti.tenant_id = chess_tags.tenant_id AND ti.tag_id = chess_tags.id
		) WHERE tenant_id = ?`
	args := []interface{}{tenantID}
	if len(tagIDs) > 0 {
		sql += " AND id IN ?"
		args = append(args, tagIDs)
	}
	return r.db.WithContext(ctx).Exec(sql, args...).Error
}

// ---- Nối thẻ với nội dung (đa hình) ----

// SetTagsFor GHI ĐÈ toàn bộ thẻ của MỘT mục nội dung theo đúng thứ tự tagIDs
// (xóa-rồi-chèn-lại trong transaction) — cùng khuôn SetTopicArticles/SetShelfBooks.
func (r *chessLibraryRepository) SetTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND chess_type = ? AND chess_id = ?", tenantID, chessType, chessID).
			Delete(&types.ChessTagItem{}).Error; err != nil {
			return err
		}
		rows := make([]types.ChessTagItem, 0, len(tagIDs))
		seen := make(map[string]bool, len(tagIDs))
		for i, tid := range tagIDs {
			if tid == "" || seen[tid] {
				continue
			}
			seen[tid] = true
			rows = append(rows, types.ChessTagItem{
				TenantID: tenantID, TagID: tid, ChessType: chessType, ChessID: chessID, SortOrder: i,
			})
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

// RemoveAllTagsFor gỡ MỘT mục nội dung khỏi mọi thẻ (khi xóa mục đó).
// BẮT BUỘC gọi trong mọi đường Delete* — quên là để lại liên kết mồ côi và
// usage_count sai vĩnh viễn.
func (r *chessLibraryRepository) RemoveAllTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND chess_type = ? AND chess_id = ?", tenantID, chessType, chessID).
		Delete(&types.ChessTagItem{}).Error
}

// RemoveTagItems xóa toàn bộ liên kết của MỘT thẻ (khi xóa thẻ) — KHÔNG đụng nội dung.
func (r *chessLibraryRepository) RemoveTagItems(ctx context.Context, tenantID uint64, tagID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Delete(&types.ChessTagItem{}).Error
}

// MergeTagItems trỏ mọi liên kết của fromTagID sang toTagID (gộp thẻ).
// Bỏ trước các liên kết mà mục đã mang SẴN thẻ đích, nếu không câu UPDATE sẽ
// vi phạm khóa chính (tag_id, chess_type, chess_id).
func (r *chessLibraryRepository) MergeTagItems(ctx context.Context, tenantID uint64, fromTagID, toTagID string) error {
	if fromTagID == toTagID {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// EXISTS tương quan thay vì row-value "(a,b) IN (SELECT ...)": row
		// value chỉ có ở SQLite từ 3.15 và cách GORM render nó không chắc
		// chắn. Cũng KHÔNG đặt bí danh bảng sau DELETE FROM (SQLite không
		// nhận), mà tham chiếu thẳng tên bảng.
		const dedupe = `DELETE FROM chess_tag_items
			WHERE tenant_id = ? AND tag_id = ?
			  AND EXISTS (SELECT 1 FROM chess_tag_items b
			              WHERE b.tenant_id = chess_tag_items.tenant_id
			                AND b.tag_id = ?
			                AND b.chess_type = chess_tag_items.chess_type
			                AND b.chess_id = chess_tag_items.chess_id)`
		if err := tx.Exec(dedupe, tenantID, fromTagID, toTagID).Error; err != nil {
			return err
		}
		return tx.Model(&types.ChessTagItem{}).
			Where("tenant_id = ? AND tag_id = ?", tenantID, fromTagID).
			Update("tag_id", toTagID).Error
	})
}

// tagWithOwner là dòng kết quả của TagsForMany (thẻ + mục đang mang thẻ).
type tagWithOwner struct {
	types.ChessTag
	OwnerID string `gorm:"column:owner_id"`
}

// TagsForMany trả thẻ của NHIỀU mục cùng loại trong MỘT truy vấn, khóa theo
// chess_id — dùng để đính chip thẻ vào một trang danh sách mà không N+1.
func (r *chessLibraryRepository) TagsForMany(ctx context.Context, tenantID uint64, chessType string, chessIDs []string) (map[string][]*types.ChessTag, error) {
	out := make(map[string][]*types.ChessTag, len(chessIDs))
	if len(chessIDs) == 0 {
		return out, nil
	}
	var rows []tagWithOwner
	err := r.db.WithContext(ctx).Model(&types.ChessTagItem{}).
		Select("chess_tags.*, chess_tag_items.chess_id AS owner_id").
		Joins("JOIN chess_tags ON chess_tags.id = chess_tag_items.tag_id").
		Where("chess_tag_items.tenant_id = ? AND chess_tag_items.chess_type = ? AND chess_tag_items.chess_id IN ?",
			tenantID, chessType, chessIDs).
		Order("chess_tag_items.sort_order ASC, chess_tags.name ASC").
		Scan(&rows).Error
	if err != nil {
		return out, err
	}
	for i := range rows {
		tag := rows[i].ChessTag
		out[rows[i].OwnerID] = append(out[rows[i].OwnerID], &tag)
	}
	return out, nil
}

// applyTagFilter thêm mệnh đề lọc theo thẻ vào một truy vấn danh sách nội
// dung. Dùng subquery trên pivot thay vì JOIN để không nhân bản hàng khi mục
// mang nhiều thẻ (JOIN sẽ làm hỏng cả LIMIT lẫn số đếm).
//
// Mode "all": GROUP BY + HAVING COUNT(DISTINCT slug) = số thẻ yêu cầu — mục
// phải mang ĐỦ mọi thẻ. Mode "any" (mặc định): chỉ cần khớp một.
//
// idColumn PHẢI là tên cột ĐỦ TÊN BẢNG (vd "chess_articles.id"): một số truy
// vấn danh sách đã có JOIN sẵn (lọc bài viết theo chuyên mục dùng JOIN
// chess_article_topic_items) nên "id" trần sẽ mập mờ và Postgres báo lỗi.
func (r *chessLibraryRepository) applyTagFilter(q *gorm.DB, tenantID uint64, chessType, idColumn string, sel types.ChessTagSelector) *gorm.DB {
	if !sel.Active() {
		return q
	}
	sub := r.db.Model(&types.ChessTagItem{}).
		Select("chess_tag_items.chess_id").
		Joins("JOIN chess_tags ON chess_tags.id = chess_tag_items.tag_id").
		Where("chess_tag_items.tenant_id = ? AND chess_tag_items.chess_type = ? AND chess_tags.slug IN ?",
			tenantID, chessType, sel.TagSlugs)
	if sel.MatchAll() {
		sub = sub.Group("chess_tag_items.chess_id").
			Having("COUNT(DISTINCT chess_tags.slug) = ?", len(sel.TagSlugs))
	}
	return q.Where(idColumn+" IN (?)", sub)
}

// ---- Tra nội dung theo thẻ (mục lục ngang xuyên loại) ----

// CountTagItems đếm số mục mang một thẻ; chessType rỗng = mọi loại.
func (r *chessLibraryRepository) CountTagItems(ctx context.Context, tenantID uint64, tagID, chessType string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&types.ChessTagItem{}).
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID)
	if chessType != "" {
		q = q.Where("chess_type = ?", chessType)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// typeCount là dòng kết quả của CountTagItemsByType.
type typeCount struct {
	ChessType string `gorm:"column:chess_type"`
	N         int64  `gorm:"column:n"`
}

// CountTagItemsByType đếm số mục mang một thẻ, tách theo loại nội dung — để
// giao diện hiện "sách 3 · bài viết 12 · ván 5" trước khi người dùng lọc.
func (r *chessLibraryRepository) CountTagItemsByType(ctx context.Context, tenantID uint64, tagID string) (map[string]int64, error) {
	out := map[string]int64{}
	var rows []typeCount
	err := r.db.WithContext(ctx).Model(&types.ChessTagItem{}).
		Select("chess_type, COUNT(*) AS n").
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Group("chess_type").Scan(&rows).Error
	if err != nil {
		return out, err
	}
	for _, r := range rows {
		out[r.ChessType] = r.N
	}
	return out, nil
}

// ListTagItems trả một TRANG liên kết của một thẻ (đã phân trang thật, khác
// mọi list cờ hiện tại vốn cắt cứng ở 500 không báo). chessType rỗng = mọi loại.
func (r *chessLibraryRepository) ListTagItems(ctx context.Context, tenantID uint64, tagID, chessType string, offset, limit int) ([]*types.ChessTagItem, error) {
	q := r.db.WithContext(ctx).Model(&types.ChessTagItem{}).
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID)
	if chessType != "" {
		q = q.Where("chess_type = ?", chessType)
	}
	var items []*types.ChessTagItem
	err := q.Order("chess_type ASC, sort_order ASC, created_at DESC").
		Offset(offset).Limit(limit).Find(&items).Error
	return items, err
}

// ---- Cột hiển thị `tags` (CSV) trên 3 bảng cũ ----

// chessTagCSVTables ánh xạ loại nội dung sang bảng CÓ cột `tags` CSV. Chỉ 3
// loại có cột này (thêm từ migration 000070/000071/000072); 5 loại còn lại
// hoàn toàn dựa vào pivot.
var chessTagCSVTables = map[string]string{
	types.ChessRefTypePosition: "chess_positions",
	types.ChessRefTypeBook:     "chess_books",
	types.ChessRefTypeArticle:  "chess_articles",
}

// UpdateEntityTagsCSV ghi lại cột hiển thị `tags` cho một mục sau khi đã
// chuẩn hóa qua hệ thẻ. Loại không có cột này thì no-op (KHÔNG phải lỗi).
//
// Cột CSV là BẢN SAO HIỂN THỊ, không phải nguồn sự thật — chỉ hàm này được
// ghi vào nó, và luôn ghi từ dữ liệu lấy ra từ chess_tag_items.
func (r *chessLibraryRepository) UpdateEntityTagsCSV(ctx context.Context, tenantID uint64, chessType, chessID, csv string) error {
	table, ok := chessTagCSVTables[chessType]
	if !ok {
		return nil
	}
	return r.db.WithContext(ctx).Table(table).
		Where("tenant_id = ? AND id = ?", tenantID, chessID).
		Update("tags", csv).Error
}
