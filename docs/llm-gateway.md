# Dùng gói thuê bao thay API cho WeKnora — runbook

> **Mục tiêu:** đưa chi phí LLM về gần 0 bằng cách tận dụng free tier hợp lệ và các gói đã trả tiền, thay vì trả thêm theo token.
> **Phạm vi đã chốt:** cá nhân Thầy Tường dùng · chỉ đi đường chính thức, không proxy dịch ngược · VPS Hetzner chỉ có CPU.

---

## 0. Gói nào thật sự dùng được (tính tới 8/2026)

Năm 2026 các hãng đã **bịt gần hết** đường "lấy gói thuê bao thay API". Hiện trạng thật:

| Gói | Đường chính thức cho ứng dụng? | Kết luận |
|---|---|---|
| **Google AI / Gemini** | **Gemini API free tier**: Flash ~1.500 request/ngày, 10 RPM, 250k TPM — miễn phí, có cả embedding | ✅ **Nền tảng chính** |
| **Claude Pro/Max** | **Agent SDK credit** (từ 15/6/2026): túi riêng **$20 (Pro) / $100 (Max 5x) / $200 (Max 20x)** mỗi tháng, tính theo giá API, **không** trừ vào hạn mức Claude Code tương tác, **không cộng dồn** sang tháng sau | ✅ **Tầng việc nặng** (Giai đoạn 3) |
| **ChatGPT Plus/Pro** | `codex exec` headless là chính thức, nhưng chỉ phủ phiên đăng nhập ChatGPT trên máy — **không có túi credit cho ứng dụng** | ⚠️ Không dùng (vùng xám) |
| **Grok / SuperGrok** | xAI tách bạch hoàn toàn — SuperGrok **không kèm** API credit nào | ❌ Không dùng được |
| **GitHub Copilot** | Không có API công khai; mọi proxy đều là dịch ngược | ❌ Loại |

**Hai điều phải biết trước:**

1. **Không gói CLI nào có endpoint embedding.** RAG cờ sống bằng embedding, nên chỗ này phải giải riêng. May là embedding rẻ đến mức không đáng bàn — index cả thư viện sách chỉ tốn vài cent.
2. **Với quy mô một người dùng, thứ tiết kiệm nhất không phải là bắc cầu subscription** mà là **Giai đoạn 1 dưới đây** — cấu hình lại model theo tác vụ. Nó không cần sửa một dòng code nào và đã giải quyết phần lớn vấn đề.

---

## 1. Giai đoạn 1 — làm ngay, không cần bật cổng LLM

Toàn bộ mục này làm trên **giao diện web + Google AI Studio**. Không build lại gì.

### 1.1. Kiểm kê trước, đừng tối ưu mù

- Vào **Cài đặt → Model**, ghi lại: model chat đang dùng là gì, `base_url` nào, model embedding là gì.
- Vào Google AI Studio kiểm tra key Gemini đang dùng thuộc **free tier** hay đã bật billing.

> Nếu key vốn đã ở free tier thì chi phí hiện tại đã gần 0 — Giai đoạn 2/3 chỉ còn là chuyện **chất lượng**, không phải tiết kiệm. Biết điều này trước sẽ đỡ làm thừa.

### 1.2. Tạo 3 model theo tầng tác vụ

Trong **Cài đặt → Model → tab "chat" → nút +** (cần quyền admin), tạo ba model loại `KnowledgeQA`:

| Tên | Dùng cho | Vì sao tách |
|---|---|---|
| `ds-fast` | viết lại truy vấn, nhận diện ý định, đặt tiêu đề hội thoại | Nhóm tốn **nhiều lượt gọi nhất** nhưng đòi hỏi chất lượng thấp nhất |
| `ds-chat` | trả lời RAG, agent "HLV Cờ vua" | Xương sống. **Bắt buộc** model hỗ trợ tool calling tốt (agent cờ gọi 7 tool) |
| `ds-smart` | Wiki synthesis, soạn nội dung dài | Ít lượt, cần chất lượng cao |

Không có giới hạn số model cùng loại trong một tenant.

### 1.3. Gán model **tường minh** — bước này không phải tùy chọn

- **Agent "HLV Cờ vua"**: đặt `model_id` = `ds-chat`, `query_understand_model_id` = `ds-fast`.
- **KB "Tri thức cờ vua"**: `summary_model_id` = `ds-chat`, model Wiki synthesis = `ds-smart`.
- **Giữ nguyên model embedding.**

> ⚠️ **Vì sao bắt buộc:** xem cạm bẫy số 1 ở mục 4. Tóm tắt: cột `is_default` **không hề được đọc lúc chạy** — mọi luồng chưa gán tường minh sẽ chọn "model KnowledgeQA đầu tiên theo thứ tự trong CSDL", tức có thể rơi trúng `ds-smart` là model đắt nhất.

### 1.4. (Tùy chọn) Bịt lỗ rerank

Tenant hiện **chưa có rerank model**, nên `knowledge_search` rơi về rerank bằng LLM — đốt token chat mỗi lần tìm kiếm. Thêm một rerank model free tier vừa rẻ hơn vừa cho kết quả tốt hơn.

---

## 2. Giai đoạn 2 — bật cổng LLM (LiteLLM)

Chỉ cần khi muốn: **fallback tự động lúc hết hạn ngạch**, **trần ngân sách**, và **một điểm vào duy nhất** để sau này cắm Claude vào.

Cổng là **opt-in**. Không bật thì stack chạy y hệt như trước.

### 2.1. Chuẩn bị `.env` trên server

```bash
# Sinh khóa bảo vệ cổng
openssl rand -hex 32
```

Thêm vào `.env`:

```dotenv
LLM_GATEWAY_ENABLED=true
LITELLM_MASTER_KEY=<chuỗi vừa sinh>
GEMINI_API_KEY=<key Google AI Studio>
# Tùy chọn — nhánh dự phòng khi Gemini hết hạn ngạch ngày.
# Bỏ trống thì hết quota sẽ lỗi thẳng thay vì âm thầm tiêu tiền.
DEEPSEEK_API_KEY=
```

> **Không** cần tự thêm `llm-gateway` vào `SSRF_WHITELIST_EXTRA` — `docker-compose.llm.yml` đã tự nối vào.

### 2.2. Đối chiếu tên model của Google

Tên model Google đổi theo thời gian. Liệt kê cái đang khả dụng với key của Thầy:

```bash
curl -s "https://generativelanguage.googleapis.com/v1beta/models?key=$GEMINI_API_KEY" \
  | grep -o '"name": "models/[^"]*"'
```

Rồi sửa `config/litellm/config.yaml` cho khớp. File này **mount từ host** → sửa xong chỉ cần `restart`, không rebuild.

### 2.3. Khởi động

```bash
scripts/deploy/dc.sh up -d
scripts/deploy/dc.sh ps          # llm-gateway phải "healthy"
scripts/deploy/dc.sh logs -f llm-gateway
```

Kiểm tra từ bên trong container `app`:

```bash
scripts/deploy/dc.sh exec app sh -c \
  'wget -qO- --header="Authorization: Bearer $LITELLM_MASTER_KEY" http://llm-gateway:4000/v1/models'
```

Phải liệt kê đủ 5 tên: `ds-fast`, `ds-chat`, `ds-chat-backup`, `ds-smart`, `ds-embed`.

### 2.4. Trỏ WeKnora sang cổng

Sửa 3 model đã tạo ở Giai đoạn 1:

| Trường | Giá trị |
|---|---|
| Nguồn | `remote` |
| Provider | **`generic`** (自定义 / OpenAI 兼容接口) |
| Base URL | `http://llm-gateway:4000/v1` |
| Model name | đúng tên alias (`ds-fast` / `ds-chat` / `ds-smart`) |
| API Key | chuỗi `LITELLM_MASTER_KEY` |

Bấm **"Kiểm tra kết nối"** trước khi lưu — nút này gửi một message thật qua đúng pipeline production, không phải chỉ ping URL.

---

## 3. Giai đoạn 3 — cắm túi credit Claude Agent SDK (`claude-bridge`)

**Cân nhắc kinh tế:** túi Agent SDK credit không cộng dồn qua tháng — bỏ không là mất. Vì `ds-smart` ở đây CHỈ dùng cho việc ít lượt (Wiki synthesis, soạn nội dung dài — xem 3.4), ngay cả gói Pro ($20/tháng) có thể đã đủ dùng; Max 5x/20x ($100/$200) thì gần như chắc chắn thừa cho một mình Thầy dùng.

Đường chính thức là **Claude Agent SDK** (`query()`), đúng kênh Anthropic công bố cho "third-party apps that authenticate with your Claude subscription through the Agent SDK". Cầu nối dùng ở đây: **[Meridian](https://github.com/rynfar/meridian)** (MIT, `@rynfar/meridian`) — gọi thẳng `query()` của SDK chính thức, không chạm OAuth token của Claude Code, không dịch ngược backend web.

Overlay đã dựng sẵn: `docker-compose.llm-claude.yml` (service `claude-bridge`, build thẳng từ tag `meridian-v1.62.7` của repo nguồn — xác nhận tag này tồn tại lúc viết tài liệu này). Việc còn lại là 4 bước dưới đây.

### 3.1. Lấy token — chỉ Thầy làm được, không ai làm thay

VPS Hetzner không có trình duyệt, nên **không dùng** `claude login` trực tiếp trên VPS. Dùng đường **OAuth-token profile** — chính hãng khuyến nghị cho "CI runners, ephemeral containers, and cross-host deployments where browser-based login isn't reachable":

Trên **máy có trình duyệt của Thầy** (máy tính cá nhân đã cài sẵn Claude Code CLI chính thức và đăng nhập bằng gói Pro/Max — KHÔNG cần cài `@rynfar/meridian` ở bước này, `setup-token` là lệnh có sẵn của chính `claude` CLI):

```bash
claude setup-token
```

Lệnh này in ra một chuỗi dài dạng `sk-ant-oat01-...`. Đây là token sống lâu (không phải token 8-giờ của phiên `claude login` thường) — **giữ bí mật như một API key thật**, không commit vào git.

### 3.2. Đặt vào `.env` trên VPS

```dotenv
CLAUDE_BRIDGE_ENABLED=true
MERIDIAN_PROFILES=[{"id":"ds","oauthToken":"sk-ant-oat01-...dan-token-that-cua-Thay"}]
MERIDIAN_DEFAULT_PROFILE=ds
```

### 3.3. Build MỘT LẦN rồi khởi động

`claude-bridge` build từ mã nguồn (không có image dựng sẵn đáng tin trên registry công khai), nên **không đi qua** `pull-deploy.sh` (script đó cố ý "không build trên VPS" — sẽ tự chặn với thông báo rõ nếu Thầy quên bước này):

```bash
scripts/deploy/dc.sh build claude-bridge   # một lần, ~1-2 phút, cần Internet ra GitHub+Docker Hub
scripts/deploy/dc.sh up -d
scripts/deploy/dc.sh logs -f claude-bridge
```

Kiểm tên model Meridian thực sự chấp nhận — **đừng đoán** (tài liệu Meridian không công bố rõ, ví dụ dưới cần đối chiếu lại bằng chính lệnh này với token thật):

```bash
scripts/deploy/dc.sh exec app sh -c \
  'wget -qO- http://claude-bridge:3456/v1/models'
```

### 3.4. Đổi khối `ds-smart` trong `config/litellm/config.yaml`

```yaml
  - model_name: ds-smart
    litellm_params:
      model: openai/<ten-model-lay-tu-buoc-3.3>   # ví dụ claude-sonnet-4-6 — XÁC NHẬN LẠI, đừng chép nguyên
      api_base: http://claude-bridge:3456/v1
      api_key: dummy   # Meridian không cần key thật ở chặng này, token thật đã nằm trong MERIDIAN_PROFILES
```

Rồi `scripts/deploy/dc.sh restart llm-gateway`.

### ⚠️ Bốn rủi ro phải biết trước

1. **Đừng giao agent cờ cho `ds-smart`.** Request mang tool từng bị trả `400 You're out of extra usage` ở một số nhóm tài khoản — Meridian ghi nhận phần lớn đã được xử lý từ 7/2026, nhưng "phần lớn" không phải "chắc chắn". Agent "HLV Cờ vua" gọi 7 tool → **giữ nguyên ở `ds-chat` (Gemini)**. `ds-smart` chỉ dùng cho sinh văn bản thuần (Wiki synthesis).
2. **Đừng dùng đường `~/.claude` volume-mount cho VPS headless** (một cách khác Meridian hỗ trợ) — nó cần `claude login` tương tác ngay trên máy chủ (mở trình duyệt), không hợp với Hetzner headless. Đường token ở mục 3.1 mới đúng cho trường hợp này.
3. **`mem_limit: 384m`** trong overlay là ước lượng cho một tiến trình Node đơn giản — nếu container bị OOM-kill (`docker inspect WeKnora-claude-bridge | grep OOMKilled`), nới lên trong `docker-compose.llm-claude.yml`.
4. **Nếu sau này mở phần mềm cho học viên/phụ huynh thì PHẢI gỡ nhánh này** và chuyển `ds-smart` sang API key thật — Anthropic nói rõ automation/production dùng chung nên dùng API key qua Claude Platform, không phải túi credit cá nhân. Kiến trúc đã tách sẵn: chỉ sửa đúng khối `ds-smart` ở mục 3.4, tắt `CLAUDE_BRIDGE_ENABLED`.

### Chưa kiểm chứng được (nói thẳng)

Mục 3.1–3.4 dựng từ tài liệu chính thức của Meridian (README + `docs/deployment.md` + `docs/profiles.md`, đọc trực tiếp lúc viết tài liệu này) và từ việc build thật Dockerfile của họ (xác nhận: build 2 tầng `oven/bun:1` → `node:22-alpine`, user `claude` UID 1000, port `3456`, endpoint `/v1/chat/completions`+`/v1/models` có thật). **Chưa** chạy thử với token thật (không có token của Thầy) nên **chưa xác nhận** được: tên model chính xác Meridian chấp nhận trong trường `model`, và toàn bộ chuỗi hoạt động đầu-cuối qua `ds-smart`. Làm đúng mục 3.3 trước khi tin vào bất kỳ tên model nào.

---

## 4. Ba cạm bẫy của riêng repo này

### Cạm bẫy 1 — `is_default` KHÔNG được đọc lúc chạy (lỗi CÂM)

`internal/types/model.go:126` có cột `is_default`, và giao diện hiện nhãn "Mặc định" — nhưng **không có logic chọn model runtime nào đọc nó**:

- `selectChatModelID` — `internal/application/service/session_knowledge_qa.go:234-311`
- `resolveChatModelID` — `internal/application/service/session_qa_helpers.go:55-77`
- `GenerateTitle` — `internal/application/service/session.go:471-490`

Cả ba đều fallback bằng **"model `KnowledgeQA` đầu tiên trong `ListModels`"**, tức **theo thứ tự CSDL**. `is_default` cũng **không set được qua giao diện** (`CreateModelRequest` không có trường này) — nó chỉ được gán từ `config/builtin_models.yaml` lúc khởi động.

⇒ **Tạo thêm model chat mà quên gán tường minh là hệ thống tự chọn model bất kỳ.** Không có lỗi, không có cảnh báo — chỉ là hóa đơn sai. Bước 1.3 vì thế bắt buộc.

### Cạm bẫy 2 — sửa agent qua giao diện làm `builtin_agents.yaml` mất tác dụng

`builtin-chess-coach` hiện **không khai `model_id`** trong `config/builtin_agents.yaml`. Muốn gán model phải sửa qua giao diện → sinh một bản ghi đè trong bảng `custom_agents` → **từ đó YAML không còn tác dụng** (`GetAgentByID` ưu tiên bản DB).

Đây đúng là lỗi đã từng xảy ra và được ghi trong `.claude/memory/04-nhat-ky-tuy-bien.md`: bản DB thiếu `knowledge_search` → log báo `tool not found: knowledge_search` dù YAML khai đủ.

⇒ **Sau khi gán model, kiểm lại ngay:** `allowed_tools` vẫn đủ `knowledge_search` + 7 tool cờ, và `kb_selection_mode` vẫn là `all`.

### Cạm bẫy 3 — đổi model embedding là phải index lại toàn bộ

Đổi model embedding là đổi số chiều vector **và** cả không gian vector ⇒ mọi vector cũ thành vô nghĩa, phải index lại toàn bộ KB "Tri thức cờ vua" (cả 8 loại nội dung).

`ds-embed` **đã khai sẵn** trong config nhưng **cố ý chưa dùng**. Nếu về sau thật sự cần đổi, làm thành một đợt riêng: backup CSDL → xóa KB → "Đẩy lại index" → xác nhận qua panel **Kho tri thức**, theo `docs/chess-rag-enable.md`.

---

## 5. Bảng chẩn đoán

| Triệu chứng | Nguyên nhân thường gặp | Xử lý |
|---|---|---|
| Lưu model báo lỗi **Base URL** | `llm-gateway` chưa vào whitelist SSRF | Kiểm `docker-compose.llm.yml` có được chồng vào không (`LLM_GATEWAY_ENABLED=true`); xác nhận bằng `scripts/deploy/dc.sh config \| grep SSRF_WHITELIST_EXTRA` |
| **404 model not found** trong log LiteLLM | Tên model Google đã đổi | Chạy lệnh liệt kê ở mục 2.2, sửa `config/litellm/config.yaml`, `restart llm-gateway` |
| **400 Bad Request** cho mọi lời gọi | `drop_params` bị tắt — Gemini không nhận `frequency_penalty` / `presence_penalty` / `parallel_tool_calls` mà WeKnora gửi kèm | Bật lại `litellm_settings.drop_params: true` |
| **429** liên tục | Hết hạn ngạch ngày của free tier | Đợi sang ngày mới, hoặc khai `DEEPSEEK_API_KEY` để có nhánh dự phòng |
| Agent cờ **không gọi được tool** | Model gán cho agent không hỗ trợ function calling | Đổi `model_id` của agent về `ds-chat` |
| `llm-gateway` **unhealthy** | Thiếu `LITELLM_MASTER_KEY`, hoặc YAML sai cú pháp | `scripts/deploy/dc.sh logs llm-gateway` |
| Container **bị OOM-kill** | `mem_limit: 768m` quá chặt | Kiểm `docker inspect WeKnora-llm-gateway \| grep OOMKilled`, nới trong `docker-compose.llm.yml` |
| Hóa đơn vẫn phát sinh | Luồng nào đó chưa gán model tường minh | Cạm bẫy 1 — rà lại toàn bộ agent và KB |
| `pull-deploy.sh` báo lỗi **"image claude-bridge chua tung build"** rồi dừng | `CLAUDE_BRIDGE_ENABLED=true` nhưng chưa build lần đầu — script này cố ý **không** build trên VPS | Chạy một lần `scripts/deploy/dc.sh build claude-bridge` trước khi deploy lại |
| `claude-bridge` không gọi được / 401 | `MERIDIAN_PROFILES` sai định dạng JSON, hoặc token đã bị Thầy thu hồi trên trang quản lý Claude | Kiểm cú pháp JSON (mảng, dấu nháy kép); sinh token mới bằng `claude setup-token` |

---

## 6. Tắt / quay lui

**Chỉ tắt Giai đoạn 3 (giữ Giai đoạn 1+2):**
```dotenv
CLAUDE_BRIDGE_ENABLED=false
```
Nhớ trả khối `ds-smart` trong `config/litellm/config.yaml` về lại Gemini trước khi tắt hẳn `claude-bridge`, tránh Wiki synthesis gọi vào một địa chỉ không còn ai lắng nghe.

**Tắt toàn bộ cổng LLM (cả hai giai đoạn):**
```dotenv
LLM_GATEWAY_ENABLED=false
CLAUDE_BRIDGE_ENABLED=false
```

```bash
scripts/deploy/dc.sh up -d --remove-orphans
```

Các model trong WeKnora vẫn còn nhưng sẽ không gọi được — sửa `base_url` của chúng về thẳng nhà cung cấp (ví dụ `https://generativelanguage.googleapis.com/v1beta/openai`) là chạy lại ngay. Không mất dữ liệu, không đụng vector store.

---

## Nguồn

- [Anthropic — Use the Claude Agent SDK with your Claude plan](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan)
- [Gemini CLI — Quotas and Pricing](https://google-gemini.github.io/gemini-cli/docs/quota-and-pricing.html)
- [Meridian — cầu nối Agent SDK ↔ OpenAI API (MIT)](https://github.com/rynfar/meridian) · [docs/deployment.md](https://github.com/rynfar/meridian/blob/main/docs/deployment.md) · [docs/profiles.md](https://github.com/rynfar/meridian/blob/main/docs/profiles.md)
- [OpenAI — Codex CLI](https://developers.openai.com/codex/cli)
