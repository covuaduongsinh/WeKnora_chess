#!/usr/bin/env node
// seed_chess_wikilink_demo.mjs — Tạo dữ liệu DEMO cho "wikilink cho thế cờ / ván
// cờ / bài giảng". Dùng Node (fetch) để giữ UTF-8 sạch end-to-end — TRÁNH lỗi
// curl --data làm hỏng tiếng Việt trên Git-Bash/Windows.
//
// YÊU CẦU: Node >= 18; backend đã migrate 000064 + 000065.
// CHẠY:
//   BASE_URL=http://localhost API_KEY='<tenant api key>' \
//   WIKI_KB_ID='<id KB bật wiki — tuỳ chọn>' \
//   node scripts/seed_chess_wikilink_demo.mjs

const BASE_URL = process.env.BASE_URL || 'http://localhost';
const API_KEY = process.env.API_KEY || '';
const TOKEN = process.env.TOKEN || '';
const WIKI_KB_ID = process.env.WIKI_KB_ID || '';

if (!API_KEY && !TOKEN) {
  console.error('ERROR: cần API_KEY hoặc TOKEN.');
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
    throw new Error(`${method} ${path} -> ${res.status} ${text.slice(0, 120)}`);
  }
  return json.data;
}

const log = (m) => console.log('\x1b[1;36m== ' + m + ' ==\x1b[0m');

async function main() {
  // 1) Ván cờ
  log('Tạo ván cờ demo');
  const g1 = await api('POST', '/api/v1/chess/games', {
    white: 'Paul Morphy', black: 'Duke Karl / Count Isouard', result: '1-0',
    eco: 'C41', event: 'Paris Opera', date: '1858.06.21',
    pgn: '1.e4 e5 2.Nf3 d6 3.d4 Bg4 4.dxe5 Bxf3 5.Qxf3 dxe5 6.Bc4 Nf6 7.Qb3 Qe7 8.Nc3 c6 9.Bg5 b5 10.Nxb5 cxb5 11.Bxb5+ Nbd7 12.O-O-O Rd8 13.Rxd7 Rxd7 14.Rd1 Qe6 15.Bxd7+ Nxd7 16.Qb8+ Nxb8 17.Rd8# 1-0',
  });
  const g2 = await api('POST', '/api/v1/chess/games', {
    white: 'Học trò', black: 'Tập sự', result: '1-0', event: 'Ván mẫu',
    pgn: '1.e4 e5 2.Bc4 Nc6 3.Qh5 Nf6 4.Qxf7# 1-0',
  });
  console.log('  game/' + g1.slug + '\n  game/' + g2.slug);

  // 2) Thế cờ / bài tập
  log('Tạo thế cờ / bài tập demo');
  const p1 = await api('POST', '/api/v1/chess/puzzles', {
    title: 'Chiếu bí hàng ngang', fen: '6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1',
    solution: 'Ra8#', theme: 'chiếu hết', difficulty: 'de',
  });
  const p2 = await api('POST', '/api/v1/chess/puzzles', {
    title: 'Thế cờ khai cuộc Ý', fen: 'r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4',
    solution: 'Ng5', theme: 'khai cuộc', difficulty: 'trung-binh',
  });
  console.log('  puzzle/' + p1.slug + '\n  puzzle/' + p2.slug);

  // 3) Khóa học + bài giảng (chip + nhúng)
  log('Tạo khóa học + bài giảng demo');
  const course = await api('POST', '/api/v1/chess/courses', {
    title: 'DEMO — Wikilink cờ vua',
    description: 'Minh hoạ chip & nhúng cho ván/thế cờ/bài giảng', level: 'co-ban',
  });
  const l1 = await api('POST', `/api/v1/chess/courses/${course.id}/lessons`, {
    title: 'Bài 1 — Chip vs Nhúng', sort_order: 0,
    content: [
      '## Chip vs Nhúng',
      '',
      `**Chip** (bấm để mở popup bàn cờ): ván nổi tiếng [[game/${g1.slug}|Ván Opera 1858]] và thế cờ [[puzzle/${p1.slug}|Chiếu bí hàng ngang]].`,
      '',
      '**Nhúng** (bàn cờ tương tác ngay trong bài) — thêm dấu `!` phía trước:',
      '',
      `![[game/${g1.slug}]]`,
      '',
      'Một thế cờ chiến thuật để luyện:',
      '',
      `![[puzzle/${p2.slug}]]`,
    ].join('\n'),
  });
  await api('POST', `/api/v1/chess/courses/${course.id}/lessons`, {
    title: 'Bài 2 — Liên kết bài giảng', sort_order: 1,
    content: [
      '## Liên kết bài giảng',
      '',
      `Xem lại [[lesson/${l1.slug}|Bài 1 — Chip vs Nhúng]] trước khi làm bài tập.`,
      '',
      `Bài này thuộc khóa: [[course/${course.slug}|DEMO — Wikilink cờ vua]]`,
      '',
      `Ván cờ chiếu bí nhanh: ![[game/${g2.slug}]]`,
    ].join('\n'),
  });
  console.log('  course id=' + course.id + '  (Bài 1 = lesson/' + l1.slug + ')');

  // 4) (Tùy chọn) Trang wiki tham chiếu cờ → backlink + đồ thị
  if (WIKI_KB_ID) {
    log('Tạo trang wiki demo (backlink + đồ thị) trong KB ' + WIKI_KB_ID);
    await api('POST', `/api/v1/knowledgebase/${WIKI_KB_ID}/wiki/pages`, {
      slug: 'concept/demo-co-vua', title: 'Demo cờ vua (wikilink)',
      page_type: 'concept', status: 'published',
      content: [
        '# Khai cuộc & chiến thuật (demo)',
        '',
        'Trang wiki này tham chiếu trực tiếp tới đối tượng cờ:',
        '',
        `- Ván minh hoạ: [[game/${g1.slug}|Ván Opera]]`,
        `- Bài tập: [[puzzle/${p1.slug}]]`,
        `- Bài giảng liên quan: [[lesson/${l1.slug}]]`,
        `- Khóa học: [[course/${course.slug}|DEMO — Wikilink cờ vua]]`,
        '',
        'Bàn cờ nhúng ngay trong trang wiki:',
        '',
        `![[game/${g1.slug}]]`,
      ].join('\n'),
    });
    console.log('  wiki page: concept/demo-co-vua');
  }

  log('XONG — cách xem từng trường hợp');
  console.log(`1. Chip & Nhúng (bài giảng):  ${BASE_URL} → Quản lý cờ vua → tab Khóa học → "DEMO — Wikilink cờ vua" → Bài 1.
2. Liên kết bài giảng:        mở "Bài 2" → chip [[lesson/..]] mở popup bài giảng.
3. Sao chép wikilink:         tab Kho ván / Ngân hàng bài tập → nút 🔗 mỗi dòng.
4. Bộ chọn chèn:              sửa một bài giảng → nút "Chèn ván/thế cờ".` +
    (WIKI_KB_ID ? `
5. Trong trang wiki:          KB "Bách khoa Cờ vua" → Wiki → "Demo cờ vua (wikilink)": chip + bàn cờ nhúng; popup chip hiện "Được liên kết bởi".
6. Đồ thị:                    KB → Wiki → Graph: node cờ (màu riêng) nối từ trang, bấm mở bàn cờ.` : ''));
}

main().catch((e) => { console.error('SEED FAILED:', e.message); process.exit(1); });
