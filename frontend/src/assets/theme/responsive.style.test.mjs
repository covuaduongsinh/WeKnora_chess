/**
 * Khoá các bất biến của lớp giao diện điện thoại.
 *
 * Đây là test TĨNH (đọc source dạng text) theo đúng khuôn có sẵn trong repo
 * (`views/chat/components/docInfo.style.test.mjs`). Nó KHÔNG kiểm chứng được
 * "trông có đẹp không" — việc đó phải xem bằng DevTools device emulation. Nó chỉ
 * chặn 6 kiểu hồi quy CÂM đã biết là dễ xảy ra nhất khi sửa tiếp về sau.
 */
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const src = join(here, '..', '..')

const responsiveCss = readFileSync(join(here, 'duongsinh-responsive.css'), 'utf8')
const breakpointTs = readFileSync(join(src, 'composables', 'useBreakpoint.ts'), 'utf8')

/** Bỏ comment `/* … *​/` và `// …` để không soi nhầm phần văn xuôi giải thích. */
function stripComments(css) {
  return css.replace(/\/\*[\s\S]*?\*\//g, '').replace(/^[ \t]*\/\/.*$/gm, '')
}

/**
 * Liệt kê các dòng khớp `re`. Dùng thay cho `assert.doesNotMatch(cảFile, …)` —
 * assert kia in ra NGUYÊN file khi hỏng, làm log CI không đọc nổi.
 */
function linesMatching(text, re) {
  return text
    .split(/\r?\n/)
    .map((line, i) => [i + 1, line])
    .filter(([, line]) => re.test(line))
    .map(([n, line]) => `  ${n}: ${line.trim()}`)
}

test('mọi @media đều có tiền tố `screen and` — chặn rò vào @media print của BookPrint/ArticlePrint', () => {
  const medias = stripComments(responsiveCss).match(/@media[^{]+/g) || []
  assert.ok(medias.length > 0, 'không tìm thấy @media nào — file có còn đúng không?')
  for (const m of medias) {
    assert.match(
      m.trim(),
      /^@media\s+screen\s+and\b/,
      `@media thiếu "screen and" (sẽ rò vào bản in, phá .bkp-page 180mm): ${m.trim()}`,
    )
  }
})

test('không dùng đơn vị viewport — <html> mang CSS zoom nên vw/vh sai lệch', () => {
  const offenders = linesMatching(stripComments(responsiveCss), /\d\s*(vw|vh|dvh|svh|lvh|dvw|svw|lvw)\b/)
  assert.deepEqual(
    offenders,
    [],
    `dùng 100% + min-width: 0 thay cho đơn vị viewport (xem utils/zoom.ts):\n${offenders.join('\n')}`,
  )
})

test('breakpoint trong CSS khớp hằng BP.mobile của useBreakpoint.ts', () => {
  assert.match(responsiveCss, /max-width:\s*767px/)
  assert.match(breakpointTs, /mobile:\s*767\b/)
})

test('menu.vue giữ cam kết 0 dòng sửa cho responsive (drawer làm hoàn toàn bằng CSS ngoài)', () => {
  const menu = readFileSync(join(src, 'components', 'menu.vue'), 'utf8')
  const offenders = linesMatching(menu, /@media/)
  assert.deepEqual(
    offenders,
    [],
    `nếu buộc phải sửa menu.vue, cập nhật lại nhật ký tuỳ biến mục C3:\n${offenders.join('\n')}`,
  )
})

test('drawer mobile còn bám đúng 2 class của upstream (.main > .aside_box)', () => {
  // Nếu upstream đổi tên class hoặc đổi .main thành scoped, drawer chết CÂM.
  // Test này là lưới an toàn duy nhất khi merge upstream.
  const platform = readFileSync(join(src, 'views', 'platform', 'index.vue'), 'utf8')
  const menu = readFileSync(join(src, 'components', 'menu.vue'), 'utf8')
  assert.match(platform, /class="main"/)
  assert.match(menu, /class="aside_box"/)
  assert.match(responsiveCss, /\.main\s*>\s*\.aside_box/)
})

test('chat/index.vue không còn tính bề rộng bằng 100vw (sai dưới zoom, vỡ khi sidebar thành ngăn kéo)', () => {
  const chat = stripComments(readFileSync(join(src, 'views', 'chat', 'index.vue'), 'utf8'))
  const offenders = linesMatching(chat, /100vw/)
  assert.deepEqual(
    offenders,
    [],
    `dùng max-width: 100% (containing block đã trừ sidebar sẵn):\n${offenders.join('\n')}`,
  )
})

test('cả 6 trang cờ đều có nhánh mobile', () => {
  for (const name of ['ArticleBank', 'BookLibrary', 'GameLibrary', 'PuzzleBank', 'PositionBank', 'ChessCourses']) {
    const page = readFileSync(join(src, 'views', 'chess', `${name}.vue`), 'utf8')
    assert.match(
      page,
      /@media screen and \(max-width:\s*767px\)/,
      `${name}.vue thiếu nhánh mobile`,
    )
  }
})
