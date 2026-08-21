<template>
  <t-dialog
    :visible="visible"
    header="Gắn thẻ"
    :confirm-btn="{ content: 'Lưu thẻ', loading: saving }"
    @update:visible="(v: boolean) => emit('update:visible', v)"
    @confirm="save"
  >
    <p v-if="itemTitle" class="ctad-target">
      <span class="ctad-type">{{ chessTypeLabel(chessType) }}</span>
      <b>{{ itemTitle }}</b>
    </p>
    <ChessTagInput v-model="tags" hint="Gõ tên thẻ mới rồi Enter để tạo. Bỏ hết thẻ = gỡ thẻ khỏi mục này." />
  </t-dialog>
</template>

<script setup lang="ts">
/**
 * ChessTagAssignDialog — gắn thẻ cho MỘT mục thuộc BẤT KỲ loại nội dung nào.
 *
 * Tồn tại vì 5 trong 8 loại (ván cờ, bài tập, bài giảng, khóa học, chương)
 * không có cột `tags` trong thân request create/update của chúng — và ván cờ
 * thì thậm chí không có hộp thoại sửa nào (ván nhập từ PGN). Thay vì đổi chữ ký
 * API của từng loại, mọi nơi dùng chung một hộp thoại gọi PUT /chess/tags/assign.
 *
 * Ba loại CÓ cột `tags` (thế cờ/sách/bài viết) vẫn gắn thẻ ngay trong form của
 * chúng; hộp thoại này là đường bổ sung, không thay thế — cả hai cùng đi qua
 * một bộ chuẩn hóa ở backend nên không lệch nhau.
 */
import { ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { assignChessTags, getChessTagsOf, type ChessTag } from '@/api/chess';
import { chessTypeLabel } from '@/utils/chessTaxonomy';
import ChessTagInput from './ChessTagInput.vue';

const props = defineProps<{
  visible: boolean;
  chessType: string;
  chessId: string;
  itemTitle?: string;
}>();
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'saved', tags: ChessTag[]): void;
}>();

const tags = ref('');
const saving = ref(false);

watch(
  () => [props.visible, props.chessId] as const,
  async ([vis, id]) => {
    if (!vis || !id) return;
    tags.value = '';
    try {
      const res: any = await getChessTagsOf(props.chessType, id);
      tags.value = (res?.data || []).map((t: ChessTag) => t.name).join(', ');
    } catch {
      tags.value = '';
    }
  },
  { immediate: true },
);

async function save() {
  saving.value = true;
  try {
    const names = tags.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    const res: any = await assignChessTags(props.chessType, props.chessId, names);
    emit('saved', res?.data || []);
    emit('update:visible', false);
    MessagePlugin.success('Đã lưu thẻ');
  } catch (e: any) {
    MessagePlugin.error(e?.message || 'Lưu thẻ thất bại');
  } finally {
    saving.value = false;
  }
}
</script>

<style scoped>
.ctad-target {
  margin: 0 0 12px;
  font-size: 13px;
}
.ctad-type {
  color: var(--td-text-color-placeholder);
  margin-right: 6px;
}
</style>
