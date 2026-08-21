package modelcontext

// Model-handle policy cho 7 tool cờ vua (fork Dương Sinh).
//
// Upstream 0.7.2 đặt ra hợp đồng: MỌI tool lộ ra UI phải khai policy, kể cả
// policy RỖNG — "im lặng" bị coi là lỗi vì một tool mới thêm sẽ vô tình chọn
// đường thoát khỏi cơ chế nén handle (xem tool_policy_coverage_test.go).
//
// Đăng ký qua init() ở FILE RIÊNG thay vì sửa thẳng map trong tool_policy.go:
// giữ 0 dòng thay đổi ở file upstream nên merge sau không đụng — đúng nguyên
// tắc "code cờ để riêng" của AGENTS.md.
//
// Vì sao policy RỖNG là đúng cho cả 7 tool:
//   - Tham số đầu vào là FEN / PGN / slug cờ, KHÔNG phải knowledge_id,
//     chunk_id hay knowledge_base_id → không có sourceIDKeys nào để khai.
//   - Kết quả là đánh giá thế cờ + bàn cờ tương tác do frontend render, không
//     phải đoạn trích tài liệu để LLM trích dẫn → sourceOutput = false.
//
// Nếu sau này thêm tool cờ ĐỌC kho tri thức (vd trả về knowledge_id của sách),
// phải khai sourceIDKeys/sourceOutput tương ứng tại đây, không để rỗng.
func init() {
	for _, name := range []string{
		"chess_analyze_position",
		"chess_best_move",
		"chess_evaluate_game",
		"chess_explain_move",
		"chess_lookup_opening",
		"chess_generate_puzzle",
		"chess_lookup_position",
	} {
		if _, exists := toolHandlePolicies[name]; !exists {
			toolHandlePolicies[name] = toolHandlePolicy{}
		}
	}
}
