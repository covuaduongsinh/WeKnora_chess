<template>
  <div class="atb">
    <div class="atb-toolbar">
      <t-select v-model="filter.category" :options="articleCategoryOptions" placeholder="Thể loại" clearable
        style="width:170px" @change="load" />
      <t-select v-model="filter.level" :options="articleLevelOptions" placeholder="Cấp độ" clearable
        style="width:120px" @change="load" />
      <t-select v-model="filter.status" :options="articleStatusOptions" placeholder="Trạng thái" clearable
        style="width:150px" @change="load" />
      <t-input v-model="filter.q" placeholder="Tìm theo tên/bí danh/tag…" clearable style="width:200px" @change="load" />
      <div style="flex:1"></div>
      <t-button variant="outline" size="small" @click="doExport">
        <template #icon><t-icon name="download" /></template>Export
      </t-button>
      <t-button variant="outline" size="small" @click="doImport">
        <template #icon><t-icon name="upload" /></template>Import
      </t-button>
      <t-button variant="outline" size="small" @click="topicManagerVisible = true">
        <template #icon><t-icon name="folder" /></template>Quản lý chuyên mục
      </t-button>
      <t-button theme="primary" size="small" @click="createDialog.visible = true">
        <template #icon><t-icon name="add" /></template>Bài viết mới
      </t-button>
    </div>

    <div class="atb-body">
      <div class="atb-topics">
        <div class="atb-topic-row" :class="{ active: !filter.topic_id }" @click="selectTopic('')">
          <span>Tất cả</span>
        </div>
        <template v-for="node in topicTree" :key="node.id">
          <div class="atb-topic-row" :class="{ active: filter.topic_id === node.id }" @click="selectTopic(node.id)">
            <span>{{ node.title }}</span>
            <span class="atb-topic-count">{{ node.article_count || 0 }}</span>
          </div>
          <div v-for="child in node.children" :key="child.id" class="atb-topic-row atb-topic-row--child"
            :class="{ active: filter.topic_id === child.id }" @click="selectTopic(child.id)">
            <span>{{ child.title }}</span>
            <span class="atb-topic-count">{{ child.article_count || 0 }}</span>
          </div>
        </template>
        <div v-if="topics.length === 0" class="atb-empty">Chưa có chuyên mục.</div>
      </div>

      <div class="atb-list">
        <div v-if="articles.length === 0" class="atb-empty">Chưa có bài viết. Nhấn "Bài viết mới".</div>
        <div v-for="a in articles" :key="a.id" class="atb-row" :class="{ active: selected && selected.id === a.id }"
          @click="select(a)">
          <div class="atb-row-main">
            <div class="atb-title">{{ a.title || '(không tiêu đề)' }}</div>
            <div class="atb-meta">
              <span v-if="a.category" class="atb-tag">{{ articleCategoryLabel(a.category) }}</span>
              <span v-if="a.level" class="atb-tag atb-tag--level">{{ articleLevelLabel(a.level) }}</span>
              <span class="atb-tag" :class="a.status === 'published' ? 'atb-tag--published' : 'atb-tag--draft'">
                {{ articleStatusLabel(a.status) }}
              </span>
            </div>
          </div>
          <span class="atb-actions">
            <t-button size="small" variant="text" :title="t('chess.ref.copyLink')" @click.stop="copyWikilink(a)">
              <t-icon name="link" />
            </t-button>
            <t-button size="small" variant="text" title="Đổi slug" @click.stop="renameSlug(a)"><t-icon name="tag" /></t-button>
            <t-button size="small" variant="text" theme="danger" @click.stop="remove(a)"><t-icon name="delete" /></t-button>
          </span>
        </div>
      </div>

      <div class="atb-viewer">
        <template v-if="selected">
          <div class="atb-header">
            <h3 class="atb-header-title">{{ selected.title || '(không tiêu đề)' }}</h3>
            <div class="atb-header-actions">
              <template v-if="!editing">
                <t-button size="small" variant="outline" @click="openPrintView">
                  <template #icon><t-icon name="print" /></template>Trang in
                </t-button>
                <t-button size="small" variant="outline" @click="openHistory">
                  <template #icon><t-icon name="history" /></template>Lịch sử
                </t-button>
                <t-button size="small" variant="outline" @click="startEdit">
                  <template #icon><t-icon name="edit-1" /></template>Sửa
                </t-button>
              </template>
              <template v-else>
                <t-button size="small" variant="outline" @click="cancelEdit">Hủy</t-button>
                <t-button size="small" theme="primary" @click="saveEdit">Lưu</t-button>
              </template>
            </div>
          </div>

          <template v-if="!editing">
            <ChessBacklinks v-if="selected.slug" ref-type="article" :slug="selected.slug" show-empty class="atb-backlinks" />
            <div v-if="selected.summary" class="atb-summary">{{ selected.summary }}</div>
            <div v-if="selected.aliases" class="atb-aliases">Tên gọi khác: {{ selected.aliases }}</div>
            <div class="atb-content" @click="onContentClick">
              <template v-for="(seg, si) in contentSegments(selected)" :key="si">
                <ChessBoardDisplay v-if="seg.type === 'board'" :data="seg.board" />
                <ChessRefEmbed v-else-if="seg.type === 'ref'" :ref-str="seg.refStr" />
                <div v-else class="atb-md-segment" v-html="seg.html"></div>
              </template>
              <div v-if="!selected.content" class="atb-empty">Bài viết chưa có nội dung. Nhấn "Sửa" để soạn.</div>
            </div>
          </template>

          <template v-else>
            <div class="atb-form">
              <label>Tiêu đề</label>
              <t-input v-model="editDraft.title" placeholder="VD: Ghim (Pin) là gì?" />
              <div class="atb-row2">
                <div>
                  <label>Thể loại</label>
                  <t-select v-model="editDraft.category" :options="articleCategoryOptions" clearable filterable />
                </div>
                <div>
                  <label>Cấp độ</label>
                  <t-select v-model="editDraft.level" :options="articleLevelOptions" clearable />
                </div>
              </div>
              <div class="atb-row2">
                <div>
                  <label>Trạng thái</label>
                  <t-select v-model="editDraft.status" :options="articleStatusOptions" />
                </div>
                <div>
                  <label>Thẻ (tùy chọn, cách nhau dấu phẩy)</label>
                  <t-input v-model="editDraft.tags" placeholder="chien-thuat, co-ban" />
                </div>
              </div>
              <label>Tóm tắt ngắn (hiện ở danh sách + popup wikilink)</label>
              <t-textarea v-model="editDraft.summary" :autosize="{ minRows: 2, maxRows: 4 }" />
              <label>Bí danh / từ đồng nghĩa (tùy chọn, cách nhau dấu phẩy)</label>
              <t-input v-model="editDraft.aliases" placeholder="Pin, Đóng đinh" />
              <label>Ảnh bìa (URL, tùy chọn)</label>
              <t-input v-model="editDraft.cover_url" placeholder="https://…" />
              <div class="atb-content-label">
                <label>Nội dung (markdown — gõ [[ để chèn liên kết ván/thế cờ/bài giảng/sách; dán/kéo-thả ảnh)</label>
                <div class="atb-editor-toolbar">
                  <t-button v-for="btn in editor.toolbarButtons" :key="btn.key" size="small" variant="text"
                    :title="btn.title" @click="btn.action">
                    <t-icon :name="btn.icon" />
                  </t-button>
                  <t-button size="small" variant="text" title="Chèn ảnh" @click="imageInputEl?.click()">
                    <t-icon name="image" />
                  </t-button>
                </div>
              </div>
              <t-textarea ref="contentRef" v-model="editDraft.content" :autosize="{ minRows: 10 }"
                placeholder="Viết nội dung bài… Gõ [[ để gợi ý ván/thế cờ/bài tập/sách/chương. Dán/kéo-thả ảnh vào đây." />
              <input ref="imageInputEl" type="file" accept="image/png,image/jpeg,image/gif,image/webp"
                style="display:none" @change="onImageInputChange" />
              <ChessWikiLinkSuggest :textarea="contentTextareaEl" v-model="editDraft.content" />
              <label>Ghi chú thay đổi (tùy chọn — chỉ lưu lịch sử khi tiêu đề/nội dung thực sự đổi)</label>
              <t-input v-model="editDraft.revisionNote" placeholder="VD: Bổ sung ví dụ minh họa" />
            </div>
          </template>
        </template>
        <div v-else class="atb-empty atb-empty--big">Chọn một bài viết ở bên trái.</div>
      </div>
    </div>

    <t-dialog v-model:visible="createDialog.visible" header="Bài viết mới" :on-confirm="confirmCreate" width="480px">
      <div class="atb-form">
        <label>Tiêu đề *</label>
        <t-input v-model="createDialog.title" placeholder="VD: Ghim (Pin) là gì?" />
        <label>Thể loại</label>
        <t-select v-model="createDialog.category" :options="articleCategoryOptions" clearable filterable />
      </div>
    </t-dialog>

    <ChessRefDialog v-model:visible="refDialog.visible" :ref-str="refDialog.refStr" />
    <ChessArticleTopicManager v-model:visible="topicManagerVisible" @changed="loadTopics" />
    <ChessArticleHistory v-model:visible="historyVisible" :article-id="historyArticleId" @restored="onRestored" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { marked } from 'marked';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import ChessBacklinks from '@/views/chess/components/ChessBacklinks.vue';
import ChessRefEmbed from '@/views/chess/components/ChessRefEmbed.vue';
import ChessRefDialog from '@/views/chess/components/ChessRefDialog.vue';
import ChessWikiLinkSuggest from '@/views/chess/components/ChessWikiLinkSuggest.vue';
import ChessArticleTopicManager from '@/views/chess/components/ChessArticleTopicManager.vue';
import ChessArticleHistory from '@/views/chess/components/ChessArticleHistory.vue';
import {
  listArticles, getArticleBySlug, createArticle, updateArticle, deleteArticle,
  renameArticleSlug, exportArticles, importArticles, type ChessArticle,
  listArticleTopics, uploadArticleImage, type ChessArticleTopic,
} from '@/api/chess';
import { downloadText, pickTextFile } from '@/utils/fileTransfer';
import { splitChessContent, renderChessChips } from '@/utils/chessBlocks';
import {
  articleCategoryOptions, articleLevelOptions, articleStatusOptions,
  articleCategoryLabel, articleLevelLabel, articleStatusLabel,
} from '@/utils/chessArticleOptions';
import { useChessEditor } from '@/views/chess/composables/useChessEditor';

const { t } = useI18n();
const router = useRouter();

// Deep-link "Mở trong thư viện": chọn sẵn bài viết theo slug (từ [[article/<slug>]]).
const props = defineProps<{ focusSlug?: string }>();
async function focusBySlug(slug?: string) {
  if (!slug) return;
  try {
    const res: any = await getArticleBySlug(slug);
    if (res?.data) { selected.value = res.data; editing.value = false; }
  } catch { /* không tìm thấy → bỏ qua */ }
}
onMounted(() => focusBySlug(props.focusSlug));
watch(() => props.focusSlug, (s) => focusBySlug(s));

const articles = ref<ChessArticle[]>([]);
const selected = ref<ChessArticle | null>(null);
const filter = reactive({ category: '', level: '', status: '', q: '', topic_id: '' });
const editing = ref(false);

async function load() {
  try {
    const res: any = await listArticles(filter);
    articles.value = res?.data || [];
  } catch { MessagePlugin.error('Tải ngân hàng bài viết thất bại'); }
}
function select(a: ChessArticle) { selected.value = a; editing.value = false; }

// ---- Chuyên mục (trục DỌC điều hướng — trục NGANG là bộ lọc category/level/
// status/q ở trên). Cây tối đa 2 tầng: gốc + con. Chọn "Tất cả" (topic_id='')
// quay về đúng danh sách phẳng đã lọc. ----
const topics = ref<ChessArticleTopic[]>([]);
const topicManagerVisible = ref(false);
interface TopicNode extends ChessArticleTopic { children: ChessArticleTopic[] }
const topicTree = computed<TopicNode[]>(() => {
  const roots = topics.value.filter((t) => !t.parent_id);
  return roots.map((r) => ({ ...r, children: topics.value.filter((t) => t.parent_id === r.id) }));
});
async function loadTopics() {
  try {
    const res: any = await listArticleTopics();
    topics.value = res?.data || [];
  } catch { MessagePlugin.error('Tải danh sách chuyên mục thất bại'); }
}
function selectTopic(topicId: string) {
  filter.topic_id = topicId;
  load();
}

// Sao chép wikilink [[article/<slug>]] để dán vào nội dung bài khác/wiki.
async function copyWikilink(a: ChessArticle) {
  if (!a.slug) { MessagePlugin.warning('Bài viết chưa có slug'); return; }
  const link = `[[article/${a.slug}|${a.title || a.slug}]]`;
  try {
    await navigator.clipboard.writeText(link);
    MessagePlugin.success(t('chess.ref.copied'));
  } catch {
    MessagePlugin.info(link);
  }
}

// ---- Nội dung: markdown + wikilink chip/nhúng (giống chú giải thế cờ) ----
function renderContentMarkdown(md: string): string {
  let pre = renderChessChips(md || '');
  pre = pre.replace(/\[\[([^\]]+)\]\]/g, (_, inner: string) => {
    const pipe = inner.indexOf('|');
    return (pipe > 0 ? inner.slice(pipe + 1) : inner).trim();
  });
  return marked.parse(pre, { breaks: true, async: false }) as string;
}
function contentSegments(a: ChessArticle) {
  return splitChessContent(a.content || '').map((seg) => {
    if (seg.type === 'board') return { type: 'board' as const, board: seg.board! };
    if (seg.type === 'ref') return { type: 'ref' as const, refStr: seg.ref!.ref };
    return { type: 'markdown' as const, html: renderContentMarkdown(seg.markdown || '') };
  });
}
const refDialog = reactive<{ visible: boolean; refStr: string }>({ visible: false, refStr: '' });
function onContentClick(e: MouseEvent) {
  const el = (e.target as HTMLElement).closest('a.chess-ref-link') as HTMLElement | null;
  if (!el) return;
  e.preventDefault();
  const ref = el.getAttribute('data-chess-ref');
  if (ref) { refDialog.refStr = ref; refDialog.visible = true; }
}

// ---- Tạo bài viết mới: chỉ hỏi Tiêu đề + Thể loại → lưu ngay draft → mở khung
// soạn — bài đã có id trước khi gõ chữ đầu tiên (dán ảnh sau này không vướng
// ràng buộc "phải lưu trước"). ----
const createDialog = reactive({ visible: false, title: '', category: '' });
async function confirmCreate() {
  if (!createDialog.title.trim()) { MessagePlugin.warning('Nhập tiêu đề bài viết'); return; }
  try {
    const res: any = await createArticle({
      title: createDialog.title, category: createDialog.category, status: 'draft',
    });
    createDialog.visible = false;
    createDialog.title = '';
    createDialog.category = '';
    await load();
    const created = res?.data;
    if (created) { selected.value = created; startEdit(); }
    MessagePlugin.success('Đã tạo bài viết');
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Tạo bài viết thất bại');
  }
}

// ---- Sửa (inline trong khung phải, không dùng modal — bài viết là văn bản dài) ----
interface ArticleEditDraft {
  title: string; summary: string; aliases: string; category: string; level: string;
  tags: string; status: string; cover_url: string; content: string; revisionNote: string;
}
const editDraft = reactive<ArticleEditDraft>({
  title: '', summary: '', aliases: '', category: '', level: '',
  tags: '', status: 'draft', cover_url: '', content: '', revisionNote: '',
});
function startEdit() {
  if (!selected.value) return;
  editDraft.title = selected.value.title || '';
  editDraft.summary = selected.value.summary || '';
  editDraft.aliases = selected.value.aliases || '';
  editDraft.category = selected.value.category || '';
  editDraft.level = selected.value.level || '';
  editDraft.tags = selected.value.tags || '';
  editDraft.status = selected.value.status || 'draft';
  editDraft.cover_url = selected.value.cover_url || '';
  editDraft.content = selected.value.content || '';
  editDraft.revisionNote = '';
  editing.value = true;
}
function cancelEdit() { editing.value = false; }
async function saveEdit() {
  if (!selected.value) return;
  if (!editDraft.title.trim()) { MessagePlugin.warning('Nhập tiêu đề bài viết'); return; }
  try {
    const id = selected.value.id;
    const { revisionNote, ...fields } = editDraft;
    const res: any = await updateArticle(id, { ...fields, revision_note: revisionNote });
    await load();
    const updated = articles.value.find((a) => a.id === id) || res?.data;
    if (updated) selected.value = updated;
    editing.value = false;
    MessagePlugin.success('Đã lưu bài viết');
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Lưu thất bại');
  }
}

// ---- Trang in (mở tab mới rồi Ctrl+P — khuôn openPrintView của BookLibrary.vue) ----
function openPrintView() {
  if (!selected.value) return;
  const routeData = router.resolve({ name: 'chessArticlePrint', params: { id: selected.value.id } });
  window.open(routeData.href, '_blank');
}

// ---- Lịch sử phiên bản ----
const historyVisible = ref(false);
const historyArticleId = ref('');
function openHistory() {
  if (!selected.value) return;
  historyArticleId.value = selected.value.id;
  historyVisible.value = true;
}
function onRestored(article: ChessArticle) {
  selected.value = article;
  load();
}

// Đổi slug bài viết (power-feature cho HLV): link cũ [[article/<cũ>]] vẫn sống nhờ alias.
async function renameSlug(a: ChessArticle) {
  if (!a.slug) { MessagePlugin.warning('Bài viết chưa có slug'); return; }
  const next = window.prompt(`Đổi slug cho "${a.title || a.slug}" (link cũ vẫn sống nhờ alias):`, a.slug);
  if (next == null) return;
  const v = next.trim();
  if (!v || v === a.slug) return;
  try {
    const res: any = await renameArticleSlug(a.id, v);
    await load();
    if (selected.value?.id === a.id && res?.data) selected.value = res.data;
    MessagePlugin.success(`Đã đổi slug → ${res?.data?.slug || v}`);
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Đổi slug thất bại');
  }
}

function remove(a: ChessArticle) {
  DialogPlugin.confirm({
    header: 'Xóa bài viết', body: `Xóa "${a.title || 'bài viết'}"?`,
    theme: 'warning', confirmBtn: { content: 'Xóa', theme: 'danger' },
    onConfirm: async () => {
      try {
        await deleteArticle(a.id);
        if (selected.value?.id === a.id) { selected.value = null; editing.value = false; }
        await load();
        MessagePlugin.success('Đã xóa');
      } catch { MessagePlugin.error('Xóa thất bại'); }
    },
  });
}

// ---- Export / Import ----
async function doExport() {
  try {
    const res: any = await exportArticles(filter);
    const items = res?.data || [];
    if (!items.length) { MessagePlugin.info('Không có bài viết để xuất'); return; }
    downloadText(`bai-viet-${new Date().toISOString().slice(0, 10)}.json`, JSON.stringify(items, null, 2), 'application/json');
    MessagePlugin.success(`Đã xuất ${items.length} bài viết`);
  } catch { MessagePlugin.error('Xuất thất bại'); }
}
async function doImport() {
  const text = await pickTextFile('.json,application/json');
  if (text == null) return;
  let arr: any;
  try { arr = JSON.parse(text); } catch { MessagePlugin.error('File JSON không hợp lệ'); return; }
  if (!Array.isArray(arr)) { MessagePlugin.error('File phải là một mảng JSON bài viết'); return; }
  try {
    const res: any = await importArticles(arr);
    await load();
    MessagePlugin.success(`Đã nhập ${res?.data?.imported || 0} bài viết`);
  } catch (e: any) { MessagePlugin.error(e?.error || e?.message || 'Import thất bại'); }
}

// ---- Autocomplete "[[" + soạn thảo (toolbar, dán/kéo-thả ảnh) trong ô nội dung ----
const contentRef = ref<any>(null);
const contentTextareaEl = ref<HTMLTextAreaElement | null>(null);
const imageInputEl = ref<HTMLInputElement | null>(null);

const contentModel = computed({
  get: () => editDraft.content,
  set: (v: string) => { editDraft.content = v; },
});
const editor = useChessEditor({
  getTextarea: () => contentTextareaEl.value,
  content: contentModel,
  upload: async (file: File) => {
    if (!selected.value) throw new Error('Chưa chọn bài viết');
    const res: any = await uploadArticleImage(selected.value.id, file);
    return res?.data?.url as string;
  },
  onUploadError: (e: any) => MessagePlugin.error(e?.error || e?.message || 'Tải ảnh lên thất bại'),
});
function onImageInputChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) editor.uploadAndInsert(file);
  input.value = '';
}

watch(editing, async (isEditing) => {
  if (!isEditing) { contentTextareaEl.value = null; editor.bindTextarea(null); return; }
  await nextTick();
  const root = contentRef.value?.$el || contentRef.value;
  contentTextareaEl.value = (root?.querySelector?.('textarea') as HTMLTextAreaElement) || null;
  editor.bindTextarea(contentTextareaEl.value);
});

load();
loadTopics();
</script>

<style lang="less" scoped>
.atb { display: flex; flex-direction: column; height: 100%; }
.atb-toolbar { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; flex-wrap: wrap; }
.atb-body { display: flex; gap: 16px; flex: 1; min-height: 0; }
.atb-topics { width: 220px; flex: 0 0 220px; overflow-y: auto; border-right: 1px solid var(--td-component-stroke); padding-right: 10px; }
.atb-topic-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 8px; border-radius: 6px; cursor: pointer; font-size: 13px; margin-bottom: 2px;
  &:hover { background: var(--td-bg-color-container-hover); }
  &.active { background: var(--td-bg-color-secondarycontainer); color: var(--td-brand-color); font-weight: 600; }
}
.atb-topic-row--child { padding-left: 20px; font-size: 12px; }
.atb-topic-count { color: var(--td-text-color-placeholder); font-size: 12px; }
.atb-list { width: 320px; flex: 0 0 320px; overflow-y: auto; border-right: 1px solid var(--td-component-stroke); padding-right: 12px; }
.atb-viewer { flex: 1; overflow-y: auto; }
.atb-backlinks { margin: 0 0 12px; }
.atb-empty { color: var(--td-text-color-placeholder); font-size: 14px; padding: 16px 4px; }
.atb-empty--big { text-align: center; padding-top: 80px; }
.atb-row { display: flex; align-items: center; justify-content: space-between; padding: 8px 10px; border: 1px solid var(--td-component-stroke); border-radius: 8px; margin-bottom: 6px; cursor: pointer;
  &:hover { background: var(--td-bg-color-container-hover); }
  &.active { background: var(--td-bg-color-secondarycontainer); border-color: var(--td-brand-color); } }
.atb-title { font-weight: 600; color: var(--td-text-color-primary); }
.atb-meta { display: flex; gap: 6px; margin-top: 4px; flex-wrap: wrap; }
.atb-tag { background: var(--td-brand-color-light); color: var(--td-brand-color); padding: 0 6px; border-radius: 4px; font-size: 12px; }
.atb-tag--level { background: #e6f4ff; color: #2275b4; }
.atb-tag--draft { background: var(--td-bg-color-secondarycontainer); color: var(--td-text-color-secondary); }
.atb-tag--published { background: #e8f8f0; color: #3dbb95; }
.atb-actions { display: flex; }
.atb-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 8px; }
.atb-header-title { margin: 0; font-size: 18px; }
.atb-header-actions { display: flex; gap: 8px; flex: none; }
.atb-summary { color: var(--td-text-color-secondary); font-style: italic; margin-bottom: 8px; }
.atb-aliases { font-size: 12px; color: var(--td-text-color-secondary); margin-bottom: 12px; }
.atb-content { line-height: 1.6; color: var(--td-text-color-primary); }
.atb-md-segment :deep(.chess-ref-link) {
  color: var(--td-brand-color);
  background: var(--td-brand-color-light);
  padding: 0 6px; border-radius: 4px; text-decoration: none; font-weight: 600;
  &::before { content: '♟ '; }
  &:hover { text-decoration: underline; }
}
.atb-form { display: flex; flex-direction: column; gap: 6px; label { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 6px; } }
.atb-content-label { margin-top: 6px; display: flex; align-items: center; justify-content: space-between; }
.atb-editor-toolbar { display: flex; gap: 2px; }
.atb-row2 { display: flex; gap: 12px; > div { flex: 1; min-width: 0; } }
</style>
