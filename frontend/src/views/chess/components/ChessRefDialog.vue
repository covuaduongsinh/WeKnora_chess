<template>
  <t-dialog
    :visible="visible"
    :header="resolved?.title || t('chess.ref.dialogTitle')"
    :footer="false"
    width="560px"
    @close="close"
    @update:visible="(v: boolean) => !v && close()"
  >
    <div class="crd-body">
      <div v-if="loading" class="crd-state">{{ t('chess.ref.loading') }}</div>
      <div v-else-if="!resolved || !resolved.found" class="crd-state crd-missing">
        {{ t('chess.ref.notFound', { ref: refStr }) }}
      </div>
      <template v-else>
        <ChessBoardDisplay v-if="resolved.board" :data="resolved.board" />
        <div v-else class="crd-state">{{ t('chess.ref.noBoard') }}</div>
        <ChessBacklinks :ref-type="resolved.type" :slug="resolved.slug" @navigate="close" />
        <div class="crd-actions">
          <t-button size="small" theme="primary" variant="outline" @click="openInLibrary">
            {{ t('chess.ref.openInLibrary') }}
          </t-button>
        </div>
      </template>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import ChessBacklinks from './ChessBacklinks.vue';
import { resolveChessRef, type ResolvedChessRef } from '@/utils/chessRef';

// Popup hiển thị bàn cờ khi bấm CHIP [[game/<slug>]]. Host điều khiển qua
// v-model:visible + refStr.
const props = defineProps<{ visible: boolean; refStr: string }>();
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>();
const { t } = useI18n();
const router = useRouter();

const loading = ref(false);
const resolved = ref<ResolvedChessRef | null>(null);

function close() {
  emit('update:visible', false);
}

async function load() {
  if (!props.refStr) return;
  loading.value = true;
  resolved.value = null;
  try {
    resolved.value = await resolveChessRef(props.refStr);
  } catch {
    resolved.value = null;
  } finally {
    loading.value = false;
  }
}

function openInLibrary() {
  router.push({ name: 'chessCourses', query: { ref: props.refStr } });
  close();
}

// Nạp lại mỗi khi mở popup hoặc đổi tham chiếu.
watch(
  () => [props.visible, props.refStr] as const,
  ([vis]) => {
    if (vis) load();
  },
  { immediate: true },
);
</script>

<style lang="less" scoped>
.crd-body { min-height: 60px; }
.crd-state { color: var(--td-text-color-placeholder); font-size: 13px; padding: 8px 0; }
.crd-missing { color: var(--td-warning-color); }
.crd-actions { margin-top: 12px; display: flex; justify-content: flex-end; }
.crd-backlinks { margin-top: 12px; }
.crd-backlinks-title { font-size: 12px; color: var(--td-text-color-secondary); margin-bottom: 6px; }
.crd-backlink {
  display: inline-block; margin: 0 6px 6px 0; padding: 2px 8px; font-size: 12px;
  background: var(--td-bg-color-secondarycontainer); border-radius: 4px; color: var(--td-text-color-primary);
}
</style>
