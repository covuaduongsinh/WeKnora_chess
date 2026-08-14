<template>
  <t-dialog v-model:visible="localVisible" header="Kho tri thức cờ vua" :footer="false" width="680px">
    <div class="ckb">
      <div v-if="loading" class="ckb-empty">Đang tải…</div>

      <div v-else-if="forbidden" class="ckb-empty">
        Cần quyền <b>Contributor</b> trở lên để xem trạng thái kho tri thức.
      </div>

      <div v-else-if="loadError" class="ckb-empty">
        Không tải được trạng thái: {{ loadError }}
      </div>

      <template v-else-if="status">
        <!-- 4 điều kiện tiên quyết. Điều kiện nào đỏ thì mọi thứ bên dưới vô nghĩa,
             nên hiện kèm luôn câu xử lý thay vì bắt Thầy tra runbook. -->
        <div class="ckb-section-title">Điều kiện hoạt động</div>
        <div v-for="c in conditions" :key="c.key" class="ckb-cond" :class="c.ok ? 'is-ok' : 'is-bad'">
          <t-icon :name="c.ok ? 'check-circle-filled' : 'error-circle-filled'" />
          <div class="ckb-cond-body">
            <div class="ckb-cond-label">{{ c.label }}</div>
            <div v-if="!c.ok" class="ckb-cond-fix">{{ c.fix }}</div>
          </div>
        </div>

        <template v-if="status.kb_exists">
          <div class="ckb-section-title">Đã đẩy vào kho (theo loại)</div>
          <div v-if="byTypeRows.length === 0" class="ckb-empty ckb-empty--sm">
            Chưa có mục nào được đẩy vào kho. Bấm "Đẩy lại index" bên dưới.
          </div>
          <div v-else class="ckb-types">
            <div v-for="row in byTypeRows" :key="row.type" class="ckb-type">
              <span class="ckb-type-label">{{ row.label }}</span>
              <span class="ckb-type-count">{{ row.count }}</span>
            </div>
          </div>

          <div class="ckb-section-title">Tiến độ embedding</div>
          <!-- Hai con số này KHÁC nhau và đây đúng là chỗ từng bị hiểu nhầm thành
               "thành công giả": đẩy vào kho xong KHÔNG có nghĩa đã embed. Chỉ khi
               'Xong' bằng tổng thì agent mới thực sự tìm thấy. -->
          <div class="ckb-progress">
            <span class="ckb-pill ckb-pill--ok">Xong: {{ status.completed }}</span>
            <span class="ckb-pill" :class="{ 'ckb-pill--warn': status.pending > 0 }">Đang chờ: {{ status.pending }}</span>
            <span class="ckb-pill" :class="{ 'ckb-pill--bad': status.failed > 0 }">Lỗi: {{ status.failed }}</span>
            <span class="ckb-pill">Tổng: {{ status.total }}</span>
          </div>
          <div v-if="status.pending > 0" class="ckb-hint">
            Embedding chạy nền — chờ khoảng một phút rồi bấm "Làm mới". Chỉ khi "Xong" bằng "Tổng" thì
            HLV Cờ vua mới trích dẫn được.
          </div>
          <div v-if="status.disabled_docs > 0" class="ckb-hint ckb-hint--warn">
            {{ status.disabled_docs }} tài liệu đang tắt → không truy hồi được. Thử đẩy lại index.
          </div>
          <div v-if="status.sample_error" class="ckb-error">Lỗi mẫu: {{ status.sample_error }}</div>
        </template>

        <div v-if="reindexResult" class="ckb-result">
          Đã đẩy đi <b>{{ reindexResult.enqueued }}</b> mục
          (bài viết {{ reindexResult.articles_total }} · sách {{ reindexResult.books_total }} ·
          chương {{ reindexResult.chapters_total }} · ván {{ reindexResult.games_total }} ·
          bài tập {{ reindexResult.puzzles_total }} · thế cờ {{ reindexResult.positions_total }}).
          <span v-if="reindexResult.failed > 0" class="ckb-result-failed">Lỗi: {{ reindexResult.failed }}.</span>
          Embedding chạy nền — chờ ~1 phút rồi bấm "Làm mới".
        </div>

        <div class="ckb-actions">
          <t-button variant="outline" :loading="loading" @click="load">
            <template #icon><t-icon name="refresh" /></template>Làm mới
          </t-button>
          <t-button theme="primary" :loading="reindexing" :disabled="!status.enabled" @click="doReindex">
            <template #icon><t-icon name="cloud-upload" /></template>Đẩy lại index
          </t-button>
        </div>
      </template>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { getChessIndexStatus, reindexChessKB, type ChessIndexStatus, type ChessReindexResult } from '@/api/chess';

// Panel chẩn đoán + bảo trì kho tri thức cờ. Trước đây hai endpoint này chỉ gọi
// được bằng curl (không có UI nào) nên không cách gì biết bài viết/sách đã vào
// kho hay chưa. Các câu xử lý bên dưới lấy thẳng từ bảng chẩn đoán §4b của
// docs/chess-rag-enable.md — giữ đồng bộ khi sửa runbook.
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>();

const localVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
});

const loading = ref(false);
const reindexing = ref(false);
const forbidden = ref(false);
const loadError = ref('');
const status = ref<ChessIndexStatus | null>(null);
const reindexResult = ref<ChessReindexResult | null>(null);

// Nhãn tiếng Việt cho chess_type. Loại lạ (thêm về sau mà quên cập nhật đây)
// vẫn hiện được — rơi về chính chuỗi type thay vì biến mất khỏi bảng.
const TYPE_LABELS: Record<string, string> = {
  article: 'Bài viết', book: 'Sách', chapter: 'Chương',
  game: 'Ván cờ', puzzle: 'Bài tập', position: 'Thế cờ', lesson: 'Bài giảng',
};
const byTypeRows = computed(() => {
  const bt = status.value?.by_type || {};
  return Object.entries(bt)
    .filter(([, n]) => n > 0)
    .map(([type, count]) => ({ type, label: TYPE_LABELS[type] || type, count }))
    .sort((a, b) => b.count - a.count);
});

const conditions = computed(() => {
  const s = status.value;
  if (!s) return [];
  return [
    {
      key: 'enabled', ok: s.enabled,
      label: 'Đã bật index kho tri thức (CHESS_KB_INDEX)',
      fix: 'Chưa bật trên máy chủ. Đặt CHESS_KB_INDEX=true trong .env của service app rồi khởi động lại app. Khi chưa bật, mọi thao tác index đều bị bỏ qua âm thầm.',
    },
    {
      key: 'kb_exists', ok: s.kb_exists,
      label: 'Kho "Tri thức cờ vua" đã tồn tại',
      fix: 'Chưa có. Cần ít nhất một kho tri thức khác đã cấu hình embedding model để sao chép cấu hình, rồi bấm "Đẩy lại index".',
    },
    {
      key: 'embedding', ok: s.embedding_configured,
      label: 'Kho đã có embedding model',
      fix: 'Kho cờ chưa có embedding model → không embed được. Cấu hình embedding cho một kho bất kỳ, xóa kho "Tri thức cờ vua" rồi đẩy lại index.',
    },
    {
      key: 'searchable', ok: s.searchable,
      label: 'Agent nhìn thấy được kho (vector hoặc keyword)',
      fix: 'Kho đang TẮT cả vector lẫn keyword → HLV Cờ vua hoàn toàn không thấy kho này, và nội dung cũng không được embed. Xóa kho "Tri thức cờ vua" trong mục Kho tri thức rồi bấm "Đẩy lại index" (kho tạo mới sẽ bật sẵn).',
    },
  ];
});

async function load() {
  loading.value = true;
  forbidden.value = false;
  loadError.value = '';
  try {
    const res: any = await getChessIndexStatus();
    status.value = res?.data || null;
  } catch (e: any) {
    if (e?.status === 403) forbidden.value = true;
    else loadError.value = e?.message || e?.error || 'lỗi không rõ';
  } finally {
    loading.value = false;
  }
}

async function doReindex() {
  reindexing.value = true;
  try {
    const res: any = await reindexChessKB();
    reindexResult.value = res?.data || null;
    MessagePlugin.success('Đã đẩy lại index — embedding đang chạy nền');
    await load();
  } catch (e: any) {
    MessagePlugin.error(e?.message || e?.error || 'Đẩy lại index thất bại');
  } finally {
    reindexing.value = false;
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    reindexResult.value = null;
    load();
  }
});
</script>

<style scoped lang="less">
.ckb { display: flex; flex-direction: column; gap: 4px; }
.ckb-empty { color: var(--td-text-color-placeholder); font-size: 14px; padding: 16px 4px; }
.ckb-empty--sm { padding: 6px 4px; }
.ckb-section-title {
  font-size: 13px; font-weight: 600; color: var(--td-text-color-secondary);
  margin: 14px 0 6px;
}
.ckb-cond {
  display: flex; gap: 8px; align-items: flex-start;
  padding: 8px 10px; border-radius: 8px; margin-bottom: 6px;
  border: 1px solid var(--td-component-stroke);
  &.is-ok { color: var(--td-success-color); }
  &.is-bad { color: var(--td-error-color); border-color: var(--td-error-color-3); }
}
.ckb-cond-body { flex: 1; min-width: 0; }
.ckb-cond-label { font-size: 14px; color: var(--td-text-color-primary); }
.ckb-cond-fix { font-size: 12px; color: var(--td-text-color-secondary); margin-top: 4px; line-height: 1.5; }
.ckb-types { display: flex; flex-wrap: wrap; gap: 8px; }
.ckb-type {
  display: flex; gap: 6px; align-items: baseline;
  padding: 4px 10px; border-radius: 6px; background: var(--td-bg-color-container);
  border: 1px solid var(--td-component-stroke);
}
.ckb-type-label { font-size: 13px; color: var(--td-text-color-secondary); }
.ckb-type-count { font-size: 15px; font-weight: 600; color: var(--td-brand-color); }
.ckb-progress { display: flex; flex-wrap: wrap; gap: 8px; }
.ckb-pill {
  padding: 3px 10px; border-radius: 12px; font-size: 13px;
  background: var(--td-bg-color-container); color: var(--td-text-color-secondary);
  &--ok { color: var(--td-success-color); }
  &--warn { color: var(--td-warning-color); }
  &--bad { color: var(--td-error-color); }
}
.ckb-hint {
  font-size: 12px; color: var(--td-text-color-secondary); margin-top: 8px; line-height: 1.5;
  &--warn { color: var(--td-warning-color); }
}
.ckb-error {
  font-size: 12px; color: var(--td-error-color); margin-top: 8px;
  word-break: break-word; line-height: 1.5;
}
.ckb-result {
  margin-top: 14px; padding: 10px; border-radius: 8px; font-size: 13px; line-height: 1.6;
  background: var(--td-bg-color-container); color: var(--td-text-color-primary);
}
.ckb-result-failed { color: var(--td-error-color); }
.ckb-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 18px; }
</style>
