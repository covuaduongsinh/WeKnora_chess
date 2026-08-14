/**
 * Điều hướng trên điện thoại: biến sidebar cố định 260px thành ngăn kéo (drawer)
 * trượt từ trái, phủ lên nội dung.
 *
 * TẠI SAO KHÔNG SỬA `components/menu.vue` (2069 dòng, file dùng chung upstream):
 * toàn bộ phần trượt/ẩn/hiện làm bằng CSS trong `assets/theme/duongsinh-responsive.css`
 * (khối A) nhắm vào `.main > .aside_box`. Giữ menu.vue ở 0 dòng sửa → merge
 * upstream không đụng độ. Đánh đổi: nếu upstream đổi tên class thì drawer chết
 * CÂM — `responsive.style.test.mjs` có test khoá đúng 2 class đó làm lưới an toàn.
 *
 * TẠI SAO PHẢI "CHE" (shadow) `sidebarCollapsed`:
 * menu.vue dùng `v-if="!uiStore.sidebarCollapsed"` ở 7 chỗ để ẩn/hiện NỘI DUNG
 * (logo, tenant selector, danh sách phiên…). Nếu vào mobile khi cờ đang `true`
 * (người dùng từng thu gọn trên desktop, đã lưu localStorage), drawer sẽ mở ra
 * một dải icon 60px vô dụng — CSS không cứu được vì đó là `v-if`, không phải style.
 * Nên khi vào mobile ta gán TRỰC TIẾP `ui.sidebarCollapsed = false` (KHÔNG gọi
 * `expandSidebar()` — action đó ghi localStorage, sẽ xoá mất ưu tiên desktop của
 * người dùng), và khi ra desktop thì đọc lại localStorage để khôi phục.
 */
import { ref, watch } from 'vue';
import { useUIStore } from '@/stores/ui';
import { isMobile } from './useBreakpoint';

const SIDEBAR_STORAGE_KEY = 'sidebar_collapsed';

/** Drawer đang mở hay không. Dùng chung toàn app (singleton, ngoài hàm). */
const navOpen = ref(false);

/** Chỉ cài watcher một lần dù composable được gọi ở nhiều component. */
let installed = false;

export function useMobileNav() {
  const ui = useUIStore();

  if (!installed) {
    installed = true;

    watch(
      isMobile,
      (mobile) => {
        // Đổi chế độ thì luôn đóng drawer, tránh trạng thái treo lơ lửng.
        navOpen.value = false;
        if (mobile) {
          ui.sidebarCollapsed = false;
        } else {
          ui.sidebarCollapsed = localStorage.getItem(SIDEBAR_STORAGE_KEY) === 'true';
        }
      },
      { immediate: true },
    );

    watch(navOpen, (open) => {
      // Class trên <html> vừa để CSS trượt drawer, vừa để khoá cuộn nền.
      document.documentElement.classList.toggle('ds-nav-open', open);
    });
  }

  return {
    navOpen,
    isMobile,
    openNav: () => {
      navOpen.value = true;
    },
    closeNav: () => {
      navOpen.value = false;
    },
    toggleNav: () => {
      navOpen.value = !navOpen.value;
    },
  };
}
