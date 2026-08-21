<template>
  <div v-if="total > 0" class="clf">
    <span class="clf-count">
      Đang xem <b>{{ loaded }}</b>/<b>{{ total }}</b>
    </span>
    <t-button v-if="hasMore" size="small" variant="outline" :loading="loading" @click="emit('more')">
      Tải thêm
    </t-button>
  </div>
</template>

<script setup lang="ts">
/**
 * ChessListFooter — chân danh sách hiển thị "đang xem N/Tổng" + nút tải thêm.
 *
 * Trước đợt này danh sách bị cắt ở 500 bản ghi mà không có bất kỳ dấu hiệu
 * nào; người dùng không có cách nào biết mình đang nhìn thiếu. Dòng đếm này là
 * phần quan trọng nhất — nút "Tải thêm" chỉ là hệ quả.
 */
withDefaults(defineProps<{ loaded: number; total: number; hasMore: boolean; loading?: boolean }>(), {
  loading: false,
});
const emit = defineEmits<{ (e: 'more'): void }>();
</script>

<style scoped>
.clf {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 10px 0 4px;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}
.clf-count b {
  color: var(--td-text-color-secondary);
  font-weight: 600;
}
</style>
