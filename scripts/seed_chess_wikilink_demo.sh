#!/usr/bin/env bash
#
# seed_chess_wikilink_demo.sh — Tạo dữ liệu DEMO cho tính năng "wikilink cho thế
# cờ / ván cờ / bài giảng". Tạo ván cờ + thế cờ + khóa học/bài giảng (và tùy chọn
# một trang wiki) có dùng [[game/<slug>]] (chip) và ![[game/<slug>]] (nhúng), bao
# phủ MỌI trường hợp để bạn xem trực tiếp trên UI.
#
# YÊU CẦU: backend ĐÃ rebuild + chạy migration 000064 (slug) và 000065
#          (wiki_chess_refs). Script bắt slug do server tự sinh nên link luôn khớp.
#
# XÁC THỰC — chọn MỘT trong hai:
#   API_KEY : khoá API của tenant (ổn định, không hết hạn). Lấy ở UI phần Cài đặt
#             > API (hoặc nơi hiển thị "API Key"). Gửi qua header X-API-Key.
#   TOKEN   : JWT từ DevTools > Console: localStorage.weknora_token
#             (kèm TENANT_ID nếu cần: localStorage.weknora_selected_tenant_id)
#
# CHẠY (ví dụ với API key):
#     BASE_URL=http://localhost \
#     API_KEY='<tenant api key>' \
#     WIKI_KB_ID='<tuỳ chọn — id KB đã bật wiki, để demo backlink + đồ thị>' \
#     bash scripts/seed_chess_wikilink_demo.sh
#
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost}"
API_KEY="${API_KEY:-}"
TOKEN="${TOKEN:-}"
TENANT_ID="${TENANT_ID:-}"
WIKI_KB_ID="${WIKI_KB_ID:-}"

if [[ -z "$API_KEY" && -z "$TOKEN" ]]; then
  echo "ERROR: cần API_KEY hoặc TOKEN. Xem hướng dẫn ở đầu file." >&2
  exit 1
fi

# --- helper gọi API + trích JSON (dùng node, đã có sẵn) ---------------------
api() {                      # api METHOD PATH [JSON_BODY]
  local method="$1" path="$2" body="${3:-}"
  local args=(-sS -X "$method" "$BASE_URL$path" -H "Content-Type: application/json")
  if [[ -n "$API_KEY" ]]; then
    args+=(-H "X-API-Key: $API_KEY")            # khoá API map sẵn tenant
  else
    args+=(-H "Authorization: Bearer $TOKEN")
    [[ -n "$TENANT_ID" ]] && args+=(-H "X-Tenant-ID: $TENANT_ID")
  fi
  [[ -n "$body" ]] && args+=(--data "$body")
  curl "${args[@]}"
}
jget() {                     # đọc JSON từ stdin, in giá trị field (vd .data.slug)
  node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{try{const o=JSON.parse(s);const p=process.argv[1].split(".").filter(Boolean);let v=o;for(const k of p)v=v?.[k];process.stdout.write(v==null?"":String(v));}catch(e){process.stdout.write("")}})' "$1"
}
say() { printf '\n\033[1;36m== %s ==\033[0m\n' "$1"; }

# --- 1) Ván cờ (Kho ván đấu) -------------------------------------------------
say "Tạo ván cờ demo"
G1=$(api POST /api/v1/chess/games '{
  "white":"Paul Morphy","black":"Duke Karl / Count Isouard","result":"1-0",
  "eco":"C41","event":"Paris Opera","date":"1858.06.21",
  "pgn":"1.e4 e5 2.Nf3 d6 3.d4 Bg4 4.dxe5 Bxf3 5.Qxf3 dxe5 6.Bc4 Nf6 7.Qb3 Qe7 8.Nc3 c6 9.Bg5 b5 10.Nxb5 cxb5 11.Bxb5+ Nbd7 12.O-O-O Rd8 13.Rxd7 Rxd7 14.Rd1 Qe6 15.Bxd7+ Nxd7 16.Qb8+ Nxb8 17.Rd8# 1-0"
}')
G1_SLUG=$(printf '%s' "$G1" | jget .data.slug)

G2=$(api POST /api/v1/chess/games '{
  "white":"Học trò","black":"Tập sự","result":"1-0","event":"Ván mẫu",
  "pgn":"1.e4 e5 2.Bc4 Nc6 3.Qh5 Nf6 4.Qxf7# 1-0"
}')
G2_SLUG=$(printf '%s' "$G2" | jget .data.slug)
echo "  game/$G1_SLUG"
echo "  game/$G2_SLUG"

# --- 2) Thế cờ / Bài tập (Ngân hàng bài tập) ---------------------------------
say "Tạo thế cờ / bài tập demo"
P1=$(api POST /api/v1/chess/puzzles '{
  "title":"Chiếu bí hàng ngang","fen":"6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1",
  "solution":"Ra8#","theme":"chiếu hết","difficulty":"de"
}')
P1_SLUG=$(printf '%s' "$P1" | jget .data.slug)

P2=$(api POST /api/v1/chess/puzzles '{
  "title":"Thế cờ khai cuộc Ý","fen":"r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
  "solution":"Ng5","theme":"khai cuộc","difficulty":"trung-binh"
}')
P2_SLUG=$(printf '%s' "$P2" | jget .data.slug)
echo "  puzzle/$P1_SLUG"
echo "  puzzle/$P2_SLUG"

# --- 3) Khóa học + bài giảng (dùng chip + nhúng) -----------------------------
say "Tạo khóa học + bài giảng demo"
COURSE=$(api POST /api/v1/chess/courses '{
  "title":"DEMO — Wikilink cờ vua","description":"Minh hoạ chip & nhúng cho ván/thế cờ/bài giảng","level":"co-ban"
}')
COURSE_ID=$(printf '%s' "$COURSE" | jget .data.id)

# Bài 1: minh hoạ CHIP (bấm mở popup) và NHÚNG (bàn cờ inline).
L1_CONTENT=$(cat <<EOF
## Chip vs Nhúng

**Chip** (bấm để mở popup bàn cờ): ván nổi tiếng [[game/$G1_SLUG|Ván Opera 1858]] và thế cờ [[puzzle/$P1_SLUG|Chiếu bí hàng ngang]].

**Nhúng** (bàn cờ tương tác ngay trong bài) — thêm dấu \`!\` phía trước:

![[game/$G1_SLUG]]

Một thế cờ chiến thuật để luyện:

![[puzzle/$P2_SLUG]]
EOF
)
L1=$(api POST "/api/v1/chess/courses/$COURSE_ID/lessons" \
  "$(node -e 'const c=process.argv[1];process.stdout.write(JSON.stringify({title:"Bài 1 — Chip vs Nhúng",content:c,sort_order:0}))' "$L1_CONTENT")")
L1_SLUG=$(printf '%s' "$L1" | jget .data.slug)

# Bài 2: minh hoạ LIÊN KẾT BÀI GIẢNG (lesson -> lesson).
L2_CONTENT=$(cat <<EOF
## Liên kết bài giảng

Xem lại [[lesson/$L1_SLUG|Bài 1 — Chip vs Nhúng]] trước khi làm bài tập.

Ván cờ chiếu bí nhanh: ![[game/$G2_SLUG]]
EOF
)
api POST "/api/v1/chess/courses/$COURSE_ID/lessons" \
  "$(node -e 'const c=process.argv[1];process.stdout.write(JSON.stringify({title:"Bài 2 — Liên kết bài giảng",content:c,sort_order:1}))' "$L2_CONTENT")" >/dev/null
echo "  course id=$COURSE_ID  (Bài 1 = lesson/$L1_SLUG)"

# --- 4) (Tùy chọn) Trang wiki tham chiếu cờ → demo backlink + đồ thị ---------
if [[ -n "$WIKI_KB_ID" ]]; then
  say "Tạo trang wiki demo (backlink + đồ thị) trong KB $WIKI_KB_ID"
  WIKI_CONTENT=$(cat <<EOF
# Khai cuộc & chiến thuật (demo)

Trang wiki này tham chiếu trực tiếp tới đối tượng cờ:

- Ván minh hoạ: [[game/$G1_SLUG|Ván Opera]]
- Bài tập: [[puzzle/$P1_SLUG]]
- Bài giảng liên quan: [[lesson/$L1_SLUG]]

Bàn cờ nhúng ngay trong trang wiki:

![[game/$G1_SLUG]]
EOF
)
  api POST "/api/v1/knowledgebase/$WIKI_KB_ID/wiki/pages" \
    "$(node -e 'const c=process.argv[1];process.stdout.write(JSON.stringify({slug:"concept/demo-co-vua",title:"Demo cờ vua (wikilink)",page_type:"concept",content:c,status:"published"}))' "$WIKI_CONTENT")" >/dev/null
  echo "  wiki page: concept/demo-co-vua"
fi

# --- Tổng kết: xem ở đâu -----------------------------------------------------
say "XONG — cách xem từng trường hợp"
cat <<EOF
1. Chip & Nhúng (bài giảng):  http://localhost → Quản lý cờ vua → tab Khóa học
     → "DEMO — Wikilink cờ vua" → mở "Bài 1": chip bấm mở popup, ![[..]] hiện bàn cờ inline.
2. Liên kết bài giảng:        mở "Bài 2" → chip [[lesson/..]] mở popup bài giảng.
3. Sao chép wikilink:         tab Kho ván / Ngân hàng bài tập → nút 🔗 mỗi dòng.
4. Bộ chọn chèn:              sửa một bài giảng → nút "Chèn ván/thế cờ".
EOF
if [[ -n "$WIKI_KB_ID" ]]; then
cat <<EOF
5. Trong trang wiki:          mở KB → tab Wiki → trang "Demo cờ vua (wikilink)":
     chip + bàn cờ nhúng. Popup chip hiện "Được liên kết bởi".
6. Đồ thị:                    KB → Wiki → Graph: node cờ (màu riêng) nối từ trang,
     bấm node mở bàn cờ.
EOF
else
echo "(Đặt WIKI_KB_ID=<id KB bật wiki> để seed thêm trang wiki demo + backlink + đồ thị.)"
fi
