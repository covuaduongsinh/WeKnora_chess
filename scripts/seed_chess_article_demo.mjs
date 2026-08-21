#!/usr/bin/env node
// seed_chess_article_demo.mjs — Tạo 1 bài viết DEMO trong "Ngân hàng bài
// viết" (khái niệm "Ghim" — Pin) kèm bí danh + wikilink trỏ ra một thế cờ
// minh họa, để kiểm tra tay: CRUD, resolve bí danh, wikilink 2 chiều,
// backlinks, trang in (ArticlePrint.vue). Xem .claude/memory/04-nhat-ky-tuy-bien.md
// mục "Ngân hàng bài viết".
//
// Dùng Node (fetch) để giữ UTF-8 sạch end-to-end — TRÁNH lỗi curl --data làm
// hỏng tiếng Việt trên Git-Bash/Windows (cùng mẫu seed_chess_wikilink_demo.mjs
// / seed_chess_book_print_demo.mjs).
//
// Bài viết tạo ra ở status="draft" (Bản thảo) — KHÔNG bị index vào KB "Tri
// thức cờ vua" (chỉ status="published" mới index), để dữ liệu demo không lẫn
// vào RAG. Xoá thử nghiệm xong bằng nút Xoá trong Ngân hàng bài viết, hoặc:
//   curl -X DELETE -H "X-API-Key: $API_KEY" $BASE_URL/api/v1/chess/articles/<id>
//
// YÊU CẦU: Node >= 18; backend đã migrate 000910 (chess_articles); tài
// khoản/API key có vai trò Contributor trở lên (tạo bài viết cần Contributor).
// CHẠY:
//   BASE_URL=http://localhost API_KEY='<tenant api key>' \
//   node scripts/seed_chess_article_demo.mjs

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

// Thế cờ minh họa: Mã trắng ghim Hậu đen vào Vua đen (ghim tuyệt đối) — thế cờ
// giản lược, đủ để minh họa khái niệm mà không cần dựng cả ván.
const positionFEN = '4k3/8/8/8/8/3n4/2Q5/4K3 w - - 0 1';

const articleContent = [
  '**Ghim (Pin)** là một chiến thuật cơ bản: một quân của bạn tấn công một quân đối phương',
  'sao cho nếu quân đó di chuyển, một quân giá trị hơn (hoặc Vua) phía sau nó sẽ bị lộ ra',
  'và bị ăn/chiếu.',
  '',
  '- **Ghim tuyệt đối (absolute pin)** — quân bị ghim là chính Vua đối phương phía sau; quân',
  '  đứng trước KHÔNG ĐƯỢC PHÉP di chuyển (luật cờ cấm để lộ Vua đang bị chiếu).',
  '- **Ghim tương đối (relative pin)** — quân phía sau không phải Vua, mà là quân giá trị',
  '  hơn (vd Hậu, Xe). Quân bị ghim VẪN ĐƯỢC PHÉP di chuyển, nhưng thường là nước xấu.',
  '',
  '```chess',
  positionFEN,
  '```',
  '',
  'Xem thêm ván minh họa ghim trong thực chiến: [[game/|]] (thay bằng slug ván thật khi có),',
  'hoặc kỹ thuật liên quan: [[article/xien|Xiên (Skewer)]] — đòn "ghim ngược".',
].join('\n');

async function main() {
  log('Tạo bài viết demo');
  const article = await api('POST', '/api/v1/chess/articles', {
    title: 'Ghim (Pin) là gì?',
    summary: 'Ghim là chiến thuật khiến một quân không thể di chuyển vì phía sau nó có quân giá trị hơn (hoặc Vua) sẽ bị lộ ra.',
    aliases: 'Pin, Đóng đinh',
    category: 'thuat-ngu',
    level: 'ma',
    tags: 'chien-thuat, co-ban',
    status: 'draft',
    content: articleContent,
  });
  console.log('  article id=' + article.id + '  slug=' + article.slug);

  log('XONG — cách xem');
  console.log(`1. Xem trong app:  ${BASE_URL} → Quản lý cờ vua → tab Ngân hàng bài viết → "Ghim (Pin) là gì?".
2. Thử wikilink:   gõ [[article/${article.slug}]] hoặc [[article/pin]] (bí danh) trong một bài viết/chương khác.
3. In thử ngay:    ${BASE_URL}/article-print/${article.id}  → nút "In / Lưu PDF (Ctrl+P)".
4. Dọn sau khi thử xong: mở bài trong Ngân hàng bài viết → nút Xoá,
   hoặc: curl -X DELETE -H "X-API-Key: $API_KEY" ${BASE_URL}/api/v1/chess/articles/${article.id}`);
}

main().catch((e) => { console.error('SEED FAILED:', e.message); process.exit(1); });
