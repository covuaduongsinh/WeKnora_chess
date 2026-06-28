// Tiện ích tải xuống / chọn file văn bản phía trình duyệt, dùng cho các chức năng
// import/export cờ vua (khóa học JSON, ván đấu PGN, bài tập JSON...).

// downloadText tạo 1 file văn bản và kích hoạt tải xuống trong trình duyệt.
export function downloadText(filename: string, text: string, mime = "text/plain;charset=utf-8"): void {
  const blob = new Blob([text], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

// pickTextFile mở hộp thoại chọn 1 file và đọc nội dung văn bản.
// Trả về null nếu người dùng hủy hoặc đọc lỗi.
export function pickTextFile(accept: string): Promise<string | null> {
  return new Promise((resolve) => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = accept;
    input.onchange = () => {
      const file = input.files && input.files[0];
      if (!file) {
        resolve(null);
        return;
      }
      const reader = new FileReader();
      reader.onload = () => resolve(String(reader.result || ""));
      reader.onerror = () => resolve(null);
      reader.readAsText(file);
    };
    input.click();
  });
}
