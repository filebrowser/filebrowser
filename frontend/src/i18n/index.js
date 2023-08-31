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
import tr from "./tr.json";
import uk from "./uk.json";
import svSE from "./sv-se.json";
import zhCN from "./zh-cn.json";
import zhTW from "./zh-tw.json";

export function detectLocale() {
  // locale is an RFC 5646 language tag
  // https://developer.mozilla.org/en-US/docs/Web/API/Navigator/language
  let locale = navigator.language.toLowerCase();
  switch (true) {
    case /^he\b/.test(locale):
      locale = "he";
      break;
    case /^hu\b/.test(locale):
      locale = "hu";
      break;
    case /^ar\b/.test(locale):
      locale = "ar";
      break;
    case /^es\b/.test(locale):
      locale = "es";
      break;
    case /^en\b/.test(locale):
      locale = "en";
      break;
    case /^is\b/.test(locale):
      locale = "is";
      break;
    case /^it\b/.test(locale):
      locale = "it";
      break;
    case /^fr\b/.test(locale):
      locale = "fr";
      break;
    case /^pt-br\b/.test(locale):
      locale = "pt-br";
      break;
    case /^pt\b/.test(locale):
      locale = "pt";
      break;
    case /^ja\b/.test(locale):
      locale = "ja";
      break;
    case /^zh-tw\b/.test(locale):
      locale = "zh-tw";
      break;
    case /^zh-cn\b/.test(locale):
    case /^zh\b/.test(locale):
      locale = "zh-cn";
      break;
    case /^de\b/.test(locale):
      locale = "de";
      break;
    case /^ro\b/.test(locale):
      locale = "ro";
      break;
    case /^ru\b/.test(locale):
      locale = "ru";
      break;
    case /^pl\b/.test(locale):
      locale = "pl";
      break;
    case /^ko\b/.test(locale):
      locale = "ko";
      break;
    case /^sk\b/.test(locale):
      locale = "sk";
      break;
    case /^tr\b/.test(locale):
      locale = "tr";
      break;
    // ua wasnt a valid locale for ukraine
    case /^uk\b/.test(locale):
      locale = "uk";
      break;
    case /^sv-se\b/.test(locale):
    case /^sv\b/.test(locale):
      locale = "sv";
      break;
    case /^nl-be\b/.test(locale):
      locale = "nl-be";
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
    tr: removeEmpty(tr),
    uk: removeEmpty(uk),
    "zh-cn": removeEmpty(zhCN),
    "zh-tw": removeEmpty(zhTW),
  },
  // expose i18n.global for outside components
  legacy: true,
});

export default i18n;
