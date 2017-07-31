import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en.json'
import pt from './pt.json'

Vue.use(VueI18n)

export default new VueI18n({
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    'en': en,
    'pt': pt
  }
})
