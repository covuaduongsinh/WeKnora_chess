#!/usr/bin/env node
// seed_chess_book_print_demo.mjs — Tạo 1 sách DEMO trong "Thư viện sách cờ vua"
// với 1 chương nhúng khối ```chess chứa PGN (ván cờ nhiều nước), để kiểm tra
// tính năng IN SÁCH (BookPrint.vue) hiển thị đủ danh sách nước đi trên giấy —
// xem .claude/memory/04-nhat-ky-tuy-bien.md mục "BookPrint hides move-list...".
//
// Dùng Node (fetch) để giữ UTF-8 sạch end-to-end — TRÁNH lỗi curl --data làm
// hỏng tiếng Việt trên Git-Bash/Windows (cùng mẫu scripts/seed_chess_wikilink_demo.mjs).
//
// Sách tạo ra ở status="draft" (Bản thảo) — KHÔNG bị index vào KB "Tri thức cờ
// vua" (chỉ status="published" mới index), để dữ liệu demo không lẫn vào RAG.
// Xoá thử nghiệm xong bằng nút Xoá trong Thư viện sách, hoặc:
//   curl -X DELETE -H "X-API-Key: $API_KEY" $BASE_URL/api/v1/chess/books/<id>
//
// YÊU CẦU: Node >= 18; backend đã migrate 000909 (chess_books); tài khoản/API
// key có vai trò Contributor trở lên (tạo sách/chương cần Contributor).
// CHẠY:
//   BASE_URL=http://localhost API_KEY='<tenant api key>' \
//   node scripts/seed_chess_book_print_demo.mjs
//
// (Đổi BASE_URL=https://weknora.covuaduongsinh.com để tạo thẳng trên production
// — chỉ làm vậy SAU KHI đã deploy bản sửa move-list/BookPrint lên đó.)

const BASE_URL = process.env.BASE_URL || 'http://localhost';
const API_KEY = process.env.API_KEY || '';
const TOKEN = process.env.TOKEN || '';

if (!API_KEY && !TOKEN) {
  console.error('ERROR: cần API_KEY hoặc TOKEN. Xem hướng dẫn ở đầu file.');
  process.exit(1);
}

const headers = { 'Content-Type': 'application/json' };
if (API_KEY) headers['X-API-Key'] = API_KEY;
else headers['Authorization'] = `Bearer ${TOKEN}`;

async function api(method, path, body) {
  const res = await fetch(BASE_URL + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let json;
  try { json = JSON.parse(text); } catch { json = null; }
  if (!res.ok || !json) {
    throw new Error(`${method} ${path} -> ${res.status} ${text.slice(0, 200)}`);
  }
  return json.data;
}

const log = (m) => console.log('\x1b[1;36m== ' + m + ' ==\x1b[0m');

// Ván Opera (Morphy, 1858) — công khai (public domain), 17 nước đầy đủ (33
// nửa-nước), đủ dài để kiểm tra move-list KHÔNG bị cắt/ẩn khi in, đủ ngắn để
// gọn trên một trang. Cùng ván đã dùng trong seed_chess_wikilink_demo.mjs.
const chapterContent = [
  '## Ván cờ kinh điển — Ván Opera (Morphy, 1858)',
  '',
  'Paul Morphy chơi ván này ngẫu hứng tại nhà hát Opera Paris, đối đầu liên minh',
  'Công tước Karl và Bá tước Isouard — mẫu mực về phát triển quân nhanh, hy sinh',
  'chất lượng để giành thế chủ động, kết thúc bằng một đòn phối hợp chiếu hết.',
  '',
  '```chess',
  '[Event "Paris Opera"]',
  '[Site "Paris"]',
  '[Date "1858.06.21"]',
  '[White "Paul Morphy"]',
  '[Black "Duke Karl / Count Isouard"]',
  '[Result "1-0"]',
  '',
  '1.e4 e5 2.Nf3 d6 3.d4 Bg4 4.dxe5 Bxf3 5.Qxf3 dxe5 6.Bc4 Nf6 7.Qb3 Qe7 8.Nc3 c6 9.Bg5 b5 10.Nxb5 cxb5 11.Bxb5+ Nbd7 12.O-O-O Rd8 13.Rxd7 Rxd7 14.Rd1 Qe6 15.Bxd7+ Nxd7 16.Qb8+ Nxb8 17.Rd8# 1-0',
  '```',
  '',
  'Trắng chiếu hết ở nước 17. Một trong những ván được phân tích nhiều nhất',
  'trong lịch sử cờ vua — thường dùng để dạy về tốc độ phát triển quân và',
  'giá trị của thời gian trong khai cuộc.',
].join('\n');

async function main() {
  log('Tạo sách demo');
  const book = await api('POST', '/api/v1/chess/books', {
    title: 'DEMO — In sách (khối PGN)',
    subtitle: 'Dữ liệu demo để kiểm tra tính năng in sách — xoá được sau khi thử xong',
    level: 'hau',
    phase: 'tactics',
    status: 'draft',
    description: 'Sách demo tạo tự động để kiểm tra BookPrint hiển thị đủ move-list khi chương nhúng khối ```chess chứa PGN.',
    tags: 'demo',
  });
  console.log('  book id=' + book.id + '  slug=' + book.slug);

  log('Tạo chương nhúng khối PGN');
  const chapter = await api('POST', `/api/v1/chess/books/${book.id}/chapters`, {
    title: 'Ván Opera (Morphy, 1858)',
    level: 'hau',
    sort_order: 0,
    content: chapterContent,
  });
  console.log('  chapter id=' + chapter.id + '  slug=' + chapter.slug);

  log('XONG — cách xem');
  console.log(`1. Xem trong app:  ${BASE_URL} → Quản lý cờ vua → tab Thư viện sách → "DEMO — In sách (khối PGN)".
2. In thử ngay:    ${BASE_URL}/book-print/${book.id}  → nút "In / Lưu PDF (Ctrl+P)".
   → Move-list "1. e4 e5  2. Nf3 d6  ..." phải hiện đủ trên trang/PDF (không bị ẩn/cắt).
3. Dọn sau khi thử xong: mở sách trong Thư viện sách → nút Xoá (xoá cascade cả chương),
   hoặc: curl -X DELETE -H "X-API-Key: $API_KEY" ${BASE_URL}/api/v1/chess/books/${book.id}`);
}

main().catch((e) => { console.error('SEED FAILED:', e.message); process.exit(1); });
