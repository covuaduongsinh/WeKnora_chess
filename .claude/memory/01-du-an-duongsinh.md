# 01 — Dự án Dương Sinh (bối cảnh nền)

## Công ty & người chủ trì
- **Công ty CP Cờ vua Dương Sinh** — slogan **"Vui trí tuệ"**.
- **Phùng Đức Tường** ("Thầy Tường") — Nhà sáng lập & CEO; HLV & trọng tài cờ vua quốc gia; tác giả sách cờ vua; chủ nhiệm CLB Cờ vua Dương Sinh.

## Triết lý cốt lõi (chi phối mọi quyết định sản phẩm)
> Dùng **cờ vua làm công cụ giáo dục**, phát triển **tư duy – nhân cách** cho trẻ em.
> **Ưu tiên phong trào hơn thành tích.**

Khi thiết kế tính năng/nội dung: nếu phải chọn, ưu tiên phục vụ **học viên nhỏ tuổi, phụ huynh, HLV phong trào** hơn đấu thủ đỉnh cao.

## Hệ sinh thái Dương Sinh
CLB & đào tạo; hợp tác trường học; tổ chức giải đấu; xuất bản sách + hiệu sách cờ vua; website/app học trực tuyến.

→ **WeKnora_chess** là **lớp tri thức + công cụ phân tích cờ** bắc qua hệ sinh thái: tra cứu, hỏi-đáp, phân tích thế cờ, quản lý khóa học/ván cờ/bài tập, và sinh Wiki cờ vua — cho cả nội bộ (HLV, biên tập sách) lẫn người dùng cuối (phụ huynh, học viên).

## Lộ trình đào tạo 6 cấp (xương sống nội dung)
Khung phân loại để tổ chức knowledge base, tag tài liệu, phân tầng câu trả lời theo trình độ:

| Cấp | Tên | Ý nghĩa (gợi ý) |
|---|---|---|
| 1 | **Tốt** (Pawn) | Vỡ lòng: bàn cờ, quân, nước đi, luật cơ bản |
| 2 | **Mã** (Knight) | Nước đi đặc biệt, ăn quân, chiếu, chiến thuật sơ cấp |
| 3 | **Tượng** (Bishop) | Giá trị quân, chiến thuật (fork/pin/skewer), khai cuộc cơ bản |
| 4 | **Xe** (Rook) | Tàn cuộc cơ bản, kế hoạch, khai cuộc mở rộng |
| 5 | **Hậu** (Queen) | Chiến lược, trung cuộc, phối hợp |
| 6 | **Vua** (King) | Nâng cao, hệ thống khai cuộc, tàn cuộc lý thuyết, thi đấu |

> Mỗi bài giảng / ván / bài tập / trang Wiki nên gắn tag cấp độ (Tốt…Vua) để agent trả lời đúng độ sâu người hỏi.

## Nhận diện thương hiệu (áp cho mọi tùy biến UI/ấn phẩm)
> **Nguồn chuẩn:** package `@ds/brand` tại `C:\Users\duongsinh\Documents\code\covuaduongsinh\packages\brand` (theme.css token, logo SVG, pattern, fonts). Lấy màu/asset từ đây, đừng đoán.
- **Màu chủ đạo:** `#2B3990` (navy) + tông **xanh** bổ trợ (teal `#3dbb95`, blue `#2275b4`). ⚠️ **KHÔNG dùng cam/gold** cho nhận diện — amber/đỏ chỉ biểu thị trạng thái (warning/error). *(Đính chính: ghi chú "gold" trước đây là SAI so với brand guide.)*
- **Font:** Roboto / Calibri.
- **Họa tiết:** ô vuông bàn cờ (checkerboard), logo Dương Sinh (`logo-symbol.svg` = biểu tượng ô cờ, `logo-full.svg` = logo đầy đủ).
- **Đã áp vào frontend (WS4a):** `frontend/src/assets/theme/duongsinh-brand.css` đè thang `--td-brand-color-*` sang navy (import sau theme.css ở `main.ts`/`embed-main.ts`); favicon + title ở `index.html`; logo copy vào `frontend/public/duongsinh-*.svg`.

## Bối cảnh kỹ thuật của Thầy Tường
- Soạn sách cờ vua trong **Obsidian** (vault OBSIDIAN2026), tích hợp AI. Nguồn giáo trình gốc **tiếng Nga**, đang **Việt hóa** theo cấu trúc bài giảng có sẵn.
- Công cụ: **Claude Desktop (MCP)**, **Claude Code CLI**, **Node.js**; quen `.md`, `.docx` (thư viện `docx` Node), quy trình **PDF→Markdown**.
- → Tính năng **Chess Wikilink** (`[[game/<slug>|Nhãn]]`) trong repo này khớp đúng workflow Obsidian: liên kết chéo giữa bài giảng và ván/thế cờ/bài tập. Nguồn nạp chủ yếu là Markdown Việt hóa + PDF sách cờ → pipeline cần khỏe với tiếng Việt có dấu và ký hiệu cờ.

## Hai hệ thống tri thức của Dương Sinh (KHÔNG trộn)
| | **WeKnora_chess** (repo này) | **Arkon** |
|---|---|---|
| Stack | Go + Vue | Python/FastAPI + Gemini |
| Production | `weknora.covuaduongsinh.com` | `app.covuaduongsinh.com` |
| Repo | `WeKnora_chess` (fork Tencent/WeKnora) | `arkon_duongsinh` |
| Mục đích | Chuyên cờ vua: phân tích + tri thức + khóa học | KB tổng quát self-hosted |
| Vận hành | tài liệu/skill riêng | skill `arkon-deploy` |

Hai hệ chạy song song. Tuyệt đối **không** áp lệnh deploy/schema/cấu trúc của hệ này sang hệ kia.
