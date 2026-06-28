<template>
  <div v-if="links.length" class="chess-backlinks">
    <div class="cbl-title">{{ t('chess.ref.backlinks') }}</div>
    <a
      v-for="(b, i) in links"
      :key="i"
      href="#"
      class="cbl-item"
      :class="b.source_type === 'lesson' ? 'cbl-lesson' : 'cbl-wiki'"
      :title="t('chess.ref.openBacklink')"
      @click.prevent="open(b)"
    >
      <t-icon :name="b.source_type === 'lesson' ? 'books' : 'file'" size="14px" />
      {{ b.page_title || b.page_slug }}
    </a>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { resolveChessBacklinks, type ChessBacklink, type ChessRefType } from '@/utils/chessRef';

// Hiển thị "Được liên kết bởi" cho một đối tượng cờ + bấm để về nguồn.
// Nguồn 'lesson' → mở bài giảng trong Khóa học; 'wiki' → mở trang wiki trong KB.
const props = defineProps<{ refType: ChessRefType; slug?: string }>();
const emit = defineEmits<{ (e: 'navigate'): void }>();
const { t } = useI18n();
const router = useRouter();

const links = ref<ChessBacklink[]>([]);

async function load() {
  links.value = [];
  if (!props.slug) return;
  links.value = await resolveChessBacklinks(`${props.refType}/${props.slug}`);
}

function open(b: ChessBacklink) {
  if (b.source_type === 'lesson') {
    router.push({ name: 'chessCourses', query: { ref: `lesson/${b.page_slug}` } });
  } else {
    router.push({ name: 'knowledgeBaseDetail', params: { kbId: b.kb_id }, query: { tab: 'wiki', slug: b.page_slug } });
  }
  emit('navigate');
}

watch(() => [props.refType, props.slug] as const, load, { immediate: true });
</script>

<style lang="less" scoped>
.chess-backlinks { margin-top: 12px; }
.cbl-title { font-size: 12px; color: var(--td-text-color-secondary); margin-bottom: 6px; }
.cbl-item {
  display: inline-flex; align-items: center; gap: 4px;
  margin: 0 6px 6px 0; padding: 2px 8px; font-size: 12px; border-radius: 4px;
  text-decoration: none; cursor: pointer;
  background: var(--td-bg-color-secondarycontainer); color: var(--td-text-color-primary);
  &:hover { background: var(--td-bg-color-container-hover); color: var(--td-brand-color); }
}
</style>
