<template>
  <div class="pb">
    <div class="pb-toolbar">
      <t-select v-model="filter.theme" :options="themeOptions" placeholder="Chủ đề" clearable
        style="width:160px" @change="load" />
      <t-select v-model="filter.difficulty" :options="diffOptions" placeholder="Độ khó" clearable
        style="width:140px" @change="load" />
      <div style="flex:1"></div>
      <t-button variant="outline" size="small" @click="practice">
        <template #icon><t-icon name="play-circle" /></template>Luyện ngẫu nhiên
      </t-button>
      <t-button variant="outline" size="small" @click="doExport">
        <template #icon><t-icon name="download" /></template>Export
      </t-button>
      <t-button variant="outline" size="small" @click="doImport">
        <template #icon><t-icon name="upload" /></template>Import
      </t-button>
      <t-button theme="primary" size="small" @click="openDialog()">
        <template #icon><t-icon name="add" /></template>Tạo bài tập
      </t-button>
    </div>

    <div class="pb-body">
      <div class="pb-list">
        <div v-if="puzzles.length === 0" class="pb-empty">Chưa có bài tập. Nhấn "Tạo bài tập".</div>
        <div v-for="p in puzzles" :key="p.id" class="pb-row" :class="{ active: selected && selected.id === p.id }"
          @click="select(p)">
          <div class="pb-row-main">
            <div class="pb-title">{{ p.title || '(không tiêu đề)' }}</div>
            <div class="pb-meta">
              <span v-if="p.theme" class="pb-tag">{{ p.theme }}</span>
              <span v-if="p.difficulty" class="pb-tag pb-tag--diff">{{ diffLabel(p.difficulty) }}</span>
            </div>
          </div>
          <span class="pb-actions">
            <t-button size="small" variant="text" :title="t('chess.ref.copyLink')" @click.stop="copyWikilink(p)">
              <t-icon name="link" />
            </t-button>
            <t-button size="small" variant="text" @click.stop="openDialog(p)"><t-icon name="edit" /></t-button>
            <t-button size="small" variant="text" theme="danger" @click.stop="remove(p)"><t-icon name="delete" /></t-button>
          </span>
        </div>
      </div>
      <div class="pb-viewer">
        <template v-if="selected">
          <ChessBacklinks v-if="selected.slug" ref-type="puzzle" :slug="selected.slug" show-empty class="pb-backlinks" />
          <ChessBoardDisplay :key="selected.id + revealKey" :data="viewerData" />
          <div class="pb-solution">
            <t-button v-if="!revealed" size="small" variant="outline" @click="revealed = true">Hiện đáp án</t-button>
            <div v-else class="pb-solution-text">
              <b>Đáp án:</b> {{ selected.solution || '(chưa nhập — hãy hỏi HLV Cờ vua phân tích thế cờ này)' }}
            </div>
          </div>
        </template>
        <div v-else class="pb-empty pb-empty--big">Chọn một bài tập hoặc bấm "Luyện ngẫu nhiên".</div>
      </div>
    </div>

    <t-dialog v-model:visible="dialog.visible" :header="dialog.id ? 'Sửa bài tập' : 'Tạo bài tập'"
      :on-confirm="save" width="600px">
      <div class="pb-form">
        <label>Tiêu đề</label>
        <t-input v-model="dialog.title" />
        <label>Thế cờ FEN *</label>
        <t-input v-model="dialog.fen" placeholder="rnbqkbnr/.../ w KQkq - 0 1" />
        <label>Lời giải (SAN/UCI, tùy chọn)</label>
        <t-input v-model="dialog.solution" placeholder="VD: Qxf7#" />
        <label>Chủ đề</label>
        <t-select v-model="dialog.theme" :options="themeOptions" creatable filterable clearable />
        <label>Độ khó</label>
        <t-select v-model="dialog.difficulty" :options="diffOptions" clearable />
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import ChessBacklinks from '@/views/chess/components/ChessBacklinks.vue';
import type { ChessBoardData } from '@/types/tool-results';
import {
  listPuzzles, getPuzzleBySlug, createPuzzle, updatePuzzle, deletePuzzle, randomPuzzle,
  exportPuzzles, importPuzzles, type ChessPuzzle,
} from '@/api/chess';
import { downloadText, pickTextFile } from '@/utils/fileTransfer';

const { t } = useI18n();

// Deep-link "Mở trong thư viện": chọn sẵn thế cờ theo slug (từ [[puzzle/<slug>]]).
const props = defineProps<{ focusSlug?: string }>();
async function focusBySlug(slug?: string) {
  if (!slug) return;
  try {
    const res: any = await getPuzzleBySlug(slug);
    if (res?.data) { selected.value = res.data; revealed.value = false; revealKey.value++; }
  } catch { /* không tìm thấy → bỏ qua */ }
}
onMounted(() => focusBySlug(props.focusSlug));
watch(() => props.focusSlug, (s) => focusBySlug(s));

// Sao chép wikilink [[puzzle/<slug>]] để dán vào nội dung wiki/bài giảng.
async function copyWikilink(p: ChessPuzzle) {
  if (!p.slug) { MessagePlugin.warning('Bài tập chưa có slug'); return; }
  const link = `[[puzzle/${p.slug}|${p.title || p.slug}]]`;
  try {
    await navigator.clipboard.writeText(link);
    MessagePlugin.success(t('chess.ref.copied'));
  } catch {
    MessagePlugin.info(link);
  }
}

const themeOptions = [
  { label: 'Chiếu hết', value: 'chiếu hết' }, { label: 'Chiến thuật', value: 'chiến thuật' },
  { label: 'Khai cuộc', value: 'khai cuộc' }, { label: 'Tàn cuộc', value: 'tàn cuộc' },
];
const diffOptions = [
  { label: 'Dễ', value: 'de' }, { label: 'Trung bình', value: 'trung-binh' }, { label: 'Khó', value: 'kho' },
];
const diffLabel = (v: string) => diffOptions.find(o => o.value === v)?.label || v;

const puzzles = ref<ChessPuzzle[]>([]);
const selected = ref<ChessPuzzle | null>(null);
const filter = reactive({ theme: '', difficulty: '' });
const revealed = ref(false);
const revealKey = ref(0);

const viewerData = computed<ChessBoardData>(() => ({
  display_type: 'chess_board',
  fen: selected.value?.fen || 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
  caption: selected.value?.title || 'Bài tập',
}));

async function load() {
  try {
    const res: any = await listPuzzles(filter);
    puzzles.value = res?.data || [];
  } catch { MessagePlugin.error('Tải bài tập thất bại'); }
}
function select(p: ChessPuzzle) { selected.value = p; revealed.value = false; revealKey.value++; }
async function practice() {
  try {
    const res: any = await randomPuzzle(filter);
    if (res?.data) { selected.value = res.data; revealed.value = false; revealKey.value++; }
    else MessagePlugin.info('Chưa có bài tập phù hợp');
  } catch { MessagePlugin.info('Chưa có bài tập phù hợp bộ lọc'); }
}

async function doExport() {
  try {
    const res: any = await exportPuzzles(filter);
    const items = res?.data || [];
    if (!items.length) { MessagePlugin.info('Không có bài tập để xuất'); return; }
    downloadText(`baitap-${new Date().toISOString().slice(0, 10)}.json`, JSON.stringify(items, null, 2), 'application/json');
    MessagePlugin.success(`Đã xuất ${items.length} bài tập`);
  } catch { MessagePlugin.error('Xuất thất bại'); }
}
async function doImport() {
  const text = await pickTextFile('.json,application/json');
  if (text == null) return;
  let arr: any;
  try { arr = JSON.parse(text); } catch { MessagePlugin.error('File JSON không hợp lệ'); return; }
  if (!Array.isArray(arr)) { MessagePlugin.error('File phải là một mảng JSON bài tập'); return; }
  try {
    const res: any = await importPuzzles(arr);
    await load();
    MessagePlugin.success(`Đã nhập ${res?.data?.imported || 0} bài tập`);
  } catch (e: any) { MessagePlugin.error(e?.error || e?.message || 'Import thất bại'); }
}

const dialog = reactive<any>({ visible: false, id: '', title: '', fen: '', solution: '', theme: '', difficulty: '' });
function openDialog(p?: ChessPuzzle) {
  dialog.visible = true;
  dialog.id = p?.id || '';
  dialog.title = p?.title || '';
  dialog.fen = p?.fen || '';
  dialog.solution = p?.solution || '';
  dialog.theme = p?.theme || '';
  dialog.difficulty = p?.difficulty || '';
}
async function save() {
  if (!dialog.fen.trim()) { MessagePlugin.warning('Nhập thế cờ FEN'); return; }
  const payload = {
    title: dialog.title, fen: dialog.fen, solution: dialog.solution,
    theme: dialog.theme, difficulty: dialog.difficulty,
  };
  try {
    if (dialog.id) await updatePuzzle(dialog.id, payload);
    else await createPuzzle(payload);
    dialog.visible = false;
    await load();
    MessagePlugin.success('Đã lưu bài tập');
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Lưu thất bại (kiểm tra FEN)');
  }
}
function remove(p: ChessPuzzle) {
  DialogPlugin.confirm({
    header: 'Xóa bài tập', body: `Xóa "${p.title || 'bài tập'}"?`,
    theme: 'warning', confirmBtn: { content: 'Xóa', theme: 'danger' },
    onConfirm: async () => {
      try {
        await deletePuzzle(p.id);
        if (selected.value?.id === p.id) selected.value = null;
        await load();
        MessagePlugin.success('Đã xóa');
      } catch { MessagePlugin.error('Xóa thất bại'); }
    },
  });
}
load();
</script>

<style lang="less" scoped>
.pb { display: flex; flex-direction: column; height: 100%; }
.pb-toolbar { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; flex-wrap: wrap; }
.pb-body { display: flex; gap: 16px; flex: 1; min-height: 0; }
.pb-list { width: 340px; flex: 0 0 340px; overflow-y: auto; border-right: 1px solid var(--td-component-stroke); padding-right: 12px; }
.pb-viewer { flex: 1; overflow-y: auto; }
.pb-backlinks { margin: 0 0 12px; }
.pb-empty { color: var(--td-text-color-placeholder); font-size: 14px; padding: 16px 4px; }
.pb-empty--big { text-align: center; padding-top: 80px; }
.pb-row { display: flex; align-items: center; justify-content: space-between; padding: 8px 10px; border: 1px solid var(--td-component-stroke); border-radius: 8px; margin-bottom: 6px; cursor: pointer;
  &:hover { background: var(--td-bg-color-container-hover); }
  &.active { background: var(--td-bg-color-secondarycontainer); border-color: var(--td-brand-color); } }
.pb-title { font-weight: 600; color: var(--td-text-color-primary); }
.pb-meta { display: flex; gap: 6px; margin-top: 4px; }
.pb-tag { background: var(--td-brand-color-light); color: var(--td-brand-color); padding: 0 6px; border-radius: 4px; font-size: 12px; }
.pb-tag--diff { background: var(--td-warning-color-light, #fff3e0); color: var(--td-warning-color, #e37318); }
.pb-actions { display: flex; }
.pb-solution { margin-top: 10px; }
.pb-solution-text { color: var(--td-text-color-primary); font-size: 14px; }
.pb-form { display: flex; flex-direction: column; gap: 6px; label { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 6px; } }
</style>
