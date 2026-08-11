import { nextTick } from 'vue';

// useChessEditor.ts — tiện ích soạn thảo markdown THUẦN (chèn tại con trỏ,
// bọc lựa chọn, khối ```chess, dán/kéo-thả ảnh) cho MỘT <textarea>. Tách
// riêng cho ArticleBank.vue — KHÔNG đụng vào phần soạn thảo nội bộ của
// BookLibrary.vue (đợt trước, đã chạy ổn định) để tránh rủi ro hồi quy khi
// refactor file vừa ship. Gộp lại là backlog, không phải việc của đợt này.

export interface ChessEditorContent {
  value: string;
}

export interface ChessEditorOptions {
  getTextarea: () => HTMLTextAreaElement | null;
  content: ChessEditorContent; // thường truyền vào bằng computed({get,set})
  // upload nhận 1 file ảnh, trả URL để chèn vào markdown dạng ![tên](url).
  upload: (file: File) => Promise<string>;
  onUploadError?: (err: unknown) => void;
}

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

export function useChessEditor(opts: ChessEditorOptions) {
  function insertAtCursor(text: string) {
    const ta = opts.getTextarea();
    const v = opts.content.value;
    if (!ta) {
      opts.content.value = v + text;
      return;
    }
    const start = ta.selectionStart ?? v.length;
    const end = ta.selectionEnd ?? v.length;
    opts.content.value = v.slice(0, start) + text + v.slice(end);
    nextTick(() => {
      ta.focus();
      const pos = start + text.length;
      ta.setSelectionRange(pos, pos);
    });
  }

  // Bọc đoạn đang chọn bằng prefix/suffix (VD **đậm**); nếu không chọn gì thì
  // chèn placeholder rồi bôi đen để gõ đè.
  function wrapSelectionSimple(prefix: string, suffix: string, placeholder: string) {
    const ta = opts.getTextarea();
    const v = opts.content.value;
    if (!ta) {
      insertAtCursor(prefix + placeholder + suffix);
      return;
    }
    const start = ta.selectionStart ?? 0;
    const end = ta.selectionEnd ?? 0;
    const selected = v.slice(start, end) || placeholder;
    opts.content.value = v.slice(0, start) + prefix + selected + suffix + v.slice(end);
    nextTick(() => {
      ta.focus();
      ta.setSelectionRange(start + prefix.length, start + prefix.length + selected.length);
    });
  }

  function insertChessBlock() {
    insertAtCursor(`\n\`\`\`chess\n${START_FEN}\n\`\`\`\n`);
  }

  async function uploadAndInsert(file: File) {
    if (!file.type.startsWith('image/')) return;
    try {
      const url = await opts.upload(file);
      insertAtCursor(`![${file.name}](${url})`);
    } catch (e) {
      opts.onUploadError?.(e);
    }
  }

  function onPaste(e: ClipboardEvent) {
    const items = e.clipboardData?.items;
    if (!items) return;
    for (const it of Array.from(items)) {
      if (it.kind === 'file' && it.type.startsWith('image/')) {
        const file = it.getAsFile();
        if (file) {
          e.preventDefault();
          uploadAndInsert(file);
        }
      }
    }
  }
  function onDrop(e: DragEvent) {
    const files = e.dataTransfer?.files;
    if (!files || !files.length) return;
    e.preventDefault();
    for (const f of Array.from(files)) {
      if (f.type.startsWith('image/')) uploadAndInsert(f);
    }
  }
  function onDragOver(e: DragEvent) {
    e.preventDefault();
  }

  // Gắn/gỡ listener dán+kéo-thả cho MỘT textarea — gọi lại mỗi khi textarea
  // đổi (VD mở lại khung soạn) để không gắn trùng.
  let bound: HTMLTextAreaElement | null = null;
  function bindTextarea(ta: HTMLTextAreaElement | null) {
    if (bound === ta) return;
    if (bound) {
      bound.removeEventListener('paste', onPaste as EventListener);
      bound.removeEventListener('drop', onDrop as EventListener);
      bound.removeEventListener('dragover', onDragOver as EventListener);
    }
    if (ta) {
      ta.addEventListener('paste', onPaste as EventListener);
      ta.addEventListener('drop', onDrop as EventListener);
      ta.addEventListener('dragover', onDragOver as EventListener);
    }
    bound = ta;
  }

  const toolbarButtons = [
    { key: 'bold', icon: 'textformat-bold', title: 'Đậm', action: () => wrapSelectionSimple('**', '**', 'đậm') },
    { key: 'italic', icon: 'textformat-italic', title: 'Nghiêng', action: () => wrapSelectionSimple('*', '*', 'nghiêng') },
    { key: 'h2', icon: 'numbers-2', title: 'Tiêu đề', action: () => insertAtCursor('\n## Tiêu đề\n') },
    { key: 'ul', icon: 'view-list', title: 'Danh sách', action: () => insertAtCursor('\n- ') },
    { key: 'chess', icon: 'grid', title: 'Khối bàn cờ (FEN/PGN)', action: insertChessBlock },
  ];

  return {
    insertAtCursor, wrapSelectionSimple, insertChessBlock,
    uploadAndInsert, onPaste, onDrop, onDragOver, bindTextarea,
    toolbarButtons,
  };
}
