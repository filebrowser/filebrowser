import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en.yaml'
import fr from './fr.yaml'
import pt from './pt.yaml'
import ja from './ja.yaml'
import zhCN from './zh-cn.yaml'
import zhTW from './zh-tw.yaml'

Vue.use(VueI18n)

let locale = (navigator.language || navigator.browserLangugae).toLowerCase()
switch (true) {
  case /^en.*/i.test(locale):
    locale = 'en'
    break
  case /^fr.*/i.test(locale):
    locale = 'fr'
    break
  case /^pt.*/i.test(locale):
    locale = 'pr'
    break
  case /^ja.*/i.test(locale):
    locale = 'ja'
    break
  case /^zh-CN/i.test(locale):
    locale = 'zh-cn'
    break
  case /^zh-TW/i.test(locale):
    locale = 'zh-tw'
    break
  case /^zh.*/i.test(locale):
    locale = 'zh-cn'
    break
  default:
    locale = 'en'
}

const i18n = new VueI18n({
  locale: locale,
  fallbackLocale: 'en',
  messages: {
    'en': en,
    'fr': fr,
    'pt': pt,
    'ja': ja,
    'zh-cn': zhCN,
    'zh-tw': zhTW
  }
})

export default i18n
