<template>
  <div class="csp">
    <div class="csp-bar">
      <t-input
        v-model="q"
        placeholder="Tìm trong toàn bộ kho cờ — sách, chương, bài viết, bài giảng, ván, thế cờ, bài tập…"
        clearable
        size="large"
        @change="onType"
        @enter="run(1)"
      >
        <template #prefix-icon><t-icon name="search" /></template>
      </t-input>
    </div>

    <div class="csp-filters">
      <t-select v-model="level" :options="chessLevelOptions" placeholder="Cấp độ" clearable
        style="width:120px" @change="run(1)" />
      <t-select v-model="status" :options="statusOptions" placeholder="Trạng thái" clearable
        style="width:140px" @change="run(1)" />
      <div class="csp-tagfilter">
        <ChessTagInput v-model="tag" placeholder="Lọc theo thẻ…" />
      </div>
      <div class="csp-spacer"></div>
      <span class="csp-hint">Gõ không dấu vẫn tìm được — “tan cuoc” ra “Tàn cuộc”.</span>
    </div>

    <div v-if="typeCounts.length" class="csp-types">
      <button class="csp-type" :class="{ 'is-active': type === '' }" @click="setType('')">
        Tất cả <span>{{ page?.total ?? 0 }}</span>
      </button>
      <button v-for="[ct, n] in typeCounts" :key="ct" class="csp-type"
        :class="{ 'is-active': type === ct }" @click="setType(ct)">
        {{ chessTypeLabel(ct) }} <span>{{ n }}</span>
      </button>
    </div>

    <p v-if="page?.truncated" class="csp-warn">
      Kết quả có thể chưa đầy đủ vì từ khóa quá rộng — thử thêm chữ hoặc lọc theo loại/thẻ.
    </p>

    <div v-if="loading" class="csp-empty">Đang tìm…</div>
    <div v-else-if="!hasCriteria" class="csp-empty">
      Nhập từ khóa, hoặc chọn một thẻ / cấp độ để duyệt.
    </div>
    <div v-else-if="!page?.items?.length" class="csp-empty">
      Không tìm thấy gì khớp. Thử bỏ bớt bộ lọc, hoặc gõ ngắn hơn.
    </div>
    <div v-else class="csp-results">
      <a
        v-for="it in page.items"
        :key="`${it.chess_type}:${it.chess_id}`"
        class="csp-hit"
        :href="chessRefLink(it.chess_type, it.slug)"
        @click.prevent="open(it)"
      >
        <span class="csp-hit-type">{{ chessTypeLabel(it.chess_type) }}</span>
        <span class="csp-hit-body">
          <span class="csp-hit-title">{{ it.title || it.slug }}</span>
          <span v-if="it.subtitle" class="csp-hit-sub">{{ it.subtitle }}</span>
          <span v-if="it.snippet" class="csp-hit-snippet">{{ it.snippet }}</span>
          <ChessTagChips v-if="it.tags?.length" :tags="it.tags" @pick="pickTag" />
        </span>
        <span class="csp-hit-meta">
          <span v-if="it.level" class="csp-badge">{{ chessLevelLabel(it.level) }}</span>
          <span v-if="it.status === 'draft'" class="csp-badge csp-badge--draft"
            title="Bản thảo — chưa vào kho tri thức">Bản thảo</span>
        </span>
      </a>

      <t-pagination
        v-if="page.total > page.page_size"
        class="csp-pager"
        :total="page.total"
        :page-size="page.page_size"
        :current="page.page"
        :show-jumper="false"
        :show-page-size="false"
        @current-change="run"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * SearchPanel — ô tìm kiếm HỢP NHẤT của khu Quản lý cờ vua.
 *
 * Trước đợt này không có chỗ nào gõ một từ khóa mà ra được cả sách + bài viết
 * + ván cờ: mỗi tab tự lọc riêng, và endpoint gộp duy nhất
 * (/chess/refs/search) chỉ phục vụ autocomplete wikilink — nó không chấm điểm
 * và nối kết quả theo thứ tự cứng của từng loại trong code.
 *
 * Bấm một kết quả sẽ dùng lại deep-link ?ref= sẵn có của ChessManage, nên
 * không cần dựng route riêng cho từng loại.
 */
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { searchChess, type ChessSearchPage, type ChessSearchHit } from '@/api/chess';
import { chessLevelOptions, chessLevelLabel, chessTypeLabel, chessRefLink } from '@/utils/chessTaxonomy';
import ChessTagInput from '@/views/chess/components/ChessTagInput.vue';
import ChessTagChips from '@/views/chess/components/ChessTagChips.vue';
import { debounceFn } from '@/views/chess/composables/useChessPaging';

const props = withDefaults(defineProps<{ initialQuery?: string }>(), { initialQuery: '' });

const router = useRouter();
const q = ref(props.initialQuery);
const level = ref('');
const status = ref('');
const tag = ref('');
const type = ref('');
const page = ref<ChessSearchPage | null>(null);
const loading = ref(false);
// reqSeq chặn race: phản hồi của một từ khóa CŨ về sau có thể ghi đè kết quả
// mới hơn. Cùng kỹ thuật với popup gợi ý wikilink.
let reqSeq = 0;

const statusOptions = [
  { label: 'Bản thảo', value: 'draft' },
  { label: 'Đã xuất bản', value: 'published' },
];

const hasCriteria = computed(() => !!q.value.trim() || !!tag.value.trim() || !!level.value);

const typeCounts = computed(() =>
  Object.entries(page.value?.by_type || {}).filter(([, n]) => n > 0),
);

async function run(p = 1) {
  if (!hasCriteria.value) {
    page.value = null;
    return;
  }
  const seq = ++reqSeq;
  loading.value = true;
  try {
    const res: any = await searchChess(q.value, {
      type: type.value,
      level: level.value,
      status: status.value,
      tag: tag.value,
      page: p,
      page_size: 20,
    });
    if (seq !== reqSeq) return; // đã có truy vấn mới hơn
    page.value = res?.data || null;
  } catch {
    if (seq === reqSeq) MessagePlugin.error('Tìm kiếm thất bại');
  } finally {
    if (seq === reqSeq) loading.value = false;
  }
}

const runDebounced = debounceFn(() => run(1));
function onType() {
  runDebounced();
}

function setType(ct: string) {
  // Đổi bộ lọc loại KHÔNG tính lại by_type (nó phản ánh toàn bộ tập kết quả),
  // nên các con số trên chip giữ nguyên — đúng như người dùng mong đợi.
  type.value = ct;
  run(1);
}

function pickTag(name: string) {
  tag.value = name;
  run(1);
}

function open(it: ChessSearchHit) {
  router.push({ name: 'chessCourses', query: { ref: `${it.chess_type}/${it.slug}` } });
}

watch(() => props.initialQuery, (v) => {
  if (v && v !== q.value) {
    q.value = v;
    run(1);
  }
});

if (props.initialQuery) run(1);
</script>

<style scoped>
.csp {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}
.csp-bar {
  margin-bottom: 12px;
}
.csp-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.csp-tagfilter {
  width: 200px;
}
.csp-spacer {
  flex: 1;
}
.csp-hint {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}
.csp-types {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}
.csp-type {
  border: none;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
  border-radius: 12px;
  padding: 3px 10px;
  font-size: 12px;
  cursor: pointer;
}
.csp-type.is-active {
  background: var(--td-brand-color);
  color: #fff;
}
.csp-type span {
  opacity: 0.7;
  margin-left: 2px;
}
.csp-warn {
  margin: 0 0 10px;
  font-size: 12px;
  color: var(--td-warning-color);
}
.csp-results {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.csp-hit {
  display: flex;
  gap: 10px;
  padding: 10px;
  border-radius: 8px;
  text-decoration: none;
  color: inherit;
  cursor: pointer;
  border-bottom: 1px solid var(--td-border-level-1-color);
}
.csp-hit:hover {
  background: var(--td-bg-color-container-hover);
}
.csp-hit-type {
  flex: 0 0 76px;
  font-size: 11px;
  color: var(--td-text-color-placeholder);
  padding-top: 2px;
}
.csp-hit-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.csp-hit-title {
  font-size: 14px;
  font-weight: 500;
}
.csp-hit-sub {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.csp-hit-snippet {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.csp-hit-meta {
  display: flex;
  align-items: flex-start;
  gap: 4px;
  flex-shrink: 0;
}
.csp-badge {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
  white-space: nowrap;
}
.csp-badge--draft {
  background: var(--td-warning-color-light);
  color: var(--td-warning-color);
}
.csp-empty {
  color: var(--td-text-color-placeholder);
  font-size: 13px;
  padding: 24px 0;
}
.csp-pager {
  margin: 16px 0;
}

/* Điện thoại: nhãn loại xuống dòng riêng để tiêu đề có đủ bề ngang.
   Tiền tố "screen and" là BẮT BUỘC — thiếu chữ screen thì luật rò sang
   @media print của BookPrint/ArticlePrint. */
@media screen and (max-width: 767px) {
  .csp-hit {
    flex-wrap: wrap;
  }
  .csp-hit-type {
    flex: 0 0 100%;
  }
  .csp-tagfilter {
    width: 100%;
  }
  .csp-hint {
    display: none;
  }
}
</style>
