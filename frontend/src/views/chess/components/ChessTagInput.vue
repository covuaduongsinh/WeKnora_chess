<template>
  <div class="cti">
    <t-select
      :value="selected"
      multiple
      filterable
      creatable
      clearable
      :min-collapsed-num="0"
      :placeholder="placeholder"
      :loading="loading"
      @change="onChange"
      @create="onCreate"
    >
      <t-option-group v-if="groupTags.length" label="Nhóm nội dung">
        <t-option v-for="t in groupTags" :key="t.slug" :value="t.name" :label="t.name" />
      </t-option-group>
      <t-option-group v-if="freeTags.length" label="Thẻ">
        <t-option v-for="t in freeTags" :key="t.slug" :value="t.name" :label="t.name" />
      </t-option-group>
    </t-select>
    <div v-if="hint" class="cti-hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
/**
 * ChessTagInput — ô nhập THẺ dùng chung cho mọi trang cờ.
 *
 * Vì sao là t-select `multiple creatable filterable` chứ không phải component
 * chip tự viết: repo đã dùng đúng khuôn này ở GraphSettings.vue / PuzzleBank.vue,
 * nên hành vi bàn phím, giao diện và chủ đề tối/sáng khớp sẵn — và từ điển thẻ
 * của một tenant đủ nhỏ để tải một lần, không cần tìm kiếm phía máy chủ với
 * debounce như popup wikilink.
 *
 * Giá trị vào/ra là TÊN THẺ có dấu ("Khai cuộc"), không phải slug. Backend tự
 * quy về slug đã khử dấu, nên gõ "khai cuoc" vẫn khớp đúng thẻ "Khai cuộc" —
 * lần lưu sau ô này sẽ hiện lại tên chuẩn.
 *
 * v-model là CSV để cắm thẳng vào các form đang lưu cột `tags` dạng chuỗi
 * (thế cờ/sách/bài viết); nơi không có cột đó thì đọc/ghi qua props `modelValue`
 * rồi gọi assignChessTags.
 */
import { computed, onMounted, ref, watch } from 'vue';
import { listChessTags, type ChessTag } from '@/api/chess';

const props = withDefaults(
  defineProps<{
    /** CSV tên thẻ, vd "Ghim, Khai cuộc". */
    modelValue?: string;
    placeholder?: string;
    hint?: string;
    /** Chỉ hiện thẻ nhóm (dùng cho ô "Nhóm nội dung" riêng). */
    onlyGroups?: boolean;
    /** Bỏ thẻ nhóm khỏi gợi ý (dùng cho ô "Thẻ" khi đã có ô nhóm riêng). */
    excludeGroups?: boolean;
  }>(),
  { modelValue: '', placeholder: 'Chọn hoặc gõ thẻ mới…', hint: '', onlyGroups: false, excludeGroups: false },
);
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>();

const tags = ref<ChessTag[]>([]);
const loading = ref(false);
// Thẻ người dùng vừa gõ nhưng CHƯA tồn tại trong từ điển: phải giữ lại làm
// option, nếu không t-select hiển thị giá trị trống cho mục vừa tạo.
const adhoc = ref<string[]>([]);

const selected = computed(() =>
  props.modelValue
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean),
);

const visible = computed(() => {
  let list = tags.value;
  if (props.onlyGroups) list = list.filter((t) => t.kind === 'group');
  else if (props.excludeGroups) list = list.filter((t) => t.kind !== 'group');
  const extra = adhoc.value
    .filter((name) => !list.some((t) => t.name === name))
    .map((name) => ({ id: '', slug: name, name, kind: 'free', description: '', color: '', usage_count: 0, sort_order: 0 }) as ChessTag);
  return [...list, ...extra];
});

const groupTags = computed(() => visible.value.filter((t) => t.kind === 'group'));
const freeTags = computed(() => visible.value.filter((t) => t.kind !== 'group'));

function onChange(v: string[]) {
  emit('update:modelValue', (v || []).join(', '));
}

function onCreate(v: string) {
  const name = String(v || '').trim();
  if (!name) return;
  if (!adhoc.value.includes(name)) adhoc.value.push(name);
  if (!selected.value.includes(name)) {
    emit('update:modelValue', [...selected.value, name].join(', '));
  }
}

async function load() {
  loading.value = true;
  try {
    const res: any = await listChessTags();
    tags.value = res?.data || [];
  } catch {
    tags.value = [];
  } finally {
    loading.value = false;
  }
}

// Giá trị hiện có mà từ điển chưa biết (vd dữ liệu cũ chưa backfill) vẫn phải
// hiện thành chip, nên nạp luôn vào adhoc.
watch(
  () => props.modelValue,
  () => {
    for (const name of selected.value) {
      if (!tags.value.some((t) => t.name === name) && !adhoc.value.includes(name)) {
        adhoc.value.push(name);
      }
    }
  },
  { immediate: true },
);

onMounted(load);
defineExpose({ reload: load });
</script>

<style scoped>
.cti {
  width: 100%;
}
.cti-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}
</style>
