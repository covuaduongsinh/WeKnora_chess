<template>
  <div class="bkp">
    <div class="bkp-toolbar no-print">
      <t-radio-group v-model="columns" variant="default-filled">
        <t-radio-button :value="1">1 cột</t-radio-button>
        <t-radio-button :value="2">2 cột</t-radio-button>
      </t-radio-group>
      <t-radio-group v-model="headerLevel" variant="default-filled">
        <t-radio-button value="full">Đầy đủ</t-radio-button>
        <t-radio-button value="brief">Gọn</t-radio-button>
        <t-radio-button value="none">Không</t-radio-button>
      </t-radio-group>
      <t-button variant="outline" :disabled="!chapters.length" @click="openChapterPicker">
        <template #icon><t-icon name="view-list" /></template>Chọn chương ({{ selectedIds.size }}/{{ chapters.length }})
      </t-button>
      <t-button theme="primary" @click="doPrint">
        <template #icon><t-icon name="print" /></template>In / Lưu PDF (Ctrl+P)
      </t-button>
    </div>
    <div v-if="loading" class="bkp-empty">Đang tải…</div>
    <div v-else-if="!book" class="bkp-empty">Không tìm thấy sách.</div>
    <article v-else class="bkp-page" :class="{ 'bkp-page--cols-2': columns === 2 }">
      <template v-if="headerLevel !== 'none'">
        <h1>{{ book.title }}</h1>
        <p v-if="headerLevel === 'full' && book.subtitle" class="bkp-subtitle">{{ book.subtitle }}</p>
        <ul class="bkp-meta">
          <li v-if="book.author">Tác giả: {{ book.author }}</li>
          <li v-if="book.translator">Dịch giả: {{ book.translator }}</li>
          <li v-if="book.publisher">Nhà xuất bản: {{ book.publisher }}</li>
          <li v-if="book.year">Năm: {{ book.year }}</li>
          <li v-if="book.level">Cấp độ: {{ bookLevelLabel(book.level) }}</li>
          <li v-if="book.phase">Giai đoạn: {{ bookPhaseLabel(book.phase) }}</li>
        </ul>
        <div v-if="headerLevel === 'full' && book.description" class="bkp-description" v-html="renderMd(book.description)"></div>
      </template>

      <div v-if="chapters.length > 0 && printedChapters.length === 0" class="bkp-empty">Chưa chọn chương nào để in.</div>
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

    <!-- Dialog chọn chương để in — khuôn giống ChessShelfManager.vue (dialog +
         ô tìm + list t-checkbox + confirm), state nháp riêng để bấm Huỷ không
         mất lựa chọn cũ. class="no-print" phòng khi dialog còn mở lúc Ctrl+P. -->
    <t-dialog
      v-model:visible="pickerDialog.visible"
      header="Chọn chương để in"
      width="560px"
      :on-confirm="applyChapterPicker"
      class="no-print"
    >
      <t-input v-model="pickerDialog.search" placeholder="Tìm chương…" clearable style="margin-bottom: 10px" />
      <div class="bkp-picker-actions">
        <t-button variant="text" size="small" theme="primary" @click="pickerSelectAll">Chọn tất cả</t-button>
        <t-button variant="text" size="small" theme="primary" @click="pickerSelectNone">Bỏ chọn tất cả</t-button>
      </div>
      <div class="bkp-picker-list">
        <template v-for="(group, gi) in pickerGroups" :key="gi">
          <div v-if="group.part" class="bkp-picker-part">{{ group.part }}</div>
          <t-checkbox v-for="ch in group.items" :key="ch.id" v-model="pickerDialog.selected[ch.id]">
            {{ ch.title }}
          </t-checkbox>
        </template>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { marked } from 'marked';
import { MessagePlugin } from 'tdesign-vue-next';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import type { ChessBoardData } from '@/types/tool-results';
import { splitChessSegments } from '@/utils/chessBlocks';
import { safeMarkdownToHTML, sanitizeHTML } from '@/utils/security';
import { getBook, listChapters, type ChessBook, type ChessBookChapter } from '@/api/chess';
import { bookLevelLabel, bookPhaseLabel } from '@/utils/chessBookOptions';

// Trang IN sách: mọi chương trên một trang dài, mở tab mới rồi Ctrl+P → PDF.
// Không dùng generator PDF server-side (repo chưa có) — @media print lo layout.
const route = useRoute();
const router = useRouter();
const loading = ref(true);
const book = ref<ChessBook | null>(null);
const chapters = ref<ChessBookChapter[]>([]); // TẤT CẢ chương của sách — nguồn cho dialog chọn

// Chương sẽ IN — mặc định (không có ?chapters trên URL) là toàn bộ, khớp hành
// vi cũ. Lọc theo id (KHÔNG theo slug — ChessBookChapter.slug là optional).
const selectedIds = ref<Set<string>>(new Set());
const printedChapters = computed(() => chapters.value.filter((ch) => selectedIds.value.has(ch.id)));

const QUERY_KEY = 'chapters';

// Đọc lựa chọn từ ?chapters=id1,id2 lúc mount. Rỗng hoặc id không thuộc sách
// (đã xoá/gõ sai) → rơi về CHỌN TẤT CẢ thay vì trang trắng khó hiểu.
function applySelectionFromQuery() {
  const raw = String(route.query[QUERY_KEY] || '');
  const ids = raw.split(',').map((s) => s.trim()).filter(Boolean);
  const valid = ids.filter((id) => chapters.value.some((ch) => ch.id === id));
  const allIds = chapters.value.map((ch) => ch.id);
  selectedIds.value = new Set(valid.length ? valid : allIds);
}

// Ghi ngược lựa chọn ra URL (F5 / dán link cho người khác vẫn đúng chương đã
// chọn) — chọn đủ TẤT CẢ thì xoá hẳn query cho URL sạch, khớp hành vi mặc định.
function syncQueryFromSelection() {
  const allIds = chapters.value.map((ch) => ch.id);
  const isAll = selectedIds.value.size === allIds.length && allIds.every((id) => selectedIds.value.has(id));
  const query = { ...route.query };
  if (isAll) delete query[QUERY_KEY];
  else query[QUERY_KEY] = Array.from(selectedIds.value).join(',');
  router.replace({ path: route.path, query });
}

// Bố cục cột — nhớ theo trình duyệt (localStorage), không đụng DB/sách cụ
// thể. Vùng nội dung dùng đơn vị mm khớp khổ A4 (xem <style>) nên bố cục lúc
// XEM và lúc IN giống hệt nhau — bàn cờ co đúng MỘT lần ngay khi đổi cột,
// không reflow lúc in. cm-chessboard ghi kích thước cố định (px) qua
// ResizeObserver bị hoãn tới macrotask; window.print() chụp layout đồng bộ
// TRƯỚC khi macrotask đó kịp chạy, nên đổi cột chỉ trong @media print sẽ để
// bàn cờ tràn/cắt khỏi cột — không được lặp lại cách đó. Tùy chọn "phần đầu
// sách" bên dưới áp dụng CÙNG nguyên tắc (đổi ở cả xem trước lẫn khi in).
const COLUMNS_STORAGE_KEY = 'weknora_book_print_columns';
function readStoredColumns(): 1 | 2 {
  return localStorage.getItem(COLUMNS_STORAGE_KEY) === '2' ? 2 : 1;
}
const columns = ref<1 | 2>(readStoredColumns());
watch(columns, (v) => {
  localStorage.setItem(COLUMNS_STORAGE_KEY, String(v));
});

// Mức hiển thị phần đầu sách khi in — hữu ích khi chỉ in 1-2 chương lẻ (photo
// một bài học) và không cần nhắc lại mô tả cả cuốn mỗi lần.
type HeaderLevel = 'full' | 'brief' | 'none';
const HEADER_LEVEL_STORAGE_KEY = 'weknora_book_print_header_level';
function readStoredHeaderLevel(): HeaderLevel {
  const v = localStorage.getItem(HEADER_LEVEL_STORAGE_KEY);
  return v === 'brief' || v === 'none' ? v : 'full';
}
const headerLevel = ref<HeaderLevel>(readStoredHeaderLevel());
watch(headerLevel, (v) => {
  localStorage.setItem(HEADER_LEVEL_STORAGE_KEY, v);
});

const chapterGroups = computed(() => groupByPart(printedChapters.value));
function groupByPart(list: ChessBookChapter[]) {
  const groups: { part: string; items: ChessBookChapter[] }[] = [];
  let last: string | null = null;
  for (const ch of list) {
    const part = ch.part || '';
    if (last === null || part !== last || groups.length === 0) groups.push({ part, items: [] });
    groups[groups.length - 1].items.push(ch);
    last = part;
  }
  return groups;
}

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

// ---- Dialog chọn chương ----
interface PickerDialogState {
  visible: boolean;
  search: string;
  selected: Record<string, boolean>;
}
const pickerDialog = reactive<PickerDialogState>({ visible: false, search: '', selected: {} });

function openChapterPicker() {
  const sel: Record<string, boolean> = {};
  for (const ch of chapters.value) sel[ch.id] = selectedIds.value.has(ch.id);
  pickerDialog.selected = sel;
  pickerDialog.search = '';
  pickerDialog.visible = true;
}

const pickerFilteredChapters = computed(() => {
  const q = pickerDialog.search.trim().toLowerCase();
  if (!q) return chapters.value;
  return chapters.value.filter((ch) => (ch.title || '').toLowerCase().includes(q));
});
const pickerGroups = computed(() => groupByPart(pickerFilteredChapters.value));

function pickerSelectAll() {
  for (const ch of pickerFilteredChapters.value) pickerDialog.selected[ch.id] = true;
}
function pickerSelectNone() {
  for (const ch of pickerFilteredChapters.value) pickerDialog.selected[ch.id] = false;
}

function applyChapterPicker() {
  const ids = Object.entries(pickerDialog.selected).filter(([, v]) => v).map(([id]) => id);
  if (ids.length === 0) {
    MessagePlugin.warning('Chọn ít nhất 1 chương');
    return;
  }
  selectedIds.value = new Set(ids);
  syncQueryFromSelection();
  pickerDialog.visible = false;
}

onMounted(async () => {
  const id = String(route.params.id || '');
  if (!id) { loading.value = false; return; }
  try {
    const [br, cr]: any[] = await Promise.all([getBook(id), listChapters(id)]);
    book.value = br?.data || null;
    chapters.value = cr?.data || [];
    applySelectionFromQuery();
  } finally {
    loading.value = false;
  }
});
</script>

<style lang="less" scoped>
@page { size: A4 portrait; margin: 15mm; }

.bkp { padding: 24px; overflow-x: auto; }
.bkp-toolbar { position: sticky; top: 0; padding: 10px 0; background: var(--td-bg-color-page); z-index: 10; display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
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

// Không cắt đôi BÀN CỜ giữa 2 cột/2 trang — nhưng KHÔNG áp cho cả khối
// .chess-board-display nữa (đổi so với trước): từ khi move-list có thể mang
// bình luận rất dài, ép "avoid" ở cấp cả khối sẽ đẩy NGUYÊN khối (bàn cờ +
// bình luận dài) sang cột/trang sau, bỏ trắng cả cột trước — đúng cái bẫy mà
// comment ở .bkp-chapter phía trên đã cảnh báo, nhưng ở cấp CHƯƠNG; giờ lặp
// lại đúng vấn đề đó ở cấp một khối bàn cờ. Chỉ giữ nguyên vẹn phần .chess-body
// (bàn cờ + thanh eval) — phần cho phép ngắt là move-list (chữ, ngắt được).
:deep(.chess-board-display) { break-inside: auto; }
:deep(.chess-body) { break-inside: avoid; }
:deep(.chess-nav),
:deep(.chess-nav-single),
:deep(.eval-bar),
:deep(.chess-eval-line) { display: none !important; }

// .move-list (danh sách nước đi + chú giải PGN nếu có — bình luận/NAG/nhánh
// phụ, xem utils/pgnAnnotated.ts) nằm NGOÀI .chess-nav (xem ChessBoardDisplay.vue)
// nên KHÔNG bị ẩn bởi luật trên — đây là nội dung sách cờ cần in ra giấy. Chỉ
// bỏ khung cuộn 120px (vô nghĩa khi không còn bấm/cuộn được) và tắt hiệu ứng
// con trỏ/hover — cả 2 nơi (xem preview) lẫn khi in, cùng lý do nhất quán như trên.
:deep(.move-list) {
  max-height: none;
  overflow: visible;
}
:deep(.move-item) {
  cursor: default;

  &:hover { background: none; }

  // Nền navy của nước "đang xem" vô nghĩa trên giấy — in đậm thay vì tô nền.
  &.active {
    background: none !important;
    font-weight: 700;
  }
}
// Bình luận dài không mồ côi 1 dòng đầu/cuối khi bị ngắt qua trang/cột.
:deep(.move-comment) { orphans: 2; widows: 2; }

// Dialog "Chọn chương" — teleport ra ngoài .bkp nhưng CSS scoped vẫn áp dụng
// qua data-v attribute nên style bình thường như mọi nơi khác trong file.
.bkp-picker-actions { display: flex; gap: 4px; margin-bottom: 8px; }
.bkp-picker-list {
  max-height: 360px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.bkp-picker-part {
  margin-top: 8px;
  font-weight: 600;
  color: var(--td-text-color-secondary);
  font-size: 13px;

  &:first-child { margin-top: 0; }
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

  // Chú giải PGN (bình luận/số hiệu nước/ký hiệu NAG/dấu ngoặc nhánh phụ) —
  // cùng lý do :root{color-scheme:light} ở trên: chế độ tối sẽ in ra chữ trắng
  // trên giấy trắng nếu không ép lại theo danh sách selector cụ thể này.
  :deep(.move-comment),
  :deep(.move-item),
  :deep(.move-no),
  :deep(.move-nag),
  :deep(.move-paren),
  :deep(.move-var) { color: #000 !important; }
}
</style>
