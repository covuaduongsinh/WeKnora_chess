# Demo & giải thích: cải thiện chất lượng Wikilink cờ vua

Tài liệu này hướng dẫn xem từng cải tiến wikilink cờ (4 pha) cùng **giải thích lý do**.

## Chạy dữ liệu demo

Cần Node ≥ 18 và **API key của tenant** (lấy ở UI: Cài đặt → API key — không đọc
được từ DB vì lưu mã hoá). KB wiki là tuỳ chọn (để xem backlink/đồ thị).

```bash
BASE_URL=https://weknora.covuaduongsinh.com \
API_KEY='<tenant api key>' \
WIKI_KB_ID='<id KB bật wiki — tuỳ chọn>' \
node scripts/seed_chess_wikilink_demo.mjs
```

Script in ra danh sách "cách xem từng trường hợp" sau khi tạo xong.

---

## Pha 1 — Soạn thảo dễ hơn

### 1.1 Autocomplete khi gõ `[[`
**Xem:** Quản lý cờ vua → Khóa học → sửa một bài giảng → trong ô nội dung **gõ `[[`** →
hiện dropdown gợi ý ván/thế cờ/bài/khóa; chọn → tự chèn `[[game/<slug>|Tiêu đề]]`.
Cũng hoạt động trong trình soạn trang wiki (Knowledge thủ công).

**Giải thích:** slug được sinh tự động và bất biến nên người viết *không thể đoán*.
Trước đây phải nhớ slug hoặc mở picker. Autocomplete tra cứu sống qua
`GET /api/v1/chess/refs/search?q=` và chèn đúng `slug|nhãn`, loại bỏ nguồn sai số 1
khi tạo link. Dropdown bám theo con trỏ (kỹ thuật "mirror-div") và teleport ra ngoài
để không bị hộp thoại cắt.

### 1.2 Picker có xem trước bàn cờ
**Xem:** trong dialog bài giảng → "Chèn ván/thế cờ" → di chuột vào một mục → **bàn cờ
hiện ở khung phải**.

**Giải thích:** chọn đúng ván/thế cờ trực quan hơn khi thấy bàn cờ thay vì chỉ slug.

---

## Pha 2 — Link bền vững (xem ở "Bài 3 — Demo độ bền link")

### 2.1 Fuzzy tự nắn slug gần đúng
**Xem:** Bài 3, dòng (1) — chip có slug **thiếu một dấu gạch nối** nhưng bấm vẫn mở
đúng ván Opera.

**Giải thích:** khi không khớp slug chính xác, backend thử `exact → alias → fuzzy`.
Fuzzy tái dùng đúng thuật toán đã dùng cho trang wiki (`resolveDeadSlug`): bỏ gạch
nối/hoa-thường rồi so khớp, và bigram-Jaccard ≥ 0.8 cho lỗi gõ nhẹ. Nhờ vậy link do
người/LLM gõ sai một chút **không bị gãy**.

### 2.2 Gợi ý "Ý bạn là…?"
**Xem:** Bài 3, dòng (2) — chip chỉ gõ mỗi `morphy` → bấm hiện "không tìm thấy" **kèm
gợi ý ván Paul Morphy** để chọn (bấm là đổi sang đúng ván).

**Giải thích:** khi slug quá khác để fuzzy tự nắn, ta vẫn giúp người dùng bằng cách
tìm theo từ khoá và đề xuất ứng viên gần nhất, thay vì báo lỗi cụt.

### 2.3 Link gãy hẳn + nút "Tạo mới"
**Xem:** Bài 3, dòng (3) — slug không liên quan → hiện "không tìm thấy" + nút **Tạo
mới** (điều hướng tới khu quản lý đúng loại).

**Giải thích:** biến ngõ cụt thành hành động — tạo nhanh đối tượng còn thiếu.

### 2.4 Alias/redirect khi đổi slug
**Giải thích (hạ tầng):** bảng `chess_slug_aliases` (migration 000068) cho phép map
`slug cũ → slug mới`; khâu resolve kiểm tra alias *trước* fuzzy. Hiện chưa có nút đổi
tên nên bảng đang trống — đây là nền cho tính năng đổi tên/re-import sau này để link
cũ vẫn sống. (Chưa có demo trực quan.)

---

## Pha 3 — AI/RAG hiểu wikilink (⚠️ mặc định TẮT)

Gate sau biến môi trường **`CHESS_KB_INDEX`** (mặc định tắt) vì cần model embedding +
vector store + worker — phải kiểm thử trên full stack trước khi bật thật.

**Bật & xem:**
1. Đặt `CHESS_KB_INDEX=true` cho service `app`, đảm bảo tenant đã có ≥1 KB cấu hình
   embedding (KB cờ sẽ **sao chép** cấu hình đó).
2. Tạo mới hoặc sửa một ván/thế cờ/bài giảng.
3. KB **"Tri thức cờ vua"** tự sinh, chứa bản ghi tri thức của đối tượng đó.
4. Gắn KB này vào agent HLV → hỏi về ván vừa tạo → câu trả lời truy hồi được nội dung.

**Giải thích:**
- **Part A — expand wikilink khi embed:** trước khi đưa nội dung đi nhúng vector, thay
  `[[game/slug|Nhãn]]`/`![[…]]` bằng *văn bản có nghĩa* (nhãn/tiêu đề), **giữ nguyên
  raw để hiển thị**. Vector nắm "Paul Morphy – Duke Karl, Opera 1858" thay vì chuỗi
  slug nhiễu.
- **Part B — index đối tượng cờ:** ván/thế cờ/bài giảng vốn KHÔNG nằm trong KB nên HLV
  không truy hồi được. Indexer tạo bản ghi manual-knowledge (tái dùng toàn bộ pipeline
  chunk + embedding sẵn có) và lưu ánh xạ ở `chess_kb_index` (migration 000069). Chạy
  **best-effort, không bao giờ chặn** thao tác tạo/sửa cờ; import PGN hàng loạt **không**
  trigger để tránh "bão" embedding.

---

## Pha 4 — Đồ thị & backlink

### 4.1 Empty-state backlink
**Xem:** Kho ván → chọn **"Aronian – Anand"** (ván mồ côi) → khung chi tiết hiện
"Chưa có trang/bài giảng nào tham chiếu."

**Giải thích:** trước đây mục backlink ẩn hẳn khi rỗng → khó biết "thật sự chưa có" hay
"lỗi tải". Empty-state làm rõ trạng thái.

### 4.2 Node cờ trong đồ thị wiki
**Xem:** KB wiki → Wiki → Graph → node ván/thế cờ có **màu riêng theo loại**, bấm vào
mở popup bàn cờ.

**Giải thích:** đồ thị tri thức nay thể hiện cả mạng tham chiếu cờ, không chỉ trang wiki.
(Phần tô màu + click đã có sẵn; đợt này bổ sung nhãn loại cờ cho tag/legend.)
