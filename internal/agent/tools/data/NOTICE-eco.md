# Dữ liệu khai cuộc (ECO) nhúng — `eco.tsv`

- **Nguồn:** [lichess-org/chess-openings](https://github.com/lichess-org/chess-openings)
- **Giấy phép:** CC0-1.0 (Public Domain Dedication) — tự do dùng, không cần ghi công, nhưng ở đây vẫn ghi nguồn cho minh bạch.
- **Định dạng:** TSV 3 cột `eco<TAB>name<TAB>pgn` (pgn là chuỗi SAN có số nước, ví dụ `1. e4 c5 2. Nf3`).
- **Dùng ở đâu:** nạp một lần lúc khởi động bởi `chess_openings_data.go` (`//go:embed`) để tool `chess_lookup_opening` nhận diện hàng nghìn khai cuộc/biến.
- **Quan hệ với bảng Việt hoá:** map `ecoOpenings` trong `chess_lookup_opening.go` được **overlay đè lên** dataset này → khai cuộc phổ biến hiển thị tên tiếng Việt, phần còn lại dùng tên tiếng Anh từ dataset.

## Cập nhật dataset (khi cần)
```bash
out=internal/agent/tools/data/eco.tsv
printf 'eco\tname\tpgn\n' > "$out"
for f in a b c d e; do
  curl -sL "https://raw.githubusercontent.com/lichess-org/chess-openings/master/$f.tsv" | tail -n +2 >> "$out"
done
```
