import { createI18n } from "vue-i18n";

import("dayjs/locale/ar");
import("dayjs/locale/de");
import("dayjs/locale/en");
import("dayjs/locale/es");
import("dayjs/locale/fr");
import("dayjs/locale/he");
import("dayjs/locale/hu");
import("dayjs/locale/is");
import("dayjs/locale/it");
import("dayjs/locale/ja");
import("dayjs/locale/ko");
import("dayjs/locale/nl-be");
import("dayjs/locale/pl");
import("dayjs/locale/pt-br");
import("dayjs/locale/pt");
import("dayjs/locale/ro");
import("dayjs/locale/ru");
import("dayjs/locale/sk");
import("dayjs/locale/sv");
import("dayjs/locale/tr");
import("dayjs/locale/uk");
import("dayjs/locale/zh-cn");
import("dayjs/locale/zh-tw");

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
import uk from "./uk.json";
import svSE from "./sv-se.json";
import zhCN from "./zh-cn.json";
import zhTW from "./zh-tw.json";

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
    case /^pt-BR.*/i.test(locale):
      locale = "pt-br";
      break;
    case /^pt.*/i.test(locale):
      locale = "pt";
      break;
    case /^ja.*/i.test(locale):
      locale = "ja";
      break;
    case /^zh-TW/i.test(locale):
      locale = "zh-tw";
      break;
    case /^zh-CN/i.test(locale):
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
    case /^uk.*/i.test(locale):
    case /^ua.*/i.test(locale):
      locale = "uk";
      break;
    case /^sv-SE.*/i.test(locale):
    case /^sv.*/i.test(locale):
      locale = "sv";
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

export const i18n = createI18n({
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
    sv: removeEmpty(svSE),
    uk: removeEmpty(uk),
    "zh-cn": removeEmpty(zhCN),
    "zh-tw": removeEmpty(zhTW),
  },
  legacy: true,
});

export default i18n;
