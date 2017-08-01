import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en.yaml'
import pt from './pt.yaml'
import zhCN from './zh-cn.yaml'

Vue.use(VueI18n)

const i18n = new VueI18n({
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    'en': en,
    'pt': pt,
    'zh-cn': zhCN
  }
})

export default i18n
