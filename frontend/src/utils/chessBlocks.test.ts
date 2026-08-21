import assert from 'node:assert/strict'
import test from 'node:test'

import {
  extractChessBlocks, findChessBlockIndexAt, findChessBlocks, isValidFEN, parseFenLine,
} from './chessBlocks.ts'

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

/** Bọc nội dung vào một khối ```chess như trong nội dung chương/bài viết thật. */
function block(body: string): string {
  return 'Chữ dẫn.\n\n```chess\n' + body + '\n```\n\nChữ sau.'
}

function firstBoard(body: string) {
  const boards = extractChessBlocks(block(body))
  assert.equal(boards.length, 1, 'phải bóc được đúng 1 bàn cờ')
  return boards[0]
}

// ─── parseFenLine: bóc nhãn dẫn đầu ────────────────────────────────────────
// Đây là gốc của lỗi "bàn cờ lệch 1 hàng": cm-chessboard tách FEN bằng
// split(/\/|\s/) rồi đọc ngược parts[7-part] mà không kiểm số nhóm, nên nhãn
// "fen:" đứng trước trở thành một token thừa và đẩy mọi hàng xuống 1 nấc.

test('bóc nhãn "fen:" — đúng ca gây lỗi hiển thị trong chương sách', () => {
  assert.equal(parseFenLine('fen: ' + START_FEN), START_FEN)
})

test('nhãn viết hoa, có/không dấu hai chấm, và các nhãn đa ngữ', () => {
  for (const prefix of ['FEN: ', 'Fen ', 'fen=', 'position: ', 'Posición: ', 'posicion:', 'Thế cờ: ']) {
    assert.equal(parseFenLine(prefix + START_FEN), START_FEN, `hỏng với tiền tố ${JSON.stringify(prefix)}`)
  }
})

test('gỡ nhiễu markdown quanh dòng FEN', () => {
  assert.equal(parseFenLine('> ' + START_FEN), START_FEN)
  assert.equal(parseFenLine('- ' + START_FEN), START_FEN)
  assert.equal(parseFenLine('`' + START_FEN + '`'), START_FEN)
  assert.equal(parseFenLine('"fen: ' + START_FEN + '"'), START_FEN)
})

test('dòng không phải FEN trả null', () => {
  assert.equal(parseFenLine('Khi bắt đầu ván cờ, mỗi người chơi…'), null)
  assert.equal(parseFenLine(''), null)
  // Thiếu một hàng → không được nhận, nếu không cm-chessboard sẽ vẽ lệch.
  assert.equal(parseFenLine('rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1'), null)
})

// ─── extractChessBlocks: nhánh FEN ─────────────────────────────────────────

test('khối ```chess có nhãn fen: cho ra FEN sạch, không cờ lỗi', () => {
  const b = firstBoard('fen: ' + START_FEN)
  assert.equal(b.fen, START_FEN)
  assert.equal(b.fen_invalid, undefined)
  assert.equal(b.pgn, undefined)
})

test('chữ đứng trước dòng FEN trở thành caption, không bị nuốt mất', () => {
  const b = firstBoard('Thế cờ ban đầu\nfen: ' + START_FEN)
  assert.equal(b.fen, START_FEN)
  assert.equal(b.caption, 'Thế cờ ban đầu')
})

test('chú thích chứa "1." KHÔNG được cướp khối FEN sang nhánh PGN', () => {
  // looksLikePGN cũ khớp /\d+\./ nên "Diagrama 1." từng đẩy cả khối sang nhánh
  // PGN → bàn cờ hiện thế ban đầu, mất hẳn thế cờ thật của trang sách.
  const b = firstBoard('Diagrama 1. Posición inicial\nfen: 8/8/8/4k3/8/8/4P3/4K3 w - - 0 1')
  assert.equal(b.pgn, undefined, 'không được rơi vào nhánh PGN')
  assert.equal(b.fen, '8/8/8/4k3/8/8/4P3/4K3 w - - 0 1')
  assert.equal(b.caption, 'Diagrama 1. Posición inicial')
})

test('FEN chỉ có trường bố cục quân vẫn được nhận', () => {
  const b = firstBoard('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR')
  assert.equal(b.fen, 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR')
  assert.equal(b.fen_invalid, undefined)
})

test('thế cờ KHÔNG có Vua vẫn được nhận (Ngân hàng thế cờ cố ý cho phép)', () => {
  const kingless = '8/8/8/3ppp2/3PPP2/8/8/8 w - - 0 1'
  assert.equal(isValidFEN(kingless), true)
  const b = firstBoard(kingless)
  assert.equal(b.fen, kingless)
  assert.equal(b.fen_invalid, undefined)
})

// ─── FEN hỏng thật → hộp báo lỗi, KHÔNG vẽ bàn cờ sai ───────────────────────

test('FEN thiếu hàng bị đánh dấu fen_invalid thay vì vẽ bàn cờ sai', () => {
  const b = firstBoard('rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')
  assert.equal(b.fen_invalid, true)
})

test('hàng không đủ 8 ô bị đánh dấu fen_invalid', () => {
  const b = firstBoard('rnbqkbn/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')
  assert.equal(b.fen_invalid, true)
})

test('khối không nhận dạng được thì bị bỏ qua (giữ nguyên thành code block)', () => {
  assert.deepEqual(extractChessBlocks(block('chỉ là chữ, không có gì giống cờ')), [])
})

// ─── nhánh PGN vẫn nguyên vẹn ──────────────────────────────────────────────

test('PGN có thẻ vẫn vào nhánh PGN', () => {
  const b = firstBoard('[Event "Opera Game"]\n[White "Morphy"]\n\n1. e4 e5 2. Nf3 d6')
  assert.equal(b.fen, START_FEN)
  assert.ok(b.pgn?.includes('1. e4 e5'))
})

test('PGN trần (chỉ nước đi, không thẻ) vẫn vào nhánh PGN', () => {
  const b = firstBoard('1. e4 e5 2. Nf3 Nc6 3. Bb5 a6')
  assert.equal(b.fen, START_FEN)
  assert.ok(b.pgn?.includes('Bb5'))
})

test('PGN bắt đầu từ thế cờ giữa ván: thẻ FEN quyết định, không phải nhánh FEN', () => {
  const b = firstBoard('[SetUp "1"]\n[FEN "8/8/8/4k3/8/8/4P3/4K3 w - - 0 24"]\n\n24. e4 Kd5')
  assert.ok(b.pgn?.includes('24. e4'), 'phải giữ nguyên PGN để hiện được chuỗi nước đi')
})

// ─── Nhiều khối trong một tài liệu ─────────────────────────────────────────
// Dải xem trước khi soạn (ChessEditorBoards.vue) dựng danh sách bàn cờ bằng
// extractChessBlocks, nên hành vi nhiều-khối là hợp đồng của nó.

test('nhiều khối ```chess cho ra đúng số bàn cờ, đúng thứ tự', () => {
  const md = [
    'Mở đầu.',
    '```chess',
    'fen: 8/8/8/4k3/8/8/4P3/4K3 w - - 0 1',
    '```',
    'Giữa bài.',
    '```chessboard',
    START_FEN,
    '```',
    'Cuối bài.',
    '```chess',
    '[Event "Opera Game"]',
    '',
    '1. e4 e5',
    '```',
  ].join('\n')

  const boards = extractChessBlocks(md)
  assert.equal(boards.length, 3)
  assert.equal(boards[0].fen, '8/8/8/4k3/8/8/4P3/4K3 w - - 0 1')
  assert.equal(boards[1].fen, START_FEN)
  assert.ok(boards[2].pgn?.includes('1. e4 e5'), 'khối thứ ba phải là PGN')
})

test('một khối hỏng giữa chừng KHÔNG nuốt mất các khối còn lại', () => {
  const md = [
    '```chess',
    START_FEN,
    '```',
    '```chess',
    'rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
    '```',
    '```chess',
    '8/8/8/4k3/8/8/4P3/4K3 w - - 0 1',
    '```',
  ].join('\n')

  const boards = extractChessBlocks(md)
  assert.equal(boards.length, 3, 'vẫn phải đủ 3 mục')
  assert.equal(boards[0].fen_invalid, undefined)
  assert.equal(boards[1].fen_invalid, true, 'mục hỏng được đánh dấu để hiện hộp lỗi đúng vị trí')
  assert.equal(boards[2].fen_invalid, undefined)
})

test('khối CHƯA đóng ba backtick thì chưa tính là bàn cờ', () => {
  // Khoá hành vi "bàn cờ chỉ hiện sau khi gõ xong dấu đóng" — đúng như dòng
  // hướng dẫn hiển thị trong ChessEditorBoards.vue.
  const closed = '```chess\n' + START_FEN + '\n```'
  const open = '```chess\n' + START_FEN
  assert.equal(extractChessBlocks(closed).length, 1)
  assert.equal(extractChessBlocks(open).length, 0)
})

// ─── Vị trí khối + tra theo con trỏ ────────────────────────────────────────
// Panel bàn cờ cạnh ô soạn dùng hai hàm này để biết con trỏ đang ở khối nào.

const DOC = [
  'Đoạn mở đầu.',            // 0
  '',
  '```chess',
  START_FEN,
  '```',
  '',
  'Đoạn giữa hai khối.',
  '',
  '```chess',
  '8/8/8/4k3/8/8/4P3/4K3 w - - 0 1',
  '```',
  '',
  'Đoạn kết.',
].join('\n')

test('findChessBlocks trả khoảng ký tự bao đúng khối', () => {
  const blocks = findChessBlocks(DOC)
  assert.equal(blocks.length, 2)
  for (const b of blocks) {
    const slice = DOC.slice(b.start, b.end)
    assert.ok(slice.startsWith('```chess'), `khoảng phải bắt đầu tại dấu mở: ${JSON.stringify(slice.slice(0, 12))}`)
    assert.ok(slice.trimEnd().endsWith('```'), 'khoảng phải kết thúc sau dấu đóng')
  }
  assert.ok(DOC.slice(blocks[0].start, blocks[0].end).includes(START_FEN))
  assert.ok(DOC.slice(blocks[1].start, blocks[1].end).includes('4k3'))
})

test('findChessBlockIndexAt: con trỏ giữa khối trả đúng chỉ số', () => {
  const blocks = findChessBlocks(DOC)
  const mid0 = Math.floor((blocks[0].start + blocks[0].end) / 2)
  const mid1 = Math.floor((blocks[1].start + blocks[1].end) / 2)
  assert.equal(findChessBlockIndexAt(DOC, mid0), 0)
  assert.equal(findChessBlockIndexAt(DOC, mid1), 1)
})

test('findChessBlockIndexAt: ranh giới đầu/cuối vẫn tính là trong khối', () => {
  const blocks = findChessBlocks(DOC)
  assert.equal(findChessBlockIndexAt(DOC, blocks[0].start), 0)
  assert.equal(findChessBlockIndexAt(DOC, blocks[0].end), 0)
})

test('findChessBlockIndexAt: con trỏ ngoài mọi khối trả -1', () => {
  assert.equal(findChessBlockIndexAt(DOC, 0), -1, 'đầu tài liệu')
  assert.equal(findChessBlockIndexAt(DOC, DOC.indexOf('Đoạn giữa') + 3), -1, 'đoạn văn giữa hai khối')
  assert.equal(findChessBlockIndexAt(DOC, DOC.length), -1, 'cuối tài liệu')
})

test('extractChessBlocks vẫn khớp findChessBlocks sau khi viết lại', () => {
  assert.deepEqual(extractChessBlocks(DOC), findChessBlocks(DOC).map((b) => b.data))
})
