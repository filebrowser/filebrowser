import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en.yaml'
import it from './it.yaml'
import fr from './fr.yaml'
import pt from './pt.yaml'
import ptBR from './pt-br.yaml'
import ja from './ja.yaml'
import zhCN from './zh-cn.yaml'
import zhTW from './zh-tw.yaml'
import es from './es.yaml'
import de from './de.yaml'
import ru from './ru.yaml'
import pl from './pl.yaml'

Vue.use(VueI18n)

export function detectLocale () {
  let locale = (navigator.language || navigator.browserLangugae).toLowerCase()
  switch (true) {
    case /^en.*/i.test(locale):
      locale = 'en'
      break
    case /^it.*/i.test(locale):
      locale = 'it'
      break
    case /^fr.*/i.test(locale):
      locale = 'fr'
      break
    case /^pt.*/i.test(locale):
      locale = 'pt'
      break
    case /^pt-BR.*/i.test(locale):
      locale = 'pt-br'
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
    case /^es.*/i.test(locale):
      locale = 'es'
      break
    case /^de.*/i.test(locale):
      locale = 'de'
      break
    case /^ru.*/i.test(locale):
      locale = 'ru'
      break
    case /^pl.*/i.test(locale):
      locale = 'pl'
      break
    default:
      locale = 'en'
  }

  return locale
}

const i18n = new VueI18n({
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {
    'en': en,
    'it': it,
    'fr': fr,
    'pt': pt,
    'pt-br': ptBR,
    'ja': ja,
    'zh-cn': zhCN,
    'zh-tw': zhTW,
    'es': es,
    'de': de,
    'ru': ru,
    'pl': pl
  }
})

export default i18n
