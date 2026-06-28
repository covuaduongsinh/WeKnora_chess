package service

// Giải mã "mềm" slug đối tượng cờ: khi [[<type>/<slug>]] không khớp chính xác,
// thử nắn về slug sống bằng cùng cơ chế đã dùng cho trang wiki (resolveDeadSlug:
// normalize hyphen/case + bigram Jaccard ≥ 0.8). Tránh "không tìm thấy tham chiếu
// cờ" khi tác giả/LLM gõ slug sai nhẹ (thiếu/thừa gạch nối, sai một vài ký tự).
//
// resolveChessSlugFuzzy KHÔNG dùng titleToSlug (đặt nil): nhãn trong
// [[game/slug|Nhãn]] do người viết tự đặt nên không đáng tin để tra ngược; chỉ
// dựa trên độ giống slug — bảo thủ, thà trả "không thấy" còn hơn trỏ nhầm.
func resolveChessSlugFuzzy(deadSlug string, liveSlugs []string) (string, bool) {
	if deadSlug == "" || len(liveSlugs) == 0 {
		return "", false
	}
	set := make(map[string]struct{}, len(liveSlugs))
	for _, s := range liveSlugs {
		if s != "" {
			set[s] = struct{}{}
		}
	}
	return resolveDeadSlug(deadSlug, "", set, nil)
}
