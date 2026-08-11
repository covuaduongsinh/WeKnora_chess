<template>
  <t-dialog v-model:visible="localVisible" header="Quản lý chuyên mục bài viết" :footer="false" width="640px">
    <div class="catm">
      <div class="catm-toolbar">
        <t-button theme="primary" size="small" @click="openTopicDialog()">
          <template #icon><t-icon name="add" /></template>Tạo chuyên mục
        </t-button>
      </div>
      <div v-if="tree.length === 0" class="catm-empty">Chưa có chuyên mục nào.</div>
      <template v-for="node in tree" :key="node.id">
        <div class="catm-row">
          <div class="catm-row-main">
            <div class="catm-row-title">{{ node.title }}</div>
            <div class="catm-row-meta">
              <span class="catm-count">{{ node.article_count || 0 }} bài</span>
            </div>
          </div>
          <div class="catm-row-actions">
            <t-button size="small" variant="text" title="Gán bài viết" @click="openAssignDialog(node)"><t-icon name="link" /></t-button>
            <t-button size="small" variant="text" @click="openTopicDialog(node)"><t-icon name="edit" /></t-button>
            <t-button size="small" variant="text" theme="danger" @click="removeTopic(node)"><t-icon name="delete" /></t-button>
          </div>
        </div>
        <div v-for="child in node.children" :key="child.id" class="catm-row catm-row--child">
          <div class="catm-row-main">
            <div class="catm-row-title">↳ {{ child.title }}</div>
            <div class="catm-row-meta">
              <span class="catm-count">{{ child.article_count || 0 }} bài</span>
            </div>
          </div>
          <div class="catm-row-actions">
            <t-button size="small" variant="text" title="Gán bài viết" @click="openAssignDialog(child)"><t-icon name="link" /></t-button>
            <t-button size="small" variant="text" @click="openTopicDialog(child)"><t-icon name="edit" /></t-button>
            <t-button size="small" variant="text" theme="danger" @click="removeTopic(child)"><t-icon name="delete" /></t-button>
          </div>
        </div>
      </template>
    </div>

    <t-dialog v-model:visible="topicDialog.visible" :header="topicDialog.id ? 'Sửa chuyên mục' : 'Tạo chuyên mục'"
      :on-confirm="saveTopic" width="480px">
      <div class="catm-form">
        <label>Tên chuyên mục *</label>
        <t-input v-model="topicDialog.title" placeholder="VD: Chiến thuật, Khai cuộc" />
        <label>Chuyên mục cha (để trống = chuyên mục gốc)</label>
        <t-select v-model="topicDialog.parent_id" :options="rootTopicOptions" clearable
          :disabled="topicDialog.hasChildren"
          :placeholder="topicDialog.hasChildren ? 'Chuyên mục đang có con — luôn là gốc' : 'Không (chuyên mục gốc)'" />
        <label>Mô tả</label>
        <t-textarea v-model="topicDialog.description" :autosize="{ minRows: 2 }" />
        <label>Thứ tự</label>
        <t-input-number v-model="topicDialog.sort_order" />
      </div>
    </t-dialog>

    <t-dialog v-model:visible="assignDialog.visible" :header="`Gán bài viết vào chuyên mục「${assignDialog.title}」`"
      :on-confirm="saveAssign" width="520px">
      <div class="catm-assign">
        <t-input v-model="assignDialog.search" placeholder="Tìm bài viết…" clearable style="margin-bottom:8px" />
        <div class="catm-assign-list">
          <t-checkbox v-for="a in filteredArticles" :key="a.id" v-model="assignDialog.selected[a.id]">
            {{ a.title }}
          </t-checkbox>
        </div>
      </div>
    </t-dialog>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import {
  listArticleTopics, createArticleTopic, updateArticleTopic, deleteArticleTopic, setTopicArticles,
  listArticles, type ChessArticleTopic, type ChessArticle,
} from '@/api/chess';

// Quản lý chuyên mục (CRUD, cây tối đa 2 tầng) + gán bài viết vào chuyên mục
// (nhiều-nhiều). Gán thực hiện TỪ PHÍA CHUYÊN MỤC (chọn nhiều bài cho 1
// chuyên mục, gọi setTopicArticles 1 lần) — cùng khuôn ChessShelfManager.vue.
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void; (e: 'changed'): void }>();

const localVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
});

const topics = ref<ChessArticleTopic[]>([]);

interface TopicNode extends ChessArticleTopic {
  children: ChessArticleTopic[];
}
const tree = computed<TopicNode[]>(() => {
  const roots = topics.value.filter((t) => !t.parent_id);
  return roots.map((r) => ({
    ...r,
    children: topics.value.filter((t) => t.parent_id === r.id),
  }));
});
const rootTopicOptions = computed(() =>
  topics.value.filter((t) => !t.parent_id && t.id !== topicDialog.id)
    .map((t) => ({ label: t.title, value: t.id })),
);

async function load() {
  try {
    const res: any = await listArticleTopics();
    topics.value = res?.data || [];
  } catch {
    MessagePlugin.error('Tải danh sách chuyên mục thất bại');
  }
}
watch(() => props.visible, (v) => { if (v) load(); });

// ---- Dialog tạo/sửa chuyên mục ----
interface TopicDialogState {
  visible: boolean; id: string; title: string; parent_id: string; description: string;
  sort_order: number; hasChildren: boolean;
}
const topicDialog = reactive<TopicDialogState>({
  visible: false, id: '', title: '', parent_id: '', description: '', sort_order: 0, hasChildren: false,
});
function openTopicDialog(node?: ChessArticleTopic) {
  topicDialog.visible = true;
  topicDialog.id = node?.id || '';
  topicDialog.title = node?.title || '';
  topicDialog.parent_id = node?.parent_id || '';
  topicDialog.description = node?.description || '';
  topicDialog.sort_order = node?.sort_order || 0;
  // Chuyên mục ĐANG có con thì phải giữ ở tầng gốc (ràng buộc 2 tầng ở backend).
  topicDialog.hasChildren = node ? topics.value.some((t) => t.parent_id === node.id) : false;
}
async function saveTopic() {
  if (!topicDialog.title.trim()) { MessagePlugin.warning('Nhập tên chuyên mục'); return; }
  const payload = {
    title: topicDialog.title, parent_id: topicDialog.hasChildren ? '' : topicDialog.parent_id,
    description: topicDialog.description, sort_order: topicDialog.sort_order,
  };
  try {
    if (topicDialog.id) await updateArticleTopic(topicDialog.id, payload);
    else await createArticleTopic(payload);
    topicDialog.visible = false;
    await load();
    emit('changed');
    MessagePlugin.success('Đã lưu chuyên mục');
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Lưu thất bại');
  }
}
function removeTopic(node: ChessArticleTopic) {
  DialogPlugin.confirm({
    header: 'Xóa chuyên mục', body: `Xóa "${node.title}"? Bài viết trong chuyên mục KHÔNG bị xóa, chỉ gỡ khỏi chuyên mục này.`,
    theme: 'warning', confirmBtn: { content: 'Xóa', theme: 'danger' },
    onConfirm: async () => {
      try {
        await deleteArticleTopic(node.id);
        await load();
        emit('changed');
        MessagePlugin.success('Đã xóa chuyên mục');
      } catch (e: any) {
        MessagePlugin.error(e?.error || e?.message || 'Xóa thất bại (chuyên mục còn con?)');
      }
    },
  });
}

// ---- Dialog gán bài viết ----
interface AssignDialogState {
  visible: boolean; topicId: string; title: string; search: string; selected: Record<string, boolean>;
}
const assignDialog = reactive<AssignDialogState>({ visible: false, topicId: '', title: '', search: '', selected: {} });
const allArticles = ref<ChessArticle[]>([]);

async function openAssignDialog(node: ChessArticleTopic) {
  assignDialog.visible = true;
  assignDialog.topicId = node.id;
  assignDialog.title = node.title;
  assignDialog.search = '';
  assignDialog.selected = {};
  try {
    const [all, onTopic]: any[] = await Promise.all([listArticles(), listArticles({ topic_id: node.id })]);
    allArticles.value = all?.data || [];
    const onTopicIds = new Set((onTopic?.data || []).map((a: ChessArticle) => a.id));
    const sel: Record<string, boolean> = {};
    for (const a of allArticles.value) sel[a.id] = onTopicIds.has(a.id);
    assignDialog.selected = sel;
  } catch {
    MessagePlugin.error('Tải danh sách bài viết thất bại');
  }
}
const filteredArticles = computed(() => {
  const q = assignDialog.search.trim().toLowerCase();
  if (!q) return allArticles.value;
  return allArticles.value.filter((a) => (a.title || '').toLowerCase().includes(q));
});
async function saveAssign() {
  const articleIds = Object.entries(assignDialog.selected).filter(([, v]) => v).map(([id]) => id);
  try {
    await setTopicArticles(assignDialog.topicId, articleIds);
    assignDialog.visible = false;
    await load();
    emit('changed');
    MessagePlugin.success(`Đã gán ${articleIds.length} bài viết vào chuyên mục`);
  } catch (e: any) {
    MessagePlugin.error(e?.error || e?.message || 'Gán bài viết thất bại');
  }
}
</script>

<style scoped lang="less">
.catm-toolbar { display: flex; justify-content: flex-end; margin-bottom: 10px; }
.catm-empty { color: var(--td-text-color-placeholder); font-size: 14px; padding: 16px 4px; }
.catm-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px; border-radius: 8px; margin-bottom: 6px; border: 1px solid var(--td-component-stroke);
  &:hover { background: var(--td-bg-color-container-hover); }
}
.catm-row--child { margin-left: 20px; background: var(--td-bg-color-container-select); }
.catm-row-title { font-weight: 600; color: var(--td-text-color-primary); }
.catm-row-meta { display: flex; gap: 8px; align-items: center; margin-top: 4px; font-size: 12px; }
.catm-count { color: var(--td-text-color-secondary); }
.catm-row-actions { display: flex; }
.catm-form { display: flex; flex-direction: column; gap: 6px; label { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 6px; } }
.catm-assign-list { max-height: 320px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
</style>
