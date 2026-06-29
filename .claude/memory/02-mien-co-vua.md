# 02 — Miền tri thức cờ vua (chess domain)

Giúp AI agent hiểu **cấu trúc tri thức cờ vua** để tổ chức nội dung, đặt tag, viết prompt, và xử lý ký hiệu cờ đúng cách. (Hạ tầng kỹ thuật của lớp cờ xem `03-kien-truc-weknora.md`.)

## 2.1. Nhóm nội dung (dùng làm knowledge base / tag)
1. **Luật cờ (Rules)** — FIDE + luật VN, cờ nhanh/chớp, xử lý trọng tài.
2. **Khai cuộc (Openings)** — lý thuyết, bẫy, theo ECO; FEN/PGN minh họa. *(Có tool `chess_lookup_opening`.)*
3. **Trung cuộc (Middlegame)** — chiến lược, kế hoạch, cấu trúc tốt.
4. **Tàn cuộc (Endgame)** — cơ bản → lý thuyết.
5. **Chiến thuật (Tactics)** — fork/pin/skewer/discovered/mate; kho bài tập. *(Có Puzzle Bank + `chess_generate_puzzle`.)*
6. **Giáo trình & giáo án (Curriculum)** — bài giảng 6 cấp Tốt→Vua, giáo án, worksheet, flashcard. *(Có Chess Courses.)*
7. **Văn hóa & lịch sử (Culture/History)** — kỳ thủ, ván kinh điển, chuyện truyền cảm hứng. *(Có Game Library.)*
8. **Vận hành Dương Sinh (Ops)** — quy trình CLB, hợp tác trường, giải đấu (KB riêng, quyền hạn chế).

## 2.2. Phân tầng độ sâu theo 6 cấp
Gắn cấp **Tốt / Mã / Tượng / Xe / Hậu / Vua** cho mỗi mục để agent điều chỉnh độ sâu:
- Phụ huynh / cấp Tốt–Mã → đơn giản, ví dụ gần gũi, ít thuật ngữ, động viên.
- HLV / cấp Hậu–Vua → cho phép thuật ngữ chuyên sâu, biến (variation), đánh giá thế cờ.

## 2.3. Thuật ngữ song ngữ chuẩn (Việt – Anh)
| Tiếng Việt | English | | Tiếng Việt | English |
|---|---|---|---|---|
| Khai cuộc | Opening | | Ghim | Pin |
| Trung cuộc | Middlegame | | Xiên | Skewer |
| Tàn cuộc | Endgame | | Bắt đôi | Fork |
| Chiến thuật | Tactics | | Đòn mở | Discovered attack |
| Chiến lược | Strategy | | Thí quân | Sacrifice |
| Nước đi | Move | | Cấu trúc tốt | Pawn structure |
| Bắt/ăn quân | Capture | | Tốt thông | Passed pawn |
| Chiếu | Check | | Tốt cô lập | Isolated pawn |
| Chiếu hết | Checkmate / Mate | | Tốt chồng | Doubled pawns |
| Hòa | Draw | | Ưu thế | Advantage |
| Hết nước (bí) | Stalemate | | Thế cờ | Position |
| Nhập thành | Castling (O-O/O-O-O) | | Biến (nhánh) | Variation / Line |
| Bắt tốt qua đường | En passant | | Hệ thống | System |
| Phong cấp | Promotion | | Bẫy khai cuộc | Opening trap |

**Quân cờ:** Vua=King(K), Hậu=Queen(Q), Xe=Rook(R), Tượng=Bishop(B), Mã=Knight(N), Tốt=Pawn(P).

## 2.4. Ký hiệu kỹ thuật — đã có hạ tầng xử lý
Lớp cờ trong repo đã hỗ trợ sẵn, nên KHI tạo nội dung hãy giữ ký hiệu raw để hệ thống nhận diện:
- **FEN** — chuỗi mô tả 1 thế cờ. **Không tách dòng, không dịch.**
- **PGN** — toàn bộ ván + metadata. Dùng cho `chess_evaluate_game` (chấm ván).
- **SAN (nước đi):** `Nf3`, `O-O`, `exd5`, `Qxh7#`, `e8=Q`; đánh giá `!` `?` `!?` `?!` `+` `#`.
- **Khối ```chess** — chèn FEN hoặc PGN trong khối mã ` ```chess ` để frontend tự render **bàn cờ tương tác** (xem `ChessBoardDisplay.vue`, `utils/chessBlocks.ts`). Đây là cách minh họa thế cờ không cần gọi tool.
- **Chess Wikilink** — `[[game/<slug>|Nhãn]]`, `![[…]]` để nhúng. Slug **bất biến**, sinh tự động; có autocomplete khi gõ `[[`. (Xem `02.6` và `docs/chess-wikilink-demo.md`.)

## 2.5. Ontology cho Wiki Mode / Knowledge Graph
Các loại trang để định hướng Wiki:
- **Concept** — "Ghim", "Tốt thông", "Nhập thành"…
- **Opening** — theo tên + ECO, liên kết biến con.
- **Endgame** — "Vua+Xe đấu Vua", "Vua+Tốt đấu Vua"…
- **Lesson** — bài giảng, gắn cấp 6 bậc.
- **Game / Player** — ván kinh điển, kỳ thủ (phần văn hóa).
Liên kết: *bài giảng* → dạy → *concept* → minh họa bằng → *ván/bài tập*. Node cờ hiển thị trong đồ thị wiki (màu theo loại, click mở bàn cờ).

## 2.6. Thực thể cờ & slug (đặc thù repo này)
Bốn loại thực thể có slug bất biến, tham chiếu qua wikilink: **game** (ván), **position** (thế cờ), **lesson** (bài giảng), **course** (khóa học) — cộng **puzzle** (bài tập).
- Resolve link theo thứ tự **exact → alias → fuzzy** (bigram-Jaccard ≥ 0.8) nên link gõ sai nhẹ vẫn sống.
- Đổi slug dùng bảng alias (`chess_slug_aliases`) để link cũ không gãy.
- Khi viết bài giảng/nội dung: ưu tiên dùng wikilink thay vì chép cứng tên ván → giữ liên kết chéo và backlink.
