<template>
  <div class="chess-manage">
    <div class="cm-header">
      <h1 class="cm-title">Quản lý cờ vua</h1>
      <!-- Ô tìm đặt ở HEADER (ngoài mọi tab) vì nó tra CẢ 8 loại nội dung —
           giống nút Kho tri thức bên cạnh. Enter chuyển sang tab Tìm kiếm. -->
      <t-input v-model="headerQuery" class="cm-search" placeholder="Tìm trong kho cờ…" clearable
        @enter="gotoSearch">
        <template #prefix-icon><t-icon name="search" /></template>
      </t-input>
      <div class="cm-spacer"></div>
      <!-- Kho tri thức phục vụ CẢ 7 loại thực thể (không riêng bài viết) nên nút
           đặt ở đầu trang, ngoài mọi tab. -->
      <t-button size="small" variant="outline" @click="kbStatusVisible = true">
        <template #icon><t-icon name="system-storage" /></template>Kho tri thức
      </t-button>
    </div>
    <t-tabs v-model="tab" class="cm-tabs">
      <t-tab-panel value="courses" label="Khóa học">
        <div class="cm-pane"><ChessCourses :focus-lesson-slug="focusLessonSlug" :focus-course-slug="focusCourseSlug" /></div>
      </t-tab-panel>
      <t-tab-panel value="games" label="Kho ván đấu">
        <div class="cm-pane"><GameLibrary v-if="tab === 'games'" :focus-slug="focusGameSlug" /></div>
      </t-tab-panel>
      <t-tab-panel value="puzzles" label="Ngân hàng bài tập">
        <div class="cm-pane"><PuzzleBank v-if="tab === 'puzzles'" :focus-slug="focusPuzzleSlug" /></div>
      </t-tab-panel>
      <t-tab-panel value="positions" label="Ngân hàng thế cờ">
        <div class="cm-pane"><PositionBank v-if="tab === 'positions'" :focus-slug="focusPositionSlug" /></div>
      </t-tab-panel>
      <t-tab-panel value="books" label="Thư viện sách">
        <div class="cm-pane">
          <BookLibrary v-if="tab === 'books'" :focus-book-slug="focusBookSlug" :focus-chapter-slug="focusChapterSlug" />
        </div>
      </t-tab-panel>
      <t-tab-panel value="articles" label="Ngân hàng bài viết">
        <div class="cm-pane"><ArticleBank v-if="tab === 'articles'" :focus-slug="focusArticleSlug" /></div>
      </t-tab-panel>
      <!-- Thẻ là trục phân loại NGANG phủ cả 8 loại, nên tab này không phải
           "loại nội dung thứ 7" mà là MỤC LỤC NGANG của 6 tab kia. -->
      <t-tab-panel value="tags" label="Thẻ">
        <div class="cm-pane"><TagBrowser v-if="tab === 'tags'" /></div>
      </t-tab-panel>
      <t-tab-panel value="search" label="Tìm kiếm">
        <div class="cm-pane"><SearchPanel v-if="tab === 'search'" :initial-query="searchQuery" /></div>
      </t-tab-panel>
    </t-tabs>
    <ChessKBStatusDialog v-model:visible="kbStatusVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import ChessCourses from './ChessCourses.vue';
import GameLibrary from './GameLibrary.vue';
import PuzzleBank from './PuzzleBank.vue';
import PositionBank from './PositionBank.vue';
import BookLibrary from './BookLibrary.vue';
import ArticleBank from './ArticleBank.vue';
import TagBrowser from './TagBrowser.vue';
import SearchPanel from './SearchPanel.vue';
import ChessKBStatusDialog from './components/ChessKBStatusDialog.vue';

const route = useRoute();
const tab = ref('courses');
const kbStatusVisible = ref(false);

// headerQuery là ô ở header; searchQuery là thứ THỰC SỰ được gửi đi. Tách hai
// biến để gõ dở trong header không kích hoạt tìm kiếm — chỉ Enter mới chuyển tab.
const headerQuery = ref('');
const searchQuery = ref('');
function gotoSearch() {
  if (!headerQuery.value.trim()) return;
  searchQuery.value = headerQuery.value;
  tab.value = 'search';
}

// Deep-link "Mở trong thư viện": ?ref=game/<slug> | puzzle/<slug> | lesson/<slug>
// | course/<slug> | position/<slug> | book/<slug> | chapter/<slug> | article/<slug>
// → chuyển sang đúng tab và bảo component con chọn/mở đối tượng tương ứng.
const focusGameSlug = ref('');
const focusPuzzleSlug = ref('');
const focusLessonSlug = ref('');
const focusCourseSlug = ref('');
const focusPositionSlug = ref('');
const focusBookSlug = ref('');
const focusChapterSlug = ref('');
const focusArticleSlug = ref('');

const focusRef = computed(() => String(route.query.ref || ''));
function parseRef(r: string): { type: string; slug: string } | null {
  const i = r.indexOf('/');
  if (i <= 0) return null;
  return { type: r.slice(0, i), slug: r.slice(i + 1) };
}

watch(
  focusRef,
  (r) => {
    const p = parseRef(r);
    if (!p) return;
    if (p.type === 'game') {
      tab.value = 'games';
      focusGameSlug.value = p.slug;
    } else if (p.type === 'puzzle') {
      tab.value = 'puzzles';
      focusPuzzleSlug.value = p.slug;
    } else if (p.type === 'lesson') {
      tab.value = 'courses';
      focusLessonSlug.value = p.slug;
    } else if (p.type === 'course') {
      tab.value = 'courses';
      focusCourseSlug.value = p.slug;
    } else if (p.type === 'position') {
      tab.value = 'positions';
      focusPositionSlug.value = p.slug;
    } else if (p.type === 'book') {
      tab.value = 'books';
      focusBookSlug.value = p.slug;
    } else if (p.type === 'chapter') {
      tab.value = 'books';
      focusChapterSlug.value = p.slug;
    } else if (p.type === 'article') {
      tab.value = 'articles';
      focusArticleSlug.value = p.slug;
    }
  },
  { immediate: true },
);
</script>

<style lang="less" scoped>
.chess-manage {
  display: flex; flex-direction: column; height: 100%; box-sizing: border-box;
  padding: 12px 20px 0; overflow: hidden;
  /* Họa tiết ô cờ (checkerboard) rất nhạt — nhận diện Dương Sinh, không cản đọc. */
  background-image:
    linear-gradient(45deg, rgba(43, 57, 144, 0.025) 25%, transparent 25%, transparent 75%, rgba(43, 57, 144, 0.025) 75%),
    linear-gradient(45deg, rgba(43, 57, 144, 0.025) 25%, transparent 25%, transparent 75%, rgba(43, 57, 144, 0.025) 75%);
  background-size: 44px 44px;
  background-position: 0 0, 22px 22px;
}
.cm-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 8px; }
.cm-title { font-size: 20px; margin: 0; color: var(--td-text-color-primary); }
/* justify-content của .cm-header là space-between; .cm-spacer đẩy nút Kho tri
   thức về sát mép phải để ô tìm nằm ngay cạnh tiêu đề. */
.cm-search { width: 260px; }
.cm-spacer { flex: 1; }
.cm-tabs { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.cm-tabs :deep(.t-tabs__content) { flex: 1; min-height: 0; }
.cm-tabs :deep(.t-tab-panel) { height: 100%; }
.cm-pane { height: 100%; min-height: 0; }

/* ── Điện thoại ─────────────────────────────────────────────────────────────
   `screen and` là BẮT BUỘC (xem PuzzleBank.vue). Thanh 6 tab nhãn tiếng Việt dài
   sẽ tự cuộn ngang — TDesign đã hỗ trợ sẵn, chỉ cần thu nhỏ tiêu đề + lề. */
@media screen and (max-width: 767px) {
  .chess-manage { padding: 8px 0 0; }
  .cm-header { margin: 0 12px 6px; }
  .cm-title { font-size: 17px; }
  /* Điện thoại: ô tìm xuống hàng riêng và chiếm trọn bề ngang. */
  .cm-header { flex-wrap: wrap; }
  .cm-search { width: 100%; order: 9; }
  .cm-spacer { display: none; }
  .cm-tabs :deep(.t-tabs__nav-item) { font-size: 14px; }
}
</style>
