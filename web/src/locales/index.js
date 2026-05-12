import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS
}

// Get saved locale or browser locale
function getDefaultLocale() {
  const saved = localStorage.getItem('locale')
  if (saved && messages[saved]) {
    return saved
  }

  // Try to match browser language
  const browserLang = navigator.language
  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'en-US',
  messages
})

// Helper to change locale
export function setLocale(locale) {
  if (messages[locale]) {
    i18n.global.locale.value = locale
    localStorage.setItem('locale', locale)
    document.documentElement.lang = locale
  }
}

export function getLocale() {
  return i18n.global.locale.value
}

export default i18n
