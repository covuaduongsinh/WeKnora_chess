import assert from 'node:assert/strict'
import test from 'node:test'

import {
  buildAnnotatedPgn, cleanComment, glyphOf, needsSpaceBefore, tokensFromPositions,
  type PgnToken,
} from './pgnAnnotated.ts'

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

// Chuỗi hoá token để so sánh một phát, cùng dạng ký hiệu sách cờ.
function render(tokens: PgnToken[]): string {
  return tokens
    .map((t) => {
      if (t.kind === 'move') return `${t.num ? t.num + ' ' : ''}${t.san}${t.glyph}`
      if (t.kind === 'comment') return `{${t.text}}`
      return t.kind === 'open' ? '(' : ')'
    })
    .join(' ')
    .replace(/\( /g, '(')
    .replace(/ \)/g, ')')
}

// PGN thật của Thầy Tường (rút gọn) — ván bắt đầu từ thế cờ giữa ván (nước 6,
// Trắng đi), chứa cả 3 loại chú giải: bình luận, NAG, nhánh phụ; kèm dữ liệu
// máy [%evp ...] lẫn trong bình luận mở đầu.
const THAY_TUONG_PGN = `[SetUp "1"]
[FEN "r1bqk2r/ppp2ppp/2np1n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQK2R w KQkq - 0 6"]

{[%evp 0,15,7,-4] This is the first position I would like to discuss with you.} 6. Na4 $1 {In one of his books, Dvoretsky teaches.} Bb6 {theory says.} 7. a3 (7. O-O O-O 8. a3) 7... O-O 8. O-O Be6 9. Bxe6 fxe6 {structure.} 10. c3 $1 Ne7 11. b4 $14 {Black is worse.} *`

test('PGN của Thầy: bình luận + NAG + nhánh phụ hiện đầy đủ, dữ liệu máy bị lọc', () => {
  const res = buildAnnotatedPgn(THAY_TUONG_PGN, START_FEN)
  assert.ok(res)
  assert.equal(res!.degraded, false)
  assert.equal(res!.truncated, false)
  assert.equal(
    render(res!.tokens),
    '{This is the first position I would like to discuss with you.} 6. Na4! ' +
      '{In one of his books, Dvoretsky teaches.} 6... Bb6 {theory says.} ' +
      '7. a3 (7. O-O O-O 8. a3) 7... O-O 8. O-O Be6 9. Bxe6 fxe6 {structure.} ' +
      '10. c3! Ne7 11. b4⩲ {Black is worse.}',
  )
})

test('bình luận CHỈ chứa dữ liệu máy ([%clk]) không sinh token và không làm Đen in số hiệu thừa', () => {
  const pgn = '1. e4 {[%clk 0:05:00]} e5 {[%clk 0:04:58]} 2. Nf3 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  assert.equal(
    render(res!.tokens),
    '1. e4 e5 2. Nf3',
  )
})

test('PGN thường (không chú giải) tương thích ngược với hiển thị cũ', () => {
  const res = buildAnnotatedPgn('1. e4 e5 2. Nf3 Nc6 *', START_FEN)
  assert.ok(res)
  const moves = res!.tokens.filter((t) => t.kind === 'move') as Extract<PgnToken, { kind: 'move' }>[]
  assert.deepEqual(moves.map((t) => t.num), ['1.', '', '2.', ''])
  assert.ok(moves.every((t) => t.glyph === ''))
  assert.equal(render(res!.tokens), '1. e4 e5 2. Nf3 Nc6')
})

test('nhánh phụ lồng 2 tầng: depth đúng, posIndex luôn null trong nhánh', () => {
  const pgn = '1. e4 e5 (1... c5 2. Nf3 (2. c3 d5) d6) 2. Nf3 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  const moves = res!.tokens.filter((t) => t.kind === 'move') as Extract<PgnToken, { kind: 'move' }>[]
  assert.deepEqual(moves.map((t) => t.depth), [0, 0, 1, 1, 2, 2, 1, 0])
  assert.deepEqual(
    moves.filter((t) => t.depth > 0).map((t) => t.posIndex),
    [null, null, null, null, null],
  )
  // mainline không hề bị ảnh hưởng bởi nhánh phụ chèn giữa
  assert.deepEqual(
    moves.filter((t) => t.depth === 0).map((t) => t.san),
    ['e4', 'e5', 'Nf3'],
  )
})

test('posIndex mainline = 1..N và positions.length = N + 1 (thế cờ ban đầu do caller tự chèn)', () => {
  const res = buildAnnotatedPgn(THAY_TUONG_PGN, START_FEN)
  assert.ok(res)
  assert.equal(res!.positions.length, 11)
  const mainlineMoves = res!.tokens.filter(
    (t) => t.kind === 'move' && t.depth === 0,
  ) as Extract<PgnToken, { kind: 'move' }>[]
  assert.deepEqual(
    mainlineMoves.map((t) => t.posIndex),
    Array.from({ length: 11 }, (_, i) => i + 1),
  )
})

test('grammar throw (ngoặc nhọn lồng) → tầng "đã xoá bình luận": còn nước, mất bình luận', () => {
  const pgn = '1. e4 {a {b} c} e5 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  assert.equal(res!.degraded, true)
  assert.equal(res!.positions.length, 2)
  assert.ok(!res!.tokens.some((t) => t.kind === 'comment'))
})

test('glyphOf: gộp NAG, khử trùng lặp, ký hiệu lạ giữ nguyên $N', () => {
  assert.equal(glyphOf({ nag: ['1', '14'] }), '!⩲')
  // suffix "!?" và NAG $5 (cũng map thành "!?") không được lặp lại
  assert.equal(glyphOf({ suffix: ['!', '?'], nag: ['5'] }), '!?')
  assert.equal(glyphOf({ nag: ['99'] }), '$99')
  assert.equal(glyphOf({}), '')
})

test('ván bắt đầu bằng lượt Đen (FEN giữa ván) → token đầu in đủ "N..."', () => {
  const pgn = '[FEN "4k3/8/8/8/8/8/8/4K3 b - - 0 24"]\n\n24... Kd7 25. Kd2 Ke6 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  const moves = res!.tokens.filter((t) => t.kind === 'move') as Extract<PgnToken, { kind: 'move' }>[]
  assert.deepEqual(
    moves.map((t) => t.num),
    ['24...', '25.', ''],
  )
})

test('PGN rỗng hoặc rác → null (caller rơi về FEN đơn)', () => {
  assert.equal(buildAnnotatedPgn('', START_FEN), null)
  assert.equal(buildAnnotatedPgn('không phải PGN chút nào!!', START_FEN), null)
})

test('nước sai luật giữa ván → giữ phần mainline đã dựng được (truncated), không mất sạch', () => {
  // Ng3 hợp lệ về CÚ PHÁP SAN nhưng không mã nào đi được tới g3 từ vị trí này.
  const pgn = '1. e4 e5 2. Ng3 Nc6 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  assert.equal(res!.truncated, true)
  assert.equal(res!.positions.length, 2)
})

test('bình luận chứa markup không bị escape hay lược bỏ (an toàn nhờ template dùng {{ }}, không v-html)', () => {
  const pgn = '1. e4 {<img src=x onerror=alert(1)>} e5 *'
  const res = buildAnnotatedPgn(pgn, START_FEN)
  assert.ok(res)
  const comment = res!.tokens.find((t) => t.kind === 'comment')
  assert.equal(comment && 'text' in comment ? comment.text : null, '<img src=x onerror=alert(1)>')
})

test('cleanComment: lọc [%evp]/[%clk], giữ nguyên chữ; rỗng sau khi lọc → chuỗi rỗng', () => {
  assert.equal(cleanComment('[%clk 0:05:00] Good move'), 'Good move')
  assert.equal(cleanComment('[%evp 0,15,7,-4]'), '')
  assert.equal(cleanComment(undefined), '')
  assert.equal(cleanComment('  nhiều   khoảng   trắng  '), 'nhiều khoảng trắng')
})

test('needsSpaceBefore: có khoảng trắng giữa các nước, KHÔNG có quanh ngoặc nhánh phụ', () => {
  const move = (depth = 0): PgnToken => ({ kind: 'move', num: '', san: 'e4', glyph: '', posIndex: 1, depth })
  const open = (depth = 1): PgnToken => ({ kind: 'open', depth })
  const close = (depth = 1): PgnToken => ({ kind: 'close', depth })
  const comment = (depth = 0): PgnToken => ({ kind: 'comment', text: 'x', depth })

  assert.equal(needsSpaceBefore(null, move()), false)          // token đầu tiên
  assert.equal(needsSpaceBefore(move(), move()), true)         // giữa 2 nước
  assert.equal(needsSpaceBefore(move(), open()), true)         // trước "(" vẫn có
  assert.equal(needsSpaceBefore(open(), move(1)), false)       // ngay sau "(" thì không
  assert.equal(needsSpaceBefore(move(1), close()), false)      // trước ")" thì không
  assert.equal(needsSpaceBefore(close(), move()), true)        // sau ")" có lại
  // Bình luận mainline là KHỐI (tự xuống dòng) → không cần khoảng trắng hai bên
  assert.equal(needsSpaceBefore(move(), comment()), false)
  assert.equal(needsSpaceBefore(comment(), move()), false)
  // Bình luận TRONG nhánh phụ hiển thị inline → vẫn cần khoảng trắng
  assert.equal(needsSpaceBefore(move(1), comment(1)), true)
})

test('tokensFromPositions: sinh token khớp hiển thị hiện có cho nhánh plies (không chú giải)', () => {
  const list = [
    { fen: 'f1', label: '1. e4', san: 'e4', moveNumber: 1, side: 'w' as const },
    { fen: 'f2', label: '1... e5', san: 'e5', moveNumber: 1, side: 'b' as const },
  ]
  const tokens = tokensFromPositions(list)
  assert.deepEqual(
    tokens.map((t) => (t.kind === 'move' ? [t.num, t.san, t.posIndex] : null)),
    [
      ['1.', 'e4', 1],
      ['', 'e5', 2],
    ],
  )
})
