import { createI18n } from 'vue-i18n'
import viVN from './locales/vi-VN.ts'
import ruRU from './locales/ru-RU.ts'
import enUS from './locales/en-US.ts'
import koKR from './locales/ko-KR.ts'

const messages = {
  'vi-VN': viVN,
  'en-US': enUS,
  'ru-RU': ruRU,
  'ko-KR': koKR
}

// Lấy ngôn ngữ đã lưu từ localStorage, mặc định dùng tiếng Việt
const savedLocale = localStorage.getItem('locale') || 'vi-VN'

const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  // Chuỗi fallback 2 tầng, CỐ Ý không để 'vi-VN' trần: trỏ fallback về chính
  // locale mặc định nghĩa là khoá nào vi-VN chưa có sẽ hiện KEY THÔ
  // ("menu.integrations") thay vì rơi về tiếng Anh — lỗi CÂM mỗi lần merge
  // upstream thêm khoá mới. 'en-US' đón khoá upstream chưa dịch; 'vi-VN' đứng
  // cuối để các khoá CHỈ fork có (chess.*) vẫn hiện khi đang xem locale khác.
  fallbackLocale: ['en-US', 'vi-VN'],
  globalInjection: true,
  // Some translations intentionally embed `<strong>` markup (e.g. agent step summaries).
  // We render them via v-html with our own sanitization, so silence vue-i18n's HTML warning
  // to avoid flooding the console and slowing renders during history loads.
  warnHtmlMessage: false,
  messages
})

export default i18n
