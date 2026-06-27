import type { ChessBoardData } from '@/types/tool-results';

// Tiện ích bóc tách khối ``` chess ``` trong nội dung câu trả lời để render bàn
// cờ tương tác. Dùng hàm THUẦN (không đụng DOM) — botmsg.vue render bàn cờ bằng
// component Vue thật trong template, nằm NGOÀI vùng v-stable-html, nên không bị
// directive đó morph/ghi đè.

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

// Khớp một khối mã fenced ```chess ... ``` (đã đóng).
const CHESS_BLOCK_RE = /```[ \t]*chess[^\n]*\n([\s\S]*?)```/gi;

function looksLikePGN(text: string): boolean {
  return /\[\s*\w+\s+"[^"]*"\s*\]/.test(text) || /(^|\s)\d+\.(\.\.)?\s*\S/.test(text);
}

function parseChessSource(src: string): ChessBoardData | null {
  const text = (src || '').trim();
  if (!text) return null;
  if (looksLikePGN(text)) {
    return { display_type: 'chess_board', fen: START_FEN, pgn: text };
  }
  const firstLine = text.split('\n').map((s) => s.trim()).find(Boolean) || '';
  if (firstLine.includes('/')) {
    return { display_type: 'chess_board', fen: firstLine };
  }
  return null;
}

/** Bóc tất cả khối ```chess đã đóng trong markdown thành danh sách dữ liệu bàn cờ. */
export function extractChessBlocks(markdown: string): ChessBoardData[] {
  if (!markdown || typeof markdown !== 'string' || !markdown.includes('chess')) return [];
  const boards: ChessBoardData[] = [];
  CHESS_BLOCK_RE.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = CHESS_BLOCK_RE.exec(markdown)) !== null) {
    const data = parseChessSource(m[1]);
    if (data) boards.push(data);
  }
  return boards;
}

/** Loại bỏ các khối ```chess đã đóng khỏi markdown (để không hiện thành code block trùng). */
export function stripChessBlocks(markdown: string): string {
  if (!markdown || typeof markdown !== 'string' || !markdown.includes('chess')) return markdown;
  return markdown.replace(CHESS_BLOCK_RE, '').replace(/\n{3,}/g, '\n\n');
}
