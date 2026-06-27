<template>
  <div class="chess-courses">
    <!-- Cột trái: danh sách khóa học -->
    <div class="cc-left">
      <div class="cc-header">
        <h2>Khóa học cờ vua</h2>
        <t-button theme="primary" size="small" @click="openCourseDialog()">
          <template #icon><t-icon name="add" /></template>
          Tạo khóa học
        </t-button>
      </div>
      <div v-if="courses.length === 0" class="cc-empty">Chưa có khóa học. Nhấn "Tạo khóa học".</div>
      <div v-for="c in courses" :key="c.id" class="cc-course-row"
        :class="{ active: selectedCourse && selectedCourse.id === c.id }" @click="selectCourse(c)">
        <div class="cc-course-main">
          <div class="cc-course-title">{{ c.title }}</div>
          <div class="cc-course-meta">
            <span v-if="c.level" class="cc-tag">{{ levelLabel(c.level) }}</span>
            <span class="cc-count">{{ c.lesson_count || 0 }} bài học</span>
          </div>
        </div>
        <div class="cc-course-actions">
          <t-button size="small" variant="text" @click.stop="openCourseDialog(c)"><t-icon name="edit" /></t-button>
          <t-button size="small" variant="text" theme="danger" @click.stop="removeCourse(c)"><t-icon name="delete" /></t-button>
        </div>
      </div>
    </div>

    <!-- Cột phải: bài học của khóa đang chọn -->
    <div class="cc-right">
      <template v-if="selectedCourse">
        <div class="cc-header">
          <div>
            <h2>{{ selectedCourse.title }}</h2>
            <div class="cc-desc" v-if="selectedCourse.description">{{ selectedCourse.description }}</div>
          </div>
          <t-button theme="primary" size="small" @click="openLessonDialog()">
            <template #icon><t-icon name="add" /></template>
            Thêm bài học
          </t-button>
        </div>
        <div v-if="lessons.length === 0" class="cc-empty">Chưa có bài học.</div>
        <div v-for="(l, idx) in lessons" :key="l.id" class="cc-lesson">
          <div class="cc-lesson-head" @click="toggleLesson(l.id)">
            <span class="cc-lesson-idx">{{ idx + 1 }}.</span>
            <span class="cc-lesson-title">{{ l.title }}</span>
            <span class="cc-lesson-actions">
              <t-button size="small" variant="text" @click.stop="openLessonDialog(l)"><t-icon name="edit" /></t-button>
              <t-button size="small" variant="text" theme="danger" @click.stop="removeLesson(l)"><t-icon name="delete" /></t-button>
              <t-icon :name="expandedLessons.has(l.id) ? 'chevron-up' : 'chevron-down'" />
            </span>
          </div>
          <div v-if="expandedLessons.has(l.id)" class="cc-lesson-body">
            <div class="cc-lesson-content" v-if="l.content">{{ l.content }}</div>
            <ChessBoardDisplay v-if="l.fen || l.pgn" :data="boardData(l)" />
          </div>
        </div>
      </template>
      <div v-else class="cc-empty cc-empty--big">Chọn một khóa học bên trái để xem bài học.</div>
    </div>

    <!-- Dialog khóa học -->
    <t-dialog v-model:visible="courseDialog.visible" :header="courseDialog.id ? 'Sửa khóa học' : 'Tạo khóa học'"
      :on-confirm="saveCourse" width="520px">
      <div class="cc-form">
        <label>Tên khóa học *</label>
        <t-input v-model="courseDialog.title" placeholder="VD: Khai cuộc cho người mới" />
        <label>Mô tả</label>
        <t-textarea v-model="courseDialog.description" :autosize="{ minRows: 2 }" />
        <label>Trình độ</label>
        <t-select v-model="courseDialog.level" :options="levelOptions" clearable />
        <label>Thứ tự</label>
        <t-input-number v-model="courseDialog.sort_order" />
      </div>
    </t-dialog>

    <!-- Dialog bài học -->
    <t-dialog v-model:visible="lessonDialog.visible" :header="lessonDialog.id ? 'Sửa bài học' : 'Thêm bài học'"
      :on-confirm="saveLesson" width="640px">
      <div class="cc-form">
        <label>Tên bài học *</label>
        <t-input v-model="lessonDialog.title" />
        <label>Nội dung (văn bản/markdown)</label>
        <t-textarea v-model="lessonDialog.content" :autosize="{ minRows: 4 }"
          placeholder="Nội dung bài giảng..." />
        <label>Thế cờ FEN (tùy chọn — sẽ hiện bàn cờ)</label>
        <t-input v-model="lessonDialog.fen" placeholder="rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1" />
        <label>Ván minh họa PGN (tùy chọn)</label>
        <t-textarea v-model="lessonDialog.pgn" :autosize="{ minRows: 2 }" placeholder="1. e4 e5 2. Nf3 Nc6 ..." />
        <label>Thứ tự</label>
        <t-input-number v-model="lessonDialog.sort_order" />
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import ChessBoardDisplay from '@/views/chat/components/tool-results/ChessBoardDisplay.vue';
import type { ChessBoardData } from '@/types/tool-results';
import {
  listCourses, createCourse, updateCourse, deleteCourse,
  listLessons, createLesson, updateLesson, deleteLesson,
  type ChessCourse, type ChessLesson,
} from '@/api/chess';

const courses = ref<ChessCourse[]>([]);
const selectedCourse = ref<ChessCourse | null>(null);
const lessons = ref<ChessLesson[]>([]);
const expandedLessons = ref<Set<string>>(new Set());

const levelOptions = [
  { label: 'Cơ bản', value: 'co-ban' },
  { label: 'Trung cấp', value: 'trung-cap' },
  { label: 'Nâng cao', value: 'nang-cao' },
];
const levelLabel = (v: string) => levelOptions.find(o => o.value === v)?.label || v;

const STARTFEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';
function boardData(l: ChessLesson): ChessBoardData {
  if (l.pgn && l.pgn.trim()) {
    return { display_type: 'chess_board', fen: STARTFEN, pgn: l.pgn, caption: l.title };
  }
  return { display_type: 'chess_board', fen: l.fen || STARTFEN, caption: l.title };
}

async function loadCourses() {
  try {
    const res: any = await listCourses();
    courses.value = res?.data || [];
  } catch {
    MessagePlugin.error('Tải khóa học thất bại');
  }
}

async function selectCourse(c: ChessCourse) {
  selectedCourse.value = c;
  expandedLessons.value = new Set();
  try {
    const res: any = await listLessons(c.id);
    lessons.value = res?.data || [];
  } catch {
    lessons.value = [];
  }
}

function toggleLesson(id: string) {
  const next = new Set(expandedLessons.value);
  next.has(id) ? next.delete(id) : next.add(id);
  expandedLessons.value = next;
}

// ---- Course dialog ----
const courseDialog = reactive<any>({ visible: false, id: '', title: '', description: '', level: '', sort_order: 0 });
function openCourseDialog(c?: ChessCourse) {
  courseDialog.visible = true;
  courseDialog.id = c?.id || '';
  courseDialog.title = c?.title || '';
  courseDialog.description = c?.description || '';
  courseDialog.level = c?.level || '';
  courseDialog.sort_order = c?.sort_order || 0;
}
async function saveCourse() {
  if (!courseDialog.title.trim()) { MessagePlugin.warning('Nhập tên khóa học'); return; }
  const payload = {
    title: courseDialog.title, description: courseDialog.description,
    level: courseDialog.level, sort_order: courseDialog.sort_order,
  };
  try {
    if (courseDialog.id) await updateCourse(courseDialog.id, payload);
    else await createCourse(payload);
    courseDialog.visible = false;
    await loadCourses();
    MessagePlugin.success('Đã lưu khóa học');
  } catch {
    MessagePlugin.error('Lưu thất bại');
  }
}
function removeCourse(c: ChessCourse) {
  DialogPlugin.confirm({
    header: 'Xóa khóa học', body: `Xóa "${c.title}" và toàn bộ bài học?`,
    theme: 'warning', confirmBtn: { content: 'Xóa', theme: 'danger' },
    onConfirm: async () => {
      try {
        await deleteCourse(c.id);
        if (selectedCourse.value?.id === c.id) { selectedCourse.value = null; lessons.value = []; }
        await loadCourses();
        MessagePlugin.success('Đã xóa');
      } catch { MessagePlugin.error('Xóa thất bại'); }
    },
  });
}

// ---- Lesson dialog ----
const lessonDialog = reactive<any>({ visible: false, id: '', title: '', content: '', fen: '', pgn: '', sort_order: 0 });
function openLessonDialog(l?: ChessLesson) {
  lessonDialog.visible = true;
  lessonDialog.id = l?.id || '';
  lessonDialog.title = l?.title || '';
  lessonDialog.content = l?.content || '';
  lessonDialog.fen = l?.fen || '';
  lessonDialog.pgn = l?.pgn || '';
  lessonDialog.sort_order = l?.sort_order || lessons.value.length;
}
async function saveLesson() {
  if (!selectedCourse.value) return;
  if (!lessonDialog.title.trim()) { MessagePlugin.warning('Nhập tên bài học'); return; }
  const payload = {
    title: lessonDialog.title, content: lessonDialog.content,
    fen: lessonDialog.fen, pgn: lessonDialog.pgn, sort_order: lessonDialog.sort_order,
  };
  try {
    if (lessonDialog.id) await updateLesson(lessonDialog.id, payload);
    else await createLesson(selectedCourse.value.id, payload);
    lessonDialog.visible = false;
    await selectCourse(selectedCourse.value);
    await loadCourses(); // cập nhật số bài học
    MessagePlugin.success('Đã lưu bài học');
  } catch {
    MessagePlugin.error('Lưu thất bại');
  }
}
function removeLesson(l: ChessLesson) {
  DialogPlugin.confirm({
    header: 'Xóa bài học', body: `Xóa "${l.title}"?`,
    theme: 'warning', confirmBtn: { content: 'Xóa', theme: 'danger' },
    onConfirm: async () => {
      try {
        await deleteLesson(l.id);
        if (selectedCourse.value) await selectCourse(selectedCourse.value);
        await loadCourses();
        MessagePlugin.success('Đã xóa');
      } catch { MessagePlugin.error('Xóa thất bại'); }
    },
  });
}

loadCourses();
</script>

<style lang="less" scoped>
.chess-courses {
  display: flex;
  height: 100%;
  gap: 16px;
  padding: 16px 20px;
  box-sizing: border-box;
  overflow: hidden;
}
.cc-left {
  width: 340px;
  flex: 0 0 340px;
  overflow-y: auto;
  border-right: 1px solid var(--td-component-stroke);
  padding-right: 12px;
}
.cc-right { flex: 1 1 auto; overflow-y: auto; }
.cc-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;
  h2 { font-size: 18px; margin: 0; color: var(--td-text-color-primary); }
}
.cc-desc { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 2px; }
.cc-empty { color: var(--td-text-color-placeholder); font-size: 14px; padding: 16px 4px; }
.cc-empty--big { text-align: center; padding-top: 80px; }
.cc-course-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px; border-radius: 8px; cursor: pointer; margin-bottom: 6px;
  border: 1px solid var(--td-component-stroke);
  &:hover { background: var(--td-bg-color-container-hover); }
  &.active { background: var(--td-bg-color-secondarycontainer); border-color: var(--td-brand-color); }
}
.cc-course-title { font-weight: 600; color: var(--td-text-color-primary); }
.cc-course-meta { display: flex; gap: 8px; align-items: center; margin-top: 4px; font-size: 12px; }
.cc-tag { background: var(--td-brand-color-light); color: var(--td-brand-color); padding: 0 6px; border-radius: 4px; }
.cc-count { color: var(--td-text-color-secondary); }
.cc-course-actions { display: flex; }
.cc-lesson { border: 1px solid var(--td-component-stroke); border-radius: 8px; margin-bottom: 8px; overflow: hidden; }
.cc-lesson-head {
  display: flex; align-items: center; gap: 8px; padding: 10px 12px; cursor: pointer;
  &:hover { background: var(--td-bg-color-container-hover); }
}
.cc-lesson-idx { color: var(--td-text-color-secondary); }
.cc-lesson-title { font-weight: 600; flex: 1; color: var(--td-text-color-primary); }
.cc-lesson-actions { display: flex; align-items: center; gap: 2px; }
.cc-lesson-body { padding: 12px; border-top: 1px solid var(--td-component-stroke); }
.cc-lesson-content { white-space: pre-wrap; color: var(--td-text-color-primary); margin-bottom: 12px; line-height: 1.6; }
.cc-form {
  display: flex; flex-direction: column; gap: 6px;
  label { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 6px; }
}
</style>
