import { ref } from 'vue';

/**
 * useChessPaging — trạng thái phân trang dùng chung cho các trang cờ.
 *
 * VÌ SAO CẦN: trước đợt này mọi danh sách cờ bị backend cắt cứng ở 500 bản ghi
 * mà KHÔNG có dấu hiệu gì — nội dung thứ 501 trở đi biến mất im lặng. Nay
 * endpoint trả thêm khối `meta` (total/page/page_size/has_more), nên giao diện
 * hiện được "đang xem 50/137" và có nút tải thêm.
 *
 * Cố ý KHÔNG bọc luôn cả việc gọi API: mỗi trang cờ có bộ lọc, hàm tải và các
 * lối gọi lại (sau khi tạo/sửa/xóa, deep-link ?ref=) rất khác nhau — gộp hết
 * vào một composable sẽ phải nhận quá nhiều tham số và dễ gây hồi quy trên
 * những trang đã chạy ổn định.
 */
export function useChessPaging(pageSize = 50) {
  const page = ref(1);
  const total = ref(0);
  const hasMore = ref(false);
  const loadingMore = ref(false);

  /** Tham số truyền lên API cho trang hiện tại. */
  function params(p = page.value) {
    return { page: p, page_size: pageSize };
  }

  /**
   * Đọc khối `meta` của phản hồi. Phản hồi CŨ (chưa có meta) vẫn dùng được:
   * khi đó coi như không còn trang nào và tổng = số mục đang có.
   */
  function applyMeta(res: any, loadedCount: number) {
    const meta = res?.meta;
    if (!meta) {
      total.value = loadedCount;
      hasMore.value = false;
      return;
    }
    total.value = Number(meta.total ?? loadedCount);
    hasMore.value = Boolean(meta.has_more);
  }

  function reset() {
    page.value = 1;
    hasMore.value = false;
  }

  return { page, total, hasMore, loadingMore, pageSize, params, applyMeta, reset };
}

/**
 * debounceFn — hoãn gọi cho tới khi người dùng ngừng gõ.
 *
 * Repo không có tiện ích debounce dùng chung (không lodash, không
 * @vueuse/core), và các ô tìm của trang cờ đang gắn `@change` của t-input —
 * vốn phát MỖI PHÍM GÕ, nên mỗi ký tự là một request HTTP và không hủy request
 * cũ. Với danh sách lớn, phản hồi về không theo thứ tự còn làm kết quả nhảy.
 *
 * 300ms: đủ để gõ trọn một từ tiếng Việt có dấu mà vẫn thấy tức thì.
 */
export function debounceFn<T extends (...args: any[]) => void>(fn: T, wait = 300) {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const wrapped = (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => fn(...args), wait);
  };
  wrapped.cancel = () => {
    if (timer) clearTimeout(timer);
    timer = undefined;
  };
  return wrapped as T & { cancel: () => void };
}
