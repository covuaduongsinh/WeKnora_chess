<template>
  <span v-if="items.length" class="ctc">
    <button
      v-for="t in items"
      :key="t.key"
      class="ctc-chip"
      :class="{ 'is-group': t.group }"
      :title="`Lọc theo thẻ ${t.name}`"
      @click.stop="emit('pick', t.key)"
    >
      {{ t.name }}
    </button>
  </span>
</template>

<script setup lang="ts">
/**
 * ChessTagChips — dải chip thẻ hiển thị trên MỘT hàng danh sách, bấm được để
 * lọc. Trước đợt này thẻ hoàn toàn vô hình trong mọi danh sách cờ (chỉ được in
 * ra ở ArticlePrint.vue), nên người soạn gõ thẻ xong không bao giờ thấy lại.
 *
 * Nhận HAI dạng nguồn để dùng chung được cho cả 8 loại:
 *   - `tags`: mảng đối tượng thẻ (từ POST /chess/tags/of, đủ slug + kind)
 *   - `csv` : chuỗi CSV có sẵn trên thế cờ/sách/bài viết (không có slug)
 * Với dạng CSV, khóa lọc là chính TÊN thẻ — backend tự khử dấu khi so khớp nên
 * "Khai cuộc" vẫn khớp đúng thẻ slug "khai-cuoc".
 */
import { computed } from 'vue';
import type { ChessTag } from '@/api/chess';
import { isChessGroupSlug } from '@/utils/chessTaxonomy';

const props = withDefaults(
  defineProps<{ tags?: ChessTag[]; csv?: string; max?: number }>(),
  { tags: () => [], csv: '', max: 6 },
);
const emit = defineEmits<{ (e: 'pick', slugOrName: string): void }>();

const items = computed(() => {
  const out: { key: string; name: string; group: boolean }[] = [];
  for (const t of props.tags || []) {
    out.push({ key: t.slug || t.name, name: t.name, group: t.kind === 'group' });
  }
  if (!out.length && props.csv) {
    for (const raw of props.csv.split(',')) {
      const name = raw.trim();
      if (name) out.push({ key: name, name, group: isChessGroupSlug(name) });
    }
  }
  return out.slice(0, props.max);
});
</script>

<style scoped>
.ctc {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
}
.ctc-chip {
  border: 1px solid var(--td-border-level-2-color);
  background: transparent;
  color: var(--td-text-color-secondary);
  border-radius: 10px;
  padding: 0 7px;
  font-size: 11px;
  line-height: 18px;
  cursor: pointer;
}
.ctc-chip:hover {
  border-color: var(--td-brand-color);
  color: var(--td-brand-color);
}
.ctc-chip.is-group {
  border-style: dashed;
  font-weight: 600;
}
</style>
