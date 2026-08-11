<template>
  <div class="chess-board-display">
    <div v-if="caption" class="chess-caption">{{ caption }}</div>

    <div class="chess-body">
      <!-- Thanh đánh giá (nếu có điểm engine) -->
      <div v-if="hasEval" class="eval-bar" :title="evalText">
        <div class="eval-bar-fill" :style="evalFillStyle"></div>
      </div>

      <div ref="boardEl" class="chess-board"></div>
    </div>

    <!-- Thông tin đánh giá engine -->
    <div v-if="hasEval" class="chess-eval-line">
      <span class="eval-score" :class="{ 'eval-mate': data.is_mate }">{{ evalText }}</span>
      <span v-if="bestMoveLabel" class="eval-best">
        {{ t('chess.bestMove') }}: <strong>{{ bestMoveLabel }}</strong>
      </span>
      <span v-if="data.depth" class="eval-depth">{{ t('chess.depth') }} {{ data.depth }}</span>
    </div>

    <!-- Điều hướng nước đi (khi có nhiều thế cờ) -->
    <div v-if="positions.length > 1" class="chess-nav">
      <div class="chess-nav-controls">
        <button class="nav-btn" :disabled="currentIndex === 0" @click="goTo(0)" :title="t('chess.start')">⏮</button>
        <button class="nav-btn" :disabled="currentIndex === 0" @click="goTo(currentIndex - 1)" :title="t('chess.prev')">◀</button>
        <span class="nav-label">{{ currentLabel }}</span>
        <button class="nav-btn" :disabled="currentIndex === positions.length - 1" @click="goTo(currentIndex + 1)" :title="t('chess.next')">▶</button>
        <button class="nav-btn" :disabled="currentIndex === positions.length - 1" @click="goTo(positions.length - 1)" :title="t('chess.end')">⏭</button>
        <button class="nav-btn" @click="flip" :title="t('chess.flip')">⇅</button>
      </div>
    </div>

    <div v-else class="chess-nav-single">
      <button class="nav-btn" @click="flip" :title="t('chess.flip')">⇅ {{ t('chess.flip') }}</button>
    </div>

    <!-- Danh sách nước đi (kèm chú giải PGN nếu có: bình luận/NAG/nhánh phụ, xem
         utils/pgnAnnotated.ts) — TÁCH RIÊNG khỏi .chess-nav để BookPrint.vue có
         thể ẩn nút bấm điều hướng (vô nghĩa trên giấy) mà vẫn in được (xem
         :deep(.chess-nav) trong BookPrint.vue). Điều kiện dùng tokens (không
         phải positions) để PGN chỉ có bình luận, không nước nào, vẫn hiện được. -->
    <div v-if="tokens.length" class="move-list" :class="{ 'move-list--annotated': hasAnnotations }">
      <template v-for="(tk, i) in displayTokens" :key="i">
        <span v-if="tk.sep" class="move-sep">{{ ' ' }}</span>
        <span v-if="tk.kind === 'comment'" class="move-comment" :class="varClass(tk.depth)">{{ tk.text }}</span>
        <span v-else-if="tk.kind === 'open'" class="move-paren move-paren--open" :class="varClass(tk.depth)">(</span>
        <span v-else-if="tk.kind === 'close'" class="move-paren move-paren--close" :class="varClass(tk.depth)">)</span>
        <span v-else class="move-group" :class="varClass(tk.depth)">
          <span v-if="tk.num" class="move-no">{{ tk.num }}</span>
          <span
            class="move-item"
            :class="{ active: tk.posIndex !== null && tk.posIndex === currentIndex, 'move-item--static': tk.posIndex === null }"
            @click="tk.posIndex !== null && goTo(tk.posIndex)"
          >{{ tk.san }}<span v-if="tk.glyph" class="move-nag">{{ tk.glyph }}</span></span>
        </span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Chessboard, COLOR } from 'cm-chessboard/src/Chessboard.js';
import piecesUrl from 'cm-chessboard/assets/pieces/standard.svg?url';
import 'cm-chessboard/assets/chessboard.css';
import type { ChessBoardData } from '@/types/tool-results';
import {
  buildAnnotatedPgn, needsSpaceBefore, tokensFromPositions,
  type AnnotatedPos, type PgnToken,
} from '@/utils/pgnAnnotated';

const props = defineProps<{ data: ChessBoardData }>();
const { t } = useI18n();

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

const caption = computed(() => props.data.caption || '');

interface BoardModel { positions: AnnotatedPos[]; tokens: PgnToken[] }

// Xây mô hình hiển thị: ưu tiên plies từ backend (không kèm chú giải), rồi tới
// parse PGN ở client — ĐẦY ĐỦ chú giải (bình luận/NAG/nhánh phụ, xem
// utils/pgnAnnotated.ts thay cho chess.js loadPgn() vốn âm thầm vứt bỏ 3 thứ
// đó), cuối cùng là một FEN đơn lẻ. MỘT computed duy nhất sinh cả `positions`
// (mainline — dùng để tua bàn cờ, giữ NGUYÊN ngữ nghĩa cũ mà GameLibrary.vue
// đang phụ thuộc qua defineExpose bên dưới) lẫn `tokens` (toàn bộ, dùng để
// render move-list) — tránh parse PGN hai lần và tránh hai bộ hiển thị trôi
// lệch nhau theo thời gian.
const model = computed<BoardModel>(() => {
  const fallbackFen = props.data.fen || START_FEN;
  const startEntry = (fen: string): AnnotatedPos => (
    { fen, label: t('chess.startPosition'), san: '', moveNumber: 0, side: 'w' }
  );

  if (props.data.plies && props.data.plies.length) {
    const list: AnnotatedPos[] = props.data.plies.map((p) => {
      const side = p.side === 'b' ? 'b' : 'w';
      return {
        fen: p.fen_after, san: p.san, moveNumber: p.move_number, side,
        label: `${p.move_number}${side === 'w' ? '.' : '...'} ${p.san}`,
      };
    });
    return { positions: [startEntry(fallbackFen), ...list], tokens: tokensFromPositions(list) };
  }

  if (props.data.pgn) {
    const res = buildAnnotatedPgn(props.data.pgn, fallbackFen);
    if (res) {
      return { positions: [startEntry(res.startFen), ...res.positions], tokens: res.tokens };
    }
    // parse thất bại cả 2 tầng fallback (xem buildAnnotatedPgn) → rơi về FEN đơn.
  }

  return { positions: [startEntry(fallbackFen)], tokens: [] };
});

const positions = computed(() => model.value.positions);
const tokens = computed(() => model.value.tokens);
// Có chú giải (bình luận/nhánh phụ) hay chỉ toàn nước đi trần — quyết định có
// bỏ khung cuộn 120px của move-list hay không (xem CSS .move-list--annotated).
const hasAnnotations = computed(() => tokens.value.some((tk) => tk.kind !== 'move'));

// Dạng PHẲNG (không union) của PgnToken để render trong template. vue-tsc
// không thu hẹp được kiểu union qua NHIỀU tầng v-else-if nối tiếp (chỉ loại
// được nhánh đầu, các nhánh else-if sau vẫn giữ nguyên union) — đã xác nhận
// bằng cách chạy type-check thật, không phải suy đoán. Thay vì ép kiểu tay
// (dễ che giấu lỗi thật), làm phẳng dữ liệu TRƯỚC khi vào template: field nào
// không áp dụng cho loại token đó thì để giá trị mặc định an toàn ('' / null).
interface DisplayToken {
  kind: PgnToken['kind'];
  num: string;
  san: string;
  glyph: string;
  posIndex: number | null;
  depth: number;
  text: string; // nội dung bình luận — rỗng với token không phải comment
  sep: boolean; // có chèn khoảng trắng THẬT trước token này không (xem needsSpaceBefore)
}
const displayTokens = computed<DisplayToken[]>(() => {
  const src = tokens.value;
  return src.map((tk, i) => ({
    kind: tk.kind,
    num: tk.kind === 'move' ? tk.num : '',
    san: tk.kind === 'move' ? tk.san : '',
    glyph: tk.kind === 'move' ? tk.glyph : '',
    posIndex: tk.kind === 'move' ? tk.posIndex : null,
    depth: tk.depth,
    text: tk.kind === 'comment' ? tk.text : '',
    sep: needsSpaceBefore(i > 0 ? src[i - 1] : null, tk),
  }));
});

function varClass(depth: number) {
  return depth > 0 ? ['move-var', `move-var--d${Math.min(depth, 2)}`] : undefined;
}

const currentIndex = ref(0);
const currentLabel = computed(() => positions.value[currentIndex.value]?.label ?? '');
// FEN của thế cờ đang xem (nước hiện tại nếu đang tua ván). Lộ ra qua
// defineExpose bên dưới để component cha (vd GameLibrary) "kéo" FEN tại đúng
// thời điểm bấm nút — thay vì phát emit mỗi lần đổi nước, tránh sinh state
// trùng và tự miễn nhiễm với watch(positions) reset currentIndex bên dưới.
const currentFen = computed(() => positions.value[currentIndex.value]?.fen ?? '');

// ---- Đánh giá engine ----
const hasEval = computed(() =>
  props.data.is_mate === true ||
  typeof props.data.eval_cp === 'number' ||
  !!props.data.best_move,
);

const evalText = computed(() => {
  if (props.data.is_mate) {
    const n = props.data.mate_in ?? 0;
    return n >= 0 ? `#${n}` : `#-${Math.abs(n)}`;
  }
  if (typeof props.data.eval_cp === 'number') {
    // Quy về góc nhìn Trắng cho nhất quán hiển thị.
    const white = props.data.side_to_move === 'b' ? -props.data.eval_cp : props.data.eval_cp;
    const pawns = white / 100;
    return (pawns > 0 ? '+' : '') + pawns.toFixed(2);
  }
  return '';
});

const bestMoveLabel = computed(() => props.data.best_move_san || props.data.best_move || '');

// Vị trí con trỏ thanh đánh giá: 0% (Đen thắng) → 100% (Trắng thắng).
const evalFillStyle = computed(() => {
  let whiteAdvantage = 0.5;
  if (props.data.is_mate) {
    const n = props.data.mate_in ?? 0;
    const sideWhite = props.data.side_to_move !== 'b';
    const whiteMating = (n >= 0) === sideWhite;
    whiteAdvantage = whiteMating ? 1 : 0;
  } else if (typeof props.data.eval_cp === 'number') {
    const white = props.data.side_to_move === 'b' ? -props.data.eval_cp : props.data.eval_cp;
    // Hàm sigmoid mềm để giới hạn về [0,1].
    whiteAdvantage = 1 / (1 + Math.exp(-white / 400));
  }
  return { height: `${(whiteAdvantage * 100).toFixed(1)}%` };
});

// ---- cm-chessboard ----
const boardEl = ref<HTMLElement | null>(null);
let board: any = null;
const orientation = ref<'white' | 'black'>(props.data.orientation || 'white');

function buildBoard() {
  if (!boardEl.value) return;
  board = new Chessboard(boardEl.value, {
    position: positions.value[currentIndex.value]?.fen || START_FEN,
    orientation: orientation.value === 'black' ? COLOR.black : COLOR.white,
    assetsUrl: '/',
    style: {
      pieces: { file: piecesUrl },
      borderType: 'frame',
      showCoordinates: true,
    },
    responsive: true,
  });
}

function applyPosition() {
  const fen = positions.value[currentIndex.value]?.fen;
  if (board && fen) {
    board.setPosition(fen, true);
  }
}

function goTo(idx: number) {
  if (idx < 0 || idx >= positions.value.length) return;
  currentIndex.value = idx;
  applyPosition();
}

function flip() {
  orientation.value = orientation.value === 'white' ? 'black' : 'white';
  if (board) {
    board.setOrientation(orientation.value === 'black' ? COLOR.black : COLOR.white);
  }
}

onMounted(async () => {
  await nextTick();
  buildBoard();
});

onBeforeUnmount(() => {
  if (board) {
    try { board.destroy(); } catch { /* noop */ }
    board = null;
  }
});

// Khi dữ liệu đổi (stream cập nhật), dựng lại danh sách và đồng bộ bàn cờ.
watch(positions, () => {
  currentIndex.value = 0;
  applyPosition();
});

// Cho phép component cha đọc thế cờ đang xem (vd nút "Lưu thế cờ này" trong
// GameLibrary khi tua ván) mà không cần biết chỉ số nước nội bộ.
defineExpose({ currentFen, currentIndex, currentLabel });
</script>

<style lang="less" scoped>
.chess-board-display {
  margin: 12px 0;
  max-width: 420px;
}

.chess-caption {
  font-size: 13px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin-bottom: 8px;
}

.chess-body {
  display: flex;
  align-items: stretch;
  gap: 8px;
}

.eval-bar {
  width: 12px;
  border-radius: 4px;
  background: #3a3a3a;
  overflow: hidden;
  display: flex;
  flex-direction: column-reverse;
  flex: 0 0 auto;
}

.eval-bar-fill {
  width: 100%;
  background: #f0f0f0;
  transition: height 0.3s ease;
}

.chess-board {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
}

.chess-eval-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  font-size: 13px;
  color: var(--td-text-color-secondary);

  .eval-score {
    font-weight: 700;
    font-family: var(--app-font-family-mono);
    color: var(--td-text-color-primary);
  }

  .eval-mate {
    color: var(--td-error-color, #d54941);
  }
}

.chess-nav {
  margin-top: 10px;
}

.chess-nav-controls {
  display: flex;
  align-items: center;
  gap: 6px;
}

.chess-nav-single {
  margin-top: 8px;
}

.nav-btn {
  border: 1px solid var(--td-component-stroke);
  background: var(--td-bg-color-container);
  color: var(--td-text-color-primary);
  border-radius: 6px;
  padding: 2px 8px;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.6;

  &:hover:not(:disabled) {
    background: var(--td-bg-color-secondarycontainer);
  }

  &:disabled {
    opacity: 0.4;
    cursor: default;
  }
}

.nav-label {
  flex: 1 1 auto;
  text-align: center;
  font-family: var(--app-font-family-mono);
  font-size: 13px;
  color: var(--td-text-color-primary);
}

.move-list {
  // display:block (không phải flex) — bình luận là khối chữ chảy dòng thật, và
  // ")" của nhánh phụ không được có khoảng trắng phía trước; khoảng cách giữa
  // các nhóm nước dùng margin-left thay cho gap (xem .move-group bên dưới).
  display: block;
  margin-top: 8px;
  line-height: 1.7;
  max-height: 120px;
  overflow-y: auto;

  // Có chú giải (bình luận/nhánh phụ) ⇒ đây là nội dung sách, phải hiện đủ,
  // không nhốt trong khung cuộn 120px nghĩ cho ván cờ trần.
  &.move-list--annotated {
    max-height: none;
    overflow: visible;
  }

  // Khoảng trắng giữa các nước PHẢI là text node thật (span này) chứ không phải
  // margin: trình biên dịch Vue xoá sạch whitespace giữa các thẻ anh em trong
  // template (chế độ condense mặc định — đã xác minh bằng cách biên dịch thật),
  // mà margin thì KHÔNG tạo điểm ngắt dòng. Thiếu nó, cả chuỗi nước đi dài
  // thành MỘT dòng không ngắt được → tràn khỏi cột và tràn ngang khỏi khổ giấy
  // khi in (trình duyệt sinh thêm trang trắng). Xem needsSpaceBefore().
  .move-sep { white-space: normal; }

  .move-group {
    display: inline;
    white-space: nowrap; // "7." và "a3" không bị tách qua 2 dòng
  }

  .move-no {
    font-family: var(--app-font-family-mono);
    font-size: 12px;
    color: var(--td-text-color-placeholder);
    user-select: none;
    margin-right: 4px; // 4 + 2(pad move-item) = 6px giữa "7." và "a3"
  }

  .move-item {
    font-family: var(--app-font-family-mono);
    font-size: 12px;
    color: var(--td-text-color-secondary);
    cursor: pointer;
    padding: 0 2px;
    border-radius: 3px;

    &:hover {
      background: var(--td-bg-color-secondarycontainer);
    }

    &.active {
      background: var(--td-brand-color, #0052d9);
      color: #fff;
    }

    // Nhánh phụ (RAV): chữ tĩnh, không bấm tua bàn cờ được (quyết định đã chốt
    // — tránh đụng state điều hướng currentIndex/currentFen ngoài mainline).
    &.move-item--static {
      cursor: default;
      &:hover { background: none; }
    }
  }

  // Ký hiệu NAG (⩲ ⩱ ∓ …) thiếu glyph trong nhiều font mono trên Windows.
  .move-nag {
    font-family: 'Segoe UI Symbol', 'DejaVu Sans', 'Noto Sans Symbols 2', sans-serif;
    font-weight: 600;
    margin-left: 1px;
  }

  // Bình luận: khối chữ riêng (bình luận sách cờ thường rất dài).
  .move-comment {
    display: block;
    margin: 6px 0;
    font-size: 12.5px;
    line-height: 1.65;
    color: var(--td-text-color-primary);
    text-align: justify;
  }

  // Ngoặc nhánh phụ dính sát nước bên trong — "(7. O-O O-O 8. a3)" chứ không
  // phải "( 7. O-O O-O 8. a3 )"; do needsSpaceBefore() bỏ khoảng trắng ở 2 chỗ đó.
  .move-paren {
    font-family: var(--app-font-family-mono);
    font-size: 12px;
    color: var(--td-text-color-placeholder);
  }

  // Nhánh phụ: nghiêng + nhạt màu để phân biệt mainline. Bình luận NẰM TRONG
  // nhánh phụ thì hiện inline (không xuống dòng riêng) — đúng cách sách in.
  .move-var {
    color: var(--td-text-color-placeholder);

    .move-item, .move-no { color: inherit; font-style: italic; }

    &.move-comment {
      display: inline;
      margin: 0 0 0 4px;
      font-style: italic;
      text-align: inherit;
      color: inherit;
    }
  }
  .move-var--d2 { font-size: 11px; }
}
</style>
