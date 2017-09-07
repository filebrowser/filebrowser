import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en.yaml'
import fr from './fr.yaml'
import pt from './pt.yaml'
import ja from './ja.yaml'
import zhCN from './zh-cn.yaml'
import zhTW from './zh-tw.yaml'

Vue.use(VueI18n)

const i18n = new VueI18n({
  locale: 'en',
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
