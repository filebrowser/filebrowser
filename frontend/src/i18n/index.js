import Vue from "vue";
import VueI18n from "vue-i18n";

import he from "./he.json";
import hu from "./hu.json";
import ar from "./ar.json";
import de from "./de.json";
import en from "./en.json";
import es from "./es.json";
import fr from "./fr.json";
import is from "./is.json";
import it from "./it.json";
import ja from "./ja.json";
import ko from "./ko.json";
import nlBE from "./nl-be.json";
import pl from "./pl.json";
import pt from "./pt.json";
import ptBR from "./pt-br.json";
import ro from "./ro.json";
import ru from "./ru.json";
import sk from "./sk.json";
import ua from "./ua.json";
import svSE from "./sv-se.json";
import zhCN from "./zh-cn.json";
import zhTW from "./zh-tw.json";

Vue.use(VueI18n);

export function detectLocale() {
  let locale = (navigator.language || navigator.browserLangugae).toLowerCase();
  switch (true) {
    case /^he.*/i.test(locale):
      locale = "he";
      break;
    case /^hu.*/i.test(locale):
      locale = "hu";
      break;
    case /^ar.*/i.test(locale):
      locale = "ar";
      break;
    case /^es.*/i.test(locale):
      locale = "es";
      break;
    case /^en.*/i.test(locale):
      locale = "en";
      break;
    case /^it.*/i.test(locale):
      locale = "it";
      break;
    case /^fr.*/i.test(locale):
      locale = "fr";
      break;
    case /^pt.*/i.test(locale):
      locale = "pt";
      break;
    case /^pt-BR.*/i.test(locale):
      locale = "pt-br";
      break;
    case /^ja.*/i.test(locale):
      locale = "ja";
      break;
    case /^zh-CN/i.test(locale):
      locale = "zh-cn";
      break;
    case /^zh-TW/i.test(locale):
      locale = "zh-tw";
      break;
    case /^zh.*/i.test(locale):
      locale = "zh-cn";
      break;
    case /^de.*/i.test(locale):
      locale = "de";
      break;
    case /^ru.*/i.test(locale):
      locale = "ru";
      break;
    case /^pl.*/i.test(locale):
      locale = "pl";
      break;
    case /^ko.*/i.test(locale):
      locale = "ko";
      break;
    case /^sk.*/i.test(locale):
      locale = "sk";
      break;
    case /^ua.*/i.test(locale):
      locale = "ua";
      break;
    default:
      locale = "en";
  }

  return locale;
}

const removeEmpty = (obj) =>
  Object.keys(obj)
    .filter((k) => obj[k] !== null && obj[k] !== undefined && obj[k] !== "") // Remove undef. and null and empty.string.
    .reduce(
      (newObj, k) =>
        typeof obj[k] === "object"
          ? Object.assign(newObj, { [k]: removeEmpty(obj[k]) }) // Recurse.
          : Object.assign(newObj, { [k]: obj[k] }), // Copy value.
      {}
    );

export const rtlLanguages = ["he", "ar"];

const i18n = new VueI18n({
  locale: detectLocale(),
  fallbackLocale: "en",
  messages: {
    he: removeEmpty(he),
    hu: removeEmpty(hu),
    ar: removeEmpty(ar),
    de: removeEmpty(de),
    en: en,
    es: removeEmpty(es),
    fr: removeEmpty(fr),
    is: removeEmpty(is),
    it: removeEmpty(it),
    ja: removeEmpty(ja),
    ko: removeEmpty(ko),
    "nl-be": removeEmpty(nlBE),
    pl: removeEmpty(pl),
    "pt-br": removeEmpty(ptBR),
    pt: removeEmpty(pt),
    ru: removeEmpty(ru),
    ro: removeEmpty(ro),
    sk: removeEmpty(sk),
    "sv-se": removeEmpty(svSE),
    ua: removeEmpty(ua),
    "zh-cn": removeEmpty(zhCN),
    "zh-tw": removeEmpty(zhTW),
  },
});

export default i18n;
