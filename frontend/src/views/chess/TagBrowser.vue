<template>
  <div class="tgb" :class="{ 'is-detail': !!activeSlug }">
    <!-- Khung trái: từ điển thẻ -->
    <div class="tgb-left">
      <div class="tgb-toolbar">
        <t-input v-model="q" placeholder="Tìm thẻ…" clearable style="width: 200px" />
        <t-select v-model="kind" :options="kindOptions" clearable placeholder="Loại thẻ" style="width: 150px" />
        <t-checkbox v-model="onlyUsed">Chỉ thẻ đang dùng</t-checkbox>
        <div class="tgb-spacer"></div>
        <t-button size="small" variant="outline" @click="openCreate">
          <template #icon><t-icon name="add" /></template>Thẻ mới
        </t-button>
        <t-button size="small" variant="outline" :loading="backfilling" @click="doBackfill">
          <template #icon><t-icon name="download" /></template>Nạp thẻ từ dữ liệu cũ
        </t-button>
        <t-button size="small" variant="text" :loading="loading" @click="load">
          <template #icon><t-icon name="refresh" /></template>Làm mới
        </t-button>
      </div>

      <div v-if="loading && !tags.length" class="tgb-empty">Đang tải…</div>
      <div v-else-if="!filtered.length" class="tgb-empty">
        Chưa có thẻ nào khớp. Bấm <b>Nạp thẻ từ dữ liệu cũ</b> để chuyển các thẻ đã gõ trước đây
        (ở thế cờ, sách, bài viết) thành thẻ thật.
      </div>
      <div v-else class="tgb-cloud">
        <button
          v-for="t in filtered"
          :key="t.slug"
          class="tgb-chip"
          :class="{ 'is-group': t.kind === 'group', 'is-active': t.slug === activeSlug }"
          :title="t.description || t.name"
          @click="selectTag(t)"
        >
          <span class="tgb-chip-name">{{ t.name }}</span>
          <span class="tgb-chip-count">{{ t.usage_count }}</span>
        </button>
      </div>
    </div>

    <!-- Khung phải: nội dung mang thẻ, gộp mọi loại -->
    <div class="tgb-right">
      <div v-if="!activeSlug" class="tgb-empty tgb-placeholder">
        Chọn một thẻ để xem mọi nội dung mang thẻ đó — ván cờ, bài tập, thế cờ, bài giảng,
        khóa học, sách, chương và bài viết đều hiện chung một danh sách.
      </div>
      <template v-else>
        <div class="tgb-head">
          <t-button size="small" variant="text" class="tgb-back" @click="activeSlug = ''">‹ Danh sách</t-button>
          <h3 class="tgb-head-title">{{ activeTag?.name || activeSlug }}</h3>
          <t-tag v-if="activeTag?.kind === 'group'" theme="primary" variant="light" size="small">Nhóm nội dung</t-tag>
          <div class="tgb-spacer"></div>
          <t-button v-if="activeTag && activeTag.kind !== 'group'" size="small" variant="outline" @click="openRename">
            Đổi tên
          </t-button>
          <t-button v-if="activeTag && activeTag.kind !== 'group'" size="small" variant="outline" @click="openMerge">
            Gộp vào thẻ khác
          </t-button>
          <t-popconfirm
            v-if="activeTag && activeTag.kind !== 'group'"
            content="Xóa thẻ này? Nội dung KHÔNG bị xóa, chỉ gỡ thẻ khỏi chúng."
            @confirm="doDelete"
          >
            <t-button size="small" variant="outline" theme="danger">Xóa thẻ</t-button>
          </t-popconfirm>
        </div>
        <p v-if="activeTag?.description" class="tgb-desc">{{ activeTag.description }}</p>

        <div class="tgb-types">
          <button class="tgb-type" :class="{ 'is-active': typeFilter === '' }" @click="setType('')">
            Tất cả <span>{{ page?.total ?? 0 }}</span>
          </button>
          <button
            v-for="[ct, n] in typeCounts"
            :key="ct"
            class="tgb-type"
            :class="{ 'is-active': typeFilter === ct }"
            @click="setType(ct)"
          >
            {{ chessTypeLabel(ct) }} <span>{{ n }}</span>
          </button>
        </div>

        <div v-if="itemsLoading" class="tgb-empty">Đang tải…</div>
        <div v-else-if="!page?.items?.length" class="tgb-empty">Chưa có nội dung nào mang thẻ này.</div>
        <div v-else class="tgb-items">
          <a
            v-for="it in page.items"
            :key="`${it.chess_type}:${it.chess_id}`"
            class="tgb-item"
            :href="chessRefLink(it.chess_type, it.slug)"
            @click.prevent="openItem(it)"
          >
            <span class="tgb-item-type">{{ chessTypeLabel(it.chess_type) }}</span>
            <span class="tgb-item-body">
              <span class="tgb-item-title">{{ it.title }}</span>
              <span v-if="it.subtitle" class="tgb-item-sub">{{ it.subtitle }}</span>
            </span>
            <span v-if="it.level" class="tgb-item-level">{{ chessLevelLabel(it.level) }}</span>
            <span v-if="it.status === 'draft'" class="tgb-item-draft" title="Bản thảo — chưa vào kho tri thức">
              Bản thảo
            </span>
          </a>
        </div>

        <t-pagination
          v-if="page && page.total > page.page_size"
          class="tgb-pager"
          :total="page.total"
          :page-size="page.page_size"
          :current="page.page"
          :show-jumper="false"
          :show-page-size="false"
          @current-change="onPageChange"
        />
      </template>
    </div>

    <!-- Tạo / đổi tên thẻ -->
    <t-dialog
      v-model:visible="editDialog.visible"
      :header="editDialog.id ? 'Đổi tên thẻ' : 'Thẻ mới'"
      :confirm-btn="{ content: 'Lưu' }"
      @confirm="submitEdit"
    >
      <t-form label-align="top">
        <t-form-item label="Tên thẻ">
          <t-input v-model="editDialog.name" placeholder="vd: Ghim" />
        </t-form-item>
        <t-form-item label="Mô tả (tùy chọn)">
          <t-input v-model="editDialog.description" />
        </t-form-item>
      </t-form>
      <p class="tgb-note">
        Tên có dấu hay không dấu đều quy về cùng một thẻ — “Khai cuộc”, “khai-cuoc” và “KHAI CUOC”
        là một. Nếu tên mới trùng một thẻ đã có, hai thẻ sẽ được gộp lại.
      </p>
    </t-dialog>

    <!-- Gộp thẻ -->
    <t-dialog
      v-model:visible="mergeDialog.visible"
      header="Gộp thẻ"
      :confirm-btn="{ content: 'Gộp' }"
      @confirm="submitMerge"
    >
      <p class="tgb-note">
        Mọi nội dung đang mang <b>{{ activeTag?.name }}</b> sẽ chuyển sang thẻ đích, rồi thẻ này bị xóa.
      </p>
      <t-select v-model="mergeDialog.targetId" filterable placeholder="Chọn thẻ đích">
        <t-option
          v-for="t in mergeCandidates"
          :key="t.id"
          :value="t.id"
          :label="`${t.name} (${t.usage_count})`"
        />
      </t-select>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * TagBrowser — trang "Thẻ" trong khu Quản lý cờ vua.
 *
 * Đây là MỤC LỤC NGANG của toàn bộ kho nội dung: bấm một thẻ ra mọi loại nội
 * dung mang thẻ đó trong CÙNG một danh sách, thay vì phải mở lần lượt 6 tab và
 * tìm riêng từng nơi. Trang này cũng là chỗ dọn từ điển thẻ (đổi tên, gộp thẻ
 * trùng nghĩa, xóa thẻ rác).
 *
 * Quản lý gộp luôn vào đây thay vì tách một dialog riêng: thao tác dọn thẻ chỉ
 * có nghĩa khi đang nhìn thấy thẻ đó đang gắn cho những gì.
 */
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  listChessTags,
  listChessTagItems,
  createChessTag,
  updateChessTag,
  deleteChessTag,
  mergeChessTags,
  backfillChessTags,
  type ChessTag,
  type ChessTagItemPage,
  type ChessTagItemRef,
} from '@/api/chess';
import { chessLevelLabel, chessTypeLabel, chessRefLink } from '@/utils/chessTaxonomy';

const router = useRouter();

const tags = ref<ChessTag[]>([]);
const loading = ref(false);
const backfilling = ref(false);
const q = ref('');
const kind = ref('');
const onlyUsed = ref(false);

const kindOptions = [
  { label: 'Nhóm nội dung', value: 'group' },
  { label: 'Thẻ tự do', value: 'free' },
];

const activeSlug = ref('');
const typeFilter = ref('');
const page = ref<ChessTagItemPage | null>(null);
const itemsLoading = ref(false);

const activeTag = computed(() => tags.value.find((t) => t.slug === activeSlug.value) || null);

// Lọc phía client: từ điển thẻ của một tenant nhỏ (hàng chục tới hàng trăm),
// tải một lần rồi lọc tại chỗ cho phản hồi tức thì. Danh sách NỘI DUNG theo thẻ
// thì ngược lại — luôn phân trang phía máy chủ vì đó mới là phần phình to.
const filtered = computed(() => {
  const kw = q.value.trim().toLowerCase();
  return tags.value.filter((t) => {
    if (kind.value && t.kind !== kind.value) return false;
    if (onlyUsed.value && !t.usage_count) return false;
    if (kw && !t.name.toLowerCase().includes(kw) && !t.slug.includes(kw)) return false;
    return true;
  });
});

const typeCounts = computed(() =>
  Object.entries(page.value?.by_type || {}).filter(([, n]) => n > 0),
);

const mergeCandidates = computed(() => tags.value.filter((t) => t.slug !== activeSlug.value));

async function load() {
  loading.value = true;
  try {
    const res: any = await listChessTags();
    tags.value = res?.data || [];
  } finally {
    loading.value = false;
  }
}

async function loadItems(p = 1) {
  if (!activeSlug.value) return;
  itemsLoading.value = true;
  try {
    const res: any = await listChessTagItems(activeSlug.value, {
      type: typeFilter.value,
      page: p,
      page_size: 20,
    });
    page.value = res?.data || null;
  } finally {
    itemsLoading.value = false;
  }
}

function selectTag(t: ChessTag) {
  activeSlug.value = t.slug;
  typeFilter.value = '';
  loadItems(1);
}

function setType(ct: string) {
  typeFilter.value = ct;
  loadItems(1);
}

function onPageChange(p: number) {
  loadItems(p);
}

function openItem(it: ChessTagItemRef) {
  // Dùng lại deep-link ?ref= sẵn có của ChessManage: nó tự chuyển sang đúng tab
  // và bảo trang con chọn mục — không phải dựng route riêng cho từng loại.
  router.push({ name: 'chessCourses', query: { ref: `${it.chess_type}/${it.slug}` } });
}

const editDialog = ref({ visible: false, id: '', name: '', description: '' });

function openCreate() {
  editDialog.value = { visible: true, id: '', name: '', description: '' };
}

function openRename() {
  if (!activeTag.value) return;
  editDialog.value = {
    visible: true,
    id: activeTag.value.id,
    name: activeTag.value.name,
    description: activeTag.value.description,
  };
}

async function submitEdit() {
  const d = editDialog.value;
  if (!d.name.trim()) {
    MessagePlugin.warning('Tên thẻ không được để trống');
    return;
  }
  try {
    if (d.id) {
      // Backend có thể GỘP thẻ này vào thẻ cùng slug đã tồn tại và trả về thẻ
      // ĐÍCH — nên chọn lại theo phản hồi, đừng giả định slug/id không đổi.
      const res: any = await updateChessTag(d.id, { name: d.name, description: d.description });
      await load();
      if (res?.data?.slug) {
        activeSlug.value = res.data.slug;
        await loadItems(1);
      }
    } else {
      await createChessTag({ name: d.name, description: d.description });
      await load();
    }
    editDialog.value.visible = false;
  } catch (e: any) {
    MessagePlugin.error(e?.message || 'Lưu thẻ thất bại');
  }
}

const mergeDialog = ref({ visible: false, targetId: '' });

function openMerge() {
  mergeDialog.value = { visible: true, targetId: '' };
}

async function submitMerge() {
  if (!activeTag.value || !mergeDialog.value.targetId) {
    MessagePlugin.warning('Chọn thẻ đích trước');
    return;
  }
  try {
    const res: any = await mergeChessTags(activeTag.value.id, mergeDialog.value.targetId);
    mergeDialog.value.visible = false;
    await load();
    activeSlug.value = res?.data?.slug || '';
    if (activeSlug.value) await loadItems(1);
    MessagePlugin.success('Đã gộp thẻ');
  } catch (e: any) {
    MessagePlugin.error(e?.message || 'Gộp thẻ thất bại');
  }
}

async function doDelete() {
  if (!activeTag.value) return;
  try {
    await deleteChessTag(activeTag.value.id);
    activeSlug.value = '';
    page.value = null;
    await load();
    MessagePlugin.success('Đã xóa thẻ');
  } catch (e: any) {
    MessagePlugin.error(e?.message || 'Xóa thẻ thất bại');
  }
}

async function doBackfill() {
  backfilling.value = true;
  try {
    const res: any = await backfillChessTags();
    const d = res?.data || {};
    await load();
    const parts = [`${d.tags_created ?? 0} thẻ mới`, `${d.links_created ?? 0} liên kết`];
    MessagePlugin.success(`Đã nạp: ${parts.join(', ')}`);
    for (const w of d.warnings || []) MessagePlugin.warning(w);
  } catch (e: any) {
    MessagePlugin.error(e?.message || 'Nạp thẻ thất bại');
  } finally {
    backfilling.value = false;
  }
}

watch([kind, onlyUsed], () => {
  // Thẻ đang chọn có thể vừa bị lọc khỏi danh sách — bỏ chọn để khung phải
  // không hiện nội dung của một thẻ không còn thấy ở khung trái.
  if (activeSlug.value && !filtered.value.some((t) => t.slug === activeSlug.value)) {
    activeSlug.value = '';
    page.value = null;
  }
});

onMounted(load);
</script>

<style scoped>
.tgb {
  display: flex;
  gap: 16px;
  height: 100%;
  min-height: 0;
}
.tgb-left {
  flex: 0 0 380px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border-right: 1px solid var(--td-border-level-1-color);
  padding-right: 16px;
}
.tgb-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
}
.tgb-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.tgb-spacer {
  flex: 1;
}
.tgb-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-content: flex-start;
  overflow-y: auto;
  padding-bottom: 8px;
}
.tgb-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--td-border-level-2-color);
  background: var(--td-bg-color-container);
  color: var(--td-text-color-primary);
  border-radius: 14px;
  padding: 4px 10px;
  font-size: 13px;
  cursor: pointer;
  min-height: 28px;
}
.tgb-chip:hover {
  border-color: var(--td-brand-color);
}
.tgb-chip.is-group {
  border-style: dashed;
  font-weight: 600;
}
.tgb-chip.is-active {
  background: var(--td-brand-color-light);
  border-color: var(--td-brand-color);
  color: var(--td-brand-color);
}
.tgb-chip-count {
  font-size: 11px;
  color: var(--td-text-color-placeholder);
}
.tgb-chip.is-active .tgb-chip-count {
  color: var(--td-brand-color);
}
.tgb-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 4px;
}
.tgb-head-title {
  margin: 0;
  font-size: 16px;
}
.tgb-back {
  display: none;
}
.tgb-desc {
  margin: 0 0 12px;
  color: var(--td-text-color-secondary);
  font-size: 13px;
}
.tgb-types {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}
.tgb-type {
  border: none;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
  border-radius: 12px;
  padding: 3px 10px;
  font-size: 12px;
  cursor: pointer;
}
.tgb-type.is-active {
  background: var(--td-brand-color);
  color: #fff;
}
.tgb-type span {
  opacity: 0.7;
  margin-left: 2px;
}
.tgb-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.tgb-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 6px;
  text-decoration: none;
  color: inherit;
  cursor: pointer;
}
.tgb-item:hover {
  background: var(--td-bg-color-container-hover);
}
.tgb-item-type {
  flex: 0 0 72px;
  font-size: 11px;
  color: var(--td-text-color-placeholder);
}
.tgb-item-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.tgb-item-title {
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tgb-item-sub {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tgb-item-level,
.tgb-item-draft {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
  white-space: nowrap;
}
.tgb-item-draft {
  background: var(--td-warning-color-light);
  color: var(--td-warning-color);
}
.tgb-empty {
  color: var(--td-text-color-placeholder);
  font-size: 13px;
  padding: 16px 0;
}
.tgb-placeholder {
  max-width: 480px;
}
.tgb-note {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  margin: 8px 0 0;
}
.tgb-pager {
  margin-top: 12px;
}

/* Điện thoại: một khung tại một thời điểm (khuôn master–detail đã dùng ở 6
   trang cờ khác). Tiền tố "screen and" là BẮT BUỘC — thiếu chữ screen thì luật
   rò sang @media print của BookPrint/ArticlePrint. */
@media screen and (max-width: 767px) {
  .tgb {
    flex-direction: column;
  }
  .tgb-left {
    flex: 1 1 auto;
    border-right: none;
    padding-right: 0;
  }
  .tgb.is-detail .tgb-left {
    display: none;
  }
  .tgb:not(.is-detail) .tgb-right {
    display: none;
  }
  .tgb-back {
    display: inline-flex;
  }
}
</style>
