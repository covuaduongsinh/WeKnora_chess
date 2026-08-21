<template>
  <div class="ceb" :class="{ 'is-collapsed': !open }">
    <div class="ceb-head">
      <button type="button" class="ceb-toggle" :title="open ? 'Thu gọn' : 'Mở bảng xem trước'"
        @click="open = !open">
        <t-icon :name="open ? 'chevron-right' : 'chevron-left'" />
        <span class="ceb-toggle-text">Thế cờ</span>
        <span v-if="boards.length" class="ceb-count">{{ boards.length }}</span>
      </button>
    </div>

    <div v-if="open" ref="listEl" class="ceb-list">
      <div v-if="!boards.length" class="ceb-empty">
        Chưa có thế cờ. Bấm nút bàn cờ trên thanh công cụ để chèn một khối.
        <span class="ceb-hint">Bàn cờ chỉ hiện sau khi khối đã đóng đủ ba dấu nháy ngược.</span>
      </div>

      <div v-for="(b, i) in boards" :key="b.key" :data-idx="i" class="ceb-item"
        :class="{ 'is-active': i === activeIndex }">
        <div class="ceb-item-label">{{ b.label }}</div>
        <!--
          `:key` là BẮT BUỘC, không phải tối ưu. ChessBoardDisplay.buildBoard()
          có `if (!boardEl.value || fenError.value) return` — FEN gõ dở gần như
          luôn không hợp lệ ở nhịp đầu, khi đó `board` là null VĨNH VIỄN:
          `boardEl` nằm trong nhánh v-else nên không được render, và
          applyPosition() có `if (board && …)` nên mọi cập nhật sau rơi vào hư
          không. Không có cơ chế tự phục hồi. Key đổi theo FEN ⇒ ép remount.
          (Cùng cách hai picker wikilink đang dùng: `:key="picker.previewRef"`.)
        -->
        <ChessBoardDisplay :key="b.key" :data="b.data" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Panel bàn cờ đứng CẠNH ô soạn, dùng chung cho cả ba chỗ soạn nội dung cờ:
 * chương sách (BookLibrary), bài giảng (ChessCourses), bài viết (ArticleBank).
 *
 * Vì sao đứng cạnh chứ không nằm trong dòng chữ: `<textarea>` không chứa được
 * HTML, nên không thể vẽ bàn cờ xen giữa văn bản nếu không thay bằng trình soạn
 * thảo giàu (CodeMirror/ProseMirror). Cách khả thi là đặt cạnh và BÁM THEO CON
 * TRỎ — con trỏ đang ở khối ```chess nào thì bàn cờ đó sáng lên và tự cuộn tới.
 *
 * Vì `ChessBoardDisplay` đã tự hiện hộp báo lỗi khi FEN hỏng (xem đợt sửa nhãn
 * "fen:"), panel này đồng thời là bộ KIỂM TRA FEN TẠI CHỖ — trước đây
 * `isValidFEN` chỉ chạy lúc bấm Lưu.
 */
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import { findChessBlockIndexAt, extractChessBlocks } from '@/utils/chessBlocks';
import type { ChessBoardData } from '@/types/tool-results';

const props = defineProps<{
  /** Markdown đang gõ — mọi khối ```chess trong đó thành một bàn cờ. */
  content?: string;
  /** Ô "Thế cờ FEN minh họa" riêng (BookLibrary, ChessCourses). */
  fen?: string;
  /** Ô "Ván minh họa PGN" riêng (chỉ ChessCourses). */
  pgn?: string;
  /** Phần tử textarea NATIVE để theo dõi con trỏ. Có thể null lúc chưa mount. */
  textarea?: HTMLTextAreaElement | null;
}>();

const OPEN_STORAGE_KEY = 'weknora_chess_editor_preview_open';
// Mặc định MỞ: người dùng bật tính năng này lên là để nhìn thấy bàn cờ.
const open = ref(localStorage.getItem(OPEN_STORAGE_KEY) !== 'false');
watch(open, (v) => {
  try {
    localStorage.setItem(OPEN_STORAGE_KEY, String(v));
  } catch {
    /* chế độ riêng tư / chặn site data — bỏ qua, không để hỏng vùng soạn thảo */
  }
});

// ── Theo dõi con trỏ ───────────────────────────────────────────────────────
// Repo chưa có ref reactive nào lưu vị trí con trỏ: ChessWikiLinkSuggest và các
// hàm insertAtCursor đều đọc `selectionStart` tại chỗ. Nên tự giữ ở đây.
const caret = ref(0);

function readCaret(e: Event) {
  const ta = e.currentTarget as HTMLTextAreaElement | null;
  if (ta) caret.value = ta.selectionStart ?? 0;
}

// LƯU Ý: KHÔNG copy nhánh bỏ qua ArrowUp/ArrowDown/Enter/Tab/Escape của
// ChessWikiLinkSuggest — với popup gợi ý thì đúng (keydown đã xử lý), nhưng ở
// đây nó khiến việc di chuyển con trỏ BẰNG BÀN PHÍM không được ghi nhận, bàn cờ
// đứng im khi Thầy nhảy giữa các khối bằng phím mũi tên.
const CARET_EVENTS = ['keyup', 'click', 'input', 'focus', 'select'] as const;

function bind(ta: HTMLTextAreaElement) {
  for (const ev of CARET_EVENTS) ta.addEventListener(ev, readCaret);
  caret.value = ta.selectionStart ?? 0;
}
function unbind(ta: HTMLTextAreaElement) {
  for (const ev of CARET_EVENTS) ta.removeEventListener(ev, readCaret);
}

watch(
  () => props.textarea,
  (el, old) => {
    if (old) unbind(old);
    if (el) bind(el);
  },
  { immediate: true },
);

// ── Ảnh chụp có độ trễ của nội dung ────────────────────────────────────────
// Repo KHÔNG có tiện ích debounce dùng chung (không lodash, không @vueuse/core)
// nên viết cục bộ; khuôn lấy từ composables/useEmbedChatSession.ts. Không
// debounce thì mỗi phím gõ ép remount cả loạt bàn cờ (xem ghi chú `:key`).
const DEBOUNCE_MS = 250;
const snap = ref({ content: '', fen: '', pgn: '' });
let timer: ReturnType<typeof setTimeout> | undefined;

watch(
  () => [props.content || '', props.fen || '', props.pgn || ''] as const,
  ([content, fen, pgn]) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      snap.value = { content, fen, pgn };
    }, DEBOUNCE_MS);
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  clearTimeout(timer);
  if (props.textarea) unbind(props.textarea);
});

interface PreviewBoard {
  key: string;
  label: string;
  data: ChessBoardData;
}

/** Số bàn cờ đứng TRƯỚC danh sách khối (chỉ có ô FEN/PGN riêng, tối đa 1). */
const leadCount = computed(() => (snap.value.fen.trim() || snap.value.pgn.trim() ? 1 : 0));

const boards = computed<PreviewBoard[]>(() => {
  const out: PreviewBoard[] = [];
  const { content, fen, pgn } = snap.value;

  // Ô FEN/PGN riêng đứng trước — nó là hình minh họa mở đầu của cả chương/bài.
  if (leadCount.value) {
    const data: ChessBoardData = { display_type: 'chess_board', fen: fen.trim() };
    if (pgn.trim()) data.pgn = pgn.trim();
    out.push({ key: `own:${fen}:${pgn.length}`, label: 'Ô thế cờ minh họa', data });
  }

  extractChessBlocks(content).forEach((data, i) => {
    out.push({
      key: `blk:${i}:${data.fen}:${(data.pgn || '').length}`,
      label: data.caption || `Khối thế cờ #${i + 1}`,
      data,
    });
  });

  return out;
});

// Khối con trỏ đang đứng trong, quy về chỉ số của danh sách `boards`.
// Dùng nội dung ĐÃ debounce để khoảng vị trí khớp với danh sách đang hiển thị;
// lệch tạm vài ký tự lúc gõ không đổi được khối nào đang hoạt động.
const activeIndex = computed(() => {
  const i = findChessBlockIndexAt(snap.value.content, caret.value);
  return i < 0 ? -1 : i + leadCount.value;
});

const listEl = ref<HTMLElement | null>(null);

watch(activeIndex, async (i) => {
  if (i < 0 || !open.value) return;
  await nextTick();
  const item = listEl.value?.querySelector<HTMLElement>(`[data-idx="${i}"]`);
  // `block: 'nearest'` để panel chỉ cuộn khi bàn cờ thật sự nằm ngoài tầm nhìn —
  // không giật mỗi lần con trỏ nhích trong cùng một khối.
  item?.scrollIntoView({ block: 'nearest' });
});
</script>

<style lang="less" scoped>
/* Panel là con flex của khung soạn thảo (xem .bkl-editor-split ở BookLibrary).
   `align-self: flex-start` — KHÔNG stretch — vì textarea dùng autosize nên chiều
   cao đổi liên tục. */
.ceb {
  flex: 0 0 250px;
  max-width: 250px;
  align-self: flex-start;
  border: 1px solid var(--td-component-stroke);
  border-radius: 8px;
  background: var(--td-bg-color-container);
  box-sizing: border-box;

  /* Thu gọn: trả hết chỗ cho ô soạn, chỉ còn cái nút. */
  &.is-collapsed {
    flex: 0 0 auto;
    max-width: none;
  }
}

.ceb-head {
  padding: 4px 6px;
}

.ceb-toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  padding: 4px 2px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  color: var(--td-text-color-primary);

  &:hover {
    color: var(--td-brand-color);
  }
}

.is-collapsed .ceb-toggle-text {
  /* Thu gọn thì chỉ chừa mũi tên + số, khỏi chiếm bề ngang */
  display: none;
}

.ceb-count {
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
  border-radius: 999px;
  padding: 0 7px;
  font-size: 12px;
}

.ceb-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0 8px 8px;
  /* Trần chiều cao đặt trên CHÍNH vùng này, KHÔNG đặt lên .t-dialog__body —
     wrap của dialog đã lo cuộn, đụng vào đó là tái tạo bẫy đơn vị viewport
     dưới CSS zoom (xem khối C trong assets/theme/duongsinh-responsive.css). */
  max-height: 420px;
  overflow-y: auto;
}

.ceb-empty {
  font-size: 13px;
  color: var(--td-text-color-placeholder);
}

.ceb-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
}

.ceb-item {
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 4px;
  min-width: 0;

  /* Khối con trỏ đang đứng trong — dùng màu nhận diện navy, không phải cam. */
  &.is-active {
    border-color: var(--td-brand-color);
    background: var(--td-bg-color-secondarycontainer);
  }
}

.ceb-item-label {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  margin-bottom: 2px;
  /* Nhãn có thể là caption dài do người soạn gõ — cắt gọn thay vì phá bố cục */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Bàn cờ tự co theo khung (cm-chessboard chạy `responsive: true`); chỉ cần gỡ
   lề mặc định 12px và trần 420px của component để nó vừa cột hẹp. */
.ceb-item :deep(.chess-board-display) {
  margin: 0;
  max-width: 100%;
}

/* ── Điện thoại ────────────────────────────────────────────────────────────
   `screen and` là BẮT BUỘC — thiếu chữ đó thì luật rò sang @media print của
   BookPrint.vue/ArticlePrint.vue (xem PuzzleBank.vue). */
@media screen and (max-width: 767px) {
  .ceb {
    flex: 1 1 auto;
    max-width: none;
    align-self: stretch;
  }

  .ceb-list {
    max-height: 300px;
  }
}
</style>
