import { defineStore } from 'pinia';
import { ref } from 'vue';

// Bản nháp bài giảng được tạo từ một trang wiki cờ vua. WikiBrowser đặt bản nháp
// rồi điều hướng sang trang Quản lý cờ vua; ChessCourses "lấy" bản nháp (một lần)
// và mở sẵn hộp thoại thêm bài giảng. Đây là cách nối Wiki → LMS không cần đổi
// schema backend.
export interface ChessLessonDraft {
  title: string;
  content: string;
  // KB và slug của trang wiki nguồn (ghi lại nguồn gốc; có thể dùng sau này để
  // điều hướng ngược về trang wiki).
  sourceKbId: string;
  sourceSlug: string;
}

export const useChessWikiDraftStore = defineStore('chessWikiDraft', () => {
  const draft = ref<ChessLessonDraft | null>(null);

  function setDraft(d: ChessLessonDraft) {
    draft.value = d;
  }

  // takeDraft trả về bản nháp hiện có rồi xóa đi (tiêu thụ một lần) để không mở
  // lại hộp thoại khi người dùng quay lại trang sau này.
  function takeDraft(): ChessLessonDraft | null {
    const d = draft.value;
    draft.value = null;
    return d;
  }

  return { draft, setDraft, takeDraft };
});
