<template>
  <div class="atp">
    <div class="atp-toolbar no-print">
      <t-radio-group v-model="columns" variant="default-filled">
        <t-radio-button :value="1">1 cột</t-radio-button>
        <t-radio-button :value="2">2 cột</t-radio-button>
      </t-radio-group>
      <t-button theme="primary" @click="doPrint">
        <template #icon><t-icon name="print" /></template>In / Lưu PDF (Ctrl+P)
      </t-button>
    </div>
    <div v-if="loading" class="atp-empty">Đang tải…</div>
    <div v-else-if="!article" class="atp-empty">Không tìm thấy bài viết.</div>
    <article v-else class="atp-page" :class="{ 'atp-page--cols-2': columns === 2 }">
      <h1>{{ article.title }}</h1>
      <ul class="atp-meta">
        <li v-if="article.category">Thể loại: {{ articleCategoryLabel(article.category) }}</li>
        <li v-if="article.level">Cấp độ: {{ articleLevelLabel(article.level) }}</li>
        <li v-if="article.aliases">Tên gọi khác: {{ article.aliases }}</li>
        <li v-if="article.tags">Thẻ: {{ article.tags }}</li>
      </ul>
      <div v-if="article.summary" class="atp-summary" v-html="renderMd(article.summary)"></div>

      <template v-for="(seg, si) in contentSegments" :key="si">
        <ChessBoardDisplay v-if="seg.type === 'board'" :data="seg.board" />
        <div v-else-if="seg.type === 'markdown'" v-html="seg.html"></div>
      </template>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { marked } from 'marked';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import { splitChessSegments } from '@/utils/chessBlocks';
import { safeMarkdownToHTML, sanitizeHTML } from '@/utils/security';
import { getArticle, type ChessArticle } from '@/api/chess';
import { articleCategoryLabel, articleLevelLabel } from '@/utils/chessArticleOptions';

// Trang IN bài viết: toàn bộ nội dung trên một trang dài, mở tab mới rồi
// Ctrl+P → PDF. Rút gọn từ BookPrint.vue — bài viết là VĂN BẢN ĐƠN (không có
// khái niệm chương/phần), nên KHÔNG có dialog chọn chương lẫn tùy chọn "phần
// đầu sách". Cơ chế in (CSS @page, break rules, chess-board-display) sao y
// nguyên khối BookPrint.vue thay vì tách dùng chung — cố ý, xem lý do trong
// nhật ký tùy biến (đụng vào CSS đã dò tay kỹ càng để tiết kiệm dòng là đổi
// rủi ro hồi quy lấy code đẹp).
const route = useRoute();
const loading = ref(true);
const article = ref<ChessArticle | null>(null);

const COLUMNS_STORAGE_KEY = 'weknora_article_print_columns';
function readStoredColumns(): 1 | 2 {
  return localStorage.getItem(COLUMNS_STORAGE_KEY) === '2' ? 2 : 1;
}
const columns = ref<1 | 2>(readStoredColumns());
watch(columns, (v) => { localStorage.setItem(COLUMNS_STORAGE_KEY, String(v)); });

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
const contentSegments = computed(() => {
  if (!article.value) return [];
  return splitChessSegments(article.value.content || '').map((seg) => {
    if (seg.type === 'board') return { type: 'board' as const, board: seg.board! };
    return { type: 'markdown' as const, html: renderMd(stripWikilinkSyntax(seg.markdown || '')) };
  });
});

function doPrint() { window.print(); }

onMounted(async () => {
  const id = String(route.params.id || '');
  if (!id) { loading.value = false; return; }
  try {
    const res: any = await getArticle(id);
    article.value = res?.data || null;
  } finally {
    loading.value = false;
  }
});
</script>

<style lang="less" scoped>
@page { size: A4 portrait; margin: 15mm; }

.atp { padding: 24px; overflow-x: auto; }
.atp-toolbar { position: sticky; top: 0; padding: 10px 0; background: var(--td-bg-color-page); z-index: 10; display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
.atp-empty { color: var(--td-text-color-placeholder); padding: 40px; text-align: center; }

// Vùng nội dung = đúng bề rộng vùng in A4 (210mm − 2×15mm lề @page ở trên) —
// cùng nguyên tắc BookPrint.vue: bố cục lúc XEM và lúc IN khớp nhau tuyệt đối.
.atp-page {
  width: 180mm;
  margin: 0 auto;
  color: var(--td-text-color-primary);
  line-height: 1.7;
}
.atp-page--cols-2 {
  column-count: 2;
  column-gap: 8mm;
  column-fill: auto;
}

.atp-meta { font-size: 14px; color: var(--td-text-color-secondary); padding-left: 18px; }
.atp-summary { margin: 16px 0; padding: 12px; background: var(--td-bg-color-container); border-radius: 8px; font-style: italic; }

:deep(.chess-board-display) { break-inside: auto; max-width: 100%; }
:deep(.chess-body) { break-inside: avoid; }
:deep(.chess-nav),
:deep(.chess-nav-single),
:deep(.eval-bar),
:deep(.chess-eval-line) { display: none !important; }
:deep(.move-list) { max-height: none; overflow: visible; overflow-wrap: break-word; }
:deep(.move-item) {
  cursor: default;
  &:hover { background: none; }
  &.active { background: none !important; font-weight: 700; }
}
:deep(.move-comment) { orphans: 2; widows: 2; }
:deep(.chess-board-display:last-child) { margin-bottom: 0; }

@media print {
  .no-print { display: none !important; }
  .atp { padding: 0; overflow-x: visible; }
  .atp-page { max-width: 100%; }
  .atp-page > :last-child { margin-bottom: 0; }

  :root { color-scheme: light !important; }
  .atp-page,
  .atp-meta { color: #000 !important; }
  :deep(.chess-caption) { color: #000 !important; }
  .atp-summary { background: transparent !important; border: 1px solid #ccc; }
  :deep(.move-comment),
  :deep(.move-item),
  :deep(.move-no),
  :deep(.move-nag),
  :deep(.move-paren),
  :deep(.move-var) { color: #000 !important; }
}
</style>

<!--
  Khối style KHÔNG scoped — sao y BookPrint.vue: html/body/#app nằm NGOÀI
  component nên <style scoped> không nhắm tới được. Xem giải thích đầy đủ
  trong BookPrint.vue (App.vue đặt height:100%+transform cho #app chống xé
  hình lúc dùng bình thường — nhưng gây trang trắng thừa khi in).
-->
<style lang="less">
@media print {
  html,
  body,
  #app {
    height: auto !important;
    min-height: 0 !important;
    overflow: visible !important;
  }

  #app {
    transform: none !important;
    isolation: auto !important;
    backface-visibility: visible !important;
  }
}
</style>
