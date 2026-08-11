// Bóc PGN có chú giải (bình luận {...}, ký hiệu đánh giá NAG $N, nhánh phụ RAV
// (...)) thành token phẳng để render — dùng CHUNG cho ChessBoardDisplay.vue (mọi
// nơi hiển thị bàn cờ) thay vì chess.js `loadPgn()` + `history()`.
//
// LÝ DO: `loadPgn()` chỉ đi theo `node.variations[0]` và chỉ đọc `node.comment` —
// NAG, suffix (!?) và TOÀN BỘ nhánh phụ bị parse rồi VỨT BỎ ÂM THẦM, không báo
// lỗi. Với sách cờ, đây chính là phần giá trị nhất của một chương (Thầy Tường
// soạn chú giải rất dài). `chess.js@1.4.0` ship kèm parser ĐẦY ĐỦ hơn —
// `node_modules/chess.js/src/pgn.js` export `parse()`, chính là hàm `loadPgn()`
// gọi bên trong (chess.ts ~dòng 2159) — nên deep-import dùng thẳng nó thay vì
// thêm thư viện PGN mới. package.json của chess.js KHÔNG có `exports` field nên
// deep import hợp lệ; `moduleResolution: "bundler"` (tsconfig) resolve được
// `pgn.js` → `pgn.d.ts` đi kèm.
//
// ⚠️ RỦI RO KHI NÂNG CẤP chess.js: đường import này không phải API công khai
// (không nằm trong `dist/`, không có trong `exports`). Nếu tương lai nâng version
// và build lỗi "Cannot find module 'chess.js/src/pgn.js'" — kiểm tra lại cấu
// trúc thư mục `src/` trong gói mới (có thể đổi tên/vị trí) trước khi coi là bug.
import { Chess } from 'chess.js';
// Deep import vào nội bộ chess.js (xem ghi chú rủi ro ở đầu file) — type resolve
// được nhờ src/pgn.d.ts nằm cùng thư mục với pgn.js, không cần @ts-expect-error.
import { parse } from 'chess.js/src/pgn.js';

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

// Kiểu shipped của chess.js (src/node.ts) khai báo `suffix?: string; nag?: string`
// nhưng RUNTIME TRẢ VỀ MẢNG (grammar `nag:nag*`, `suffix:[!?]|1..2|`) — đã xác
// minh bằng cách chạy thật. Khai báo lại tại chỗ, KHÔNG phụ thuộc type sai của
// thư viện, để tự miễn nhiễm nếu chess.js sửa type ở bản sau.
interface RawPgnNode {
  move?: string;
  suffix?: string | string[];
  nag?: string | string[];
  comment?: string;
  variations: RawPgnNode[];
}
interface RawPgnGame {
  headers: Record<string, string>;
  root: RawPgnNode;
  result?: string;
}
const parsePgn = parse as unknown as (input: string) => RawPgnGame;

/** Một thế cờ trên MAINLINE — tương thích 1-1 với PosItem cũ trong ChessBoardDisplay.vue. */
export interface AnnotatedPos {
  fen: string;
  label: string; // "6. Na4" / "6... Bb6" (không kèm NAG — giữ nguyên cho title lưu ở GameLibrary)
  san: string;
  moveNumber: number;
  side: 'w' | 'b';
}

/** Token phẳng để render bằng một v-for duy nhất. */
export type PgnToken =
  | {
      kind: 'move';
      num: string; // "7." | "7..." | "" (rỗng = không in số)
      san: string;
      glyph: string; // suffix + NAG đã gộp, vd "!", "⩲", "!?"
      posIndex: number | null; // mainline: 1-based, khớp currentIndex; nhánh phụ: null (không bấm được)
      depth: number; // 0 = mainline, 1.. = nhánh phụ
    }
  | { kind: 'comment'; text: string; depth: number }
  | { kind: 'open' | 'close'; depth: number }; // "(" và ")"

/** Bình luận ở mainline hiển thị dạng KHỐI (xuống dòng riêng); bình luận nằm
 * trong nhánh phụ hiển thị INLINE — xem CSS .move-comment / .move-var.move-comment. */
function isBlockComment(tk: PgnToken): boolean {
  return tk.kind === 'comment' && tk.depth === 0;
}

/**
 * Có cần chèn một khoảng trắng THẬT (text node) trước token này không.
 *
 * BẮT BUỘC phải là text node chứ không dùng margin: trình biên dịch Vue xoá
 * sạch whitespace giữa các thẻ anh em trong template (chế độ `condense` mặc
 * định — đã xác minh bằng cách biên dịch thật), mà `margin` thì KHÔNG tạo điểm
 * ngắt dòng. Thiếu khoảng trắng thật, cả chuỗi nước đi dài trở thành MỘT dòng
 * không ngắt được → tràn khỏi cột và (khi in) tràn ngang khỏi khổ giấy, làm
 * trình duyệt sinh thêm trang trắng.
 *
 * Quy tắc theo lối trình bày sách cờ: không có khoảng trắng trước ")" và ngay
 * sau "(" — để nhánh phụ đọc là "(7. O-O O-O 8. a3)" chứ không phải "( 7. ... )".
 */
export function needsSpaceBefore(prev: PgnToken | null, cur: PgnToken): boolean {
  if (!prev) return false;
  if (isBlockComment(prev) || isBlockComment(cur)) return false; // khối tự xuống dòng
  if (cur.kind === 'close') return false;
  if (prev.kind === 'open') return false;
  return true;
}

export interface AnnotatedPgn {
  /** KHÔNG gồm thế cờ ban đầu — caller tự chèn vào vị trí 0. */
  positions: AnnotatedPos[];
  tokens: PgnToken[];
  /** FEN thực sự dùng làm gốc (headers.FEN nếu có, không thì fallbackFen truyền vào). */
  startFen: string;
  /** true nếu mainline bị cắt giữa chừng vì chess.move() throw (nước sai luật). */
  truncated: boolean;
  /** true nếu phải parse ở tầng "đã xoá bình luận" — mất hết bình luận (còn nước + NAG + nhánh phụ). */
  degraded: boolean;
}

// ---- Bảng NAG → ký hiệu (Numeric Annotation Glyph, chuẩn PGN) ----
// $1-$6: đánh giá nước đi. $10-$21: đánh giá thế cờ. Còn lại: ít dùng, thêm dần
// khi gặp. Không biết → in nguyên "$N" (không nuốt dữ liệu của Thầy).
const NAG_GLYPH: Record<string, string> = {
  '1': '!', '2': '?', '3': '!!', '4': '??', '5': '!?', '6': '?!',
  '7': '□', '8': '□', '9': '??',
  '10': '=', '11': '=', '12': '=', '13': '∞',
  '14': '⩲', '15': '⩱', '16': '±', '17': '∓', '18': '+−', '19': '−+',
  '20': '+−−', '21': '−−+',
  '22': '⨀', '23': '⨀',
  '32': '⟳', '33': '⟳',
  '36': '↑', '37': '↑',
  '40': '→', '41': '→',
  '44': '=∞', '45': '=∞',
  '132': '⇆', '133': '⇆',
  '140': '∆',
  '146': 'N',
};

function normArr(v: string | string[] | undefined): string[] {
  if (Array.isArray(v)) return v;
  return v ? [v] : [];
}

/** Chỉ phần suffix/nag của RawPgnNode — glyphOf không cần move/comment/variations,
 * tách riêng để test gọi được bằng object literal gọn mà không phải giả lập
 * cả node PGN đầy đủ (variations là field bắt buộc trên RawPgnNode). */
export interface NagSource {
  suffix?: string | string[];
  nag?: string | string[];
}

/** Gộp suffix (!?) + NAG ($1 $14) thành một chuỗi ký hiệu, khử trùng lặp. */
export function glyphOf(n: NagSource): string {
  const parts: string[] = [];
  const suffix = normArr(n.suffix).join('');
  if (suffix) parts.push(suffix);
  for (const code of normArr(n.nag)) {
    const g = NAG_GLYPH[code] ?? `$${code}`;
    if (!parts.includes(g)) parts.push(g);
  }
  return parts.join('');
}

// [%evp ...], [%clk ...], [%eval ...], [%cal ...], [%emt ...] — dữ liệu máy
// (engine eval, đồng hồ) lẫn trong bình luận. braceComment của grammar là
// `[^}]*` nên không thể có "]" lồng bên trong ⇒ regex này an toàn.
const PGN_CMD_RE = /\[%[^\]]*\]/g;

/** Lọc dữ liệu máy khỏi bình luận thô, chỉ giữ phần chữ. Rỗng → ''. */
export function cleanComment(raw?: string): string {
  if (!raw) return '';
  return raw.replace(PGN_CMD_RE, ' ').replace(/\s+/g, ' ').trim();
}

function fmtNum(n: number, side: 'w' | 'b'): string {
  return `${n}${side === 'w' ? '.' : '...'}`;
}

/** Đẩy token comment nếu sau khi lọc còn chữ. Trả về true nếu ĐÃ đẩy (⇒ ngắt mạch số hiệu nước). */
function pushComment(out: PgnToken[], raw: string | undefined, depth: number): boolean {
  const text = cleanComment(raw);
  if (!text) return false;
  out.push({ kind: 'comment', text, depth });
  return true;
}

const MAX_VARIATION_DEPTH = 6; // chặn PGN dị dạng/lồng quá sâu, tránh đệ quy vô hạn

// Nhánh phụ: CHỮ TĨNH, không dựng bàn cờ (không cần FEN vì không bấm được được —
// tránh hẳn việc một nước sai luật trong nhánh phụ làm hỏng cả phần hiển thị).
// Chỉ đếm số học (num, side) để in đúng số hiệu nước.
function emitVariation(
  ravRoot: RawPgnNode,
  num: number,
  side: 'w' | 'b',
  depth: number,
  out: PgnToken[],
): void {
  if (depth > MAX_VARIATION_DEPTH) return;
  out.push({ kind: 'open', depth });
  let needNum = true;
  pushComment(out, ravRoot.comment, depth);

  let node = ravRoot;
  let n = num;
  let s = side;
  while (node.variations.length) {
    const next = node.variations[0];
    out.push({
      kind: 'move',
      num: s === 'w' || needNum ? fmtNum(n, s) : '',
      san: next.move ?? '',
      glyph: glyphOf(next),
      posIndex: null,
      depth,
    });
    needNum = false;
    if (pushComment(out, next.comment, depth)) needNum = true;

    // Nhánh lồng trong nhánh: thay thế cho CHÍNH `next` ⇒ dùng (n, s) hiện tại,
    // in NGAY SAU `next` (đúng thứ tự đọc: nước rồi mới tới biến thể của nó).
    for (let i = 1; i < node.variations.length; i++) {
      emitVariation(node.variations[i], n, s, depth + 1, out);
      needNum = true; // sau khi đóng ngoặc, nước tiếp theo (nếu là Đen) phải in lại số
    }
    if (s === 'b') n++;
    s = s === 'w' ? 'b' : 'w';
    node = next;
  }
  out.push({ kind: 'close', depth });
}

function walk(game: RawPgnGame, startFen: string, degraded: boolean): AnnotatedPgn | null {
  let chess: Chess;
  try {
    chess = new Chess(startFen);
  } catch {
    return null;
  }

  const positions: AnnotatedPos[] = [];
  const tokens: PgnToken[] = [];
  let truncated = false;
  let needNum = true;

  pushComment(tokens, game.root.comment, 0); // bình luận TRƯỚC nước đầu tiên (vd [%evp] mở đầu ván)

  let node = game.root;
  while (node.variations.length) {
    const next = node.variations[0];
    const num = chess.moveNumber(); // tự tính — KHÔNG tin số hiệu trong PGN (grammar không kiểm tra)
    const side = chess.turn() as 'w' | 'b';

    let san: string;
    try {
      san = chess.move(next.move ?? '').san;
    } catch {
      truncated = true; // nước sai luật giữa ván → dừng mainline, GIỮ phần đã dựng được
      break;
    }

    positions.push({
      fen: chess.fen(),
      label: `${fmtNum(num, side)} ${san}`,
      san,
      moveNumber: num,
      side,
    });
    tokens.push({
      kind: 'move',
      num: side === 'w' || needNum ? fmtNum(num, side) : '',
      san,
      glyph: glyphOf(next),
      posIndex: positions.length, // = số nửa-nước đã đi = ngữ nghĩa currentIndex hiện có
      depth: 0,
    });
    needNum = false;

    if (pushComment(tokens, next.comment, 0)) needNum = true;

    for (let i = 1; i < node.variations.length; i++) {
      emitVariation(node.variations[i], num, side, 1, tokens);
      needNum = true;
    }
    node = next;
  }

  if (!positions.length && !tokens.length) return null;
  return { positions, tokens, startFen, truncated, degraded };
}

// Xoá mọi bình luận {...} và ;... — chạy 2 lượt để bóc được ngoặc nhọn lồng 1
// tầng ("{a {b} c}" → sau lượt 1 còn "{a  c}" do khớp không tham lam từ "{a "
// tới "}" đầu tiên tức sau "b", để lại " c}" thừa — lượt 2 dọn nốt).
function stripComments(pgn: string): string {
  return pgn
    .replace(/\{[^{}]*\}/g, ' ')
    .replace(/\{[\s\S]*?\}/g, ' ')
    .replace(/;[^\r\n]*/g, ' ');
}

/**
 * Phân tích PGN có chú giải đầy đủ. Thang fallback 2 tầng:
 *   1) parse() nguyên văn — đầy đủ bình luận + NAG + nhánh phụ.
 *   2) parse() sau khi xoá bình luận — cứu được nước + NAG + nhánh phụ khi PGN
 *      vi phạm 1 trong 3 giới hạn hẹp của grammar (2 comment liên tiếp, ngoặc
 *      nhọn lồng, NAG đặt sau comment — CẢ BA đều là lỗi comment).
 * Không có tầng loadPgn(): đã xác minh loadPgn() gọi CHÍNH parse() này nên throw
 * trên đúng cùng tập input — giữ nó chỉ tạo ảo giác an toàn.
 * Trả về null nếu cả 2 tầng đều throw hoặc FEN gốc không dựng được bàn cờ — caller
 * tự rơi về hiển thị 1 FEN đơn.
 */
export function buildAnnotatedPgn(pgn: string, fallbackFen: string): AnnotatedPgn | null {
  let game: RawPgnGame | null = null;
  let degraded = false;
  try {
    game = parsePgn(pgn);
  } catch {
    /* rơi xuống tầng 2 */
  }
  if (!game) {
    try {
      game = parsePgn(stripComments(pgn));
      degraded = true;
    } catch {
      return null;
    }
  }

  const fenKey = Object.keys(game.headers ?? {}).find((k) => k.toLowerCase() === 'fen');
  const headerFen = fenKey ? (game.headers[fenKey] || '').trim() : '';

  return (
    walk(game, headerFen || fallbackFen || START_FEN, degraded) ??
    (headerFen ? walk(game, fallbackFen || START_FEN, degraded) : null)
  );
}

/** Sinh token từ danh sách thế cờ đã có sẵn (không chú giải) — dùng cho nhánh
 * props.data.plies (backend) và làm điểm chung với PGN không chú giải, để MỘT
 * bộ render token phục vụ cả 3 nguồn dữ liệu thay vì hai bộ có nguy cơ trôi lệch.
 * offset cho phép positions không bắt đầu từ nửa-nước 1 (hiện chưa dùng tới,
 * giữ tham số cho rõ ý định — mọi nơi gọi hiện tại đều offset=1 mặc định). */
export function tokensFromPositions(list: AnnotatedPos[], offset = 1): PgnToken[] {
  return list.map((p, i) => ({
    kind: 'move' as const,
    num: p.side === 'w' ? `${p.moveNumber}.` : i === 0 ? `${p.moveNumber}...` : '',
    san: p.san,
    glyph: '',
    posIndex: i + offset,
    depth: 0,
  }));
}
