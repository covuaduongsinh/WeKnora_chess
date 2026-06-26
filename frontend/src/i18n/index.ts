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
  fallbackLocale: 'vi-VN',
  globalInjection: true,
  // Some translations intentionally embed `<strong>` markup (e.g. agent step summaries).
  // We render them via v-html with our own sanitization, so silence vue-i18n's HTML warning
  // to avoid flooding the console and slowing renders during history loads.
  warnHtmlMessage: false,
  messages
})

export default i18n
