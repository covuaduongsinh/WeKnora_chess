<template>
  <template v-if="isMobile">
    <header class="ds-mnav">
      <button
        type="button"
        class="ds-mnav-btn"
        :aria-label="navOpen ? t('menu.collapseSidebar') : t('menu.expandSidebar')"
        :aria-expanded="navOpen"
        @click="toggleNav"
      >
        <svg v-if="!navOpen" viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
          <path d="M4 7h16M4 12h16M4 17h16" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
          <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
      </button>
      <span class="ds-mnav-title">{{ title }}</span>
    </header>
    <div v-if="navOpen" class="ds-mnav-backdrop" @click="closeNav" />
  </template>
</template>

<script setup lang="ts">
/**
 * Thanh điều hướng cố định chỉ hiện trên điện thoại (≤767px): nút ☰ mở ngăn kéo
 * sidebar + tiêu đề trang + lớp phủ nền.
 *
 * Vì sao là thanh cố định chứ không nút nổi: các trang không có topbar dùng chung,
 * và `views/chess/ChessManage.vue` đặt <h1 class="cm-title"> ngay góc trên trái —
 * một nút nổi `top:8px; left:8px` sẽ đè lên tiêu đề đó.
 *
 * Phần trượt/ẩn của chính sidebar nằm ở khối A trong
 * `assets/theme/duongsinh-responsive.css` (nhắm `.main > .aside_box`), để
 * `components/menu.vue` giữ 0 dòng sửa — xem `composables/useMobileNav.ts`.
 */
import { computed, onMounted, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useMobileNav } from '@/composables/useMobileNav';

const route = useRoute();
const { t } = useI18n();
const { navOpen, isMobile, closeNav, toggleNav } = useMobileNav();

const TITLE_BY_ROUTE: Record<string, string> = {
  knowledgeBaseList: 'menu.knowledgeBase',
  knowledgeBaseDetail: 'menu.knowledgeBase',
  agentList: 'menu.agents',
  chessCourses: 'menu.chessCourses',
  organizationList: 'menu.organizations',
  chat: 'menu.chat',
  globalCreatChat: 'menu.newChat',
  kbCreatChat: 'menu.newChat',
  settings: 'menu.settings',
};

const title = computed(() => {
  const key = TITLE_BY_ROUTE[String(route.name || '')];
  return key ? t(key) : '';
});

// Bấm một mục trong menu → route đổi → phải tự đóng, nếu không drawer che trang mới.
watch(() => route.fullPath, () => closeNav());

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && navOpen.value) closeNav();
}

onMounted(() => window.addEventListener('keydown', onKeydown));
onUnmounted(() => window.removeEventListener('keydown', onKeydown));
</script>

<style lang="less" scoped>
/*
 * Thang z-index: thanh này (1202) > sidebar drawer (1201) > lớp phủ (1200), để
 * nút ✕ vẫn bấm được khi drawer mở. Dialog của TDesign ở 2500 nên vẫn nằm trên
 * tất cả — không đụng thang z-index của upstream.
 *
 * `env(safe-area-inset-*)` là phòng thủ: hiện `index.html` KHÔNG đặt
 * `viewport-fit=cover` nên các giá trị này bằng 0 (vô hại). Có sẵn thì mai này
 * bật cover cũng không phải sửa lại.
 */
.ds-mnav {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: calc(48px + env(safe-area-inset-top, 0px));
  padding: env(safe-area-inset-top, 0px) 8px 0 4px;
  z-index: 1202;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--td-brand-color);
  color: #fff;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.18);
}

.ds-mnav-btn {
  flex: 0 0 auto;
  /* 44px — ngưỡng vùng chạm tối thiểu (Apple HIG) */
  width: 44px;
  height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  cursor: pointer;

  &:active {
    background: rgba(255, 255, 255, 0.18);
  }
}

.ds-mnav-title {
  min-width: 0;
  font-size: 16px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ds-mnav-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.45);
}
</style>
