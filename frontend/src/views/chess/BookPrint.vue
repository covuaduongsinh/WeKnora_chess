<template>
  <div class="bkp">
    <div class="bkp-toolbar no-print">
      <t-radio-group v-model="columns" variant="default-filled">
        <t-radio-button :value="1">1 cột</t-radio-button>
        <t-radio-button :value="2">2 cột</t-radio-button>
      </t-radio-group>
      <t-button theme="primary" @click="doPrint">
        <template #icon><t-icon name="print" /></template>In / Lưu PDF (Ctrl+P)
      </t-button>
    </div>
    <div v-if="loading" class="bkp-empty">Đang tải…</div>
    <div v-else-if="!book" class="bkp-empty">Không tìm thấy sách.</div>
    <article v-else class="bkp-page" :class="{ 'bkp-page--cols-2': columns === 2 }">
      <h1>{{ book.title }}</h1>
      <p v-if="book.subtitle" class="bkp-subtitle">{{ book.subtitle }}</p>
      <ul class="bkp-meta">
        <li v-if="book.author">Tác giả: {{ book.author }}</li>
        <li v-if="book.translator">Dịch giả: {{ book.translator }}</li>
        <li v-if="book.publisher">Nhà xuất bản: {{ book.publisher }}</li>
        <li v-if="book.year">Năm: {{ book.year }}</li>
        <li v-if="book.level">Cấp độ: {{ bookLevelLabel(book.level) }}</li>
        <li v-if="book.phase">Giai đoạn: {{ bookPhaseLabel(book.phase) }}</li>
      </ul>
      <div v-if="book.description" class="bkp-description" v-html="renderMd(book.description)"></div>

      <template v-for="(group, gi) in chapterGroups" :key="gi">
        <h2 v-if="group.part" class="bkp-part">{{ group.part }}</h2>
        <section v-for="ch in group.items" :key="ch.id" class="bkp-chapter">
          <h3>{{ ch.title }}</h3>
          <ChessBoardDisplay v-if="ch.fen" :data="chapterBoardData(ch)" />
          <template v-for="(seg, si) in chapterSegments(ch)" :key="si">
            <ChessBoardDisplay v-if="seg.type === 'board'" :data="seg.board" />
            <div v-else-if="seg.type === 'markdown'" v-html="seg.html"></div>
          </template>
        </section>
      </template>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { marked } from 'marked';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import type { ChessBoardData } from '@/types/tool-results';
import { splitChessSegments } from '@/utils/chessBlocks';
import { safeMarkdownToHTML, sanitizeHTML } from '@/utils/security';
import { getBook, listChapters, type ChessBook, type ChessBookChapter } from '@/api/chess';
import { bookLevelLabel, bookPhaseLabel } from '@/utils/chessBookOptions';

// Trang IN sách: mọi chương trên một trang dài, mở tab mới rồi Ctrl+P → PDF.
// Không dùng generator PDF server-side (repo chưa có) — @media print lo layout.
const route = useRoute();
const loading = ref(true);
const book = ref<ChessBook | null>(null);
const chapters = ref<ChessBookChapter[]>([]);

// Bố cục cột — nhớ theo trình duyệt (localStorage), không đụng DB/sách cụ
// thể. Vùng nội dung dùng đơn vị mm khớp khổ A4 (xem <style>) nên bố cục lúc
// XEM và lúc IN giống hệt nhau — bàn cờ co đúng MỘT lần ngay khi đổi cột,
// không reflow lúc in. cm-chessboard ghi kích thước cố định (px) qua
// ResizeObserver bị hoãn tới macrotask; window.print() chụp layout đồng bộ
// TRƯỚC khi macrotask đó kịp chạy, nên đổi cột chỉ trong @media print sẽ để
// bàn cờ tràn/cắt khỏi cột — không được lặp lại cách đó.
const COLUMNS_STORAGE_KEY = 'weknora_book_print_columns';
function readStoredColumns(): 1 | 2 {
  return localStorage.getItem(COLUMNS_STORAGE_KEY) === '2' ? 2 : 1;
}
const columns = ref<1 | 2>(readStoredColumns());
watch(columns, (v) => {
  localStorage.setItem(COLUMNS_STORAGE_KEY, String(v));
});

const chapterGroups = computed(() => {
  const groups: { part: string; items: ChessBookChapter[] }[] = [];
  let last: string | null = null;
  for (const ch of chapters.value) {
    const part = ch.part || '';
    if (last === null || part !== last || groups.length === 0) groups.push({ part, items: [] });
    groups[groups.length - 1].items.push(ch);
    last = part;
  }
  return groups;
});

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';
function chapterBoardData(ch: ChessBookChapter): ChessBoardData {
  return { display_type: 'chess_board', fen: ch.fen || START_FEN, caption: ch.title };
}
function renderMd(md: string): string {
  return sanitizeHTML(marked.parse(safeMarkdownToHTML(md || ''), { breaks: true, async: false }) as string);
}
// KHÔNG dùng renderChessChips ở đây — bản in tĩnh, chip [[...]] chỉ cần hiện nhãn.
function stripWikilinkSyntax(md: string): string {
  return (md || '').replace(/!?\[\[([^\]]+)\]\]/g, (_m, inner: string) => {
    const pipe = inner.indexOf('|');
    return (pipe > 0 ? inner.slice(pipe + 1) : inner).trim();
  });
}
function chapterSegments(ch: ChessBookChapter) {
  return splitChessSegments(ch.content || '').map((seg) => {
    if (seg.type === 'board') return { type: 'board' as const, board: seg.board! };
    return { type: 'markdown' as const, html: renderMd(stripWikilinkSyntax(seg.markdown || '')) };
  });
}

function doPrint() { window.print(); }

onMounted(async () => {
  const id = String(route.params.id || '');
  if (!id) { loading.value = false; return; }
  try {
    const [br, cr]: any[] = await Promise.all([getBook(id), listChapters(id)]);
    book.value = br?.data || null;
    chapters.value = cr?.data || [];
  } finally {
    loading.value = false;
  }
});
</script>

<style lang="less" scoped>
@page { size: A4 portrait; margin: 15mm; }

.bkp { padding: 24px; overflow-x: auto; }
.bkp-toolbar { position: sticky; top: 0; padding: 10px 0; background: var(--td-bg-color-page); z-index: 10; display: flex; gap: 10px; align-items: center; }
.bkp-empty { color: var(--td-text-color-placeholder); padding: 40px; text-align: center; }

// Vùng nội dung = đúng bề rộng vùng in A4 (210mm − 2×15mm lề @page ở trên) để
// bố cục lúc XEM và lúc IN khớp nhau tuyệt đối — xem ghi chú readStoredColumns()
// trong <script>. Cửa sổ hẹp hơn 180mm thì .bkp cuộn ngang, KHÔNG co trang giấy.
.bkp-page {
  width: 180mm;
  margin: 0 auto;
  color: var(--td-text-color-primary);
  line-height: 1.7;
}
.bkp-page--cols-2 {
  column-count: 2;
  column-gap: 8mm; // mỗi cột = (180mm − 8mm) / 2 = 86mm
  column-fill: auto;
}

.bkp-subtitle { color: var(--td-text-color-secondary); font-style: italic; }
.bkp-meta { font-size: 14px; color: var(--td-text-color-secondary); padding-left: 18px; }
.bkp-description { margin: 16px 0; padding: 12px; background: var(--td-bg-color-container); border-radius: 8px; }

.bkp-part {
  margin-top: 32px;
  border-bottom: 2px solid var(--td-brand-color);
  padding-bottom: 6px;
  break-after: avoid; // tiêu đề Phần không mồ côi cuối cột/trang
}
.bkp-page--cols-2 .bkp-part { break-before: column; margin-top: 0; }

.bkp-chapter {
  margin: 24px 0;
  break-inside: avoid-page; // 1 cột: tránh vỡ chương giữa 2 trang
}
// 2 cột: KHÔNG cấm ngắt chương — "avoid" sẽ đẩy cả khối chương dài sang cột
// sau, để lại khoảng trắng lớn ở cột trước (phản tác dụng tiết kiệm giấy).
// Chỉ giữ tiêu đề không mồ côi + không cắt đôi bàn cờ (xem :deep bên dưới).
.bkp-page--cols-2 .bkp-chapter { break-inside: auto; }
.bkp-chapter h3 { break-after: avoid; }

// Không cắt đôi bàn cờ giữa 2 cột/2 trang; ẩn phần điều hướng tương tác (vô
// nghĩa trên giấy) — ẩn CẢ trên màn hình lẫn khi in (trang này LÀ bản xem
// trước, phải khớp bản in — ẩn cả 2 nơi cũng giữ chiều cao nhất quán giữa
// hai chế độ, tránh lệch điểm ngắt trang khi bật/tắt in).
:deep(.chess-board-display) { break-inside: avoid; }
:deep(.chess-nav),
:deep(.chess-nav-single),
:deep(.eval-bar),
:deep(.chess-eval-line) { display: none !important; }

// .move-list (danh sách nước đi dạng chữ, vd "1. e4 e5 2. Nf3 Nc6") nằm NGOÀI
// .chess-nav (xem ChessBoardDisplay.vue) nên KHÔNG bị ẩn bởi luật trên — đây
// là nội dung sách cờ cần in ra giấy. Chỉ bỏ khung cuộn 120px (vô nghĩa khi
// không còn bấm/cuộn được) và tắt hiệu ứng con trỏ/hover — cả 2 nơi (xem
// preview) lẫn khi in, cùng lý do nhất quán như trên.
:deep(.move-list) {
  max-height: none;
  overflow: visible;
}
:deep(.move-item) {
  cursor: default;

  &:hover { background: none; }
}

@media print {
  .no-print { display: none !important; }
  .bkp { padding: 0; overflow-x: visible; }

  // index.html đặt màu nền/chữ theo theme đã lưu (kể cả tối) lên html/body —
  // không ép lại thì chế độ tối in ra chữ trắng trên giấy trắng. Ép theo
  // danh sách selector cụ thể (không wildcard) để không phá màu SVG bàn cờ.
  :root { color-scheme: light !important; }
  .bkp-page,
  .bkp-subtitle,
  .bkp-meta,
  .bkp-part { color: #000 !important; }
  :deep(.chess-caption) { color: #000 !important; }
  .bkp-description { background: transparent !important; border: 1px solid #ccc; }
}
</style>
