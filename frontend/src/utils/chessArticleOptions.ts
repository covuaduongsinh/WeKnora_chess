// Danh sách phân loại + nhãn dùng chung cho Ngân hàng bài viết (ArticleBank.vue)
// — tách riêng để nhiều nơi không lệch mã code khi mở rộng, cùng mẫu
// utils/chessBookOptions.ts / chessPositionOptions.ts.

// category: thể loại bài viết.
export const articleCategoryOptions = [
  { label: 'Khái niệm', value: 'khai-niem' },
  { label: 'Thuật ngữ', value: 'thuat-ngu' },
  { label: 'Kinh nghiệm', value: 'kinh-nghiem' },
  { label: 'Hướng dẫn', value: 'huong-dan' },
  { label: 'Phương pháp dạy', value: 'phuong-phap-day' },
  { label: 'Nhân vật', value: 'nhan-vat' },
  { label: 'Lịch sử', value: 'lich-su' },
];

// level: 6 cấp lộ trình Dương Sinh Tốt→Vua (xem .claude/memory/01-du-an-duongsinh.md).
export const articleLevelOptions = [
  { label: 'Tốt', value: 'tot' },
  { label: 'Mã', value: 'ma' },
  { label: 'Tượng', value: 'tuong' },
  { label: 'Xe', value: 'xe' },
  { label: 'Hậu', value: 'hau' },
  { label: 'Vua', value: 'vua' },
];

// status: trạng thái xuất bản NỘI BỘ — draft KHÔNG được index vào KB tri thức cờ.
export const articleStatusOptions = [
  { label: 'Bản thảo', value: 'draft' },
  { label: 'Đã xuất bản', value: 'published' },
];

export function articleCategoryLabel(v: string): string {
  return articleCategoryOptions.find((o) => o.value === v)?.label || v;
}
export function articleLevelLabel(v: string): string {
  return articleLevelOptions.find((o) => o.value === v)?.label || v;
}
export function articleStatusLabel(v: string): string {
  return articleStatusOptions.find((o) => o.value === v)?.label || v;
}
