// chessTaxonomy.ts — TỪ VỰNG PHÂN LOẠI DÙNG CHUNG cho toàn bộ lớp cờ vua.
//
// Trước file này, mỗi trang cờ tự khai danh sách cấp độ của riêng nó
// (chessArticleOptions.ts, chessBookOptions.ts, chessPositionOptions.ts, và
// khai inline trong PuzzleBank.vue / GameLibrary.vue / ChessCourses.vue) —
// sáu bản sao của cùng một danh sách, chắc chắn trôi lệch theo thời gian.
//
// Ba trục phân loại thống nhất:
//   1. Cấp độ        — 6 bậc Dương Sinh, cột `level` trên mọi bảng nội dung.
//   2. Nhóm nội dung — 8 nhóm, lưu dưới dạng THẺ HỆ THỐNG (kind="group").
//   3. Thẻ tự do     — thẻ kind="free", không có từ vựng cố định.
//
// ⚠️ Bản đối ứng phía Go là internal/types/chess_tag.go (ChessTagGroupSeeds +
// hằng ChessTagGroup*). Sửa một bên PHẢI sửa bên kia, nếu không thẻ nhóm sinh
// ra ở backend sẽ không khớp nhãn/màu hiển thị ở đây.

export interface TaxonomyOption {
  label: string;
  value: string;
}

// ---- Trục 1: Cấp độ (lộ trình 6 cấp Tốt → Vua) ----
// Xem .claude/memory/01-du-an-duongsinh.md — đây là xương sống nội dung của
// Dương Sinh, dùng để agent trả lời đúng độ sâu theo trình độ người hỏi.
export const chessLevelOptions: TaxonomyOption[] = [
  { label: 'Tốt', value: 'tot' },
  { label: 'Mã', value: 'ma' },
  { label: 'Tượng', value: 'tuong' },
  { label: 'Xe', value: 'xe' },
  { label: 'Hậu', value: 'hau' },
  { label: 'Vua', value: 'vua' },
];

export function chessLevelLabel(v: string): string {
  return chessLevelOptions.find((o) => o.value === v)?.label || v;
}

// ---- Trục 2: Nhóm nội dung (8 nhóm) ----
// Bám .claude/memory/02-mien-co-vua.md mục 2.1 để phân loại nội dung khớp với
// cách tổ chức knowledge base đã thống nhất từ trước.
export const CHESS_GROUP_SLUGS = [
  'luat',
  'khai-cuoc',
  'trung-cuoc',
  'tan-cuoc',
  'chien-thuat',
  'giao-trinh',
  'van-hoa',
  'van-hanh',
] as const;

export type ChessGroupSlug = (typeof CHESS_GROUP_SLUGS)[number];

export const chessGroupOptions: TaxonomyOption[] = [
  { label: 'Luật cờ', value: 'luat' },
  { label: 'Khai cuộc', value: 'khai-cuoc' },
  { label: 'Trung cuộc', value: 'trung-cuoc' },
  { label: 'Tàn cuộc', value: 'tan-cuoc' },
  { label: 'Chiến thuật', value: 'chien-thuat' },
  { label: 'Giáo trình', value: 'giao-trinh' },
  { label: 'Văn hóa & lịch sử', value: 'van-hoa' },
  { label: 'Vận hành', value: 'van-hanh' },
];

export function chessGroupLabel(v: string): string {
  return chessGroupOptions.find((o) => o.value === v)?.label || v;
}

// isChessGroupSlug phân biệt thẻ nhóm với thẻ tự do KHI CHỈ CÓ SLUG trong tay
// (vd chuỗi lọc trên URL). Khi có cả object thẻ thì dùng `tag.kind` chính xác hơn.
export function isChessGroupSlug(slug: string): boolean {
  return (CHESS_GROUP_SLUGS as readonly string[]).includes(slug);
}

// ---- Nhãn loại nội dung (dùng cho danh sách xuyên loại theo thẻ) ----
// Giữ khớp với TYPE_LABELS trong ChessWikiLinkSuggest.vue / ChessRefMissing.vue.
export const chessTypeLabels: Record<string, string> = {
  game: 'Ván cờ',
  puzzle: 'Bài tập',
  lesson: 'Bài giảng',
  course: 'Khóa học',
  position: 'Thế cờ',
  book: 'Sách',
  chapter: 'Chương',
  article: 'Bài viết',
};

export function chessTypeLabel(v: string): string {
  return chessTypeLabels[v] || v;
}

// Đường dẫn mở một mục theo loại + slug. Tái dùng cơ chế deep-link ?ref= sẵn
// có của ChessManage.vue (nó tự chuyển tab và chọn đúng mục) thay vì dựng
// route riêng cho từng loại.
export function chessRefLink(chessType: string, slug: string): string {
  return `/platform/chess-courses?ref=${encodeURIComponent(`${chessType}/${slug}`)}`;
}
